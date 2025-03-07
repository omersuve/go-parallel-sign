package primes

import (
	"math/rand"
)

// Checks if a number is prime
func IsPrime(n int32) bool {
	if n < 2 {
		return false
	}
	// Cap i at the largest integer which has square that does not overflow int32
	for i := int32(2); i <= 46340 && i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// Generates a random prime number using the provided RNG
func GenerateRandomPrime(max int32, rng *rand.Rand) int32 {
	for {
		n := rng.Int31n(max-1) + 2 // Random positive number between 2 and max (inclusive)
		if IsPrime(n) {
			return n
		}
	}
}