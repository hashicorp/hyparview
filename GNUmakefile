export SIMULATION_COUNT=3

simulation: ## run the simulation test
	mkdir -p plot data
	cd simulation && gotestsum
	make plot
.PHONY: simulation

plot: ## make plots from simulation data
	./bin/plot-degree "In Degree" "active" $(SIMULATION_COUNT) > plot/degree.png
	./bin/plot-gossip $(SIMULATION_COUNT) > plot/gossip.png
.PHONY: plot

# via https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
