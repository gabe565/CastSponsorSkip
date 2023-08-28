package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_completeCategories(t *testing.T) {
	completions, directive := completeCategories(&cobra.Command{}, []string{}, "")
	assert.NotEqual(t, cobra.ShellCompDirectiveError, directive)
	assert.NotEmpty(t, completions)
}
