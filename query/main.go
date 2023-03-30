package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"crypto/tls"
	"crypto/x509"

	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.

	clientOptions := client.Options{
		HostPort:  os.Getenv("TEMPORAL_HOST_URL"),
		Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
	}

	if os.Getenv("TEMPORAL_MTLS_TLS_CERT") != "" && os.Getenv("TEMPORAL_MTLS_TLS_KEY") != "" {
		if os.Getenv("TEMPORAL_MTLS_TLS_CA") != "" {
			caCert, err := os.ReadFile(os.Getenv("TEMPORAL_MTLS_TLS_CA"))
			if err != nil {
				log.Fatalln("failed reading server CA's certificate", err)
			}

			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(caCert) {
				log.Fatalln("failed to add server CA's certificate", err)
			}

			cert, err := tls.LoadX509KeyPair(os.Getenv("TEMPORAL_MTLS_TLS_CERT"), os.Getenv("TEMPORAL_MTLS_TLS_KEY"))
			if err != nil {
				log.Fatalln("Unable to load certs", err)
			}

			var serverName string
			if os.Getenv("TEMPORAL_MTLS_TLS_ENABLE_HOST_VERIFICATION") == "true" {
				serverName = os.Getenv("TEMPORAL_MTLS_TLS_SERVER_NAME")
			}

			clientOptions.ConnectionOptions = client.ConnectionOptions{
				TLS: &tls.Config{
					RootCAs:      certPool,
					Certificates: []tls.Certificate{cert},
					ServerName:   serverName,
				},
			}
		} else {
			cert, err := tls.LoadX509KeyPair(os.Getenv("TEMPORAL_MTLS_TLS_CERT"), os.Getenv("TEMPORAL_MTLS_TLS_KEY"))
			if err != nil {
				log.Fatalln("Unable to load certs", err)
			}

			clientOptions.ConnectionOptions = client.ConnectionOptions{
				TLS: &tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			}
		}
	}

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowId := os.Getenv("TEMPORAL_WORKFLOW_ID")
	if workflowId == "" {
		log.Fatal("Please specify workflowId by exporting environment variable TEMPORAL_WORKFLOW_ID")
	}

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

	gameProgress, err := SendQuery(c, workflowId, "getProgress")
	if err != nil {
		log.Fatalln("Error sending the Query", err)
	}
	fmt.Println(gameProgress)
}

func SendQuery(c client.Client, workflowId, query string) (interface{}, error) {

	resp, _ := c.QueryWorkflow(context.Background(), workflowId, "", query)

	var result interface{}
	if err := resp.Get(&result); err != nil {
		return nil, err
	}

	return result, nil
}
