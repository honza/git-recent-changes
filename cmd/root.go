// git-recent-changes
// Copyright (C) 2022  Honza Pokorny <honza@pokorny.ca>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/honza/git-recent-changes/pkg/git-recent-changes"
	"github.com/spf13/cobra"
)

// CLI flags
var StartRef string
var EndRef string
var UpstreamRepo string
var DownstreamRepo string
var Indent int
var ConfigPath string
var RepoPath string
var OutputFormat string
var Since string
var Verbose bool

var rootCmd = &cobra.Command{
	Use:   "git-recent-changes",
	Short: "git-recent-changes",
	Run: func(cmd *cobra.Command, args []string) {
		var since time.Time
		var err error

		if Since != "" {
			since, err = time.Parse("2006-01-02", Since)
			if err != nil {
				fmt.Println("Failed to parse 'since' date.  Please make sure to use '2021-02-28'.")
				os.Exit(1)

			}
		}

		options := recentchanges.NewOptions(
			UpstreamRepo,
			DownstreamRepo,
			StartRef,
			EndRef,
			Indent,
			ConfigPath,
			RepoPath,
			OutputFormat,
			since,
			Verbose,
		)

		err = recentchanges.Run(options)
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ConfigPath, "config", "c", "~/.config/gitrecentchanges.ini", "Path to a config file")
	rootCmd.PersistentFlags().StringVar(&StartRef, "start-ref", "HEAD", "Where should we start?")
	rootCmd.PersistentFlags().StringVar(&EndRef, "end-ref", "HEAD~1", "Where should we stop?")
	rootCmd.PersistentFlags().StringVar(&UpstreamRepo, "upstream-repo", "", "")
	rootCmd.PersistentFlags().StringVar(&DownstreamRepo, "downstream-repo", "", "")
	rootCmd.PersistentFlags().StringVarP(&RepoPath, "dir", "d", ".", "Directory where the git repository is found")
	rootCmd.PersistentFlags().StringVarP(&OutputFormat, "output-format", "o", "plain", "options: plain, json")
	rootCmd.PersistentFlags().StringVar(&Since, "since", "", "e.g. 2021-03-21")
	rootCmd.PersistentFlags().IntVarP(&Indent, "indent", "i", 8, "How many spaces to indent?")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Show extra output")
}
