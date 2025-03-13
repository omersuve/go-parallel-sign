package main

import (
	"crypto/rsa"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/omersuve/go-parallel-sign/pkg/auth"
	"github.com/omersuve/go-parallel-sign/pkg/pool"
)

var clientCounter int32
var mu        	  sync.Mutex
var conns         []net.Conn     // Track all connections
var	connsMu       sync.Mutex     // Protect connenctions slice
var publicKeys    map[int32]*rsa.PublicKey // Client ID -> public key

func main() {
	maxNumbers := flag.Int("max", 800, "maximum number of unique primes to collect") // Default max is 800
    flag.Parse()

	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	fmt.Println("Server started on :3000")

	// Record start time
	startTime := time.Now()

	p := pool.NewNumberPool(*maxNumbers)
	publicKeys = make(map[int32]*rsa.PublicKey)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		clientID := registerClient(conn)
		go handleClient(conn, p, *maxNumbers, clientID, startTime)
	}
}

func handleClient(conn net.Conn, p *pool.NumberPool, maxNumbers int, clientID int32, startTime time.Time) {
	defer conn.Close()

	// Read client's public key
	pubBytes := make([]byte, 4096)
	n, err := conn.Read(pubBytes)
	if err != nil {
		fmt.Println("Error reading public key:", err)
		return
	}
	pubKey, err := auth.ParsePublicKey(pubBytes[:n])
	if err != nil {
		fmt.Println("Error parsing public key:", err)
		return
	}
	mu.Lock()
	publicKeys[clientID] = pubKey
	mu.Unlock()

	err = binary.Write(conn, binary.LittleEndian, clientID)
	if err != nil {
		fmt.Println("Error sending client ID:", err)
		return
	}

	var num int32
	signature := make([]byte, 256)

	for {
		err := binary.Read(conn, binary.LittleEndian, &num)
		if err != nil {
			fmt.Println("Client disconnected or error:", err)
			return
		}
		n, err := conn.Read(signature)
		if err != nil {
			fmt.Println("Error reading signature:", err)
			return
		}
		sig := signature[:n]

		var response int32
		mu.Lock()
		pubKey = publicKeys[clientID]
		mu.Unlock()
		if !auth.Verify(num, sig, pubKey) {
			fmt.Printf("Invalid signature for %d from client %d\n", num, clientID)
			response = -3 // Invalid signature response code: -3
		} else if p.Add(num, clientID) {
			fmt.Printf("Received %d from client %d, Pool length: %d\n", num, clientID, p.Len())
			if p.Len() < maxNumbers {
				response = 1 // Added successfully response code: 1
			} else if p.Len() == maxNumbers {
				notifyClientsAndshutdownServer(p, maxNumbers, startTime, conn)
				// Server is shutting down, notify clients and exit
			}
		} else {
			response = 0 // Duplicate number response code: 0
			fmt.Printf("Rejected %d (duplicate)\n", num)
		}

		// Send feedback to the client unless server is not shutting down due to having all primes collected
		err = binary.Write(conn, binary.LittleEndian, response)
		if err != nil {
			fmt.Println("Error sending feedback:", err)
			return
		}
	}
}

// registerClient assigns a client ID and tracks the connection
func registerClient(conn net.Conn) int32 {
	mu.Lock()
	clientCounter++
	clientID := clientCounter
	mu.Unlock()
	connsMu.Lock()
	conns = append(conns, conn)
	connsMu.Unlock()
	return clientID
}

// shutdownServer prints results and terminates the server
func notifyClientsAndshutdownServer(p *pool.NumberPool, maxNumbers int, startTime time.Time, triggeringConn net.Conn) {
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Printf("Collected %d numbers, final pool length: %v\n", maxNumbers, p.Len())
	scoreboard := p.GetScoreboard()
	fmt.Println("---SCORES---")
	for id, count := range scoreboard {
		fmt.Printf("Client %d: %d numbers\n", id, count)
	}
	fmt.Printf("Time taken to collect %d primes: %v\n", maxNumbers, duration)

	// Send shutdown signal to the client
	err := binary.Write(triggeringConn, binary.LittleEndian, int32(-1))
	if err != nil {
		fmt.Println("Error sending shutdown response:", err)
	}

	// Notify all other clients
	connsMu.Lock()
	for i, c := range conns {
		if c != triggeringConn { // Skip the client that triggered shutdown
			err := binary.Write(c, binary.LittleEndian, int32(-2))
			if err != nil {
				fmt.Printf("Failed to send -2 to conn %d: %v\n", i, err)
			}
		}
	}
	connsMu.Unlock()

	fmt.Println("Server shutting down")
	os.Exit(0)
}