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
	PlayerId  int    `json:"playerId"`
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"isCorrect"`
	IsFirst   bool   `json:"isFirst"`
}

type GameProgress struct {
	NumberOfQuestions int    `json:"numberOfQuestions"`
	CurrentQuestion   int    `json:"currentQuestion"`
	Stage             string `json:"stage"`
}

type ScoreBoard struct {
	Player string `json:"value"`
	Score  int    `json:"key"`
}

type Player struct {
	Id    int `json:"id"`
	Score int `json:"score"`
}
