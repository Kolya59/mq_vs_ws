package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jessevdk/go-flags"
	"github.com/psu/mq_vs_ws/pkg/crypt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var opts struct {
	RedisServer   string `long:"redis_server" env:"REDIS_SERVER" description:"Redis servers" required:"false"`
	RedisPassword string `long:"redis_password" env:"REDIS_PASSWORD" description:"Password for servers" required:"false"`
	RedisDatabase int    `long:"redis_database" env:"REDIS_DATABASE" description:"Redis database" required:"false"`

	Host string `long:"crud_conn_host" env:"HOST" description:"Host" required:"true"`
	Port int    `long:"crud_conn_port" env:"PORT" description:"Port" required:"true"`

	LogLevel string `long:"log_level" env:"LOG_LEVEL" description:"Log level for zerolog" required:"false"`
}

const (
	ttl = time.Minute
)

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
		log.Panic().Msgf("Could not parse flags: %v", err)
	}

	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Setup url
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", opts.Host, opts.Port), Path: "/"}
	log.Info().Msgf("connecting to %s", u.String())

	// Setup HTTPS
	caCert, err := ioutil.ReadFile("server.crt")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read server sert file")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read client sert file")
	}

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	// Setup ws dial
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("dial")
	}
	defer c.Close()

	// Setup timers
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.Info().Msg("Stop timer")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error().Err(err).Msg("Failed to write close msg")
				return
			}
			return
		default:
			t, msg, err := c.ReadMessage()
			if err != nil {
				log.Error().Err(err).Msgf("Failed to write msg: %s", msg)
				return
			}
			if t == websocket.CloseMessage {
				log.Warn().Msg("Connection was closed by server")
				return
			}
			unhashed := crypt.Decrypt(msg, string(websocket.TextMessage))
			log.Info().Msgf("Recd %s", unhashed)
		}
	}
}
