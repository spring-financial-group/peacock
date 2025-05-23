package webhookuc

import (
	"context"
	"github.com/google/go-github/v48/github"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"
	"testing"
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

		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.PendingState, domain.ValidationContext).Return(nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.Branch, ".peacock/feathers.yaml").Return(mockFeathersData, nil).Once()

		mockSCM.On("GetPRCommentsByUser", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber).Return(nil, nil).Once()
		mockSCM.On("DeleteUsersComments", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber).Return(nil).Once()
		mockSCM.On("CommentOnPR", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.PRNumber, mock.Anything).Return(nil).Once()
		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.RepoOwner, mockEvent.RepoName, mockEvent.SHA, domain.SuccessState, domain.ValidationContext).Return(nil).Once()

		mockNotesUC.On("GetReleaseNotesFromMarkdownAndTeamsInFeathers", prBody, allTeams).Return(mockNotes, nil)
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
