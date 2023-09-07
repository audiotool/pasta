package github

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type urlParseResult struct {
	owner string
	repo  string
	err   error
}

func TestParseUrls(t *testing.T) {

	tests := []struct {
		name   string
		url    string
		result urlParseResult
	}{
		// parsing of ok urls
		{
			url: "github.com/foo/bar",
			result: urlParseResult{
				owner: "foo",
				repo:  "bar",
				err:   nil,
			},
		},
		{
			url: "github.com/foo/bar/",
			result: urlParseResult{
				owner: "foo",
				repo:  "bar",
				err:   nil,
			},
		},
		{
			url: "www.github.com/foo/bar",
			result: urlParseResult{
				owner: "foo",
				repo:  "bar",
				err:   nil,
			},
		},
		{
			url: "www.github.com/foo/bar/",
			result: urlParseResult{
				owner: "foo",
				repo:  "bar",
				err:   nil,
			},
		},
		{
			url: "https://www.github.com/foo/bar/",
			result: urlParseResult{
				owner: "foo",
				repo:  "bar",
				err:   nil,
			},
		},
		{
			url: "https://github.com/foo/bar/",
			result: urlParseResult{
				owner: "foo",
				repo:  "bar",
				err:   nil,
			},
		},

		// parsing of not ok urls
		{
			url: "https://gitlab.com/foo/bar/",
			result: urlParseResult{
				owner: "",
				repo:  "",
				err:   errMalformedURL,
			},
		},
		{
			url: "https://github.com/foo/bar/baz",
			result: urlParseResult{
				owner: "",
				repo:  "",
				err:   errMalformedURL,
			},
		},
		{
			url: "https://github.com/foo/",
			result: urlParseResult{
				owner: "",
				repo:  "",
				err:   errMalformedURL,
			},
		},
	}

	for _, tt := range tests {
		t.Run("url "+tt.url, func(t *testing.T) {
			owner, repo, err := parse(tt.url)

			// if errors don't match
			if !errors.Is(err, tt.result.err) {
				t.Errorf("parse() error = %v, expected err = %v", err, tt.result.err)
				return
			}

			if owner != tt.result.owner {
				t.Errorf("parse() owner = %v, expected owner = %v", owner, tt.result.owner)
				return
			}

			if repo != tt.result.repo {
				t.Errorf("parse() repo = %v, expected repo = %v", repo, tt.result.repo)
			}
		})
	}
}

func TestGetCommitSha(t *testing.T) {
	repo := "pasta"
	owner := "audiotool"
	tests := []struct {
		name      string
		ref       string
		commitSha string
		err       error
	}{
		{
			name:      "explicit branch",
			ref:       "heads/test-branch-do-not-delete",
			commitSha: "b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			err:       nil,
		},
		{
			name:      "implicit branch",
			ref:       "test-branch-do-not-delete",
			commitSha: "b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			err:       nil,
		},

		{
			name:      "explicit commit sha",
			ref:       "commit/b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			commitSha: "b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			err:       nil,
		},
		{
			name:      "implicit commit sha",
			ref:       "b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			commitSha: "b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			err:       nil,
		},

		{
			name:      "explicit tag",
			ref:       "tags/test-tag-do-not-delete",
			commitSha: "b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			err:       nil,
		},

		{
			name:      "implicit tag",
			ref:       "test-tag-do-not-delete",
			commitSha: "b66b28518c9b5519d6da1b5bf601e665b2aacc63",
			err:       nil,
		},

		{
			name:      "unkown explicit branch",
			ref:       "heads/inexistant-marshmellow",
			commitSha: "",
			err:       errCouldNotGetRef,
		},
		{
			name:      "unkown explicit tag",
			ref:       "tags/inexistant-marshmellow",
			commitSha: "",
			err:       errCouldNotGetRef,
		},
		{
			name:      "unkown implicit ref",
			ref:       "inexistant-marshmellow",
			commitSha: "",
			err:       errCouldNotResolveRef,
		},
	}

	ctx := context.Background()
	client := createClient(ctx)

	t.Run("check repo access", func(t *testing.T) {
		if _, _, err := client.Repositories.Get(ctx, owner, repo); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				t.Errorf("can't access repo, do you have correct access rights & API key setup?")
				return
			}

			t.Errorf("error accessing repo: %v", err)
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sha, err := getCommitSha(ctx, client, owner, repo, tt.ref)
			if !errors.Is(err, tt.err) {
				t.Errorf("getCommitSha() error = %v, expected error = %v", err, tt.err)
				return
			}

			if sha != tt.commitSha {
				t.Errorf("getCommitSha() sha = %v, expected sha = %v", sha, tt.commitSha)
			}
		})
	}
}
