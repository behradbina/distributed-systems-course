package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type ValueRecord struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Version   int    `json:"version"`
	UpdatedBy string `json:"updated_by"`
	Timestamp int64  `json:"timestamp"` // UnixNano for Last-Write-Wins resolution
}

type Config struct {
	ID        string            `json:"id"`
	Port      string            `json:"port"`
	Peers     map[string]string `json:"peers"`
	IsPrimary bool             `json:"is_primary"`
}

type ReplicaServer struct {
	mu        sync.RWMutex
	store     map[string]ValueRecord
	config    Config
	latencyMs int
}

func main() {
	configPath := flag.String("config", "", "Path to the configuration JSON file")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("Error: Please provide a configuration file path using the -config flag.")
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	var cfg Config
	if err := json.NewDecoder(configFile).Decode(&cfg); err != nil {
		log.Fatalf("Failed to decode configuration JSON: %v", err)
	}

	server := &ReplicaServer{
		store: make(map[string]ValueRecord),
		config: cfg,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/get", server.handleGet)
	mux.HandleFunc("/put", server.handlePut)
	mux.HandleFunc("/replicate", server.handleReplicate)
	mux.HandleFunc("/set_latency", server.handleSetLatency)
	mux.HandleFunc("/strong_get", server.handleStrongGet)

	fmt.Printf("[%s] Server booting up on port %s...\n", server.config.ID, server.config.Port)
	if err := http.ListenAndServe(server.config.Port, mux); err != nil {
		log.Fatalf("Server crash on port %s: %v", server.config.Port, err)
	}
}

func (s *ReplicaServer) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	record, exists := s.store[key]
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key not found"})
		return
	}
	json.NewEncoder(w).Encode(record)
}

func (s *ReplicaServer) handlePut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key             string `json:"key"`
		Value           string `json:"value"`
		ConsistencyMode string `json:"consistency_mode"` // "eventual" or "strong"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	currentRecord := s.store[req.Key]
	newRecord := ValueRecord{
		Key:       req.Key,
		Value:     req.Value,
		Version:   currentRecord.Version + 1,
		UpdatedBy: s.config.ID,
		Timestamp: time.Now().UnixNano(),
	}
	s.store[req.Key] = newRecord
	s.mu.Unlock()

	if req.ConsistencyMode == "strong" {
		// Strong consistency: Synchronously confirm write with a majority quorum (2 out of 3 replicas total)
		acks := 1 // Initializing with 1 to count local write success
		var ackMu sync.Mutex
		var wg sync.WaitGroup

		for _, peerUrl := range s.config.Peers {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				if s.sendReplicationMessage(url, newRecord) {
					ackMu.Lock()
					acks++
					ackMu.Unlock()
				}
			}(peerUrl)
		}
		wg.Wait()

		// A quorum of 2 nodes out of 3 is required
		if acks >= 2 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "record": newRecord, "replicas_updated": acks})
		} else {
			// Roll back state if majority write consensus fails
			s.mu.Lock()
			if s.store[req.Key].Version == newRecord.Version && s.store[req.Key].UpdatedBy == s.config.ID {
				delete(s.store, req.Key)
			}
			s.mu.Unlock()
			http.Error(w, "Strong write failed: Quorum agreement not met", http.StatusServiceUnavailable)
		}
	} else {
		// Eventual Consistency: Asynchronously propagate updates to background workers
		for _, peerUrl := range s.config.Peers {
			go func(url string) {
				s.sendReplicationMessage(url, newRecord)
			}(peerUrl)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "record": newRecord, "replicas_updated": 1})
	}
}

func (s *ReplicaServer) handleReplicate(w http.ResponseWriter, r *http.Request) {
	var incoming ValueRecord
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Apply configured network latency simulation
	s.mu.RLock()
	latency := s.latencyMs
	s.mu.RUnlock()
	if latency > 0 {
		time.Sleep(time.Duration(latency) * time.Millisecond)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	local, exists := s.store[incoming.Key]
	if !exists || incoming.Version > local.Version {
		// Rule 1: Always update if incoming data has a higher version structure
		s.store[incoming.Key] = incoming
	} else if incoming.Version == local.Version {
		// Conflict Resolution: Last-Write-Wins strategy (LWW) utilizing UnixNano timestamps
		if incoming.Timestamp > local.Timestamp {
			s.store[incoming.Key] = incoming
		} else if incoming.Timestamp == local.Timestamp {
			// Deterministic Tie-Breaker: Choose lexicographically higher Unique Replica IDs
			if incoming.UpdatedBy > local.UpdatedBy {
				s.store[incoming.Key] = incoming
			}
		}
	} // Ignore if incoming.Version < local.Version

	w.WriteHeader(http.StatusOK)
}

func (s *ReplicaServer) handleSetLatency(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LatencyMs int `json:"latency_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.latencyMs = req.LatencyMs
	s.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (s *ReplicaServer) handleStrongGet(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	// If already primary
	if s.config.IsPrimary {

		s.mu.RLock()
		record, ok := s.store[key]
		s.mu.RUnlock()

		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(record)
		return
	}

	// Forward to primary
	resp, err := http.Get(
		"http://localhost:8081/strong_get?key=" + key,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)

	var body interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	json.NewEncoder(w).Encode(body)
}

func (s *ReplicaServer) sendReplicationMessage(peerUrl string, record ValueRecord) bool {
	jsonData, _ := json.Marshal(record)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(peerUrl+"/replicate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}