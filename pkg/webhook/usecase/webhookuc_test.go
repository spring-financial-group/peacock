package webhookuc

import (
	"context"
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
		Owner:         RepoOwner,
		RepoName:      RepoName,
		Body:          "### Notify Infra\n\nHello infra\n\n### Notify Skisocks\n\nHello skisocks",
		PRNumber:      PRNumber,
		SHA:           SHA,
		Branch:        Branch,
		DefaultBranch: DefaultBranch,
	}

	mockFeathers = &feathers.Feathers{
		Teams: []feathers.Team{
			{
				Name:        InfraTeam,
				APIKey:      "AKey",
				ContactType: models.Slack,
				Addresses:   []string{"C1234567890"},
			},
			{
				Name:        SkisocksTeam,
				APIKey:      "AnotherKey",
				ContactType: models.Webhook,
				Addresses:   []string{"skisocks@github.com"},
			},
		},
		Config: feathers.Config{
			Messages: feathers.Messages{
				Subject: "Subject",
			},
		},
	}
	mockFeathersData, _ = yaml.Marshal(mockFeathers)

	mockNotes = []models.ReleaseNote{
		{
			TeamNames: []string{"infra"},
			Content:   "Hello infra",
		},
		{
			TeamNames: []string{"Skisocks"},
			Content:   "Hello skisocks",
		},
	}
)

func TestWebHookUseCase_ValidatePeacock(t *testing.T) {
	mockSCM := mocks.NewSCM(t)
	mockFactory := mocks.NewSCMClientFactory(t)
	mockNotesUC := mocks.NewReleaseNotesUseCase(t)

	cfg := &config.SCM{
		User: RepoOwner,
	}

	uc := NewUseCase(cfg, mockFactory, mockNotesUC)

	t.Run("Happy Path", func(t *testing.T) {
		mockEvent := mockPullRequestEventDTO
		mockEvent.Body = prBody

		mockFactory.On("GetClient", mockEvent.Owner, mockEvent.RepoName, cfg.User, mockEvent.PRNumber).Return(mockSCM).Once()

		clientKey := "key"
		mockSCM.On("GetKey").Return(clientKey).Once()
		mockFactory.On("RemoveClient", clientKey).Return().Once()

		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.SHA, domain.PendingState, domain.ValidationContext).Return(nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.Branch, ".peacock/feathers.yaml").Return(mockFeathersData, nil).Once()

		mockSCM.On("GetPRCommentsByUser", mockCTX).Return(nil, nil).Once()
		mockSCM.On("DeleteUsersComments", mockCTX).Return(nil).Once()
		mockSCM.On("CommentOnPR", mockCTX, mock.Anything).Return(nil).Once()
		mockSCM.On("CreatePeacockCommitStatus", mockCTX, mockEvent.SHA, domain.SuccessState, domain.ValidationContext).Return(nil).Once()

		mockNotesUC.On("ParseNotesFromMarkdown", prBody).Return(mockNotes, nil)
		mockNotesUC.On("ValidateReleaseNotesWithFeathers", mockFeathers, mockNotes).Return(nil)
		mockNotesUC.On("GenerateHash", mockNotes).Return(mockHash, nil)
		mockNotesUC.On("GenerateBreakdown", mockNotes, mockHash, 2).Return("", nil)

		err := uc.ValidatePeacock(mockEvent)
		assert.NoError(t, err)
	})
}

func TestWebHookUseCase_RunPeacock(t *testing.T) {
	mockSCM := mocks.NewSCM(t)
	mockFactory := mocks.NewSCMClientFactory(t)
	mockNotesUC := mocks.NewReleaseNotesUseCase(t)

	cfg := &config.SCM{
		User: RepoOwner,
	}

	uc := NewUseCase(cfg, mockFactory, mockNotesUC)

	t.Run("Happy Path", func(t *testing.T) {
		mockEvent := mockPullRequestEventDTO
		mockEvent.Body = prBody

		mockFactory.On("GetClient", mockEvent.Owner, mockEvent.RepoName, cfg.User, mockEvent.PRNumber).Return(mockSCM).Once()

		clientKey := "key"
		mockSCM.On("GetKey").Return(clientKey).Once()
		mockFactory.On("RemoveClient", clientKey).Return().Once()

		defaultSHA := "default-SHA"
		mockSCM.On("GetLatestCommitSHAInBranch", mockCTX, mockEvent.DefaultBranch).Return(defaultSHA, nil).Once()
		mockSCM.On("CreatePeacockCommitStatus", mockCTX, defaultSHA, domain.PendingState, domain.ReleaseContext).Return(nil).Once()
		mockSCM.On("GetFileFromBranch", mockCTX, mockEvent.DefaultBranch, ".peacock/feathers.yaml").Return(mockFeathersData, nil).Once()

		mockSCM.On("CreatePeacockCommitStatus", mockCTX, defaultSHA, domain.SuccessState, domain.ReleaseContext).Return(nil).Once()

		mockNotesUC.On("ParseNotesFromMarkdown", prBody).Return(mockNotes, nil)
		mockNotesUC.On("ValidateReleaseNotesWithFeathers", mockFeathers, mockNotes).Return(nil)
		mockNotesUC.On("SendReleaseNotes", mockFeathers, mockNotes).Return(nil)

		err := uc.RunPeacock(mockEvent)
		assert.NoError(t, err)
	})
}
