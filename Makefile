all: mqtt-server mqtt-client ws-server ws-client

# MQTT Server
mqtt-server: build-mqtt-server run-mqtt-server

build-mqtt-server:
	go build -o ./bin/mqtt-server.bin ./cmd/mqtt/mqtt-server

run-mqtt-server:
	HOST=farmer.cloudmqtt.com PORT=12548 USER=szaejwok PASS=7_9su2mbRYiB TOPIC=time ./bin/mqtt-server.bin

# MQTT client
mqtt-client: build-mqtt-client run-mqtt-client

build-mqtt-client:
	go build -o ./bin/mqtt-client.bin ./cmd/mqtt/mqtt-client

run-mqtt-client:
	HOST=farmer.cloudmqtt.com PORT=12548 USER=szaejwok PASS=7_9su2mbRYiB TOPIC=time ./bin/mqtt-client.bin

# WebSocket server
ws-server: build-ws-server run-ws-server

build-ws-server:
	go build -o ./bin/ws-server.bin ./cmd/ws/ws-server

run-ws-server:
	HOST=localhost PORT=8081 ./bin/ws-server.bin

# WebSocket client
ws-client: build-ws-client run-ws-client

build-ws-client:
	go build -o ./bin/ws-client.bin ./cmd/ws/ws-client

run-ws-client:
	HOST=176.34.152.125 PORT=8081 ./bin/ws-client.bin