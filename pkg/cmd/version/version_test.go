package version_test

import (
	"github.com/spring-financial-group/peacock/pkg/cmd/version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := version.GetVersion()
	assert.NotEmpty(t, version)
}
