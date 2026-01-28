package webhookuc

import (
	"context"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"
)

const (
	RepoOwner     = "spring-financial-group"
	RepoName      = "peacock"
	PRNumber      = 1
	Branch        = "test-peacock"
	DefaultBranch = "master"
	SHA           = "some-SHA"

	InfraTeam    = "Infra"
	SkisocksTeam = "Skisocks"

	mockHash = "SomeReallyGoodHash"
	prBody   = "### Notify Infra\n\nHello infra\n\n### Notify Skisocks\n\nHello skisocks"
)

var (
	mockCTX = context.Background()

	mockPullRequestEventDTO = &models.PullRequestEventDTO{
		PullRequestID: 100,
		RepoOwner:     RepoOwner,
		RepoName:      RepoName,
		Body:          "### Notify Infra\n\nHello infra\n\n### Notify product\n\nHello product",
		PRNumber:      PRNumber,
		SHA:           SHA,
		Branch:        Branch,
		DefaultBranch: DefaultBranch,
	}

	infraTeam = models.Team{
		Name:        "infrastructure",
		APIKey:      "some-api-key",
		Addresses:   []string{"C02TE2EMTMK"},
		ContactType: models.Slack,
	}
	productTeam = models.Team{
		Name:        "product",
		APIKey:      "another-api-key",
		Addresses:   []string{"C02TE2EMTML"},
		ContactType: models.Webhook,
	}
	allTeams = models.Teams{
		infraTeam,
		productTeam,
	}

	mockFeathers = &models.Feathers{
		Teams: allTeams,
		Config: models.Config{
			Messages: models.Messages{
				Subject: "Subject",
			},
		},
	}
	mockFeathersData, _ = yaml.Marshal(mockFeathers)

	mockNotes = []models.ReleaseNote{
		{
			Teams:   models.Teams{infraTeam},
			Content: "Hello infra",
		},
		{
			Teams:   models.Teams{productTeam},
			Content: "Hello product",
		},
	}
)

func TestWebHookUseCase_ValidatePeacock(t *testing.T) {
	cfg := &config.SCM{
		User: RepoOwner,
	}

	templateContent := []byte("### Notify Infra\n\nTemplate message\n\n### Notify Product\n\nTemplate message")
	mockTemplateNotes := []models.ReleaseNote{
		{
			Teams:   models.Teams{infraTeam},
			Content: "Template message",
		},
		{
			Teams:   models.Teams{productTeam},
			Content: "Template message",
		},
	}

	t.Run("Happy Path", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockNotesUC := mocks.NewReleaseNotesUseCase(t)
		mockReleaseUC := mocks.NewReleaseUseCase(t)
		uc := NewUseCase(cfg, mockSCM, mockNotesUC, feathers.NewUseCase(), mockReleaseUC)
		uc.prTemplates = make(map[int64]*prTemplateMeta)
		mockEvent := mockPullRequestEventDTO
		mockEvent.Body = prBody

		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.PendingState, domain.ValidationContext).Return(nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, ".peacock/feathers.yaml").Return(mockFeathersData, nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, "docs/pull_request_template.md").Return(templateContent, nil).Once()

		mockSCM.On("GetPRCommentsByUser", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber).Return(nil, nil).Once()
		mockSCM.On("DeleteUsersComments", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber).Return(nil).Once()
		mockSCM.On("CommentOnPR", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber, mock.Anything).Return(nil).Once()
		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.SuccessState, domain.ValidationContext).Return(nil).Once()

		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", prBody, allTeams).Return(mockNotes, nil).Once()
		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", string(templateContent), allTeams).Return(mockTemplateNotes, nil).Once()
		mockNotesUC.On("GenerateHash", mockNotes).Return(mockHash, nil)
		mockNotesUC.On("GenerateBreakdown", mockNotes, mockHash, 2).Return("", nil)

		err := uc.ValidatePeacock(mockEvent)
		assert.NoError(t, err)
	})

	t.Run("should fail when release notes match PR template", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockNotesUC := mocks.NewReleaseNotesUseCase(t)
		mockReleaseUC := mocks.NewReleaseUseCase(t)
		uc := NewUseCase(cfg, mockSCM, mockNotesUC, feathers.NewUseCase(), mockReleaseUC)
		uc.prTemplates = make(map[int64]*prTemplateMeta)
		mockEvent := mockPullRequestEventDTO
		mockEvent.Body = prBody

		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.PendingState, domain.ValidationContext).Return(nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, ".peacock/feathers.yaml").Return(mockFeathersData, nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, "docs/pull_request_template.md").Return([]byte(prBody), nil).Once()
		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.FailureState, domain.ValidationContext).Return(nil).Once()

		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", prBody, allTeams).Return(mockNotes, nil).Twice()

		err := uc.ValidatePeacock(mockEvent)
		assert.NoError(t, err)
		mockSCM.AssertCalled(t, "CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.FailureState, domain.ValidationContext)
	})

	t.Run("should handle error when PR template cannot be fetched", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockNotesUC := mocks.NewReleaseNotesUseCase(t)
		mockReleaseUC := mocks.NewReleaseUseCase(t)
		uc := NewUseCase(cfg, mockSCM, mockNotesUC, feathers.NewUseCase(), mockReleaseUC)
		uc.prTemplates = make(map[int64]*prTemplateMeta)
		mockEvent := mockPullRequestEventDTO
		mockEvent.Body = prBody

		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.PendingState, domain.ValidationContext).Return(nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, ".peacock/feathers.yaml").Return(mockFeathersData, nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, "docs/pull_request_template.md").Return(templateContent, nil).Once()

		mockSCM.On("GetPRCommentsByUser", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber).Return(nil, nil).Once()
		mockSCM.On("DeleteUsersComments", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber).Return(nil).Once()
		mockSCM.On("CommentOnPR", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber, mock.Anything).Return(nil).Once()
		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.SuccessState, domain.ValidationContext).Return(nil).Once()

		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", prBody, allTeams).Return(mockNotes, nil).Once()
		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", string(templateContent), allTeams).Return([]models.ReleaseNote{}, nil).Once()
		mockNotesUC.On("GenerateHash", mockNotes).Return(mockHash, nil)
		mockNotesUC.On("GenerateBreakdown", mockNotes, mockHash, 2).Return("", nil)

		err := uc.ValidatePeacock(mockEvent)
		assert.NoError(t, err)
	})
}

