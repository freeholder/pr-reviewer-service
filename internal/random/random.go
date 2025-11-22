package random

import (
	"math/rand"
)

type Randomizer interface {
	Intn(n int) int
}

type DefaultRandomizer struct{}

func (DefaultRandomizer) Intn(n int) int {
	return rand.Intn(n)
}
