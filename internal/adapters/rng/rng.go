package rng

import (
	"math/rand"
	"time"

	"github.com/csg33k/droid-tapper/internal/domain/game"
)

type Rand struct {
	r *rand.Rand
}

func New(seed int64) *Rand {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Rand{r: rand.New(rand.NewSource(seed))}
}

func (r *Rand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return r.r.Intn(n)
}

func (r *Rand) Float64() float64 {
	return r.r.Float64()
}

var _ game.RNG = (*Rand)(nil)
