package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	triviagame "github.com/ktenzer/temporal-trivia"
	"github.com/ktenzer/temporal-trivia/resources"
	"github.com/pborman/getopt/v2"
	"go.temporal.io/sdk/client"
)

const version = "1.0.0"

func main() {
	optStartGame := getopt.BoolLong("start-game", 's', "", "Start a new game")
	optGameCategory := getopt.StringLong("category", 'c', "", "Game category general|sports|movies|geography|etc")
	optNumberOfQuestions := getopt.IntLong("questions", 'q', 5, "Total number of questions")
	optAnswerTimeout := getopt.IntLong("answer-timeout", 't', 5, "Time limit per answer phase")
	optResultTimeout := getopt.IntLong("result-timeout", 'r', 5, "Time limit per result phase")
	optMtlsCert := getopt.StringLong("mtls-cert", 'm', "", "Path to mtls cert /path/to/ca.pem")
	optMtlsKey := getopt.StringLong("mtls-key", 'k', "", "Path to mtls key /path/to/ca.key")
	optTemporalEndpoint := getopt.StringLong("temporal-endpoint", 'e', "", "The temporal namespace endpoint")
	optTemporalNamespace := getopt.StringLong("temporal-namespace", 'n', "", "The temporal namespace")
	optGetVersion := getopt.BoolLong("version", 0, "CLI version")
	optHelp := getopt.BoolLong("help", 0, "Help")
	getopt.Parse()

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}

	if *optGetVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	var category string
	var answerTimeout int
	var resultTimeout int
	var questions int
	var chatGptKey string
	var mtlsCert string
	var mtlsKey string
	var temporalEndpoint string
	var temporalNamespace string
	if *optStartGame {
		if getopt.IsSet("category") != true {
			category = "general"
		} else {
			category = *optGameCategory
		}

		if getopt.IsSet("questions") != true {
			questions = 5
		} else {
			questions = *optNumberOfQuestions
		}

		if getopt.IsSet("answer-timeout") != true {
			answerTimeout = 1000
		} else {
			answerTimeout = *optAnswerTimeout
		}

		if getopt.IsSet("result-timeout") != true {
			resultTimeout = 10
		} else {
			resultTimeout = *optResultTimeout
		}

		if os.Getenv("TEMPORAL_TRIVIA_MTLS_CERT") != "" {
			mtlsCert = os.Getenv("TEMPORAL_TRIVIA_MTLS_CERT")
		} else {
			if getopt.IsSet("mtls-cert") == true {
				mtlsCert = *optMtlsCert
			}
		}

		if os.Getenv("TEMPORAL_TRIVIA_MTLS_KEY") != "" {
			mtlsKey = os.Getenv("TEMPORAL_TRIVIA_MTLS_KEY")
		} else {
			if getopt.IsSet("mtls-key") == true {
				mtlsKey = *optMtlsKey
			}
		}

		if os.Getenv("TEMPORAL_TRIVIA_ENDPOINT") != "" {
			temporalEndpoint = os.Getenv("TEMPORAL_TRIVIA_ENDPOINT")
		} else {
			if getopt.IsSet("temporal-endpoint") == true {
				temporalEndpoint = *optTemporalEndpoint
			} else {
				fmt.Println("[ERROR] Missing parameter --temporal-endpoint")
				os.Exit(1)
			}
		}

		if os.Getenv("TEMPORAL_TRIVIA_NAMESPACE") != "" {
			temporalNamespace = os.Getenv("TEMPORAL_TRIVIA_NAMESPACE")
		} else {
			if getopt.IsSet("temporal-namespace") == true {
				temporalNamespace = *optTemporalNamespace
			} else {
				fmt.Println("[ERROR] Missing parameter --temporal-namespace")
				os.Exit(1)
			}
		}

		c := getTemporalClient(temporalEndpoint, temporalNamespace, mtlsCert, mtlsKey)
		defer c.Close()

		workflowId := startGame(c, chatGptKey, category, answerTimeout, resultTimeout, questions)

		failureCounter := 0
		for i := 0; i < questions; {
			for {
				if failureCounter > 10 {
					log.Fatalln("Error exceeded number of failures")
				}
				gameProgress, err := sendProgressQuery(c, workflowId, "getProgress")
				if err != nil {
					fmt.Println("Error sending the Query", err)
				}

				if gameProgress.CurrentQuestion > i+1 {
					fmt.Println("Time is up next question...")
					i++
					break
				}

				questions, err := sendGameQuery(c, workflowId, "getQuestions")
				if err != nil {
					fmt.Println("Error sending the Query", err)
				}

				if questions[i].Question != "" {
					fmt.Println(questions[i].Question)
					keys := make([]string, 0, len(questions[i].MultipleChoiceMap))
					for k := range questions[i].MultipleChoiceMap {
						//fmt.Println(key + " " + value)
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						fmt.Println(k + " " + questions[i].MultipleChoiceMap[k])
					}

					answer := getPlayerResponse()

					gameSignal := resources.Signal{
						Action: "Answer",
						Player: "player0",
						Answer: answer,
					}

					err = sendSignal(c, gameSignal, workflowId, "answer-signal")
					if err != nil {
						fmt.Println("Error sending the Signal", err)
					}

					fmt.Println("Correct Answer: " + questions[i].Answer + "\n")
					i++

					// sleep for showing results
					time.Sleep(time.Duration(resultTimeout) * time.Second)
					break

				}
			}
		}
		getPlayers, err := sendScoreQuery(c, workflowId, "getPlayers")
		if err != nil {
			log.Fatalln("Error sending the Query", err)
		}

		fmt.Println("***** Your Score *****")
		fmt.Println(getPlayers["player0"].Score)
	} else {
		getopt.Usage()
		os.Exit(0)
	}
}

