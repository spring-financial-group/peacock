package diffs

import (
	"context"
	"testing"

	"github.com/google/go-github/v48/github"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"
)

const (
	owner    = "spring-financial-group"
	repoName = "JX3_Azure_Vault_Dev_Cluster"
	prNumber = 82352
	baseRef  = "master"
	headRef  = "head-sha"

	schemaCMPath = "config-root/namespaces/jx-preproduction/mqube-extract-service/mqube-extract-service-origo-schema-cm.yaml"
)

var diffConfigs = models.DiffConfigs{
	{
		Paths:  []string{"config-root/**/*-schema-cm.yaml"},
		Format: models.DiffFormatConfigMap,
	},
}

func schemaCM(t *testing.T, version string, schema string) []byte {
	t.Helper()
	f := cmFixture{APIVersion: "v1", Kind: "ConfigMap"}
	f.Metadata.Name = "mqube-extract-service-origo-schema"
	f.Metadata.Annotations = map[string]string{"mqube.com/extract-schema-version": version}
	f.Data = map[string]string{"OrigoExtractEntry.json": schema}
	b, err := yaml.Marshal(f)
	assert.NoError(t, err)
	return b
}

func mockPRBasics(mockSCM *mocks.SCM, files []*github.CommitFile) {
	mockSCM.On("GetPRBaseBranchFromPRNumber", mock.Anything, owner, repoName, prNumber).Return(utils.NewPtr(baseRef), nil).Once()
	mockSCM.On("GetPRBranchSHAFromPRNumber", mock.Anything, owner, repoName, prNumber).Return(utils.NewPtr("pr-branch"), utils.NewPtr(headRef), nil).Once()
	mockSCM.On("GetFilesChangedFromPR", mock.Anything, owner, repoName, prNumber).Return(files, nil).Once()
}

