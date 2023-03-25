package resources

import "time"

type WorkflowInput struct {
	Key               string        `json:"key"`
	Category          string        `json:"category"`
	NumberOfQuestions int           `json:"numberOfQuestions"`
	NumberOfPlayers   int           `json:"numberOfPlayers"`
	QuestionTimeLimit time.Duration `json:"questionTimeLimit"`
}

type ActivityInput struct {
	Key      string `json:"key"`
	Question string `json:"question"`
}

type Signal struct {
	Action string `json:"action"`
	User   string `json:"user"`
	Answer string `json:"answer"`
}

type Result struct {
	Question       string   `json:"question"`
	AnswerDetails  string   `json:"answer"`
	CorrectAnswers []string `json:"correctAnswers"`
	WrongAnswers   []string `json:"wrongAnswers"`
	Winner         string   `json:"winner"`
}