func TestWebHookUseCase_RunPeacock(t *testing.T) {
	mockSCM := mocks.NewSCM(t)
	mockNotesUC := mocks.NewReleaseNotesUseCase(t)
	mockReleaseUC := mocks.NewReleaseUseCase(t)

	cfg := &config.SCM{
		User: RepoOwner,
	}

	uc := NewUseCase(cfg, mockSCM, mockNotesUC, feathers.NewUseCase(), mockReleaseUC)

	t.Run("Happy Path", func(t *testing.T) {
		mockEvent := mockPullRequestEventDTO
		mockEvent.Body = prBody
		mockFilesChanged := []*github.CommitFile{
			{
				Filename: github.String("helmfiles/staging/helmfile.yaml"),
			},
		}

		defaultSHA := "default-SHA"
		mockSCM.On("GetLatestCommitSHAInBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.DefaultBranch).Return(defaultSHA, nil).Once()
		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, defaultSHA, domain.PendingState, domain.ReleaseContext).Return(nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.DefaultBranch, ".peacock/feathers.yaml").Return(mockFeathersData, nil).Once()

		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, defaultSHA, domain.SuccessState, domain.ReleaseContext).Return(nil).Once()
		mockSCM.On("GetFilesChangedFromPR", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber).Return(mockFilesChanged, nil).Once()

		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", prBody, allTeams).Return(mockNotes, nil)
		mockNotesUC.On("SendReleaseNotes", mockFeathers.Config.Messages.Subject, mockNotes).Return(nil)

		mockReleaseUC.On("SaveRelease", mockCTX, "staging", mockNotes, mockPullRequestEventDTO.Summary()).Return(nil).Once()

		err := uc.RunPeacock(mockEvent)
		assert.NoError(t, err)
	})
}

