test: .build-test ## run the default tests
.PHONY: test

.build-test: plot/active.png plot/passive.png
	touch $@

plot:
	mkdir plot

data:
	mkdir data

plot/%.png: data/% plot data
	./bin/plot $^ $* > $@

# via https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
