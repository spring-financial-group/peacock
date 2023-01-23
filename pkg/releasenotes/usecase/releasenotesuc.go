package releasenotesuc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"regexp"
	"strings"
	"text/template"
)

const (
	teamNameHeaderRegex = "### Notify(.*)\\n"
	commaSeparated      = ","

	breakdownTemplate = `Successfully validated {{ len .notes }} release note(s).
{{ range $idx, $val := .notes }}
***
ReleaseNote {{ inc $idx }} will be sent to: {{ commaSep $val.TeamNames }}
<details>
<summary>ReleaseNote Breakdown</summary>

{{ $val.Content }}

</details>

{{ end -}}`
)

type UseCase struct {
	MsgClientsHandler domain.MessageHandler
}

func NewUseCase(msgClientsHandler domain.MessageHandler) *UseCase {
	return &UseCase{msgClientsHandler}
}

func (uc *UseCase) ParseNotesFromMarkdown(markdown string) ([]models.ReleaseNote, error) {
	teamNameReg, err := regexp.Compile(teamNameHeaderRegex)
	if err != nil {
		return nil, err
	}

	log.Debug("Parsing release notes from markdown")
	teamsInMessages := uc.parseTeamNames(teamNameReg, markdown)
	if len(teamsInMessages) < 1 {
		return nil, nil
	}
	log.Debugf("%d notes found in markdown", len(teamsInMessages))

	// Get the contents for each message & trim to remove any text before the first message
	contents := teamNameReg.Split(markdown, -1)
	contents = contents[1:]

	messages := make([]models.ReleaseNote, len(contents))
	for i, m := range contents {
		messages[i].Content = strings.TrimSpace(m)
		messages[i].TeamNames = teamsInMessages[i]
		log.Debugf("Found %d team(s) to notify in notes %d\n", len(messages[i].TeamNames), i+1)
	}
	return messages, nil
}

func (uc *UseCase) parseTeamNames(teamNameReg *regexp.Regexp, markdown string) [][]string {
	// Find all the notify headers
	notifyHeaders := teamNameReg.FindAllStringSubmatch(markdown, -1)
	if len(notifyHeaders) < 1 {
		return nil
	}

	teamsInNotes := make([][]string, len(notifyHeaders))
	for i, header := range notifyHeaders {
		// The actual team name is always the sub match, so it's the second element
		teamNames := strings.Split(header[1], commaSeparated)
		teamNames = utils.TrimSpaceInSlice(teamNames)
		teamsInNotes[i] = teamNames
	}
	return teamsInNotes
}

func (uc *UseCase) GenerateHash(messages []models.ReleaseNote) (string, error) {
	data, err := json.Marshal(messages)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal notes")
	}
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (uc *UseCase) GenerateBreakdown(notes []models.ReleaseNote, totalTeams int) (string, error) {
	tmplFuncs := template.FuncMap{
		"inc":      func(i int) int { return i + 1 },
		"commaSep": func(i []string) string { return utils.CommaSeparated(i) },
	}

	tpl, err := template.New("breakdown").Funcs(tmplFuncs).Parse(breakdownTemplate)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]any{
		"totalTeams": totalTeams,
		"notes":      notes,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func (uc *UseCase) SendReleaseNotes(feathers *feathers.Feathers, notes []models.ReleaseNote) error {
	return uc.MsgClientsHandler.SendMessages(feathers, notes)
}

func (uc *UseCase) ValidateReleaseNotesWithFeathers(feathers *feathers.Feathers, notes []models.ReleaseNote) error {
	// Check that the relevant communication methods have been configured for the feathers
	types := feathers.GetAllContactTypes()
	for _, t := range types {
		if !uc.MsgClientsHandler.IsInitialised(t) {
			return errors.New(fmt.Sprintf("communication method %s has not been configured", t))
		}
	}

	// Check that the teams in the releaseNotes exist in the feathers
	for _, m := range notes {
		if err := feathers.ExistsInFeathers(m.TeamNames...); err != nil {
			return err
		}
	}
	return nil
}
