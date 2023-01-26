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
	"github.com/spring-financial-group/peacock/pkg/git/comment"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"regexp"
	"strings"
	"text/template"
)

const (
	teamNameHeaderRegex = "### Notify(.*)\\n"
	commaSeparated      = ","

	breakdownTemplate = `Successfully validated {{ len .notes }} release note{{ addPlural (len .notes) }}.
{{ range $idx, $val := .notes }}
***
Release Note {{ inc $idx }} will be sent to: {{ commaSep $val.TeamNames }}
<details>
<summary>Release Note Breakdown</summary>

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

func (uc *UseCase) GetReleaseNotesFromMDAndTeams(markdown string, teamsInFeathers models.Teams) ([]models.ReleaseNote, error) {
	teamNameReg, err := regexp.Compile(teamNameHeaderRegex)
	if err != nil {
		return nil, err
	}

	log.Debug("Parsing release notes from markdown")
	teamNames := uc.parseTeamNames(teamNameReg, markdown)
	if len(teamNames) < 1 {
		return nil, nil
	}
	log.Debugf("%d notes found in markdown", len(teamNames))

	// Get the contents for each message & trim to remove any text before the first message
	contents := teamNameReg.Split(markdown, -1)
	contents = contents[1:]

	notes := make([]models.ReleaseNote, len(contents))
	for i, m := range contents {
		teamsNamesInNote := teamNames[i]
		teamsInNote, err := uc.getAndValidateTeamsByNames(teamsNamesInNote, teamsInFeathers)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get teams by name")
		}

		notes[i].Content = strings.TrimSpace(m)
		notes[i].Teams = teamsInNote
	}
	return notes, nil
}

func (uc *UseCase) getAndValidateTeamsByNames(teamNames []string, teamsInFeathers models.Teams) (models.Teams, error) {
	if err := teamsInFeathers.Contains(teamNames...); err != nil {
		return nil, err
	}
	wantedTeams := teamsInFeathers.GetTeamsByNames(teamNames...)
	for _, team := range wantedTeams {
		if !uc.MsgClientsHandler.IsInitialised(team.ContactType) {
			return nil, errors.New(fmt.Sprintf("communication method %s has not been configured", team.ContactType))
		}
	}
	return wantedTeams, nil
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

func (uc *UseCase) GenerateHash(notes []models.ReleaseNote) (string, error) {
	data, err := json.Marshal(notes)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal notes")
	}
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (uc *UseCase) GenerateBreakdown(notes []models.ReleaseNote, hash string, totalTeams int) (string, error) {
	tmplFuncs := template.FuncMap{
		"inc":      func(i int) int { return i + 1 },
		"commaSep": func(i []string) string { return utils.CommaSeparated(i) },
		"addPlural": func(i int) string {
			var plural string
			if i > 1 {
				plural = "s"
			}
			return plural
		},
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

	breakdown := strings.TrimSpace(buf.String())
	breakdown = comment.AddMetadataToComment(breakdown, hash, comment.BreakdownCommentType)
	return breakdown, nil
}

func (uc *UseCase) SendReleaseNotes(subject string, notes []models.ReleaseNote) error {
	return uc.MsgClientsHandler.SendReleaseNotes(subject, notes)
}
