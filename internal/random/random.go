package random

import (
	"math/rand"
	"time"
)

type Randomizer interface {
	Intn(n int) int
}

type DefaultRandomizer struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (DefaultRandomizer) Intn(n int) int {
	return rand.Intn(n)
}
