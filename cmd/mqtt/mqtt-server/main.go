package main

import (
	"fmt"
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	User         = "kolya59"
	Password     = "123456"
	Host         = "127.0.0.1"
	Port         = "1883"
	Topic        = "test"
	CloudMQTTUrl = "mqtt://%s:%s@%s.cloudmqtt.com:%s/%s"
)

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to broker")
	}
	return client
}

func SendTime(uri *url.URL, topic string) {
	client := connect("sub", uri)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			client.Publish(topic, 0, false, fmt.Sprintf("Current time: %v", t))
		default:
			time.Sleep(10 * time.Second)
		}
	}
}

func main() {
	uri, err := url.Parse(fmt.Sprintf(CloudMQTTUrl, User, Password, Host, Port, Topic))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse uri")
	}
	topic := uri.Path[1:len(uri.Path)]
	if topic == "" {
		topic = "test"
	}

	SendTime(uri, topic)
}
