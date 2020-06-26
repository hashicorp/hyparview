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
	peers, count := 1000, 1
	conv, err := strconv.Atoi(os.Getenv("SIMULATION_COUNT"))
	if err == nil {
		count = conv
	}

	conv, err = strconv.Atoi(os.Getenv("SIMULATION_PEERS"))
	if err == nil {
		peers = conv
	}

	var seed int64
	envSeed, err := strconv.ParseInt(os.Getenv("SIMULATION_SEED"), 10, 64)
	if err == nil {
		seed = envSeed
	}

	for i := 1; i <= count; i++ {
		if envSeed == 0 {
			seed = h.Rint64Crypto(math.MaxInt64 - 1)
		}

		testSimulation(t, i, peers, seed)
	}
}

// testSimulation is the entry point to test a single world
// World configuration and assertion goes here
func testSimulation(t *testing.T, i int, peers int, seed int64) {
	rand.Seed(seed)
	fmt.Printf("world: %d seed: %d peers: %d\n", i, seed, peers)

	w := simulation(WorldConfig{
		peers:       peers,
		payloads:    30,
		iteration:   i,
		failureRate: 10,
		gossips:     200,
	})

	err := w.Connected()
	if err != nil {
		t.Errorf("world %d: graph disconnected: %s", i, err.Error())
	}

	// This isn't an error. It's useful for working on symmetry, but because of the
	// failure rate, there's always a tail of asymmetries
	// err = w.isSymmetric()
	// if err != nil {
	// 	t.Logf("run %d: active view asymmetric: %s", i, err.Error())
	// }

	// w.debugQueue()
	// w.plotPeer("n2375")
	w.mkdir()
	w.plotSeed(seed)
	w.plotBootstrapCount()
	w.plotInDegree()
	w.plotOutDegree()
	w.plotGossip()
	w.plotGraphs()
}
