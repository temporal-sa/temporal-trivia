package triviagame

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ktenzer/temporal-trivia/resources"

	"go.temporal.io/sdk/activity"
)

func ModerationActivity(ctx context.Context, input resources.ModerationInput) (bool, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("ModerationActivity start")

	// Username Moderation
	var fullUrl string
	var flagged bool
	fullUrl = input.Url + input.Name

	logger.Info("FULL URL: " + fullUrl)
	resp, err := http.Get(fullUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Read the response body using io
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer resp.Body.Close()

	flagged, error := strconv.ParseBool(string(body))
	if error != nil {
		log.Fatal(error)
	}

	logger.Info("ModerationActivity")

	return flagged, err
}
