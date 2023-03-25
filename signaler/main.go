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
		cert, err = tls.LoadX509KeyPair(os.Getenv("TEMPORAL_TLS_CERT"), os.Getenv("TEMPORAL_TLS_KEY"))
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

	gameSignal := resources.Signal{
		Action: "Answer",
		User:   "John",
		Answer: "mercury",
	}
	err = SendSignal(c, gameSignal, "trivia_game_2741a492-bf3c-42a1-bf85-dc3ec455ff47")
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

func SendQuery(c client.Client, workflowId string) interface{} {

	resp, _ := c.QueryWorkflow(context.Background(), workflowId, "", "state")

	var result interface{}
	if err := resp.Get(&result); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}

	return result
}
