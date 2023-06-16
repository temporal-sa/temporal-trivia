package triviagame

import "go.temporal.io/sdk/workflow"

type GameSignal struct {
	Action string `json:"action"`
	Player string `json:"player"`
	Answer string `json:"answer"`
}

func (g *GameSignal) gameSignal(ctx workflow.Context) workflow.Selector {
	log := workflow.GetLogger(ctx)

	var addPlayerSelector workflow.Selector
	addPlayerSignalChan := workflow.GetSignalChannel(ctx, "start-game-signal")
	addPlayerSelector = workflow.NewSelector(ctx)
	addPlayerSelector.AddReceive(addPlayerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &g)
		log.Info("Recieved signal Action: " + g.Action + " Player: " + g.Player + " Answer: " + g.Answer)
	})

	return addPlayerSelector
}

func (g *GameSignal) answerSignal(ctx workflow.Context) workflow.Selector {
	log := workflow.GetLogger(ctx)

	answerSignalChan := workflow.GetSignalChannel(ctx, "answer-signal")
	answerSelector := workflow.NewSelector(ctx)
	answerSelector.AddReceive(answerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &g)
		log.Info("Recieved signal Action: " + g.Action + " Player: " + g.Player + " Answer: " + g.Answer)
	})

	return answerSelector
}
