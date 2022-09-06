package message

import (
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

const (
	messageHeaderRegex  = "## Message\\s*\\n"
	teamNameHeaderRegex = "### Team: (.*)\\n"
)

type Message struct {
	TeamName string
	Content  string
}

func ParseMessagesFromMarkdown(markdown string) ([]Message, error) {
	if markdown == "" {
		return nil, errors.New("no text found in markdown")
	}

	// Identify the messages using message header
	messageReg, err := regexp.Compile(messageHeaderRegex)
	if err != nil {
		return nil, err
	}
	messageSplit := messageReg.Split(markdown, -1)
	if len(messageSplit) < 2 {
		return nil, errors.Errorf("no messages found in markdown")
	}

	// Trim split to remove any text before the first message
	messageSplit = messageSplit[1:]
	log.Logger().Infof("Found %d message(s) in markdown\n", len(messageSplit))

	teamNameReg, _ := regexp.Compile(teamNameHeaderRegex)
	if err != nil {
		return nil, err
	}

	messages := make([]Message, len(messageSplit))
	for i, m := range messageSplit {
		// Find the team name for this message
		names := teamNameReg.FindAllStringSubmatch(m, -1)
		if names == nil {
			return nil, errors.Errorf("no teams found in message %d", i+1)
		}
		if len(names) > 1 {
			return nil, errors.Errorf("found %d teams in message %d, should only be 1", len(names), i+1)
		}

		// The actual team name is always the sub match, so it's the second element
		teamName := strings.TrimSpace(names[0][1])
		log.Logger().Infof("Found team \"%s\" in message %d\n", teamName, i+1)
		messages[i].TeamName = teamName

		// To find the content we can just remove the teamName heading
		content := teamNameReg.ReplaceAllString(m, "")
		messages[i].Content = strings.TrimSpace(content)
		if len(messages[i].Content) < 1 {
			return nil, errors.Errorf("no content found for message %d", i+1)
		}
		log.Logger().Infof("Found content for message %d\n", i+1)
	}
	return messages, nil
}
