.PHONY: build run test clean docker-build docker-run

# Default values
PORT ?= 8080
SUPER_ADMIN_KEY ?= super_admin_key
CONFIG_FILE_PATH ?= ./config/config.json
AUTO_SAVE_INTERVAL ?= 60

# Build the application
build:
	go build -o cron-server .

# Run the application
run: build
	PORT=$(PORT) SUPER_ADMIN_KEY=$(SUPER_ADMIN_KEY) CONFIG_FILE_PATH=$(CONFIG_FILE_PATH) AUTO_SAVE_INTERVAL=$(AUTO_SAVE_INTERVAL) ./cron-server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f cron-server

# Build Docker image
docker-build:
	docker build -t cron-server .

# Run Docker container
docker-run: docker-build
	docker run -p $(PORT):8080 \
		-e PORT=8080 \
		-e SUPER_ADMIN_KEY=$(SUPER_ADMIN_KEY) \
		-e CONFIG_FILE_PATH=$(CONFIG_FILE_PATH) \
		-e AUTO_SAVE_INTERVAL=$(AUTO_SAVE_INTERVAL) \
		-v $(PWD)/config:/config \
		cron-server

# Run with Docker Compose
docker-compose-up:
	docker-compose up -d

# Stop Docker Compose services
docker-compose-down:
	docker-compose down
