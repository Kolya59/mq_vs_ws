package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	CloudMQTTUrl = "mqtt://%s:%s@%s:%d/%s"
)

var opts struct {
	BrokerHost string `long:"host" env:"HOST" description:"Host" required:"true"`
	BrokerPort int    `long:"port" env:"PORT" description:"Port" required:"true"`
	User       string `long:"user" env:"USER" description:"Username" required:"true"`
	Password   string `long:"password" env:"PASS" description:"Password" required:"true"`
	Topic      string `long:"topic" env:"TOPIC" description:"Topic" required:"true"`

	LogLevel string `long:"log_level" env:"LOG_LEVEL" description:"Log level for zerolog" required:"false"`
}

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
	// Log initialization
	zerolog.MessageFieldName = "MESSAGE"
	zerolog.LevelFieldName = "LEVEL"
	zerolog.ErrorFieldName = "ERROR"
	zerolog.TimestampFieldName = "TIME"
	zerolog.CallerFieldName = "CALLER"
	log.Logger = log.Output(os.Stderr).With().Str("PROGRAM", "firmware-update-server").Caller().Logger()

	// Parse flags
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal().Msgf("Could not parse flags: %v", err)
	}

	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	uri, err := url.Parse(fmt.Sprintf(CloudMQTTUrl, opts.User, opts.Password, opts.BrokerHost, opts.BrokerPort, opts.Topic))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse uri")
	}
	topic := uri.Path[1:len(uri.Path)]
	if topic == "" {
		topic = "test"
	}

	SendTime(uri, topic)
}
