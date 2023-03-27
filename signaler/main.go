package main

import (
	"context"
	"log"
	"os"

	"crypto/tls"

	"github.com/ktenzer/triviagame/resources"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.

	var c client.Client
	var err error
	var cert tls.Certificate

	if os.Getenv("MTLS") == "false" {
		c, err = client.Dial(client.Options{
			HostPort:  os.Getenv("TEMPORAL_HOST_URL"),
			Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
		})
	} else {
		cert, err = tls.LoadX509KeyPair(os.Getenv("TEMPORAL_MTLS_TLS_CERT"), os.Getenv("TEMPORAL_MTLS_TLS_KEY"))
		if err != nil {
			log.Fatalln("Unable to load certs", err)
		}

		c, err = client.Dial(client.Options{
			HostPort:  os.Getenv("TEMPORAL_HOST_URL"),
			Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
			ConnectionOptions: client.ConnectionOptions{
				TLS: &tls.Config{Certificates: []tls.Certificate{cert}},
			},
		})
	}

	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowId := "trivia_game_1b5ca8b6-27e9-488f-95d6-a16d1a0265da"
	gameSignal := resources.Signal{
		Action: "Answer",
		User:   "Keith",
		Answer: "A",
	}

	err = SendSignal(c, gameSignal, workflowId)
	if err != nil {
		log.Fatalln("Error sending the Signal", err)
	}
}

func SendSignal(c client.Client, signal resources.Signal, workflowId string) error {

	err := c.SignalWorkflow(context.Background(), workflowId, "", "game-signal", signal)
	if err != nil {
		return err
	}

	log.Println("Workflow[" + workflowId + "] Signaled")

	return nil
}
