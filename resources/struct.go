package resources

type GameWorkflowInput struct {
	Category          string `json:"category"`
	NumberOfQuestions int    `json:"numberOfQuestions"`
	NumberOfPlayers   int    `json:"numberOfPlayers"`
	AnswerTimeLimit   int    `json:"answerTimeLimit"`
	StartTimeLimit    int    `json:"startTimeLimit"`
	ResultTimeLimit   int    `json:"resultTimeLimit"`
}

type AddPlayerWorkflowInput struct {
	GameWorkflowId string `json:"gameWorkflowId"`
	Player         string `json:"player"`
}

type TriviaQuestionsActivityInput struct {
	Key               string `json:"key"`
	Category          string `json:"category"`
	NumberOfQuestions int    `json:"question"`
}

type QueryPlayerActivityInput struct {
	WorkflowId string `json:"workflowId"`
	Player     string `json:"player"`
	QueryType  string `json:"queryType"`
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
	IsFirst   bool   `json:"isFirst"`
}

type ScoreBoard struct {
	Player string `json:"value"`
	Score  int    `json:"key"`
}

type Player struct {
	Score int `json:"score"`
}
