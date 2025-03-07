package pool

import (
	"maps"
	"sync"
)

type NumberPool struct {
	numbers  map[int32]bool
	clients  map[int32]int // Client ID -> count
	mu      sync.Mutex
	max     int
}

func NewNumberPool(max int) *NumberPool {
	return &NumberPool{
		numbers: make(map[int32]bool),
		clients: make(map[int32]int),
		max:     max, // Initialize with max limit
	}
}

// Adds prime number to the pool and increments client count for scoreboard
func (p *NumberPool) Add(num int32, clientID int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if len(p.numbers) >= p.max || p.numbers[num] {
        return false
    }
    p.numbers[num] = true
    p.clients[clientID]++
    return true
}

func (p *NumberPool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.numbers)
}

// Gets the prime numbers in the pool as a slice
func (p *NumberPool) Get() []int32 {
    p.mu.Lock()
    defer p.mu.Unlock()

    var nums []int32
    for num := range p.numbers {
        nums = append(nums, num)
    }
    return nums
}

// Copy to prevent the scoreboard from being modified by the client, as it's supposed to be immutable from the outside
func (p *NumberPool) GetScoreboard() map[int32]int {
    p.mu.Lock()
    defer p.mu.Unlock()

    scoreboard := make(map[int32]int)
    maps.Copy(scoreboard, p.clients) // Copy to be safe, make it unmodifiable
    return scoreboard
}
