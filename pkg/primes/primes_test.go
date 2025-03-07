package primes

import (
	"math/rand"
	"testing"
)

func TestIsPrime(t *testing.T) {
	tests := []struct {
		n    int32
		want bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{17, true},
		{0, false},
		{-1, false},
		{97, true},
	}
	for _, tt := range tests {
		if got := IsPrime(tt.n); got != tt.want {
			t.Errorf("IsPrime(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestGenerateRandomPrime(t *testing.T) {
	// Seed a local RNG for reproducibility in this test
	rng := rand.New(rand.NewSource(42)) // Fixed seed for consistent results

	// Test GenerateRandomPrime with a small max value
	max := int32(100)
	for range 10 { // Run multiple iterations to check variety
		num := GenerateRandomPrime(max, rng)
		if num < 2 || num > max {
			t.Errorf("GenerateRandomPrime(%d) = %d, out of range [2, %d]", max, num, max)
		}
		if !IsPrime(num) {
			t.Errorf("GenerateRandomPrime(%d) = %d, not prime", max, num)
		}
	}

	// Test with different seeds to ensure different outputs
	rng1 := rand.New(rand.NewSource(1))
	rng2 := rand.New(rand.NewSource(2))
	num1 := GenerateRandomPrime(max, rng1)
	num2 := GenerateRandomPrime(max, rng2)
	if num1 == num2 {
		// Not a failure, but a note—randomness should vary with seed
		t.Logf("GenerateRandomPrime produced same number (%d) with different seeds; randomness may need larger range", num1)
	}
}

func TestIsPrime_MaxInt32(t *testing.T) {
	if !IsPrime(2147483647) {
		t.Errorf("IsPrime(2147483647) = false, want true")
	}
}