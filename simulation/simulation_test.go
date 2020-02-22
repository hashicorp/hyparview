package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"testing"

	h "github.com/hashicorp/hyparview"
	"github.com/stretchr/testify/assert"
)

// TestSimulation is the test entry point
func TestSimulation(t *testing.T) {
	count := 1
	countEnv := os.Getenv("SIMULATION_COUNT")
	if countEnv != "" {
		conv, err := strconv.Atoi(countEnv)
		if err == nil {
			count = conv
		}
	}

	for i := 1; i <= count; i++ {
		testSimulation(t, i)
	}
}

// testSimulation is the entry point to test a single world
// World configuration and assertion goes here
func testSimulation(t *testing.T, i int) {
	seed := h.Rint64Crypto(math.MaxInt64 - 1)
	rand.Seed(seed)
	fmt.Printf("Seed %d\n", seed)

	w := simulation(WorldConfig{
		rounds:      5,
		peers:       1000,
		mortality:   30,
		payloads:    30,
		gossipHeat:  4,
		iteration:   i,
		shuffleFreq: 30,
		failureRate: 10,
	})

	assert.NoError(t, w.Connected())

	// w.debugQueue()
	w.plotSeed(seed)
	w.plotInDegree()
	w.plotOutDegree()
	w.plotGossip()
}
