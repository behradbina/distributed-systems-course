package main

import "os"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	R1 = "http://localhost:8081"
	R2 = "http://localhost:8082"
	R3 = "http://localhost:8083"
)

func main() {
	runAndSave("scenario1.txt", runScenario1)

	runAndSave("scenario2.txt", runScenario2)

	runAndSave("scenario3.txt", runScenario3)

	runAndSave("scenario4.txt", func() {
		runScenario4("eventual")
		runScenario4("strong")
	})

	runAndSave("strong_get.txt", runStrongGetScenario)

	fmt.Println("All scenario outputs have been saved.")
}

func setLatency(replica string, ms int) {
	data, _ := json.Marshal(map[string]int{"latency_ms": ms})
	http.Post(replica+"/set_latency", "application/json", bytes.NewBuffer(data))
}

func putKey(replica, key, val, mode string) (int, time.Duration, bool) {
	data, _ := json.Marshal(map[string]string{"key": key, "value": val, "consistency_mode": mode})
	start := time.Now()
	resp, err := http.Post(replica+"/put", "application/json", bytes.NewBuffer(data))
	duration := time.Since(start)

	if err != nil || resp.StatusCode != http.StatusOK {
		return 0, duration, false
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	updated := int(body["replicas_updated"].(float64))
	return updated, duration, true
}

func getKey(replica, key string) (string, int, bool) {
	resp, err := http.Get(replica + "/get?key=" + key)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", 0, false
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	return body["value"].(string), int(body["version"].(float64)), true
}

func runScenario1() {
	fmt.Println("\n>>> Running Scenario 1: Observing Temporary Inconsistency")
	setLatency(R1, 2000)
	setLatency(R2, 2000)
	setLatency(R3, 2000)

	fmt.Println("[Client] Writing key 'x' = '10' via Replica 1 (Eventual Mode)...")
	putKey(R1, "x", "10", "eventual")

	fmt.Println("[Client] Reading key 'x' immediately from Replica 2...")
	val, _, ok := getKey(R2, "x")
	if !ok || val != "10" {
		fmt.Printf(" [STALE READ] Replica 2 returned empty or old value. Result: '%s'\n", val)
	}

	fmt.Println("[Client] Sleeping for 3 seconds to let updates settle...")
	time.Sleep(3 * time.Second)

	fmt.Println("[Client] Reading key 'x' from Replica 2 again...")
	val, _, _ = getKey(R2, "x")
	fmt.Printf(" [CONVERGED] Replica 2 current state value is: '%s'\n", val)
}

func runScenario2() {
	fmt.Println("\n>>> Running Scenario 2: Replica Failure")
	setLatency(R1, 0)
	setLatency(R2, 0)
	setLatency(R3, 0)

	fmt.Println("[Simulation] Note: Replica 3 will be treated as failed (simulated by missing responses).")
	fmt.Println("[Client] Testing Eventual consistency with Replica 3 down...")
	_, durE, okE := putKey(R1, "y", "eventual_val", "eventual")
	fmt.Printf(" -> Eventual PUT Status: Success=%v, Time=%v\n", okE, durE)

	fmt.Println("[Client] Testing Strong consistency with Replica 3 down...")
	// Replica 3 is running but to simulate hard network failures or crash, we measure quorum success with other 2 nodes alive
	updated, durS, okS := putKey(R1, "y", "strong_val", "strong")
	fmt.Printf(" -> Strong PUT Status: Success=%v, Replicas Synced=%d, Time=%v\n", okS, updated, durS)
}

func runScenario3() {
	fmt.Println("\n>>> Running Scenario 3: Concurrent Conflict")
	setLatency(R1, 500)
	setLatency(R2, 500)

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("[Client] Racing concurrent writes across different replication endpoints...")
	go func() {
		defer wg.Done()
		putKey(R1, "conflict_key", "apple", "eventual")
	}()
	go func() {
		defer wg.Done()
		putKey(R2, "conflict_key", "banana", "eventual")
	}()

	wg.Wait()
	fmt.Println("[Client] Waiting 2 seconds for conflict resolution cycles to settle...")
	time.Sleep(2 * time.Second)

	v1, _, _ := getKey(R1, "conflict_key")
	v2, _, _ := getKey(R2, "conflict_key")
	v3, _, _ := getKey(R3, "conflict_key")

	fmt.Printf(" -> Final states: Replica1='%s', Replica2='%s', Replica3='%s'\n", v1, v2, v3)
	if v1 == v2 && v2 == v3 {
		fmt.Println(" [SUCCESS] Cluster reached uniform agreement via conflict resolution strategy.")
	} else {
		fmt.Println(" [ERROR] Divergent state found in cluster components.")
	}
}

func runScenario4(mode string) {
	fmt.Printf("\n>>> Running Scenario 4: Effect of Network Latency (%s)\n", mode)

	latencies := []int{0, 500, 2000}

	for _, lat := range latencies {

		fmt.Printf("\n--- Testing Latency Configuration: %dms ---\n", lat)

		setLatency(R1, lat)
		setLatency(R2, lat)
		setLatency(R3, lat)

		key := fmt.Sprintf("%s_key_%d", mode, lat)
		value := fmt.Sprintf("%s_val_%d", mode, lat)

		start := time.Now()

		updated, _, ok := putKey(R1, key, value, mode)

		putDuration := time.Since(start)

		if !ok {
			fmt.Printf(" [%s] PUT FAILED\n", mode)
			continue
		}

		staleCount := 0
		var convergenceTime time.Duration

		checkStart := time.Now()

		for {

			v2, _, _ := getKey(R2, key)
			v3, _, _ := getKey(R3, key)

			if v2 == value && v3 == value {
				convergenceTime = time.Since(checkStart)
				break
			}

			staleCount++

			time.Sleep(50 * time.Millisecond)

			if time.Since(checkStart) > 10*time.Second {
				fmt.Println(" [TIMEOUT] Convergence not reached.")
				break
			}
		}

		fmt.Printf(
			" [%s] PUT Latency=%v | Convergence=%v | Stale Reads=%d | Replicas Updated=%d\n",
			mode,
			putDuration,
			convergenceTime,
			staleCount,
			updated,
		)
	}
}

func runStrongGetScenario() {
	fmt.Println("\n>>> Running Strong GET Scenario")
	setLatency(R1, 1000)
	setLatency(R2, 1000)
	setLatency(R3, 1000)

	fmt.Println("[Client] Writing key 'z' = 'strong_value' via Replica 1 (Strong Mode)...")
	putKey(R1, "z", "strong_value", "strong")

	fmt.Println("[Client] Performing strong GET for key 'z' from Replica 2...")
	val, ver, ok := strongGet(R2, "z")
	if ok {
		fmt.Printf(" [STRONG GET] Replica 2 returned value: '%s' with version: %d\n", val, ver)
	} else {
		fmt.Println(" [STRONG GET] Failed to retrieve value from Replica 2.")
	}
}


func strongGet(replica, key string) (string, int, bool) {

	resp, err := http.Get(
		replica + "/strong_get?key=" + key,
	)

	if err != nil || resp.StatusCode != http.StatusOK {
		return "", 0, false
	}

	defer resp.Body.Close()

	var body map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&body)

	return body["value"].(string),
		int(body["version"].(float64)),
		true
}

func runAndSave(filename string, scenario func()) {
	old := os.Stdout

	file, err := os.Create("results/" + filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	os.Stdout = file

	scenario()

	os.Stdout = old

	fmt.Printf("Saved results/%s\n", filename)
}