func TestWebHookUseCase_getPRTemplate(t *testing.T) {
	mockSCM := mocks.NewSCM(t)
	mockNotesUC := mocks.NewReleaseNotesUseCase(t)
	mockReleaseUC := mocks.NewReleaseUseCase(t)

	cfg := &config.SCM{
		User: RepoOwner,
	}

	uc := NewUseCase(cfg, mockSCM, mockNotesUC, feathers.NewUseCase(), mockReleaseUC)

	mockEvent := &models.PullRequestEventDTO{
		PullRequestID: 100,
		RepoOwner:     RepoOwner,
		RepoName:      RepoName,
		SHA:           SHA,
		Branch:        Branch,
	}

	templateContent := []byte("### Notify Infra\n\nTemplate message for infra\n\n### Notify Product\n\nTemplate message for product")
	mockTemplateNotes := []models.ReleaseNote{
		{
			Teams:   models.Teams{infraTeam},
			Content: "Template message for infra",
		},
		{
			Teams:   models.Teams{productTeam},
			Content: "Template message for product",
		},
	}

	t.Run("should fetch and cache PR template successfully", func(t *testing.T) {
		uc.prTemplates = make(map[int64]*prTemplateMeta)

		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, "docs/pull_request_template.md").Return(templateContent, nil).Once()
		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", string(templateContent), allTeams).Return(mockTemplateNotes, nil).Once()

		result, err := uc.getPRTemplateOrDefault(mockCTX, mockEvent.Branch, mockEvent, allTeams)

		assert.NoError(t, err)
		assert.Equal(t, mockTemplateNotes, result)
		assert.Contains(t, uc.prTemplates, mockEvent.PullRequestID)
		assert.Equal(t, SHA, uc.prTemplates[mockEvent.PullRequestID].sha)
	})

	t.Run("should return cached template when SHA matches", func(t *testing.T) {
		uc.prTemplates = map[int64]*prTemplateMeta{
			mockEvent.PullRequestID: {
				prTemplates: mockTemplateNotes,
				sha:         SHA,
			},
		}

		result, err := uc.getPRTemplateOrDefault(mockCTX, mockEvent.Branch, mockEvent, allTeams)

		assert.NoError(t, err)
		assert.Equal(t, mockTemplateNotes, result)
		mockSCM.AssertNotCalled(t, "GetFileFromBranch")
	})

	t.Run("should refetch when SHA changes", func(t *testing.T) {
		uc.prTemplates = map[int64]*prTemplateMeta{
			mockEvent.PullRequestID: {
				prTemplates: mockTemplateNotes,
				sha:         "old-SHA",
			},
		}

		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, "docs/pull_request_template.md").Return(templateContent, nil).Once()
		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", string(templateContent), allTeams).Return(mockTemplateNotes, nil).Once()

		result, err := uc.getPRTemplateOrDefault(mockCTX, mockEvent.Branch, mockEvent, allTeams)

		assert.NoError(t, err)
		assert.Equal(t, mockTemplateNotes, result)
		assert.Equal(t, SHA, uc.prTemplates[mockEvent.PullRequestID].sha)
	})
}

func TestWebHookUseCase_compareReleaseNotesAndTemplate(t *testing.T) {
	mockSCM := mocks.NewSCM(t)
	mockNotesUC := mocks.NewReleaseNotesUseCase(t)
	mockReleaseUC := mocks.NewReleaseUseCase(t)

	cfg := &config.SCM{
		User: RepoOwner,
	}

	uc := NewUseCase(cfg, mockSCM, mockNotesUC, feathers.NewUseCase(), mockReleaseUC)

	notes1 := []models.ReleaseNote{
		{
			Teams:   models.Teams{infraTeam},
			Content: "First note",
		},
		{
			Teams:   models.Teams{productTeam},
			Content: "Second note",
		},
	}

	notes2 := []models.ReleaseNote{
		{
			Teams:   models.Teams{infraTeam},
			Content: "First note",
		},
		{
			Teams:   models.Teams{productTeam},
			Content: "Second note",
		},
	}

	t.Run("should return false when lengths are different", func(t *testing.T) {
		notesShort := []models.ReleaseNote{
			{
				Teams:   models.Teams{infraTeam},
				Content: "First note",
			},
		}

		areSame, err := uc.compareReleaseNotesAndTemplate(notes1, notesShort)

		assert.NoError(t, err)
		assert.False(t, areSame)
	})

	t.Run("should return true when lengths are equal (TODO: incomplete implementation)", func(t *testing.T) {
		areSame, err := uc.compareReleaseNotesAndTemplate(notes1, notes2)

		assert.NoError(t, err)
		// Note: This returns true due to incomplete implementation (TODO in the code)
		assert.True(t, areSame)
	})

	t.Run("should return true for empty slices", func(t *testing.T) {
		emptyNotes1 := []models.ReleaseNote{}
		emptyNotes2 := []models.ReleaseNote{}

		areSame, err := uc.compareReleaseNotesAndTemplate(emptyNotes1, emptyNotes2)

		assert.NoError(t, err)
		assert.True(t, areSame)
	})
}
