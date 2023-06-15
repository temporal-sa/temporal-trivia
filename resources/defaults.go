package resources

func SetDefaults(workflowInput GameWorkflowInput) GameWorkflowInput {

	if workflowInput.AnswerTimeLimit == 0 {
		workflowInput.AnswerTimeLimit = 60
	}

	if workflowInput.Category == "" {
		workflowInput.Category = "General"
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
