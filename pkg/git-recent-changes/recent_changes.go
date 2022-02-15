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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var prNum = regexp.MustCompile(`Merge pull request #(\d+)`)
var bugNum = regexp.MustCompile(`Bug (\d+):`)

type Bugzilla struct {
	ID   string
	Link string
}

type Commit struct {
	GitCommit   object.Commit
	PullRequest GitHubPullRequest
	Bugzilla    Bugzilla
}

// TODO
func (c *Commit) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"pull_request_title": c.PullRequest.Title,
		"pull_request_body":  c.PullRequest.Body,
		"hash":               c.GitCommit.Hash.String(),
		"bugzilla":           c.Bugzilla.Link,
		"author":             c.GitCommit.Author.Name,
	}
	return json.Marshal(obj)
}

func (c *Commit) GetBody(options Options) (string, error) {
	prefix := strings.Repeat(" ", options.Indent)
	w := bytes.NewBufferString("")

	fmt.Fprintln(w, c.PullRequest.Title)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "%sPR: %s\n", prefix, c.PullRequest.HTMLURL)
	if c.Bugzilla.Link != "" {
		fmt.Fprintf(w, "%sBZ: %s\n", prefix, c.Bugzilla.Link)
	}

	if c.PullRequest.Body != "" {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, formatBody(c.PullRequest.Body, options.Indent))
	}

	fmt.Fprintln(w, "")
	return w.String(), nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func formatBody(body string, indent int) string {
	res := []string{}
	body = strings.TrimSpace(body)
	lines := strings.Split(body, "\n")
	prefix := strings.Repeat(" ", indent)
	for _, line := range lines {
		if line == "" {
			res = append(res, line)
			continue
		}

		res = append(res, fmt.Sprintf("%s%s", prefix, line))
	}

	return strings.Join(res, "\n")
}

func FindCommits(options Options) ([]*object.Commit, error) {
	options.Log("find commits")
	commits := []*object.Commit{}

	r, err := git.PlainOpen(options.RepoPath)

	if err != nil {
		return commits, err
	}

	var from plumbing.Hash

	if options.StartRef == "HEAD" {
		head, err := r.Head()

		if err != nil {
			return commits, err
		}

		from = head.Hash()
	} else {
		revision, err := r.ResolveRevision(plumbing.Revision(options.StartRef))
		if err != nil {
			return commits, err
		}

		from = *revision
	}

	var end *plumbing.Hash
	if options.EndRef != "" {
		end, err = r.ResolveRevision(plumbing.Revision(options.EndRef))
		if err != nil {
			return commits, err
		}
	}
	// LogOrderCommitterTime produces an order that matches the default git-log
	cIter, err := r.Log(&git.LogOptions{From: from, Order: git.LogOrderCommitterTime})

	for {
		commit, err := cIter.Next()

		if err == io.EOF {
			break
		}

		if end != nil {
			if commit.Hash.String() == end.String() {
				break
			}
		}

		// We only care about merges
		if commit.NumParents() < 2 {
			continue
		}

		options.Log("adding", commit.Hash.String())

		commits = append(commits, commit)

		if commit.Committer.When.Before(options.Since) {
			break
		}

	}

	return commits, nil
}

func ParseCommit(config Config, options Options, commit *object.Commit) (*Commit, error) {
	options.Log("parsing commit", commit.Hash.String())
	result := Commit{
		GitCommit: *commit,
	}
	lines := strings.Split(commit.Message, "\n")
	subject := lines[0]
	prNumMatches := prNum.FindStringSubmatch(subject)

	if len(prNumMatches) == 0 {
		return nil, nil
	}

	pullRequestNumber := prNumMatches[1]

	pr, perr := FindPullRequest(config, options, pullRequestNumber, commit)

	switch perr {
	case PRResultRateLimit:
		return &result, errors.New("unlikely to succeed because github is rate limiting our requests")
	case PRResultHTTPError:
		return &result, errors.New("failed to get a response from github")
	case PRResultNotFound:
		return &result, errors.New("pull request not found")
	case PRResultFound:
	}

	result.PullRequest = pr

	bzMatches := bugNum.FindStringSubmatch(pr.Title)

	if len(bzMatches) > 0 {
		result.Bugzilla = Bugzilla{
			ID:   bzMatches[1],
			Link: fmt.Sprintf("https://bugzilla.redhat.com/show_bug.cgi?id=%s", bzMatches[1]),
		}

		result.PullRequest.Title = strings.ReplaceAll(result.PullRequest.Title, bzMatches[0], "")
	}

	return &result, nil
}

func Run(options Options) error {
	if options.UpstreamRepo == "" && options.DownstreamRepo == "" {
		return errors.New(fmt.Sprintln("at least one repository must be specified"))
	}
	config, err := LoadConfig(options.ConfigPath)
	if err != nil {
		return err
	}

	repoPathExists, err := PathExists(options.RepoPath)

	if err != nil {
		return err
	}

	if !repoPathExists {
		return errors.New(fmt.Sprintln("Path", options.RepoPath, "doesn't exist"))
	}

	commits, err := FindCommits(options)

	if err != nil {
		return err
	}

	parsedCommits := []Commit{}

	for _, commit := range commits {
		c, err := ParseCommit(config, options, commit)

		if err != nil {
			fmt.Println(err)
			return err
		}

		if c == nil {
			continue
		}

		parsedCommits = append(parsedCommits, *c)
	}

	switch options.OutputFormat {
	case "json":
		output, err := json.MarshalIndent(parsedCommits, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
	case "plain":
		for _, parsed := range parsedCommits {
			j, err := parsed.GetBody(options)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(j))
		}
	default:
		return errors.New("unknown output format")
	}

	return nil
}
