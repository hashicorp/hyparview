export SIMULATION_COUNT=1

simulation: ## run the simulation test
	mkdir -p plot data
	cd simulation && go test -v . || exit 0
	make plot
.PHONY: simulation

plot: ## make plots from simulation data
	./bin/plot-degree "In Degree" "in-active" $(SIMULATION_COUNT) > plot/in-degree.png
	./bin/plot-gossip $(SIMULATION_COUNT) > plot/gossip.png
.PHONY: plot

test:
	go test
.PHONY: test
