package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const (
	requestPipe  = "./request.fifo"
	responsePipe = "./response.fifo"
	PRINT_COLOR  = "\033[31m"
	PRINT_RESET  = "\033[0m"
)

var validOps = map[string]bool{
	"ADD": true, "SUB": true, "MUL": true, "DIV": true,
	"MOD": true, "POW": true, "MAX": true, "MIN": true,
}

type Request struct {
	Operation string
	A         float64
	B         float64
}

type Response struct {
	Status    string  `json:"status"`
	Result    float64 `json:"result,omitempty"`
	Error     string  `json:"error,omitempty"`
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

func logToTerminal(str string) {
	log.Println(string(PRINT_COLOR) + str + string(PRINT_RESET))
}

func createPipes() error {
	_ = os.Remove(requestPipe)
	if err := syscall.Mkfifo(requestPipe, 0666); err != nil {
		return fmt.Errorf("Failed to create request pipe: %s", err)
	}
	_ = os.Remove(responsePipe)
	if err := syscall.Mkfifo(responsePipe, 0666); err != nil {
		return fmt.Errorf("Failed to create response pipe: %s", err)
	}
	return nil
}

func openPipes() (*os.File, *os.File, error) {
	if _, err := os.Stat(requestPipe); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("Request pipe does not exist: %w", err)
	}
	if _, err := os.Stat(responsePipe); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("Response pipe does not exist: %w", err)
	}

	reqFile, err := os.OpenFile(requestPipe, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot open request pipe: %w", err)
	}

	resFile, err := os.OpenFile(responsePipe, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		reqFile.Close()
		return nil, nil, fmt.Errorf("Cannot open response pipe: %w", err)
	}

	return reqFile, resFile, nil
}

func parseRequest(reqstr string) (*Request, error) {
	parts := strings.Fields(strings.TrimSpace(reqstr))
	if len(parts) != 3 {
		return nil, fmt.Errorf(
			"Invalid argument count: expected 3 tokens (OP A B), got %d", len(parts))
	}

	a, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid operand A: %s is not a number", parts[1])
	}
	b, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid operand B: %s is not a number", parts[2])
	}

	return &Request{Operation: op, A: a, B: b}, nil
}

func compute(req *Request) (float64, error) {
	switch req.Operation {
	case "ADD":
		return req.A + req.B, nil
	case "SUB":
		return req.A - req.B, nil
	case "MUL":
		return req.A * req.B, nil
	case "DIV":
		if req.B == 0 {
			return 0, fmt.Errorf("Division by zero")
		}
		return req.A / req.B, nil
	case "MOD":
		if req.B == 0 {
			return 0, fmt.Errorf("Modulo by zero")
		}
		return math.Mod(req.A, req.B), nil
	case "POW":
		return math.Pow(req.A, req.B), nil
	case "MAX":
		if req.A >= req.B {
			return req.A, nil
		}
		return req.B, nil
	case "MIN":
		if req.A <= req.B {
			return req.A, nil
		}
		return req.B, nil
	default:
		ops := make([]string, 0, len(validOps))
		for k := range validOps {
			ops = append(ops, k)
		}
		return 0, fmt.Errorf("Unknown operation: %s; supported: %s",
			req.Operation, strings.Join(ops, ", "))
	}
}

func writeResponse(writer *bufio.Writer, resp Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("Failed to marshal response: %w", err)
	}
	_, err = fmt.Fprintf(writer, "%s\n", data)
	if err != nil {
		return fmt.Errorf("Failed to write response: %w", err)
	}
	return writer.Flush()
}

func main() {
	log.SetPrefix("[WORKER] ")
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	if err := createPipes(); err != nil {
		logToTerminal(err.Error())
		os.Exit(1)
	}
	logToTerminal("Pipes created.")

	reqFile, resFile, err := openPipes()
	if err != nil {
		logToTerminal(err.Error())
		os.Exit(1)
	}
	logToTerminal("Pipes opened.")
	defer reqFile.Close()
	defer resFile.Close()
	defer os.Remove(requestPipe)
	defer os.Remove(responsePipe)

	writer := bufio.NewWriter(resFile)
	reader := bufio.NewReader(reqFile)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			logToTerminal(fmt.Sprintf("Pipe closed or read error: %v", err))
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			logToTerminal(fmt.Sprintf("Empty input string read."))
			continue
		}
		input = strings.ToUpper(input)

		if input == "EXIT" {
			logToTerminal(fmt.Sprintf("Received EXIT command."))
			break
		}

		logToTerminal(fmt.Sprintf("Received request: %q", input))
		req, err := parseRequest(input)
		var resp Response
		if err != nil {
			logToTerminal(fmt.Sprintf("Parse error: %v\n", err))
			resp = Response{
				Status: "ERR",
				Error:  err.Error(),
			}
		} else {
			result, err := compute(req)
			if err != nil {
				logToTerminal(fmt.Sprintf("Compute error: %v", err))
				resp = Response{
					Status:    "ERR",
					Error:     err.Error(),
					Operation: req.Operation,
					A:         req.A,
					B:         req.B,
				}
			} else {
				logToTerminal(fmt.Sprintf(
					"Result: %g %s %g = %g", req.A, req.Operation, req.B, result))
				resp = Response{
					Status:    "OK",
					Result:    result,
					Operation: req.Operation,
					A:         req.A,
					B:         req.B,
				}
			}
		}

		if err := writeResponse(writer, resp); err != nil {
			logToTerminal(fmt.Sprintf("Failed to send response: %v\n", err))
			break
		}
	}
	logToTerminal("Worker exited.")
}