func getTemporalClient(optTemporalEndpoint, optTemporalNamespace, optMtlsCert, optMtlsKey string) client.Client {
	clientOptions := client.Options{
		HostPort:  optTemporalEndpoint,
		Namespace: optTemporalNamespace,
	}

	if optMtlsCert != "" && optMtlsKey != "" {
		cert, err := tls.LoadX509KeyPair(optMtlsCert, optMtlsKey)
		if err != nil {
			log.Fatalln("Unable to load certs", err)
		}

		clientOptions.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
	}

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}

	return c
}

func startGame(c client.Client, chatGptKey, category string, answerTimeout, resultTimeout, questions int) string {
	workflowId := "trivia_game_" + uuid.New().String()
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowId,
		TaskQueue: "trivia-game",
	}

	// Set ChatGPT API Key
	input := resources.WorkflowInput{
		Category:          category,
		NumberOfQuestions: questions,
		NumberOfPlayers:   1,
		AnswerTimeLimit:   answerTimeout,
		ResultTimeLimit:   resultTimeout,
		StartTimeLimit:    300,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, triviagame.TriviaGameWorkflow, input)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	// Add player
	addPlayerSignal := resources.Signal{
		Action: "Player",
		Player: "player0",
	}

	err = sendSignal(c, addPlayerSignal, workflowId, "start-game-signal")
	if err != nil {
		log.Fatalln("Error sending the Signal", err)
	}

	// Start game
	startGameSignal := resources.Signal{
		Action: "StartGame",
	}

	err = sendSignal(c, startGameSignal, workflowId, "start-game-signal")
	if err != nil {
		log.Fatalln("Error sending the Signal", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	return we.GetID()
}

// send game query
func sendGameQuery(c client.Client, workflowId, query string) (map[int]resources.Result, error) {
	resp, err := c.QueryWorkflow(context.Background(), workflowId, "", query)
	if err != nil {
		return nil, err
	}

	var result map[int]resources.Result
	if err := resp.Get(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// send score query
func sendScoreQuery(c client.Client, workflowId, query string) (map[string]resources.Player, error) {
	resp, err := c.QueryWorkflow(context.Background(), workflowId, "", query)
	if err != nil {
		return nil, err
	}

	var result map[string]resources.Player
	if err := resp.Get(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// send progress query
func sendProgressQuery(c client.Client, workflowId, query string) (resources.GameProgress, error) {
	resp, err := c.QueryWorkflow(context.Background(), workflowId, "", query)
	var result resources.GameProgress

	if err != nil {
		return result, err
	}

	if err := resp.Get(&result); err != nil {
		return result, err
	}

	return result, nil
}

func sendSignal(c client.Client, signal resources.Signal, workflowId, signalType string) error {

	err := c.SignalWorkflow(context.Background(), workflowId, "", signalType, signal)
	if err != nil {
		return err
	}

	return nil
}

func getPlayerResponse() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Answer: ")

	answer, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading input:", err)
	}
	answer = answer[:len(answer)-1]

	for {
		if !validateAnswer(answer) {
			fmt.Println("Invalid answer must be a, b, c or d")
			fmt.Print("Answer: ")
			answer, err = reader.ReadString('\n')
			if err != nil {
				log.Fatal("Error reading input:", err)
			}
			answer = answer[:len(answer)-1]
		} else {
			break
		}
	}

	return answer
}

// validate answer
func validateAnswer(answer string) bool {
	re := regexp.MustCompile(`^[A-Da-d]$`)
	isMatch := re.MatchString(answer)

	return isMatch
}

// Sort scores
func sortedScores(scoreMap map[string]int) []string {
	keys := make([]string, 0, len(scoreMap))
	for k := range scoreMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
