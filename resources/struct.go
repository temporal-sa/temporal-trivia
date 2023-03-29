package resources

type WorkflowInput struct {
	Key               string `json:"key"`
	Category          string `json:"category"`
	NumberOfQuestions int    `json:"numberOfQuestions"`
	NumberOfPlayers   int    `json:"numberOfPlayers"`
	NumberOfAnswers   int    `json:"numberOfAnswer"`
	QuestionTimeLimit int    `json:"questionTimeLimit"`
}

type ActivityInput struct {
	Key               string `json:"key"`
	Category          string `json:"category"`
	NumberOfQuestions int    `json:"question"`
}

type Signal struct {
	Action string `json:"action"`
	Player string `json:"player"`
	Answer string `json:"answer"`
}

type Result struct {
	Question          string                `json:"question"`
	Answer            string                `json:"answer"`
	Submissions       map[string]Submission `json:"submissions"`
	MultipleChoiceMap map[string]string     `json:"multipleChoiceAnswers"`
	Winner            string                `json:"winner"`
}

type Submission struct {
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"isCorrect"`
}

type GameProgress struct {
	NumberOfQuestions int `json:"numberOfQuestions"`
	CurrentQuestion   int `json:"currentQuestion"`
}

type ScoreBoard struct {
	Player string `json:"value"`
	Score  int    `json:"key"`
}
