package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_completeNetworkInterface(t *testing.T) {
	completions, directive := completeNetworkInterface(&cobra.Command{}, []string{}, "")
	assert.NotEqual(t, cobra.ShellCompDirectiveError, directive)
	assert.NotEmpty(t, completions)
}
