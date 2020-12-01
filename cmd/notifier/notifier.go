package main

import (
	"net/http"
	"os"

	"github.com/alexlast/stock-notifier/internal/notifier"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
		},
	})
}

func main() {
	// Load config
	config := new(notifier.Config)
	err := envconfig.Process("notifier", config)

	if err != nil {
		log.Fatalln(err)
	}

	// Dynamically set the log level
	switch config.LogLevel {
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Create a new AWS session
	session := session.New(
		&aws.Config{
			Region: aws.String(config.AWSRegion),
		},
	)

	// Build new clients
	c := &notifier.Context{
		SES:    ses.New(session),
		SNS:    sns.New(session),
		HTTP:   http.DefaultClient,
		Config: config,
	}

	// Were ready to start
	log.Infoln("Starting stock-notifier")

	// Start polling
	c.Start()
}
