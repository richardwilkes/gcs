package gurps

import (
	"context"
	"log/slog"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/client"
	"github.com/go-git/go-git/v6/plumbing/protocol/packp"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/richardwilkes/toolbox/v2/errs"
)

func discoverLatestCommit(ctx context.Context, repoURL, accessToken string) (string, error) {
	repo, err := git.CloneContext(ctx, memory.NewStorage(), nil, &git.CloneOptions{
		URL:           repoURL,
		ClientOptions: clientOptions(accessToken),
		SingleBranch:  true,
		NoCheckout:    true,
		Depth:         1,
		Tags:          plumbing.NoTags,
		Filter:        packp.FilterTreeDepth(0),
		Bare:          true,
		Progress:      &logGitProgress{},
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
	// Disable go-git's NTFS/HFS+ path protections for the checkout. They default to on for all platforms and reject
	// path components whose base name collides with a Windows reserved device name (e.g. "Con Man.gct", where "Con"
	// matches "CON"). Our libraries legitimately contain such names, GCS copies the files to disk itself afterward, and
	// this mirrors cloning with `git -c core.protectNTFS=false -c core.protectHFS=false`.
	storage := memory.NewStorage()
	cfg := config.NewConfig()
	cfg.Core.ProtectNTFS = config.NewOptBool(false)
	cfg.Core.ProtectHFS = config.NewOptBool(false)
	if err = storage.SetConfig(cfg); err != nil {
		return "", errs.NewWithCause("unable to configure clone of "+repoURL, err)
	}
	var repo *git.Repository
	repo, err = git.CloneContext(ctx, storage, fs, &git.CloneOptions{
		URL:           repoURL,
		ClientOptions: clientOptions(accessToken),
		SingleBranch:  true,
		Depth:         1,
		Tags:          plumbing.NoTags,
		Progress:      &logGitProgress{},
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

func clientOptions(accessToken string) []client.Option {
	var options []client.Option
	if accessToken != "" {
		options = append(options, client.WithHTTPAuth(&http.BasicAuth{
			Username: "gcs",
			Password: accessToken,
		}))
	}
	return options
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
