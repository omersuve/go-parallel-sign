package main

import (
	"crypto/rsa"
	"encoding/binary"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/omersuve/paribu-primes/pkg/auth"
	"github.com/omersuve/paribu-primes/pkg/pool"
	"github.com/omersuve/paribu-primes/pkg/primes"
)

// TestIntegration_ServerClient tests the full server-client interaction
func TestIntegration_ServerClient(t *testing.T) {
	// Start a lightweight server in a goroutine
	maxNumbers := 30 // Small number for quick test
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		t.Logf("Server starting on :3000...")
		listener, err := net.Listen("tcp", ":3000") // Different port to avoid conflict
		if err != nil {
			t.Errorf("Server failed to start: %v", err)
			return
		}
		defer listener.Close()

		p := pool.NewNumberPool(maxNumbers)
		publicKeys := make(map[int32]*rsa.PublicKey)
		clientCounter := int32(0)

		t.Logf("Server waiting for client connection...")

		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Server failed to accept: %v", err)
			return
		}
		defer conn.Close()

		clientCounter++
		clientID := clientCounter

		t.Logf("Server accepted client, assigned clientID: %d", clientID)

		pubBytes := make([]byte, 4096)
		n, err := conn.Read(pubBytes)
		if err != nil {
			t.Errorf("Server failed to read public key: %v", err)
			return
		}

		t.Logf("Server received public key (%d bytes)", n)

		pubKey, err := auth.ParsePublicKey(pubBytes[:n])
		if err != nil {
			t.Errorf("Server failed to parse public key: %v", err)
			return
		}
		publicKeys[clientID] = pubKey
		err = binary.Write(conn, binary.LittleEndian, clientID)
		if err != nil {
			t.Errorf("Server failed to send clientID: %v", err)
			return
		}

		t.Logf("Server sent clientID %d to client", clientID)

		var num int32
		sig := make([]byte, 256)
		for p.Len() < maxNumbers {
			err = binary.Read(conn, binary.LittleEndian, &num)
			if err != nil {
				t.Errorf("Server failed to read number: %v", err)
				return
			}

			t.Logf("Server received number: %d", num)

			n, err = conn.Read(sig)
			if err != nil {
				t.Errorf("Server failed to read signature: %v", err)
				return
			}

			t.Logf("Server received signature (%d bytes)", n)

			if auth.Verify(num, sig[:n], pubKey) {
				if p.Add(num, clientID) {

					t.Logf("Server added %d to pool, length now: %d", num, p.Len())

					binary.Write(conn, binary.LittleEndian, int32(1))
				} else {

					t.Logf("Server rejected %d (duplicate), pool length: %d", num, p.Len())

					binary.Write(conn, binary.LittleEndian, int32(0))
				}
			} else {

				t.Logf("Server rejected %d (invalid signature)", num)

				binary.Write(conn, binary.LittleEndian, int32(-3))
			}
		}

		t.Logf("Server collected %d numbers, sending shutdown signal (-1)", maxNumbers)

		binary.Write(conn, binary.LittleEndian, int32(-1)) // Signal completion
	}()

	// Wait for server to start

	t.Logf("Client waiting for server to start...")

	time.Sleep(100 * time.Millisecond)

	// Start client

	t.Logf("Client connecting to localhost:3000...")

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		t.Fatalf("Client failed to connect: %v", err)
	}
	defer conn.Close()

	// Send public key

	t.Logf("Client generating RSA keys...")

	keys, err := auth.GenerateKeys()
	if err != nil {
		t.Fatalf("Client failed to generate keys: %v", err)
	}
	pubBytes, err := auth.PublicKey2Bytes(keys.PublicKey)
	if err != nil {
		t.Fatalf("Client failed to have bytes from public key: %v", err)
	}
	_, err = conn.Write(pubBytes)
	if err != nil {
		t.Fatalf("Client failed to send public key: %v", err)
	}

	t.Logf("Client sent public key (%d bytes)", len(pubBytes))

	// Receive clientID
	var clientID int32
	err = binary.Read(conn, binary.LittleEndian, &clientID)
	if err != nil {
		t.Fatalf("Client failed to read clientID: %v", err)
	}

	t.Logf("Client received clientID: %d", clientID)

	if clientID != 1 {
		t.Errorf("Expected clientID 1, got %d", clientID)
	}

	// Send primes
	rng := rand.New(rand.NewSource(int64(clientID)))
	sent := 0
	for sent < maxNumbers {
		num := primes.GenerateRandomPrime(1000000, rng) // Smaller range for speed

		t.Logf("Client sending prime: %d", num)

		sig, err := auth.Sign(num, keys.PrivateKey)
		if err != nil {
			t.Fatalf("Client failed to sign %d: %v", num, err)
		}
		err = binary.Write(conn, binary.LittleEndian, num)
		if err != nil {
			t.Fatalf("Client failed to send %d: %v", num, err)
		}
		_, err = conn.Write(sig)
		if err != nil {
			t.Fatalf("Client failed to send signature for %d: %v", num, err)
		}

		var response int32
		err = binary.Read(conn, binary.LittleEndian, &response)
		if err != nil {
			t.Fatalf("Client failed to read response: %v", err)
		}

		t.Logf("Client received response %d for prime %d", response, num)

		if response == 1 {
			sent++
		} else if response == -1 {
			break
		} else if response != 0 { // Allow duplicates (0), fail on invalid sig (-3)
			t.Errorf("Unexpected response %d for num %d", response, num)
		}
	}

	// Verify completion
	var finalResponse int32
	binary.Read(conn, binary.LittleEndian, &finalResponse)

	t.Logf("Client received final response: %d", finalResponse)

	if finalResponse != -1 {
		t.Errorf("Expected final response -1, got %d", finalResponse)
	}

	// Ensure server shut down

	t.Logf("Checking if server shut down...")

	select {
	case <-serverDone:
		t.Logf("Server shut down successfully")
	case <-time.After(1 * time.Second):
		t.Error("Server did not shut down within 1 second")
	}
}