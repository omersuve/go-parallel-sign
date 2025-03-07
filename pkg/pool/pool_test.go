package pool

import (
	"reflect"
	"testing"
)

func TestNumberPool_Add(t *testing.T) {
	p := NewNumberPool(2)

	// Test adding unique numbers from different clients
	if !p.Add(2, 1) { // Client 1 adds 2
		t.Errorf("Add(2, 1) failed, expected true")
	}
	if !p.Add(3, 2) { // Client 2 adds 3
		t.Errorf("Add(3, 2) failed, expected true")
	}
	if p.Len() != 2 {
		t.Errorf("Len() = %d, want 2", p.Len())
	}

	// Check Get() contains both numbers (order agnostic)
	got := p.Get()
	expected := []int32{2, 3}
	if len(got) != 2 || !containsAll(got, expected) {
		t.Errorf("Get() = %v, want contains [2 3]", got)
	}

	// Check scoreboard
	scoreboard := p.GetScoreboard()
	expectedScoreboard := map[int32]int{1: 1, 2: 1}
	if !reflect.DeepEqual(scoreboard, expectedScoreboard) {
		t.Errorf("GetScoreboard() = %v, want %v", scoreboard, expectedScoreboard)
	}

	// Test adding duplicate
	if p.Add(2, 1) { // Client 1 tries duplicate
		t.Errorf("Add(2, 1) succeeded on duplicate, expected false")
	}
	if p.Len() != 2 {
		t.Errorf("Len() = %d, want 2 after duplicate", p.Len())
	}
	if scoreboard[1] != 1 {
		t.Errorf("Client 1 count = %d, want 1 after duplicate attempt", scoreboard[1])
	}

	// Test adding when pool is full
	if p.Add(5, 1) { // Client 1 tries when full
		t.Errorf("Add(5, 1) succeeded when pool full, expected false")
	}
	if p.Len() != 2 {
		t.Errorf("Len() = %d, want 2 when full", p.Len())
	}
}

// TestGetScoreboard tests the GetScoreboard method
func TestGetScoreboard(t *testing.T) {
	p := NewNumberPool(5)
	if got := p.GetScoreboard(); len(got) != 0 {
		t.Errorf("GetScoreboard() = %v, want empty map for empty pool", got)
	}

	p.Add(10, 1)
	p.Add(20, 1)
	p.Add(30, 2)
	expected := map[int32]int{1: 2, 2: 1}
	got := p.GetScoreboard()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetScoreboard() = %v, want %v", got, expected)
	}

	// Test after duplicate attempt
	p.Add(20, 1) // Duplicate
	got = p.GetScoreboard()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetScoreboard() = %v, want %v after duplicate", got, expected)
	}
}

// Helper function to check if slice contains all expected values (order agnostic)
func containsAll(got, expected []int32) bool {
    if len(got) != len(expected) {
        return false
    }
    gotMap := make(map[int32]bool)
    for _, n := range got {
        gotMap[n] = true
    }
    for _, n := range expected {
        if !gotMap[n] {
            return false
        }
    }
    return true
}