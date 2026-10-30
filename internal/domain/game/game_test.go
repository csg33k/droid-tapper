package game

import (
	"testing"
)

type fakeRNG struct {
	ints   []int
	intIdx int
}

func (f *fakeRNG) Intn(n int) int {
	if f.intIdx < len(f.ints) {
		v := f.ints[f.intIdx]
		f.intIdx++
		if n <= 0 {
			return 0
		}
		return v % n
	}
	return 0
}
func (f *fakeRNG) Float64() float64 { return 0.5 }

func TestServeCreatesMug(t *testing.T) {
	g := NewGame(&fakeRNG{}, DefaultConfig(), 0)
	g.Start()
	g.Serve()
	if len(g.Mugs) != 1 {
		t.Fatalf("expected 1 mug, got %d", len(g.Mugs))
	}
	if !g.Mugs[0].Filled {
		t.Fatal("expected filled mug")
	}
}

func TestFilledMugHitsCustomer(t *testing.T) {
	g := NewGame(&fakeRNG{}, DefaultConfig(), 0)
	g.Start()
	g.SpawnedCount = g.TargetCount
	g.Customers = []Customer{{ID: 1, BarID: 0, X: 500, State: CustomerApproaching, Speed: 50}}
	g.Mugs = []Mug{{ID: 2, BarID: 0, X: 495, Filled: true}}
	g.Tick(1.0 / 60.0)
	if g.Score != g.Config.ServeScore {
		t.Fatalf("score = %d, want %d", g.Score, g.Config.ServeScore)
	}
	if len(g.Mugs) != 1 || g.Mugs[0].Filled {
		t.Fatalf("expected one empty mug, got %+v", g.Mugs)
	}
	if len(g.Customers) != 1 || g.Customers[0].State != CustomerSatisfied {
		t.Fatalf("customer not satisfied: %+v", g.Customers)
	}
}

func TestCatchEmptyMug(t *testing.T) {
	cfg := DefaultConfig()
	g := NewGame(&fakeRNG{}, cfg, 0)
	g.Start()
	g.SpawnedCount = g.TargetCount
	g.Bartender.BarID = 0
	g.Bartender.X = 80
	g.Mugs = []Mug{{ID: 1, BarID: 0, X: 100, Filled: false}}
	g.Tick(1.0 / 60.0)
	if g.Score != cfg.CatchScore {
		t.Fatalf("score = %d, want %d", g.Score, cfg.CatchScore)
	}
	if len(g.Mugs) != 0 {
		t.Fatalf("expected mug caught/removed, got %d", len(g.Mugs))
	}
}

func TestEmptyMugBreakLosesLife(t *testing.T) {
	cfg := DefaultConfig()
	g := NewGame(&fakeRNG{}, cfg, 0)
	g.Start()
	g.SpawnedCount = g.TargetCount
	g.Bartender.BarID = 1 // elsewhere
	g.Bartender.X = 80
	g.Mugs = []Mug{{ID: 1, BarID: 0, X: 1, Filled: false}}
	before := g.Lives
	g.Tick(1.0 / 60.0)
	if g.Lives != before-1 {
		t.Fatalf("lives = %d, want %d", g.Lives, before-1)
	}
}

func TestCustomerReachLosesLife(t *testing.T) {
	cfg := DefaultConfig()
	g := NewGame(&fakeRNG{}, cfg, 0)
	g.Start()
	g.SpawnedCount = g.TargetCount
	g.Customers = []Customer{{ID: 1, BarID: 0, X: 80, State: CustomerApproaching, Speed: 50}}
	before := g.Lives
	g.Tick(1.0 / 60.0)
	if g.Lives != before-1 {
		t.Fatalf("lives = %d, want %d", g.Lives, before-1)
	}
}

func TestLevelCompletion(t *testing.T) {
	cfg := DefaultConfig()
	g := NewGame(&fakeRNG{}, cfg, 0)
	g.Start()
	g.SpawnedCount = g.TargetCount
	// no customers/mugs
	g.Tick(cfg.LevelClearDelay + 0.1)
	if g.Level != 2 {
		t.Fatalf("level = %d, want 2", g.Level)
	}
}

func TestGameOverOnZeroLives(t *testing.T) {
	cfg := DefaultConfig()
	g := NewGame(&fakeRNG{}, cfg, 0)
	g.Start()
	g.Lives = 1
	g.SpawnedCount = g.TargetCount
	g.Customers = []Customer{{ID: 1, BarID: 0, X: 80, State: CustomerApproaching, Speed: 50}}
	g.Tick(1.0 / 60.0)
	if g.Status != StatusGameOver {
		t.Fatalf("status = %v, want GameOver", g.Status)
	}
}
