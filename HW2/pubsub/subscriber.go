package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

type MemoryEvent struct {
	EventType   string `json:"event_type"`
	Service     string `json:"service"`
	MemoryMB    uint64 `json:"memory_mb"`
	ThresholdMB uint64 `json:"threshold_mb"`
	Timestamp   string `json:"timestamp"`
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var event MemoryEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("\n[ALERT] High memory usage detected!\n")
	fmt.Printf("	Service:    %s\n", event.Service)
	fmt.Printf("	Event type: %s\n", event.EventType)
	fmt.Printf("	Memory:     %d MB (threshold %d MB)\n", event.MemoryMB, event.ThresholdMB)
	fmt.Printf("	Timestamp:  %s\n", event.Timestamp)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := flag.Int("port", 8084, "Port to listen on for alerts")
	flag.Parse()

	http.HandleFunc("/event", eventHandler)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("Subscriber listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
