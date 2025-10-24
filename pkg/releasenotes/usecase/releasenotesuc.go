package releasenotesuc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
Release Note {{ inc $idx }} will be sent to: {{ getTeamNames $val.Teams }}
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

func (uc *UseCase) GetReleaseNotesFromMarkdownAndTeamsInFeathers(markdown string, teamsInFeathers models.Teams) ([]models.ReleaseNote, error) {
	_, releaseNotes, err := uc.ParseReleaseNoteFromMarkdown(markdown, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get release notes from markdown")
	}
	if err = uc.PopulateTeamsInReleaseNotes(releaseNotes, teamsInFeathers); err != nil {
		return nil, errors.Wrap(err, "failed to populate teams in release notes")
	}
	releaseNotes = uc.MergeReleaseNotes(releaseNotes)
	return releaseNotes, nil
}

func (uc *UseCase) PopulateTeamsInReleaseNotes(releaseNotes []models.ReleaseNote, teamsInFeathers models.Teams) error {
	for i, note := range releaseNotes {
		teamsInNote, err := uc.getAndValidateTeamsByNames(note.Teams.GetAllTeamNames(), teamsInFeathers)
		if err != nil {
			return errors.Wrap(err, "failed to get teams by name")
		}
		releaseNotes[i].Teams = teamsInNote
	}
	return nil
}

func (uc *UseCase) ParseReleaseNoteFromMarkdown(markdown string, sanitise bool) (preamble string, notes []models.ReleaseNote, err error) {
	teamNameReg, err := regexp.Compile(teamNameHeaderRegex)
	if err != nil {
		return "", nil, err
	}

	log.Debug().Msg("Parsing release notes from markdown")
	teamNames := uc.parseTeamNames(teamNameReg, markdown)
	if len(teamNames) < 1 {
		return markdown, nil, nil
	}
	log.Debug().Msgf("%d notes found in markdown", len(teamNames))

	// Get the contents for each message & trim to remove any text before the first message
	markdownSplit := teamNameReg.Split(markdown, -1)
	preamble = markdownSplit[0]
	markdownSplit = markdownSplit[1:]

	notes = make([]models.ReleaseNote, len(markdownSplit))
	for i, m := range markdownSplit {
		teamsNamesInNote := teamNames[i]
		if sanitise {
			m = uc.removeBotGeneratedText(m)
		}
		notes[i].Content = strings.TrimSpace(m)
		teamsInNote := make([]models.Team, 0, len(teamsNamesInNote))
		for _, teamName := range teamsNamesInNote {
			teamsInNote = append(teamsInNote, models.Team{
				Name: teamName,
			})
		}
		notes[i].Teams = teamsInNote
	}
	return preamble, notes, nil
}

func (uc *UseCase) GetMarkdownFromReleaseNotes(notes []models.ReleaseNote) string {
	var markdown string
	for _, note := range notes {
		markdown += fmt.Sprintf("### Notify %s\n%s\n\n", utils.CommaSeparated(note.Teams.GetAllTeamNames()), note.Content)
	}
	return strings.TrimSpace(markdown)
}

func (uc *UseCase) MergeReleaseNotes(notes []models.ReleaseNote) []models.ReleaseNote {
	if len(notes) < 1 {
		return nil
	}

	teamsMap := make(map[string]models.ReleaseNote)
	order := make([]string, 0, len(notes))

	for _, note := range notes {
		teamNames := utils.CommaSeparated(note.Teams.GetAllTeamNames())
		if existingNote, ok := teamsMap[teamNames]; ok {
			existingNote.AppendContent(note.Content)
			teamsMap[teamNames] = existingNote
		} else {
			teamsMap[teamNames] = note
			order = append(order, teamNames)
		}
	}

	merged := make([]models.ReleaseNote, 0, len(order))
	for _, teamNames := range order {
		merged = append(merged, teamsMap[teamNames])
	}
	return merged
}

func (uc *UseCase) AppendReleaseNotesToExistingMarkdown(existingMarkdown string, releaseNotesToAppend []models.ReleaseNote) (string, error) {
	// Parse the existing markdown to get the release notes
	preamble, existingReleaseNotes, err := uc.ParseReleaseNoteFromMarkdown(existingMarkdown, false)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse existing markdown")
	}
	preamble = addSuffixIfNotExists(preamble, "\n\n")
	mergedReleaseNotes := uc.mergeOrAppendReleaseNotes(existingReleaseNotes, releaseNotesToAppend)
	return preamble + uc.GetMarkdownFromReleaseNotes(mergedReleaseNotes), nil
}

// mergeOrAppendReleaseNotes merges or appends release notes based on the teams maintaining the original order
func (uc *UseCase) mergeOrAppendReleaseNotes(existing, new []models.ReleaseNote) []models.ReleaseNote {
	// Create a map to store the merged release notes and an array to maintain the order
	merged := make(map[string]models.ReleaseNote)
	order := make([]string, 0, len(existing))

	// Iterate over the existing release notes and add them to the map by the teams in the note
	for _, note := range existing {
		teams := utils.CommaSeparated(note.Teams.GetAllTeamNames())
		merged[teams] = note
		order = append(order, teams)
	}

	// Iterate over the new release notes and merge or append them
	for _, note := range new {
		teams := utils.CommaSeparated(note.Teams.GetAllTeamNames())
		if existingNote, ok := merged[teams]; ok {
			existingNote.AppendContent(note.Content)
			merged[teams] = existingNote
		} else {
			merged[teams] = note
			order = append(order, teams)
		}
	}

	result := make([]models.ReleaseNote, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	return result
}

var (
	// This regex is used to find all the bot generated text in the markdown
	// Bot generated text is of the form `[//]: # (some-bot-tag)`
	botGeneratedTextRegex = regexp.MustCompile(`\n?\[//\]: # \((.*)\)`)
)

func (uc *UseCase) removeBotGeneratedText(text string) string {
	// We want to remove all of these from the text and
	return botGeneratedTextRegex.ReplaceAllString(text, "")
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
		"inc":          func(i int) int { return i + 1 },
		"getTeamNames": func(ts models.Teams) string { return utils.CommaSeparated(ts.GetAllTeamNames()) },
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

func addSuffixIfNotExists(text string, suffix string) string {
	if text == "" || strings.HasSuffix(text, suffix) {
		return text
	}
	return text + suffix
}
