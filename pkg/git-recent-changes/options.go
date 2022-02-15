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

package recentchanges

import (
	"fmt"
	"time"
)

type Options struct {
	UpstreamRepo   string
	DownstreamRepo string
	StartRef       string
	EndRef         string
	Indent         int
	ConfigPath     string
	RepoPath       string
	OutputFormat   string
	Since          time.Time
	Log            func(values ...string)
}

func NewOptions(
	UpstreamRepo string,
	DownstreamRepo string,
	StartRef string,
	EndRef string,
	Indent int,
	ConfigPath string,
	RepoPath string,
	OutputFormat string,
	Since time.Time,
	Verbose bool,
) Options {
	options := Options{
		UpstreamRepo:   UpstreamRepo,
		DownstreamRepo: DownstreamRepo,
		StartRef:       StartRef,
		EndRef:         EndRef,
		Indent:         Indent,
		ConfigPath:     ConfigPath,
		RepoPath:       RepoPath,
		OutputFormat:   OutputFormat,
		Since:          Since,
		Log:            noOpLogger,
	}

	if Verbose {
		options.Log = verboseLogger
	}

	return options
}

func verboseLogger(values ...string) {
	faces := make([]interface{}, len(values))
	for i, v := range values {
		faces[i] = v
	}
	fmt.Println(faces...)
}

func noOpLogger(values ...string) {}
