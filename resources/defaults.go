package resources

import (
	"math/rand"
	"time"
)

func SetDefaults(workflowInput GameWorkflowInput) GameWorkflowInput {

	if workflowInput.AnswerTimeLimit == 0 {
		workflowInput.AnswerTimeLimit = 60
	}

	if workflowInput.Category == "" {
		workflowInput.Category = getRandomCategory()
	}

	if workflowInput.NumberOfPlayers == 0 {
		workflowInput.NumberOfPlayers = 2
	}

	if workflowInput.NumberOfQuestions == 0 {
		workflowInput.NumberOfQuestions = 5
	}

	if workflowInput.ResultTimeLimit == 0 {
		workflowInput.ResultTimeLimit = 10
	}

	if workflowInput.StartTimeLimit == 0 {
		workflowInput.StartTimeLimit = 300
	}

	return workflowInput
}

func getRandomCategory() string {
	rand.Seed(time.Now().UnixNano())

	keys := []string{"General", "Sports", "Science", "Travel", "Geography", "Capitols", "Authors", "Books", "Animals", "Plants", "Foods", "Cities"}
	randomIndex := rand.Intn(len(keys))

	return keys[randomIndex]
}
