package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"testing"

	h "github.com/hashicorp/hyparview"
)

// TestSimulation is the test entry point
func TestSimulation(t *testing.T) {
	count := 1
	conv, err := strconv.Atoi(os.Getenv("SIMULATION_COUNT"))
	if err == nil {
		count = conv
	}

	for i := 1; i <= count; i++ {
		testSimulation(t, i, 1000)
	}
}

// testSimulation is the entry point to test a single world
// World configuration and assertion goes here
func testSimulation(t *testing.T, i int, peers int) {
	seed := h.Rint64Crypto(math.MaxInt64 - 1)
	rand.Seed(seed)
	fmt.Printf("Seed %d\n", seed)

	w := simulation(WorldConfig{
		peers:       peers,
		payloads:    30,
		gossipHeat:  4,
		iteration:   i,
		shuffleFreq: 30,
		failureRate: 0,
	})

	err := w.Connected()
	if err != nil {
		t.Fail("graph disconnected: %s", err.Error())
	}

	err = w.isSymmetric()
	if err != nil {
		t.Fail("active view asymmetric: %s", err.Error())
	}

	// w.debugQueue()
	w.plotSeed(seed)
	w.plotInDegree()
	w.plotOutDegree()
	w.plotGossip()
}
