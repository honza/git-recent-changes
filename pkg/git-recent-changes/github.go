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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/ini.v1"
)

type Config struct {
	Username string
	Token    string
}

func LoadConfig(configPath string) (Config, error) {
	if strings.HasPrefix(configPath, "~") {
		homeDir := os.ExpandEnv("$HOME")
		configPath = strings.Replace(configPath, "~", homeDir, 1)
	}
	exists, err := PathExists(configPath)
	if err != nil {
		return Config{}, err
	}

	if !exists {
		fmt.Println("WARN: using an empty config")
		return Config{}, nil

	}
	cfg, err := ini.Load(configPath)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Username: cfg.Section("").Key("username").String(),
		Token:    cfg.Section("").Key("token").String(),
	}

	if config.Username == "" {
		fmt.Println("WARN: using an empty username")
	}

	if config.Token == "" {
		fmt.Println("WARN: using an empty token")
	}

	return config, nil
}

type GitHubPullRequest struct {
	Title          string
	Body           string
	MergedAt       time.Time `json:"merged_at"`
	MergeCommitSha string    `json:"merge_commit_sha"`
	HTMLURL        string    `json:"html_url"`
}

type GitHubError int

const (
	GitHubErrorNil = iota
	GitHub404Error
	GitHubRateLimitError
	GitHubHTTPError
	GitHubBodyError
	GitHubJSONParseError
)

type PRResult int

const (
	PRResultFound = iota
	PRResultNotFound
	PRResultRateLimit
	PRResultHTTPError
)

func makeGitHubRequest(config Config, options Options, upstreamRepo string, pull string) ([]byte, GitHubError) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s", upstreamRepo, pull)
	client := &http.Client{}
	options.Log("GET", apiURL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return []byte{}, GitHubHTTPError
	}
	req.Header.Add("Accept", `application/vnd.github.v3+json`)
	req.SetBasicAuth(config.Username, config.Token)

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, GitHubHTTPError
	}

	if resp.StatusCode == http.StatusNotFound {
		return []byte{}, GitHub404Error
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return []byte{}, GitHubRateLimitError
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return []byte{}, GitHubBodyError
	}

	return body, GitHubErrorNil
}

func retrievePullRequest(config Config, options Options, upstreamRepo string, pullRequestNumber string) (GitHubPullRequest, GitHubError) {
	var pr GitHubPullRequest
	body, gerr := makeGitHubRequest(config, options, upstreamRepo, pullRequestNumber)
	if gerr != GitHubErrorNil {
		return pr, gerr
	}
	err := json.Unmarshal(body, &pr)

	if err != nil {
		return pr, GitHubJSONParseError
	}

	return pr, GitHubErrorNil
}

func FindPullRequest(config Config, options Options, pullRequestNumber string, commit *object.Commit) (GitHubPullRequest, PRResult) {
	// Try to find an upstream PR
	//    if it exits, and it matches the commit sha, return it
	//    if we fail to find it or if it doesn't match the commit sha
	//
	//    then
	//
	// Try to find a downstream PR
	//    if it exits, and it matches the commit sha, return it
	//    if we fail to find it or if it doesn't match the commit sha
	//
	//    then
	//
	// Return error
	//
	// --------------
	//
	// If we encounter a GitHub rate limit error, exit early.

	reposToTry := []string{options.UpstreamRepo, options.DownstreamRepo}

	var pr GitHubPullRequest
	var gerr GitHubError

	for _, repo := range reposToTry {
		pr, gerr = retrievePullRequest(config, options, repo, pullRequestNumber)

		switch gerr {
		case GitHubRateLimitError:
			return pr, PRResultRateLimit
		case GitHubBodyError, GitHubJSONParseError, GitHubHTTPError:
			return pr, PRResultHTTPError
		case GitHub404Error:
			// Try with downstream
			continue
		case GitHubErrorNil:
			if pr.MergeCommitSha != commit.Hash.String() {
				// Try with downstream
				continue
			} else {
				return pr, PRResultFound
			}
		}
	}

	return pr, PRResultNotFound
}
