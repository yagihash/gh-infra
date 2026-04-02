package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/babarot/gh-infra/internal/infra"
)

func newImportCmd() *cobra.Command {
	var intoPath string

	cmd := &cobra.Command{
		Use:   "import <owner/repo> [owner/repo ...]",
		Short: "Export existing repository settings as YAML",
		Long: "Fetch current GitHub repository settings and output them as gh-infra YAML.\n" +
			"With --into, pull GitHub state back into existing local manifests.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if intoPath != "" {
				return runImportInto(args, intoPath)
			}
			return runImport(args)
		},
	}

	cmd.Flags().StringVar(&intoPath, "into", "",
		"Pull GitHub state into existing local manifests (dir or file path)")

	return cmd
}

func runImport(args []string) error {
	// stdout mode: Import handles all display internally.
	_, err := infra.Import(infra.ImportOptions{Args: args})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			printCancelled()
			return nil
		}
		return err
	}
	return nil
}

func runImportInto(args []string, intoPath string) error {
	result, err := infra.Import(infra.ImportOptions{Args: args, Into: intoPath})
	if err != nil {
		return err
	}

	if !result.Matched {
		return nil // message already printed by infra
	}

	if !result.HasChanges() {
		result.Printer().Message("\nNo changes detected")
		return nil
	}

	p := result.Printer()

	// Confirm with diff viewer (file-level) or simple prompt (repo-only).
	fileEntries := result.DiffEntries()

	var ok bool
	if len(fileEntries) > 0 {
		ok, err = p.ConfirmWithDiff("Apply import changes?", fileEntries)
		if err != nil {
			return err
		}
		result.MarkSkips(fileEntries)
	} else {
		ok, err = p.Confirm("Apply import changes?")
	}
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if err := result.Write(); err != nil {
		return err
	}

	p.Summary(fmt.Sprintf("Import complete! %d documents updated.", result.Plan.UpdatedDocs))
	return nil
}
