package message

import (
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"regexp"
	"strings"
)

const (
	teamNameHeaderRegex = "### Notify(.*)\\n"
	commaSeperated      = ","
)

type Message struct {
	TeamNames []string
	Content   string
}

func ParseMessagesFromMarkdown(markdown string) ([]Message, error) {
	teamNameReg, err := regexp.Compile(teamNameHeaderRegex)
	if err != nil {
		return nil, err
	}

	log.Logger().Infof("Parsing messages")
	teamsInMessages := ParseTeamNames(teamNameReg, markdown)
	if len(teamsInMessages) < 1 {
		log.Logger().Info("No Peacock messages found in markdown, exiting")
		return nil, nil
	}
	log.Logger().Infof("%d messages found in markdown", len(teamsInMessages))

	// Get the contents for each message & trim to remove any text before the first message
	contents := teamNameReg.Split(markdown, -1)
	contents = contents[1:]

	messages := make([]Message, len(contents))
	for i, m := range contents {
		messages[i].Content = strings.TrimSpace(m)
		messages[i].TeamNames = teamsInMessages[i]
		log.Logger().Infof("Found %d team(s) to notify in message %d\n", len(messages[i].TeamNames), i+1)
	}
	return messages, nil
}

func ParseTeamNames(teamNameReg *regexp.Regexp, markdown string) [][]string {
	// Find all the notify headers
	notifyHeaders := teamNameReg.FindAllStringSubmatch(markdown, -1)
	if len(notifyHeaders) < 1 {
		return nil
	}

	teamsInMessages := make([][]string, len(notifyHeaders))
	for i, header := range notifyHeaders {
		// The actual team name is always the sub match, so it's the second element
		teamNames := strings.Split(header[1], commaSeperated)
		teamNames = utils.TrimSpaceInSlice(teamNames)
		teamsInMessages[i] = teamNames
	}
	return teamsInMessages
}
