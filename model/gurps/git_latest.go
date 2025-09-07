package gurps

import (
	"context"
	"log/slog"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/protocol/packp"
	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/richardwilkes/toolbox/v2/errs"
)

func discoverLatestCommit(ctx context.Context, repoURL, accessToken string) (string, error) {
	repo, err := git.CloneContext(ctx, memory.NewStorage(), nil, &git.CloneOptions{
		URL:          repoURL,
		Auth:         gitAuthMethod(accessToken),
		SingleBranch: true,
		NoCheckout:   true,
		Depth:        1,
		Tags:         plumbing.NoTags,
		Filter:       packp.FilterTreeDepth(0),
		Bare:         true,
		Progress:     &logGitProgress{},
	})
	if err != nil {
		return "", errs.NewWithCause("unable to minimally clone "+repoURL, err)
	}
	var ref *plumbing.Reference
	if ref, err = repo.Head(); err != nil {
		return "", errs.NewWithCause("unable to get HEAD for "+repoURL, err)
	}
	return ref.Hash().String(), nil
}

func downloadLatestCommit(ctx context.Context, repoURL, accessToken string, fs billy.Filesystem) (hash string, err error) {
	var repo *git.Repository
	repo, err = git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		URL:          repoURL,
		Auth:         gitAuthMethod(accessToken),
		SingleBranch: true,
		Depth:        1,
		Tags:         plumbing.NoTags,
		Progress:     &logGitProgress{},
	})
	if err != nil {
		return "", errs.NewWithCause("unable to clone "+repoURL, err)
	}
	var ref *plumbing.Reference
	if ref, err = repo.Head(); err != nil {
		return "", errs.NewWithCause("unable to get HEAD for "+repoURL, err)
	}
	return ref.Hash().String(), nil
}

func gitAuthMethod(accessToken string) transport.AuthMethod {
	if accessToken == "" {
		return nil
	}
	return &githttp.BasicAuth{
		Username: "gcs",
		Password: accessToken,
	}
}

type logGitProgress struct{}

func (l *logGitProgress) Write(buffer []byte) (n int, err error) {
	var result []byte
	for _, b := range buffer {
		switch {
		case b == '\n' || b == '\r':
			if len(result) != 0 {
				slog.Info(string(result))
				result = result[:0]
			}
		case b < 32:
		default:
			result = append(result, b)
		}
	}
	if len(result) != 0 {
		slog.Info(string(result))
	}
	return len(buffer), nil
}
