SHELL := /bin/bash
.SILENT:
.DEFAULT_GOAL := help

VENV_DIR=./venv

.PHONY: build
## Build Go binary
build:
	go build -o noisy .

.PHONY: run
## Run the crawler
run:
	go run noisy.go --config config.json

.PHONY: fmt
## Format Go code
fmt:
	go fmt ./...

.PHONY: vet
## Run go vet
vet:
	go vet ./...

.PHONY: lint
## Run linters (fmt, vet)
lint: fmt vet

.PHONY: clean
## Clean build artifacts
clean:
	rm -rf noisy
	find . -type d -name '__pycache__' -exec rm -rf {} +
	find . -type d -name '.pytest_cache' -exec rm -rf {} +

.PHONY: docker.build
## Build docker image
docker.build:
	docker compose -f docker-compose/docker-compose.yml build

.PHONY: docker.run
## Run builded container
docker.run: docker.build
	docker compose -f docker-compose/docker-compose.yml up -d

.PHONY: docker.logs
## Show latest 100 lines of docker logs
docker.logs:
	docker compose -f docker-compose/docker-compose.yml logs --tail=100

.PHONY: docker.logf
## Show latest 100 lines of docker logs and follow
docker.logf:
	docker compose -f docker-compose/docker-compose.yml logs --tail=100 -f

.PHONY: help
## Show this help message
help:
	@echo "$$(tput bold)Available rules:$$(tput sgr0)"
	@echo
	@sed -n -e "/^## / { \
		h; \
		s/.*//; \
		:doc" \
		-e "H; \
		n; \
		s/^## //; \
		t doc" \
		-e "s/:.*//; \
		G; \
		s/\\n## /---/; \
		s/\\n/ /g; \
		p; \
	}" ${MAKEFILE_LIST} \
	| LC_ALL='C' sort --ignore-case \
	| awk -F '---' \
		-v ncol=$$(tput cols) \
		-v indent=19 \
		-v col_on="$$(tput setaf 6)" \
		-v col_off="$$(tput sgr0)" \
	'{ \
		printf "%s%*s%s ", col_on, -indent, $$1, col_off; \
		n = split($$2, words, " "); \
		line_length = ncol - indent; \
		for (i = 1; i <= n; i++) { \
			line_length -= length(words[i]) + 1; \
			if (line_length <= 0) { \
				line_length = ncol - indent - length(words[i]) - 1; \
				printf "\n%*s ", -indent, " "; \
			} \
			printf "%s ", words[i]; \
		} \
		printf "\n"; \
	}' \
	| more $(shell test $(shell uname) == Darwin && echo '--no-init --raw-control-chars')


