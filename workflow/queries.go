package triviagame

import (
	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"
)

type GameProgress struct {
	NumberOfQuestions int    `json:"numberOfQuestions"`
	CurrentQuestion   int    `json:"currentQuestion"`
	Stage             string `json:"stage"`
}

// Setup query handler for players
func initGetPlayersQuery(ctx workflow.Context) (*map[string]resources.Player, error) {
	log := workflow.GetLogger(ctx)
	getPlayers := make(map[string]resources.Player)

	err := workflow.SetQueryHandler(ctx, "getPlayers", func(input []byte) (map[string]resources.Player, error) {
		return getPlayers, nil
	})
	if err != nil {
		log.Error("SetQueryHandler failed for getPlayers: " + err.Error())
		return &getPlayers, err
	}

	return &getPlayers, nil
}

// Setup query handler for gathering game questions
func initGetQuestionsQuery(ctx workflow.Context) (*map[int]resources.Result, error) {
	log := workflow.GetLogger(ctx)
	getQuestions := make(map[int]resources.Result)

	err := workflow.SetQueryHandler(ctx, "getQuestions", func(input []byte) (map[int]resources.Result, error) {
		return getQuestions, nil
	})
	if err != nil {
		log.Error("SetQueryHandler failed for getQuestions: " + err.Error())
		return &getQuestions, err
	}

	return &getQuestions, nil
}

// Setup query handler for gathering game progress
func (gp *GameProgress) initGetProgressQuery(ctx workflow.Context, numberQuestions int) (GameProgress, error) {
	log := workflow.GetLogger(ctx)

	gp.NumberOfQuestions = numberQuestions
	gp.CurrentQuestion = 0
	gp.Stage = "start"

	err := workflow.SetQueryHandler(ctx, "getProgress", func(input []byte) (GameProgress, error) {
		return *gp, nil
	})
	if err != nil {
		log.Error("SetQueryHandler failed for getProgress: " + err.Error())
		return *gp, err
	}

	return *gp, nil
}

// Setup query handler for games
func initGamesQuery(ctx workflow.Context) (*map[string]string, error) {
	log := workflow.GetLogger(ctx)
	getGames := make(map[string]string)

	err := workflow.SetQueryHandler(ctx, "getGames", func(input []byte) (map[string]string, error) {
		return getGames, nil
	})
	if err != nil {
		log.Error("SetQueryHandler failed for getGames: " + err.Error())
		return &getGames, err
	}

	return &getGames, nil
}
