#!/bin/bash

# Configuration: Ensure logs directory exists if needed
echo "================================================================="
# Clean up background tasks instantly if the script drops out or you hit Ctrl+C
cleanup() {
    echo -e "\n[Orchestrator] Tearing down background replica processes safely..."
    kill $PID1 $PID2 $PID3 2>/dev/null
    exit 0
}
trap cleanup EXIT INT TERM

echo "[Orchestrator] Launching Replica 1 (Port 8081)..."
go run replica/main.go -config=configs/replica1.json &
PID1=$!

echo "[Orchestrator] Launching Replica 2 (Port 8082)..."
go run replica/main.go -config=configs/replica2.json &
PID2=$!

echo "[Orchestrator] Launching Replica 3 (Port 8083)..."
go run replica/main.go -config=configs/replica3.json &
PID3=$!

echo "[Orchestrator] Sleeping for 2 seconds to allow TCP sockets to bind..."
sleep 2

echo "[Orchestrator] Firing up the automated analysis engine..."
echo "-----------------------------------------------------------------"
go run client/main.go
echo "-----------------------------------------------------------------"

echo "[Orchestrator] Client simulation complete."