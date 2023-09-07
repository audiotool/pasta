package github

import (
	"github.com/audiotool/pasta/pkg/copier"
	gh "github.com/google/go-github/v53/github"
)

func toCommitInfo(com *gh.Commit) *copier.SourceInfo {
	return &copier.SourceInfo{
		Reference: com.GetSHA(),
		Message:   com.GetMessage(),
		Author: copier.Author{
			Date:  com.Committer.GetDate().Time,
			Name:  com.Committer.GetName(),
			Email: com.Committer.GetEmail(),
		},
	}
}
