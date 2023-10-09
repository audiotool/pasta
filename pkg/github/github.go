package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/audiotool/pasta/pkg/copier"
	"github.com/audiotool/pasta/pkg/utils"
	gh "github.com/google/go-github/v53/github"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/oauth2"
)

type Copier struct{}

func (*Copier) Matches(url string) bool {
	match, err := regexp.MatchString(".*github\\.com.*", url)
	return err == nil && match
}

func (*Copier) Copy(ctx context.Context, config copier.CopyConfig) (any, error) {
	client := createClient(ctx)

	// get repo owner and name from url
	owner, repo, err := parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse url %v, error: %v", config.URL, err)
	}

	// check if we have access to repo
	if _, _, err := client.Repositories.Get(ctx, owner, repo); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, fmt.Errorf("can't access repo %v, do you have correct access rights & API key setup?", config.URL)
		}

		return nil, fmt.Errorf("error accessing repo: %v", err)
	}

	// get sha from ref
	ref := config.Options["ref"]
	sha, err := getCommitSha(ctx, client, owner, repo, ref)
	if err != nil {
		return nil, fmt.Errorf("error retreiving sha: %v", err)
	}

	// fetch the tree
	tree, _, err := client.Git.GetTree(ctx, owner, repo, sha, true)
	if err != nil {
		return nil, fmt.Errorf("error getting tree: %v", err)
	}

	// download files, save to temp directory
	wg := pool.New().WithMaxGoroutines(20)
	for _, entry := range tree.Entries {
		if entry.GetType() != "blob" {
			continue
		}

		entry := entry

		// check if file path is in dest directory
		p := entry.GetPath()

		if !strings.HasPrefix(p, config.From) {
			continue
		}

		relp, err := filepath.Rel(config.From, p)
		if err != nil {
			return nil, fmt.Errorf("unable to create relp: %v", err)
		}

		// check if user wants to keep this file based on relp
		if !config.Keep(relp) {
			continue
		}

		// download concurrently
		wg.Go(func() {
			bs, _, err := client.Git.GetBlobRaw(ctx, owner, repo, entry.GetSHA())
			if err != nil {
				fmt.Println("Error fetching file", entry.GetPath(), err)
				return
			}

			err = utils.SaveFile(bs, path.Join(config.TempDir, relp))

			if err != nil {
				fmt.Println("Error downloading file", entry.GetPath(), err)
				return
			}
		})
	}

	wg.Wait()

	// create info message for pasta.result.yaml: Fetch commit
	com, _, err := client.Git.GetCommit(ctx, owner, repo, sha)
	if err != nil {
		return nil, fmt.Errorf("error getting commit info: %v", err)
	}

	return toCommitInfo(com), nil
}

// create a github client. Uses env var GITHUB_TOKEN as api token
// if set, otherwise initializes client without a token.
func createClient(ctx context.Context) *gh.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return gh.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return gh.NewClient(tc)
}

var errMalformedURL = errors.New("url is malformed, must be of shape github.com/<owner>/<repo>/")

// given a github url like github.com/<owner>/<repo-name>, returns owner & repo-name
func parse(url string) (owner string, repo string, err error) {
	// extract owner & repo using regex, assumptions:
	// repo name & owner name are only alphanumerical characters + `-`, todo: confirm
	r, err := regexp.Compile(`github\.com/(?P<owner>[a-zA-Z\-]+)/(?P<repo>[a-zA-z\-]+)/?(?P<rest>.*)$`)
	if err != nil {
		return "", "", err
	}
	ms := r.FindStringSubmatch(url)

	if len(ms) < 4 || ms[1] == "" || ms[2] == "" || ms[3] != "" {
		return "", "", errMalformedURL
	}
	return ms[1], ms[2], nil
}

var (
	errCouldNotFetchRepo  = errors.New("failed fetching repo")
	errCouldNotGetRef     = errors.New("failed getting ref")
	errCouldNotResolveRef = errors.New("unable to resolve ref; must be branch, tag, or sha")
)

// returns the sha based on ref. If ref is the empty string, returns the sha of the default branch head.
func getCommitSha(
	ctx context.Context,
	client *gh.Client,
	owner string,
	repo string,
	ref string,
) (string, error) {
	// if ref is empty string, set ref to "heads/<default-branch>"
	if ref == "" {
		repo, _, err := client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return "", fmt.Errorf("%w: %w", errCouldNotFetchRepo, err)
		}
		branch := repo.GetDefaultBranch()
		ref = "heads/" + branch
	}

	// if type of ref isn't explicitly set, try in order which is meant
	heads, _ := regexp.MatchString("heads/.*", ref)
	tags, _ := regexp.MatchString("tags/.*", ref)
	commit, _ := regexp.MatchString("commit/.*", ref)

	// if this is true, we have an actual git-style ref
	if heads || tags {
		// get ref info
		r, _, err := client.Git.GetRef(ctx, owner, repo, ref)
		if err != nil {
			return "", fmt.Errorf("%w: %w", errCouldNotGetRef, err)
		}

		// return ref's sha
		sha := r.Object.GetSHA()
		return sha, nil
	}

	// if this is true, we're already done
	if commit {
		return ref[len("commit/"):], nil
	}

	// else, try to resolve ref as one of heads, tags, commit-sha or ref-sha

	// try ref as branch
	if r, _, err := client.Git.GetRef(ctx, owner, repo, "heads/"+ref); err == nil {
		return r.GetObject().GetSHA(), nil
	}

	// try ref as tag
	if r, _, err := client.Git.GetRef(ctx, owner, repo, "tags/"+ref); err == nil {
		return r.GetObject().GetSHA(), nil
	}
	// try ref as commit SHA
	if _, _, err := client.Git.GetCommit(ctx, owner, repo, ref); err == nil {
		return ref, nil
	}

	// try ref as ref SHA
	if r, _, err := client.Git.GetRef(ctx, owner, repo, ref); err == nil {
		return r.GetObject().GetSHA(), nil
	}

	return "", errCouldNotResolveRef
}
