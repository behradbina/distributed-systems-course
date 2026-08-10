package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"syscall"
)

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password-hash"`
}

type AuthService struct {
	users map[string]string
}

type LoginArgs struct {
	Username string
	Password string
}

type LoginReply struct {
	Success bool
	Error   string
}

func (a *AuthService) Login(args *LoginArgs, reply *LoginReply) error {
	storedHash, exists := a.users[args.Username]
	hashBytes := sha256.Sum256([]byte(args.Password))
	passHash := hex.EncodeToString(hashBytes[:])

	if !exists || storedHash != passHash {
		reply.Success = false
		reply.Error = "Invalid username or password"
		return nil
	}
	reply.Success = true
	reply.Error = ""
	return nil
}

func loadUsers(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read users file: %w", err)
	}
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse users JSON: %w", err)
	}
	userMap := make(map[string]string, len(users))
	for _, u := range users {
		userMap[u.Username] = u.PasswordHash
	}
	return userMap, nil
}

func main() {
	port := flag.Int("port", 8082, "Port to listen on (e.g., 8082)")
	usersFile := flag.String("users", "users.json", "Path to users.json file")
	flag.Parse()

	users, err := loadUsers(*usersFile)
	if err != nil {
		log.Fatalf("Failed to load users: %v", err)
	}
	log.Printf("Loaded %d users from %s", len(users), *usersFile)

	authService := &AuthService{users: users}
	err = rpc.Register(authService)
	if err != nil {
		log.Fatalf("Failed to register RPC service: %v", err)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	log.Printf("Auth server listening on %s (JSON-RPC over TCP)", addr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down auth server...")
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		}
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}
