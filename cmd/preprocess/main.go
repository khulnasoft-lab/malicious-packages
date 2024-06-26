// Copyright 2023 Malicious Packages Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/renameio"

	"github.com/khulnasoft-lab/malicious-packages/internal/config"
	"github.com/khulnasoft-lab/malicious-packages/internal/report"
	"github.com/khulnasoft-lab/malicious-packages/internal/reportio"
)

var tempDir string

func main() {
	configFlag := flag.String("config", "", "the filepath to the YAML config file")
	flag.Parse()

	if *configFlag == "" {
		log.Fatalf("-config flag is required")
	}

	// read sources from config
	configFile, err := os.Open(*configFlag)
	if err != nil {
		log.Fatalf("Failed to open config file %s: %v", *configFlag, err)
	}
	c, err := config.ReadYAML(configFile)
	if err != nil {
		log.Fatalf("Failed reading config: %v", err)
	}

	// Create a temp directory for atomic writes. We use the system temp
	// directory to avoid accidentally leaving temp junk in the repository.
	tempDir, err = os.MkdirTemp("", "osv-merge-*")
	if err != nil {
		log.Fatalf("Failed creating temp dir: %v", err)
	}
	defer func() {
		// Clean up temp directory
		if err := os.RemoveAll(tempDir); err != nil {
			log.Fatalf("Failed cleaning up temp dir: %v", err)
		}
	}()

	if err := preprocessRepo(c); err != nil {
		log.Fatalf("Failed to preprocess repo: %v", err) //nolint:gocritic
	}
}

func preprocessRepo(c *config.Config) error {
	err := filepath.WalkDir(c.MaliciousPath, fs.WalkDirFunc(func(path string, info fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return filepath.SkipDir
		} else if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		p, err := filepath.Rel(c.MaliciousPath, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		reports, err := reportio.ReportsInPaths(p, c.Paths())
		if err != nil {
			return fmt.Errorf("failed getting reports: %w", err)
		}
		if len(reports) == 0 {
			// No reports means there is no work to be done.
			return nil
		}
		var noIDs []string
		var withIDs []string
		for _, n := range reports {
			if hasID(c.IDPrefix, n) {
				withIDs = append(withIDs, n)
			} else {
				noIDs = append(noIDs, n)
			}
		}
		if n := len(withIDs); n > 1 {
			// If there is more than one reports with an ID then our assumption
			// that there is only a single OSV per package is wrong.
			return fmt.Errorf("%d reports with IDs in %s (%v)", n, p, withIDs)
		}
		if len(noIDs) == 0 {
			// All IDs are assigned
			return nil
		}

		var basis string
		var toMerge []string
		if len(withIDs) == 1 {
			basis = withIDs[0]
			toMerge = noIDs
		} else {
			basis = noIDs[0]
			toMerge = noIDs[1:]
		}

		return processReports(p, basis, toMerge)
	}))
	return err
}

func hasID(prefix, name string) bool {
	base := filepath.Base(name)
	return !strings.HasPrefix(base, fmt.Sprintf("%s-0000-", prefix))
}

func processReports(path, dest string, mergeSrcs []string) error {
	log.Printf("Processing %s", path)
	log.Printf("  dest = %s", dest)

	destReport, err := report.FromFile(dest)
	if err != nil {
		return fmt.Errorf("failed loading dest %s: %w", dest, err)
	}

	if destReport.IsWithdrawn() && destReport.ID() == "" {
		if len(mergeSrcs) == 0 {
			// The destination is new and withdrawn and we aren't merging with
			// any other reports, so remove the report because we don't already
			// redacted reports.
			// NOTE: this may cause the same withdrawn report to be ingested and
			// ignored repeated times, as the origin information is lost when
			// the report is deleted. We may eventually decide to keep reports
			// even if they are withdrawn.
			if err := os.Remove(dest); err != nil {
				return fmt.Errorf("failed to remove %s: %w", dest, err)
			}
			return nil
		}
		// TODO: implement withdrawn behaviour
		return fmt.Errorf("merging new withdrawn reports is currently unsupported")
	}

	// Ensure the base report is always normalized.
	log.Printf("  normalizing %s", filepath.Base(dest))
	if err := destReport.Normalize(); err != nil {
		return fmt.Errorf("failed normalizing %s: %w", dest, err)
	}

	for _, src := range mergeSrcs {
		log.Printf("  merging %s", filepath.Base(src))

		srcReport, err := report.FromFile(src)
		if err != nil {
			return fmt.Errorf("failed loading src %s: %w", src, err)
		}

		if srcReport.IsWithdrawn() {
			// TODO: implement withdrawn behaviour
			return fmt.Errorf("merging new withdrawn reports is currently unsupported")
		}

		if err := destReport.Merge(srcReport); err != nil {
			return fmt.Errorf("failed to merge %s: %w", src, err)
		}
	}

	// Save the destination report atomically to avoid any corruption.
	t, err := renameio.TempFile(tempDir, dest)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", dest, err)
	}
	defer t.Cleanup()
	err = destReport.WriteJSON(t)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}
	if err := t.CloseAtomicallyReplace(); err != nil {
		return fmt.Errorf("atomic save failed: %w", err)
	}

	// Clean up all the files that we merged.
	for _, src := range mergeSrcs {
		if err := os.Remove(src); err != nil {
			return fmt.Errorf("failed to remove %s: %w", src, err)
		}
	}

	return nil
}
