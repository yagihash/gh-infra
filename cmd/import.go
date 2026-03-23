package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/babarot/gh-infra/internal/gh"
	"github.com/babarot/gh-infra/internal/manifest"
	"github.com/babarot/gh-infra/internal/repository"
	"github.com/babarot/gh-infra/internal/ui"
	goyaml "github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <owner/repo | owner/>",
		Short: "Export existing repository settings as YAML",
		Long:  "Fetch current GitHub repository settings and output them as gh-infra YAML.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(args[0])
		},
	}
	return cmd
}

func runImport(target string) error {
	runner := gh.NewRunner(false)
	fetcher := repository.NewFetcher(runner)

	parts := strings.SplitN(target, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target: %q (expected owner/repo or owner/)", target)
	}

	owner := parts[0]
	name := parts[1]

	if name == "" {
		return importAllRepos(owner, runner, fetcher)
	}

	return importSingleRepo(owner, name, fetcher)
}

func importSingleRepo(owner, name string, fetcher *repository.Fetcher) error {
	current, err := fetcher.FetchRepository(owner, name)
	if err != nil {
		return err
	}

	m := repository.ToManifest(current)
	data, err := goyaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	fmt.Fprint(os.Stdout, string(data))
	return nil
}

const defaultImportParallel = 5

func importAllRepos(owner string, runner gh.Runner, fetcher *repository.Fetcher) error {
	out, err := runner.Run("repo", "list", owner, "--json", "name", "--limit", manifest.DefaultMaxRepoList)
	if err != nil {
		return fmt.Errorf("list repos for %s: %w", owner, err)
	}

	var repos []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &repos); err != nil {
		return fmt.Errorf("parse repo list: %w", err)
	}

	if len(repos) == 0 {
		return nil
	}

	// Start spinner display
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = owner + "/" + r.Name
	}
	tracker := ui.RunRefresh(names)

	// Fetch all repos in parallel
	type importResult struct {
		data []byte
		err  error
	}
	results := make([]importResult, len(repos))
	sem := semaphore.NewWeighted(defaultImportParallel)
	var wg sync.WaitGroup

	for i, r := range repos {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			_ = sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			fullName := owner + "/" + name
			current, err := fetcher.FetchRepository(owner, name)
			if err != nil {
				results[idx] = importResult{err: err}
				tracker.Error(fullName, err)
				return
			}
			m := repository.ToManifest(current)
			data, err := goyaml.Marshal(m)
			results[idx] = importResult{data: data, err: err}
			if err != nil {
				tracker.Error(fullName, err)
			} else {
				tracker.Done(fullName)
			}
		}(i, r.Name)
	}

	wg.Wait()
	tracker.Wait()

	// Output in order
	first := true
	for i, r := range results {
		if r.err != nil {
			ui.SkipImportError(owner+"/"+repos[i].Name, r.err)
			continue
		}
		if !first {
			fmt.Println("---")
		}
		fmt.Fprint(os.Stdout, string(r.data))
		first = false
	}
	return nil
}
