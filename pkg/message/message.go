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
	log.Logger().Infof("Parsing messages")
	for i, m := range messageSplit {
		err = messages[i].ParseMessage(m, teamNameReg)
		if err != nil {
			return nil, errors.Errorf("failed to parse message %d", i+1)
		}
	}
	return messages, nil
}

func (m *Message) ParseMessage(messageMD string, teamNameReg *regexp.Regexp) error {
	// Find the team name for this message
	names := teamNameReg.FindAllStringSubmatch(messageMD, -1)
	if names == nil {
		return errors.New("no teams found in message markdown")
	}
	if len(names) > 1 {
		return errors.Errorf("found %d teams in message, should only be 1", len(names))
	}

	// The actual team name is always the sub match, so it's the second element
	teamName := strings.TrimSpace(names[0][1])
	log.Logger().Infof("Found team \"%s\" in message\n", teamName)
	m.TeamName = teamName

	// To find the content we can just remove the teamName heading
	content := teamNameReg.ReplaceAllString(messageMD, "")
	m.Content = strings.TrimSpace(content)
	if len(m.Content) < 1 {
		return errors.New("no content found for message")
	}
	log.Logger().Info("Found content for message\n")
	return nil
}
