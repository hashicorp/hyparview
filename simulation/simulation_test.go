package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	h "github.com/hashicorp/hyparview"
	"github.com/stretchr/testify/assert"
)

// TestSimulation is the test entry point
func TestSimulation(t *testing.T) {
	for i := 0; i < 3; i++ {
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
		rounds:     5,
		peers:      1000,
		mortality:  30,
		drainDepth: 30,
		payloads:   30,
		gossipHeat: 4,
		iteration:  i,
		fail: WorldFailureRate{
			active:      30,
			shuffle:     30,
			reply:       30,
			gossip:      5,
			gossipReply: 5,
		},
	})

	assert.True(t, w.isConnected())
	assert.Equal(t, 0, len(w.queue))

	// w.debugQueue()
	w.plotSeed(seed)
	w.plotInDegree()
	w.plotGossip()
}
