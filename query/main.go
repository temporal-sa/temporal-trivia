package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"crypto/tls"

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

	workflowId := "trivia_game_0dfabfb4-4497-4b35-8890-72cae9a5ba65"

	gameMap, err := SendQuery(c, workflowId, "getDetails")
	if err != nil {
		log.Fatalln("Error sending the Query", err)
	}
	fmt.Println(gameMap)

	scoreMap, err := SendQuery(c, workflowId, "getScore")
	if err != nil {
		log.Fatalln("Error sending the Query", err)
	}
	fmt.Println(scoreMap)
}

func SendQuery(c client.Client, workflowId, query string) (interface{}, error) {

	resp, _ := c.QueryWorkflow(context.Background(), workflowId, "", query)

	var result interface{}
	if err := resp.Get(&result); err != nil {
		return nil, err
	}

	return result, nil
}
