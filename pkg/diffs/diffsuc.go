package diffs

import (
	"context"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"path"
	"sort"
	"strings"
)

const (
	maxMatchedFiles   = 10
	maxFileBlockChars = 40000
	maxSectionChars   = 48000
)

type UseCase struct {
	scm domain.SCM
}

func NewUseCase(scm domain.SCM) *UseCase {
	return &UseCase{scm: scm}
}

// GenerateDiffSection renders the markdown file changes section for a PR, or an empty string
// when nothing matched or nothing effectively changed. Both file versions are fetched via the
// API at a single live snapshot rather than trusting the event payload, so webhook and CLI
// modes behave identically.
func (uc *UseCase) GenerateDiffSection(ctx context.Context, owner, repoName string, prNumber int, configs models.DiffConfigs) (string, error) {
	if len(configs) == 0 {
		return "", nil
	}
	baseBranch, err := uc.scm.GetPRBaseBranchFromPRNumber(ctx, owner, repoName, prNumber)
	if err != nil {
		return "", errors.Wrap(err, "failed to get PR base branch")
	}
	_, headSHA, err := uc.scm.GetPRBranchSHAFromPRNumber(ctx, owner, repoName, prNumber)
	if err != nil {
		return "", errors.Wrap(err, "failed to get PR head SHA")
	}
	files, err := uc.scm.GetFilesChangedFromPR(ctx, owner, repoName, prNumber)
	if err != nil {
		return "", errors.Wrap(err, "failed to get files changed from PR")
	}
	sort.Slice(files, func(i, j int) bool { return files[i].GetFilename() < files[j].GetFilename() })

	var blocks []string
	notShown := 0
	for _, file := range files {
		config := matchDiffConfig(file.GetFilename(), configs)
		if config == nil {
			continue
		}
		if len(blocks) >= maxMatchedFiles {
			notShown++
			continue
		}
		block, rendered, err := uc.fileBlock(ctx, owner, repoName, prNumber, derefString(baseBranch), derefString(headSHA), file, config)
		if err != nil {
			return "", err
		}
		if rendered {
			blocks = append(blocks, block)
		}
	}
	if len(blocks) == 0 {
		return "", nil
	}

	// keep the section comfortably under GitHub's 65536-char comment limit
	header := fmt.Sprintf("📊 File Changes — %d file%s", len(blocks), plural(len(blocks)))
	shown := len(blocks)
	for shown > 1 && len(header)+2+len(strings.Join(blocks[:shown], "\n\n")) > maxSectionChars {
		shown--
	}
	out := header + "\n\n" + strings.Join(blocks[:shown], "\n\n")
	if shown < len(blocks) {
		out += fmt.Sprintf("\n\n> … %d more changed file%s truncated — view the [Files tab](%s).",
			len(blocks)-shown, plural(len(blocks)-shown), filesTabURL(owner, repoName, prNumber))
	}
	if notShown > 0 {
		out += fmt.Sprintf("\n\n> … %d more matched file%s not shown.", notShown, plural(notShown))
	}
	return out, nil
}

// fileBlock renders one matched file's block, returning rendered=false for effectively unchanged files
func (uc *UseCase) fileBlock(ctx context.Context, owner, repoName string, prNumber int, baseBranch, headSHA string, file *github.CommitFile, config *models.DiffConfig) (string, bool, error) {
	handler, err := getHandler(config.Format)
	if err != nil {
		return "", false, err
	}

	basePath := file.GetFilename()
	if file.GetStatus() == "renamed" {
		basePath = file.GetPreviousFilename()
	}
	baseBytes, err := uc.fetchSide(ctx, owner, repoName, baseBranch, basePath, file.GetStatus() != "added")
	if err != nil {
		return "", false, err
	}
	headBytes, err := uc.fetchSide(ctx, owner, repoName, headSHA, file.GetFilename(), !isRemoved(file.GetStatus()))
	if err != nil {
		return "", false, err
	}

	sections, err := handler.Diff(file.GetFilename(), baseBytes, headBytes)
	if err != nil {
		// one unparseable file shouldn't take out the whole section
		log.Warnf("failed to diff %s: %s", file.GetFilename(), err.Error())
		return fmt.Sprintf("**`%s`**\n\n> ⚠ failed to diff this file: %s",
			escapeCell(path.Base(file.GetFilename())), escapeCell(err.Error())), true, nil
	}
	if len(sections) == 0 {
		return "", false, nil
	}

	filesURL := fileDiffURL(owner, repoName, prNumber, file.GetSHA())
	var rendered []string
	total := 0
	for i, s := range sections {
		r := renderSection(s, filesURL)
		if total+len(r) > maxFileBlockChars && i > 0 {
			rendered = append(rendered, fmt.Sprintf("> … %d more section%s truncated — view the [Files tab](%s).",
				len(sections)-i, plural(len(sections)-i), filesURL))
			break
		}
		rendered = append(rendered, r)
		total += len(r)
	}

	// the file header links straight to this file's diff in the Files tab when GitHub gives us the blob SHA
	name := escapeCell(path.Base(file.GetFilename()))
	header := fmt.Sprintf("**`%s`**", name)
	if file.GetSHA() != "" {
		header = fmt.Sprintf("**[`%s`](%s)**", name, filesURL)
	}
	return header + "\n\n" + strings.Join(rendered, "\n\n"), true, nil
}

// fetchSide fetches one side of a file, mapping a missing file to nil (new or deleted)
func (uc *UseCase) fetchSide(ctx context.Context, owner, repoName, ref, filePath string, want bool) ([]byte, error) {
	if !want || filePath == "" {
		return nil, nil
	}
	b, err := uc.scm.GetFileFromBranch(ctx, owner, repoName, ref, filePath)
	if err != nil {
		var notFound *domain.ErrFileNotFound
		if errors.As(err, &notFound) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to fetch %s at %s", filePath, ref)
	}
	return b, nil
}

// matchDiffConfig returns the first config whose glob matches the file; first config wins on overlap
func matchDiffConfig(filename string, configs models.DiffConfigs) *models.DiffConfig {
	for i := range configs {
		for _, pattern := range configs[i].Paths {
			if matched, err := doublestar.Match(pattern, filename); err == nil && matched {
				return &configs[i]
			}
		}
	}
	return nil
}

func renderSection(s models.DiffSection, filesURL string) string {
	if strings.TrimSpace(s.Body) == "" {
		return s.Summary
	}
	open := ""
	if s.Open {
		open = " open"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<details%s>\n<summary>%s</summary>\n\n%s\n\n</details>", open, s.Summary, strings.TrimRight(s.Body, "\n"))
	if s.TruncatedRows > 0 {
		fmt.Fprintf(&b, "\n> … %d more changed field%s truncated — view the full file diff in the [Files tab](%s).",
			s.TruncatedRows, plural(s.TruncatedRows), filesURL)
	}
	return b.String()
}

func fileDiffURL(owner, repoName string, prNumber int, sha string) string {
	url := fmt.Sprintf("https://github.com/%s/%s/pull/%d/files", owner, repoName, prNumber)
	if sha != "" {
		url += "#diff-" + sha
	}
	return url
}

func filesTabURL(owner, repoName string, prNumber int) string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d/files", owner, repoName, prNumber)
}

func isRemoved(status string) bool {
	return status == "removed" || status == "deleted"
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
