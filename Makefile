all: mqtt-server mqtt-client ws-server ws-client

# MQTT Server
mqtt-server: build-mqtt-server run-mqtt-server

build-mqtt-server:
	go build -o ./bin/mqtt-server.bin ./cmd/mqtt/mqtt-server

run-mqtt-server:
	HOST=127.0.0.1 PORT=8080 DB_HOST=localhost DB_PORT=5432 DB_NAME=easy_normalization DB_USER=kolya59 DB_PASSWORD=12334566w ./bin/mqtt-server.bin

# MQTT client
mqtt-client: build-mqtt-client run-mqtt-client

build-mqtt-client:
	./cmd/mqtt/mqtt-client

run-mqtt-client:
	HOST=127.0.0.1 PORT=8080 ./bin/mqtt-client.bin

# WebSocket server
ws-server: build-ws-server run-ws-server

build-ws-server:
	go build -o ./bin/ws-server.bin ./cmd/ws/ws-server

run-ws-server:
	HOST=127.0.0.1 PORT=8081 DB_HOST=localhost DB_PORT=5432 DB_NAME=easy_normalization DB_USER=kolya59 DB_PASSWORD=12334566w ./bin/ws-server.bin

# WebSocket client
ws-client: build-ws-client run-ws-client

build-ws-client:
	go build -o ./bin/ws-client.bin ./cmd/ws/ws-client

run-ws-client:
	HOST=ec2-176-34-152-125.eu-west-1.compute.amazonaws.com PORT=8081 ./bin/ws-client.bin