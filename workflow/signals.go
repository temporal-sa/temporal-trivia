package triviagame

import (
	"strconv"

	"go.temporal.io/sdk/workflow"
)

type GameSignal struct {
	Action string `json:"action"`
	Player string `json:"player"`
}

type AnswerSignal struct {
	Action   string `json:"action"`
	Player   string `json:"player"`
	Question int    `json:"question"`
	Answer   string `json:"answer"`
}

func (signal *GameSignal) gameSignal(ctx workflow.Context) workflow.Selector {
	log := workflow.GetLogger(ctx)

	var addPlayerSelector workflow.Selector
	addPlayerSignalChan := workflow.GetSignalChannel(ctx, "start-game-signal")
	addPlayerSelector = workflow.NewSelector(ctx)
	addPlayerSelector.AddReceive(addPlayerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &signal)
		log.Info("Recieved signal Action: " + signal.Action + " Player: " + signal.Player)
	})

	return addPlayerSelector
}

func (signal *AnswerSignal) answerSignal(ctx workflow.Context) workflow.Selector {
	log := workflow.GetLogger(ctx)

	answerSignalChan := workflow.GetSignalChannel(ctx, "answer-signal")
	answerSelector := workflow.NewSelector(ctx)
	answerSelector.AddReceive(answerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &signal)
		log.Info("Recieved signal Action: " + signal.Action + " Player: " + signal.Player + " Question: " + strconv.Itoa(signal.Question) + " Answer: " + signal.Answer)
	})

	return answerSelector
}
