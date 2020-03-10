export SIMULATION_COUNT=1

test: message-generated.go
	go test
.PHONY: test

simulation: ## run the simulation test
	mkdir -p plot data
	cd simulation && go test -v . || exit 0
	make plot
.PHONY: simulation

plot: ## make plots from simulation data
	./bin/plot-degree "In Degree" "in-active" $(SIMULATION_COUNT) > plot/in-degree.png
	./bin/plot-gossip $(SIMULATION_COUNT) > plot/gossip.png
.PHONY: plot

message-generated.go: message.go message.go.genny
	ln message.go.genny tmp.go
	go generate
	mv gen-tmp.go $@
	rm tmp.go
