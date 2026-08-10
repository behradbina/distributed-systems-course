package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	requestPipe  = "./request.fifo"
	responsePipe = "./response.fifo"
	PRINT_COLOR  = "\033[34m"
	PRINT_RESET  = "\033[0m"
)

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

func createResponse(resstr string) string {
	var response Response
	if err := json.Unmarshal([]byte(resstr), &response); err != nil {
		return fmt.Sprintf("Failed to parse the response. Raw response: %s", resstr)
	}
	if response.Status == "OK" {
		return fmt.Sprintf("%g %s %g = %g", response.A,
			strings.ToUpper(response.Operation), response.B, response.Result)
	}
	return fmt.Sprintf("Error: %s", response.Error)
}

func openPipes() (*os.File, *os.File, error) {
	if _, err := os.Stat(requestPipe); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("Request pipe does not exist: %w", err)
	}
	if _, err := os.Stat(responsePipe); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("Response pipe does not exist: %w", err)
	}

	reqFile, err := os.OpenFile(requestPipe, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot open request pipe: %w", err)
	}

	resFile, err := os.OpenFile(responsePipe, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		reqFile.Close()
		return nil, nil, fmt.Errorf("Cannot open response pipe: %w", err)
	}

	return reqFile, resFile, nil
}

func main() {
	log.SetPrefix("[INTERFACE] ")
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	reqFile, resFile, err := openPipes()
	if err != nil {
		log.Fatal(err)
	}
	logToTerminal("Pipes opened.")
	defer reqFile.Close()
	defer resFile.Close()

	writer := bufio.NewWriter(reqFile)
	reader := bufio.NewReader(resFile)
	stdin := bufio.NewReader(os.Stdin)

	for {
		input, err := stdin.ReadString('\n')
		if err != nil {
			logToTerminal("EOF detected. Sending EXIT to worker.")
			fmt.Fprintln(writer, "EXIT")
			writer.Flush()
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			logToTerminal(fmt.Sprintf("Empty input string read."))
			continue
		}
		input = strings.ToUpper(input)

		if input == "EXIT" || input == "QUIT" {
			logToTerminal(fmt.Sprintf("EXIT input detected. Sending EXIT to worker."))
			fmt.Fprintln(writer, "EXIT")
			writer.Flush()
			break
		}

		if _, err = fmt.Fprintln(writer, input); err != nil {
			logToTerminal(fmt.Sprintf("Failed to send request: %v\n", err))
			break
		}
		if err := writer.Flush(); err != nil {
			logToTerminal(fmt.Sprintf("Failed to flush request: %v\n", err))
			break
		}

		response, err := reader.ReadString('\n')
		if err != nil {
			logToTerminal(fmt.Sprintf("Failed to read response: %v\n", err))
			break
		}

		response = strings.TrimSpace(response)
		fmt.Println(string(PRINT_COLOR) + createResponse(response) + string(PRINT_RESET))
		logToTerminal("Response printed.")
	}
	fmt.Println(string(PRINT_COLOR) + "Interface exited." + string(PRINT_RESET))
}
