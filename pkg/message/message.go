package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"regexp"
	"strings"
	"text/template"
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

	log.Debug("Parsing messages")
	teamsInMessages := parseTeamNames(teamNameReg, markdown)
	if len(teamsInMessages) < 1 {
		return nil, nil
	}
	log.Debug("%d messages found in markdown", len(teamsInMessages))

	// Get the contents for each message & trim to remove any text before the first message
	contents := teamNameReg.Split(markdown, -1)
	contents = contents[1:]

	messages := make([]Message, len(contents))
	for i, m := range contents {
		messages[i].Content = strings.TrimSpace(m)
		messages[i].TeamNames = teamsInMessages[i]
		log.Debug("Found %d team(s) to notify in message %d\n", len(messages[i].TeamNames), i+1)
	}
	return messages, nil
}

func parseTeamNames(teamNameReg *regexp.Regexp, markdown string) [][]string {
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

func GenerateHash(messages []Message) (string, error) {
	data, err := json.Marshal(messages)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal messages")
	}
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func GenerateBreakdown(messages []Message, totalTeams int) (string, error) {
	breakdownTmpl := `[Peacock] Successfully validated {{ len .messages }} message(s).
{{ range $idx, $val := .messages }}
***
Message {{ inc $idx }} will be sent to: {{ commaSep $val.TeamNames }}
<details>
<summary>Message Breakdown</summary>

{{ $val.Content }}

</details>

{{ end -}}`

	tmplFuncs := template.FuncMap{
		"inc":      func(i int) int { return i + 1 },
		"commaSep": func(i []string) string { return utils.CommaSeperated(i) },
	}

	tpl, err := template.New("breakdown").Funcs(tmplFuncs).Parse(breakdownTmpl)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]any{
		"totalTeams": totalTeams,
		"messages":   messages,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
