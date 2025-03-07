package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"net"

	"github.com/omersuve/paribu-primes/pkg/auth"
	"github.com/omersuve/paribu-primes/pkg/primes"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	// Generate RSA key pair
	keys, err := auth.GenerateKeys()
	if err != nil {
		fmt.Println("Error generating keys:", err)
		return
	}

	// Send public key to server
	pubBytes, err := auth.PublicKey2Bytes(keys.PublicKey)
	if err != nil {
		fmt.Println("Error having bytes of public key:", err)
		return
	}

	// fmt.Printf("Generated Public Key (PEM):\n%s", string(pubBytes))

	_, err = conn.Write(pubBytes)
	if err != nil {
		fmt.Println("Error sending public key:", err)
		return
	}

	// Receive client ID from server
	var clientID int32
	err = binary.Read(conn, binary.LittleEndian, &clientID)
	if err != nil {
		fmt.Println("Error reading client ID:", err)
		return
	}
	fmt.Printf("Client ID: %d\n", clientID)

	// Create a local random generator seeded with clientID
	rng := rand.New(rand.NewSource(int64(clientID)))

	for {
		// Generating a random prime using the local RNG
		num := primes.GenerateRandomPrime(math.MaxInt32, rng)
		signature, err := auth.Sign(num, keys.PrivateKey)
		if err != nil {
			fmt.Println("Error signing number:", err)
			return
		}

		// fmt.Printf("Sending Number: %d, Signature (hex): %s", num, hex.EncodeToString(signature))

		err = binary.Write(conn, binary.LittleEndian, num)
		if err != nil {
			fmt.Println("Error sending:", err)
			return
		}
		_, err = conn.Write(signature)
		if err != nil {
			fmt.Println("Error sending signature:", err)
			return
		}

		var response int32
		err = binary.Read(conn, binary.LittleEndian, &response)
		if err != nil {
			fmt.Println("Error reading feedback:", err)
			return
		}

		if response == -1 {
            fmt.Printf("Sent %d: Successfully added (completing collection)\n", num)
			fmt.Println("Server has collected all numbers, exiting")
            return
		} else if response == -2 {
			fmt.Println("Server has collected all numbers, exiting")
			return
        } else if response == 1 {
            fmt.Printf("Sent %d: Successfully added\n", num)
		} else if response == 0 {
			fmt.Printf("Sent %d: Rejected (duplicate)\n", num)
		} else if response == -3 {
			fmt.Printf("Sent %d: Rejected (invalid signature)\n", num)
		}
        // time.Sleep(500 * time.Millisecond)
	}
}