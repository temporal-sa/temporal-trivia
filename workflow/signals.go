package triviagame

import (
	. "github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"
)

type GameSignal struct {
	Action string `json:"action"`
}

type AnswerSignal struct {
	Action   string `json:"action"`
	Player   string `json:"player"`
	Question int    `json:"question"`
	Answer   string `json:"answer"`
}

func (signal *GameSignal) gameSignal(ctx workflow.Context, startGameSelector workflow.Selector) {
	log := workflow.GetLogger(ctx)

	addPlayerSignalChan := workflow.GetSignalChannel(ctx, GameSignalChannelName)
	startGameSelector.AddReceive(addPlayerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &signal)
		log.Info("Recieved signal Action: " + signal.Action)
	})
}

func (signal *AnswerSignal) answerSignal(ctx workflow.Context, answerSelector workflow.Selector) {
	log := workflow.GetLogger(ctx)

	answerSignalChan := workflow.GetSignalChannel(ctx, AnswerSignalChannelName)
	answerSelector.AddReceive(answerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &signal)
		log.Info("Recieved signal Action: " + signal.Action + " Player: " + signal.Player + " Question: " + intToString(signal.Question) + " Answer: " + signal.Answer)
	})
}
