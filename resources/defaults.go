package resources

const GameSignalChannelName = "start-game-signal"
const AddPlayerSignalChannelName = "add-player-signal"
const AnswerSignalChannelName = "answer-signal"

type GameConfiguration struct {
	Category          string `json:"category"`
	NumberOfQuestions int    `json:"numberOfQuestions"`
	NumberOfPlayers   int    `json:"numberOfPlayers"`
	AnswerTimeLimit   int    `json:"answerTimeLimit"`
	StartTimeLimit    int    `json:"startTimeLimit"`
	ResultTimeLimit   int    `json:"resultTimeLimit"`
}

type GameConfigurationOption func(*GameConfiguration)

func NewGameConfiguration(opts []GameConfigurationOption) *GameConfiguration {
	g := &GameConfiguration{
		NumberOfQuestions: 5,
		NumberOfPlayers:   2,
		AnswerTimeLimit:   300,
		StartTimeLimit:    300,
		ResultTimeLimit:   10,
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

func NewGameConfigurationFromWorkflowInput(input GameWorkflowInput) *GameConfiguration {

	opts := []GameConfigurationOption{}
	if input.Category != "" {
		opts = append(opts, WithCategory(input.Category))
	}
	if input.AnswerTimeLimit > 0 {
		opts = append(opts, WithAnswerTimeLimit(input.AnswerTimeLimit))
	}
	if input.NumberOfPlayers > 0 {
		opts = append(opts, WithNumberOfPlayers(input.NumberOfPlayers))
	}
	if input.NumberOfQuestions > 0 {
		opts = append(opts, WithNUmberOfQuestions(input.NumberOfQuestions))
	}
	if input.ResultTimeLimit > 0 {
		opts = append(opts, WithResultTimeout(input.ResultTimeLimit))
	}
	if input.StartTimeLimit > 0 {
		opts = append(opts, WithStartTimeout(input.StartTimeLimit))
	}
	return NewGameConfiguration(opts)
}

func WithAnswerTimeLimit(n int) GameConfigurationOption {
	return func(cfg *GameConfiguration) {
		cfg.AnswerTimeLimit = n
	}
}

func WithNumberOfPlayers(n int) GameConfigurationOption {
	return func(cfg *GameConfiguration) {
		cfg.NumberOfPlayers = n
	}
}

func WithNUmberOfQuestions(n int) GameConfigurationOption {
	return func(cfg *GameConfiguration) {
		cfg.NumberOfQuestions = n
	}
}

func WithResultTimeout(n int) GameConfigurationOption {
	return func(cfg *GameConfiguration) {
		cfg.ResultTimeLimit = n
	}
}

func WithStartTimeout(n int) GameConfigurationOption {
	return func(cfg *GameConfiguration) {
		cfg.StartTimeLimit = n
	}
}

func WithCategory(s string) GameConfigurationOption {
	return func(cfg *GameConfiguration) {
		cfg.Category = s
	}
}
