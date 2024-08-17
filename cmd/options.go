package cmd

import "github.com/spf13/cobra"

type Option func(cmd *cobra.Command)

const (
	VersionKey = "version"
	CommitKey  = "commit"
)

func WithVersion(version string) Option {
	return func(cmd *cobra.Command) {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[VersionKey] = version
		cmd.Version, cmd.Annotations[CommitKey] = buildVersion(version)
		cmd.SetVersionTemplate("CastSponsorSkip {{ .Version }}\n")
		cmd.InitDefaultVersionFlag()
	}
}
