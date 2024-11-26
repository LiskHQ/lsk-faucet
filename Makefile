include .env

APP_NAME = lsk-faucet
PKGS=$(shell go list ./... | grep -v "/vendor/")
BLUE = \033[1;34m
GREEN = \033[1;32m
COLOR_END = \033[0;39m
BINDING_OUTPUT_DIR=$(realpath ./internal/bindings)

validate-env: # Validates if the necessary environment variables are set
ifndef ERC20_CONTRACT_FILE_PATH
	$(error `ERC20_CONTRACT_FILE_PATH` environment variable is undefined)
endif

build: build-frontend build-backend

build-backend: # Builds the application and create a binary at ./bin/
	@echo "$(BLUE)» Building $(APP_NAME) application binary... $(COLOR_END)"
	go build -a -o bin/$(APP_NAME) .
	@echo "$(GREEN)Binary successfully built$(COLOR_END)"

build-frontend: # Builds the frontned application
	@echo "$(BLUE)» Building frontend... $(COLOR_END)"
	go generate
	@echo "$(GREEN)Frontend successfully built$(COLOR_END)"

run: # Runs the application, use `make run FLAGS="--help"`
	./bin/${APP_NAME} ${FLAGS}

generate-abi: validate-env
	@echo "$(BLUE)Generating abi file... $(COLOR_END)"
	solc --overwrite --abi $(ERC20_CONTRACT_FILE_PATH) -o $(BINDING_OUTPUT_DIR)
	@echo "$(GREEN)ABI file generated successfully$(COLOR_END)"

generate-binding: generate-abi
	@echo "$(BLUE)Generating go binding for the ERC20 token smart contract... $(COLOR_END)"
	abigen --abi=$(shell find $(BINDING_OUTPUT_DIR) -type f -name '*.abi') --pkg=bindings --out=$(BINDING_OUTPUT_DIR)/erc20.go
	@echo "$(GREEN)Go binding generated successfully$(COLOR_END)"

test: # Runs tests
	@echo "Test packages"
	go test -race -shuffle=on -coverprofile=coverage.out -cover $(PKGS)

lint: # Runs golangci-lint on the repo
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

format: # Runs gofmt on the repo
	gofmt -s -w .

build-image: # Builds docker image
	@echo "$(BLUE)Building docker image...$(COLOR_END)"
	docker build -t $(APP_NAME) .

docker-start: # Runs docker image
	@echo "$(BLUE)Starting docker container $(APP_NAME)...$(COLOR_END)"
ifdef PRIVATE_KEY
	docker run --name $(APP_NAME) -p 8080:8080 -d --env-file .env $(APP_NAME)
else ifdef KEYSTORE
	docker run --name $(APP_NAME) -p 8080:8080 -d --env-file .env -v $(KEYSTORE)/keystore:/app/keystore -v $(KEYSTORE)/password.txt:/app/password.txt $(APP_NAME)
endif

docker-stop:
	@echo "$(BLUE)Stopping and removing docker container $(APP_NAME)...$(COLOR_END)"
	docker rm -f $(APP_NAME)

.PHONY: help
help: # Show help for each of the Makefile recipes
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "$(GREEN)$$(echo $$l | cut -f 1 -d':')$(COLOR_END):$$(echo $$l | cut -f 2- -d'#')\n"; done
