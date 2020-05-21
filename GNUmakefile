export SIMULATION_PEERS=1000
export SIMULATION_COUNT=3
export SIMULATION_SEED=

simulation: ## run the simulation test
	mkdir -p data
	go test -v ./simulation

plot: ## make plots from simulation data
	mkdir -p plot
	./bin/plot-degree "In Degree" "in-active" $(SIMULATION_COUNT) > plot/in-degree.png
	./bin/plot-all-gossip $(SIMULATION_COUNT) > plot/gossip.png

plot-slow: ## more plots, but these are slow
	mkdir -p plot
	./bin/plot-all-graphs

test: message-generated.go
	go test

message-generated.go: message.go message.go.genny
	ln message.go.genny tmp.go
	go generate
	mv gen-tmp.go $@
	rm tmp.go

.PHONY: simulation plot test
.DEFAULT_GOAL: test
