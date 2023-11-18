package triviagame

import (
	"errors"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"
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
