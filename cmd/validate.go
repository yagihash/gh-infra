package cmd

import (
	"github.com/babarot/gh-infra/internal/manifest"
	"github.com/babarot/gh-infra/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate YAML syntax and schema",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return runValidate(path)
		},
	}
	return cmd
}

func runValidate(path string) error {
	parsed, err := manifest.ParseAll(path)
	if err != nil {
		return err
	}

	ui.ValidateSummary(len(parsed.Repositories), len(parsed.FileSets))
	for _, r := range parsed.Repositories {
		ui.ValidateRepo(r.Metadata.FullName())
	}
	for _, fs := range parsed.FileSets {
		ui.ValidateFileSet(fs.Metadata.Name, len(fs.Spec.Files), len(fs.Spec.Repositories))
	}
	return nil
}
