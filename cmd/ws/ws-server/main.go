package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jessevdk/go-flags"
	"github.com/psu/mq_vs_ws/pkg/crypt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var opts struct {
	DbHost     string `long:"database_host" env:"DB_HOST" description:"Database host" required:"true"`
	DbPort     string `long:"database_port" env:"DB_PORT" description:"Database port" required:"true"`
	DbName     string `long:"database_name" env:"DB_NAME" description:"Database name" required:"true"`
	DbUser     string `long:"database_username" env:"DB_USER" description:"Database username" required:"true"`
	DbPassword string `long:"database_password" env:"DB_PASSWORD" description:"Database password" required:"true"`

	Host string `long:"host" env:"HOST" description:"Host" required:"true"`
	Port int    `long:"port" env:"PORT" description:"Port" required:"true"`

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
		log.Fatal().Msgf("Could not parse flags: %v", err)
	}

	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Setup routs
	r := mux.NewRouter()
	r.HandleFunc("/", handler)

	/*// Setup DB
	err = postgresdriver.InitDatabaseConnection(opts.DbHost, opts.DbPort, opts.DbUser, opts.DbPassword, opts.DbName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set DB connection")
	}*/

	caCert, err := ioutil.ReadFile("client.crt")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read client.crt")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	cfg := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caCertPool,
	}
	srv := &http.Server{
		Addr:      fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler:   r,
		TLSConfig: cfg,
	}

	// TODO Add graceful shutdown
	if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil {
		log.Error().Err(err).Msg("Failed to stop server")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrader")
		return
	}
	defer c.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	timer := time.NewTimer(ttl)
	defer timer.Stop()

	done := make(chan interface{})
	go func() {
		for {
			t, _, err := c.ReadMessage()
			if err != nil || t == websocket.CloseMessage {
				close(done)
				break
			}
		}
	}()

	for {
		select {
		case t := <-ticker.C:
			msg := fmt.Sprintf("Current time is: %v", t.String())
			hashed := crypt.Encrypt([]byte(msg), string(websocket.TextMessage))
			err = c.WriteMessage(websocket.BinaryMessage, hashed)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to write msg: %s", msg)
				return
			}
			log.Info().Msgf("Send %v", msg)
		case <-timer.C:
			log.Warn().Msg("Connection broken by ttl")
			return
		case <-done:
			log.Info().Msg("Connection has been closed by client")
			return
		}
	}
}

/*// Parse data
var newCar car.Car
if err := json.Unmarshal(message, newCar); err != nil {
	log.Error().Err(err).Msgf("Failed to unmarshal msg: %v", message)
	continue
}

// Save data
if err := postgresdriver.SendData(&newCar); err != nil {
	log.Error().Err(err).Msg("Failed to send car to pq")
	err = c.WriteMessage(0, []byte("Failed to send car to pq"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to send err response")
	}
}*/
