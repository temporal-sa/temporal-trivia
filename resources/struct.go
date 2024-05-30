package resources

type GameWorkflowConfig struct {
	ChatGptKey string `json:"chatGptKey"`
}

type GameWorkflowInput struct {
	Category          string `json:"category"`
	NumberOfQuestions int    `json:"numberOfQuestions"`
	NumberOfPlayers   int    `json:"numberOfPlayers"`
	AnswerTimeLimit   int    `json:"answerTimeLimit"`
	StartTimeLimit    int    `json:"startTimeLimit"`
	ResultTimeLimit   int    `json:"resultTimeLimit"`
}

type AddPlayerWorkflowInput struct {
	GameWorkflowId  string `json:"gameWorkflowId"`
	Player          string `json:"player"`
	NumberOfPlayers int    `json:"numberOfPlayers"`
}

type TriviaQuestionsActivityInput struct {
	Key               string `json:"key"`
	Category          string `json:"category"`
	NumberOfQuestions int    `json:"question"`
}

type QueryPlayerActivityInput struct {
	WorkflowId      string `json:"workflowId"`
	Player          string `json:"player"`
	NumberOfPlayers int    `json:"numberOfPlayers"`
	QueryType       string `json:"queryType"`
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

type ModerationInput struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type AddPlayerActivityInput struct {
	WorkflowId string `json:"workflowId"`
	Player     string `json:"player"`
}

type KapaResponse struct {
	Answer           string `json:"answer"`
	ThreadID         string `json:"thread_id"`
	QuestionAnswerID string `json:"question_answer_id"`
}