func TestUseCase_GenerateDiffSection(t *testing.T) {
	t.Run("renders the file changes section for matched files", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockPRBasics(mockSCM, []*github.CommitFile{
			{Filename: github.String("helmfiles/jx-preproduction/helmfile.yaml"), Status: github.String("modified")},
			{Filename: github.String(schemaCMPath), Status: github.String("modified"), SHA: github.String("blob-sha")},
		})
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath).Return(schemaCM(t, "1.0.0", schemaA), nil).Once()
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath).Return(schemaCM(t, "1.1.0", schemaA2), nil).Once()

		uc := NewUseCase(mockSCM)
		sections, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.NoError(t, err)
		assert.Contains(t, sections[""], "📊 File Changes — 1 file")
		assert.Contains(t, sections[""], "**[`mqube-extract-service-origo-schema-cm.yaml`](https://github.com/spring-financial-group/JX3_Azure_Vault_Dev_Cluster/pull/82352/files#diff-blob-sha)**")
		assert.Contains(t, sections[""], "<code>OrigoExtractEntry.json</code> — 2 fields changed")
		assert.Contains(t, sections[""], "| `a` | ✏ changed | string → integer |")
	})

	t.Run("new files skip the base fetch", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockPRBasics(mockSCM, []*github.CommitFile{
			{Filename: github.String(schemaCMPath), Status: github.String("added"), SHA: github.String("blob-sha")},
		})
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath).Return(schemaCM(t, "1.0.0", schemaA), nil).Once()

		uc := NewUseCase(mockSCM)
		sections, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.NoError(t, err)
		assert.Contains(t, sections[""], "<code>OrigoExtractEntry.json</code> — 1 field (new)")
		mockSCM.AssertNotCalled(t, "GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath)
	})

	t.Run("deleted files skip the head fetch", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockPRBasics(mockSCM, []*github.CommitFile{
			{Filename: github.String(schemaCMPath), Status: github.String("removed")},
		})
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath).Return(schemaCM(t, "1.0.0", schemaA), nil).Once()

		uc := NewUseCase(mockSCM)
		sections, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.NoError(t, err)
		assert.Contains(t, sections[""], "<code>OrigoExtractEntry.json</code> — removed")
		mockSCM.AssertNotCalled(t, "GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath)
	})

	t.Run("no matching files returns an empty section", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockPRBasics(mockSCM, []*github.CommitFile{
			{Filename: github.String("helmfiles/jx-preproduction/helmfile.yaml"), Status: github.String("modified")},
		})

		uc := NewUseCase(mockSCM)
		sections, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("matched but unchanged files return an empty section", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockPRBasics(mockSCM, []*github.CommitFile{
			{Filename: github.String(schemaCMPath), Status: github.String("modified")},
		})
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath).Return(schemaCM(t, "1.0.0", schemaA), nil).Once()
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath).Return(schemaCM(t, "1.1.0", schemaA), nil).Once()

		uc := NewUseCase(mockSCM)
		sections, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("output is deterministic", func(t *testing.T) {
		files := []*github.CommitFile{
			{Filename: github.String(schemaCMPath), Status: github.String("modified"), SHA: github.String("blob-sha")},
		}

		firstSCM := mocks.NewSCM(t)
		mockPRBasics(firstSCM, files)
		firstSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath).Return(schemaCM(t, "1.0.0", schemaA), nil).Once()
		firstSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath).Return(schemaCM(t, "1.1.0", schemaA2), nil).Once()

		secondSCM := mocks.NewSCM(t)
		mockPRBasics(secondSCM, files)
		secondSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath).Return(schemaCM(t, "1.0.0", schemaA), nil).Once()
		secondSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath).Return(schemaCM(t, "1.1.0", schemaA2), nil).Once()

		first, err := NewUseCase(firstSCM).GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.NoError(t, err)
		second, err := NewUseCase(secondSCM).GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.NoError(t, err)
		assert.Equal(t, first, second)
	})

	t.Run("diffs attach to the configured team", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockPRBasics(mockSCM, []*github.CommitFile{
			{Filename: github.String(schemaCMPath), Status: github.String("modified"), SHA: github.String("blob-sha")},
		})
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath).Return(schemaCM(t, "1.0.0", schemaA), nil).Once()
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath).Return(schemaCM(t, "1.1.0", schemaA2), nil).Once()

		teamConfigs := models.DiffConfigs{
			{
				Paths:  []string{"config-root/**/*-schema-cm.yaml"},
				Format: models.DiffFormatConfigMap,
				Team:   "AllClients",
			},
		}

		uc := NewUseCase(mockSCM)
		sections, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, teamConfigs)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Contains(t, sections["AllClients"], "📊 File Changes — 1 file")
		assert.NotContains(t, sections, "")
	})

	t.Run("api errors are returned", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockSCM.On("GetPRBaseBranchFromPRNumber", mock.Anything, owner, repoName, prNumber).Return(utils.NewPtr(baseRef), nil).Once()
		mockSCM.On("GetPRBranchSHAFromPRNumber", mock.Anything, owner, repoName, prNumber).Return(utils.NewPtr("pr-branch"), utils.NewPtr(headRef), nil).Once()
		mockSCM.On("GetFilesChangedFromPR", mock.Anything, owner, repoName, prNumber).Return(nil, assert.AnError).Once()

		uc := NewUseCase(mockSCM)
		_, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, diffConfigs)
		assert.Error(t, err)
	})

	t.Run("unparseable files render a warning instead of failing", func(t *testing.T) {
		mockSCM := mocks.NewSCM(t)
		mockPRBasics(mockSCM, []*github.CommitFile{
			{Filename: github.String(schemaCMPath), Status: github.String("modified")},
		})
		// valid yaml configmap wrapping invalid inner json falls back to text, but a broken *schema*
		// document exercised via json format proves the per file warning path
		jsonConfigs := models.DiffConfigs{
			{
				Paths:  []string{"config-root/**/*-schema-cm.yaml"},
				Format: models.DiffFormatJSON,
			},
		}
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, baseRef, schemaCMPath).Return([]byte("not json at all"), nil).Once()
		mockSCM.On("GetFileFromBranch", mock.Anything, owner, repoName, headRef, schemaCMPath).Return([]byte("also not json"), nil).Once()

		uc := NewUseCase(mockSCM)
		sections, err := uc.GenerateDiffSection(context.Background(), owner, repoName, prNumber, jsonConfigs)
		assert.NoError(t, err)
		assert.Contains(t, sections[""], "⚠ failed to diff this file")
	})
}
