package triviagame

import (
	"errors"
	"strings"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"
)

const (
	Starting  = "starting"
	Running   = "running"
	Completed = "completed"
)

// Setup update handler to perform player validation and add player to game state machine
func updatePlayer(ctx workflow.Context, players map[string]resources.Player) error {

	err := workflow.SetUpdateHandlerWithOptions(
		ctx,
		"AddPlayer",
		func(ctx workflow.Context, player string) error {
			//if _, ok := players[player]; !ok {
			score := resources.Player{
				Score: 0,
			}

			players[player] = score
			//}
			return nil
		},
		workflow.UpdateHandlerOptions{Validator: func(player string) error {
			log := workflow.GetLogger(ctx)

			if _, ok := players[player]; ok {
				log.Debug("Rejecting player, already exists", "Player", player)
				return errors.New("Player " + player + " already exists")
			}

			log.Debug("Adding player update accepted", "Player", player)

			return nil
		},
		},
	)

	if err != nil {
		return err
	}

	return nil
}

// Setup update handler to perform game validation and update game state
func updateGame(ctx workflow.Context, games map[string]string) error {

	err := workflow.SetUpdateHandlerWithOptions(
		ctx,
		"UpdateGame",
		func(ctx workflow.Context, game map[string]string) error {
			containsStarting := strings.Contains(game["status"], "starting")
			containsRunning := strings.Contains(game["status"], "running")
			containsCompleted := strings.Contains(game["status"], "completed")

			if containsCompleted {
				delete(games, game["gameId"])
			}

			if containsRunning || containsStarting {
				games[game["gameId"]] = game["status"]
			}

			return nil
		},
		workflow.UpdateHandlerOptions{Validator: func(gameMap map[string]string) error {
			log := workflow.GetLogger(ctx)

			if err := ValidateStatus(gameMap["status"]); err != nil {
				log.Debug("Rejecting game update, game state is not starting|running|completed", "Status", gameMap["status"])
				return errors.New("Game status is not starting|running|completed, it is " + gameMap["status"] + " ")
			}

			log.Debug("Game update accepted", "Game", gameMap["gameId"], gameMap["status"])

			return nil
		},
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func ValidateStatus(status string) error {
	switch status {
	case Starting, Running, Completed:
		return nil
	default:
		return errors.New("invalid status: " + status)
	}
}
