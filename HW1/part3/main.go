package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type Request struct {
	A  float64 `json:"a"`
	B  float64 `json:"b"`
	Op string  `json:"op"`
}

type Response struct {
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

var validOps = map[string]bool{
	"add": true, "sub": true, "mul": true, "div": true,
	"mod": true, "pow": true, "max": true, "MIN": true,
}

func calculate(w http.ResponseWriter, r *http.Request) {
	var req Request
	var res Response

	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")
	req.Op = r.URL.Query().Get("op")

	if aStr == "" || bStr == "" || req.Op == "" {
		res.Error = "Missing query parameters"
		json.NewEncoder(w).Encode(res)
		return
	}

	var err error
	req.A, err = strconv.ParseFloat(aStr, 64)
	if err != nil {
		res.Error = "Invalid parameter 'a'"
		json.NewEncoder(w).Encode(res)
		return
	}

	req.B, err = strconv.ParseFloat(bStr, 64)
	if err != nil {
		res.Error = "Invalid parameter 'b'"
		json.NewEncoder(w).Encode(res)
		return
	}

	switch req.Op {
	case "add":
		res.Result = req.A + req.B
	case "sub":
		res.Result = req.A - req.B
	case "mul":
		res.Result = req.A * req.B
	case "div":
		if req.B == 0 {
			res.Error = "Division by zero"
		} else {
			res.Result = req.A / req.B
		}
	case "mod":
		if req.B == 0 {
			res.Error = "Modulo by zero"
		}
		res.Result = math.Mod(req.A, req.B)
	case "pow":
		res.Result = math.Pow(req.A, req.B)
	case "max":
		if req.A >= req.B {
			res.Result = req.A
		}
		res.Result = req.B
	case "min":
		if req.A <= req.B {
			res.Result = req.A
		}
		res.Result = req.B
	default:
		ops := make([]string, 0, len(validOps))
		for k := range validOps {
			ops = append(ops, k)
		}
		res.Error = fmt.Sprintf("Unknown operation: %s; supported: %s",
			req.Op, strings.Join(ops, ", "))
	}
	json.NewEncoder(w).Encode(res)
}

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": fmt.Sprintf("Endpoint \"%s\" not found", r.URL.Path),
		"code":  404,
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/calculate", calculate)
	mux.HandleFunc("/health", health)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)

		if pattern == "" {
			notFound(w, r)
			return
		}

		mux.ServeHTTP(w, r)
	})

	port := 8081
	log.Printf("Server running on port %d...\n", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), handler))
}
