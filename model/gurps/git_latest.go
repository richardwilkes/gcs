// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/client"
	"github.com/go-git/go-git/v6/plumbing/protocol/packp"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/richardwilkes/toolbox/v2/errs"
)

func discoverLatestCommit(ctx context.Context, repoURL, accessToken string) (string, error) {
	repo, err := git.CloneContext(ctx, memory.NewStorage(), nil, &git.CloneOptions{
		URL:           repoURL,
		ClientOptions: clientOptions(accessToken, nil),
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

// downloadLatestCommit clones the newest commit of the repository into fs. received, which may be nil, is called with
// the size of each chunk of data that arrives, so that the caller can show how far along the clone is.
func downloadLatestCommit(ctx context.Context, repoURL, accessToken string, fs billy.Filesystem, received func(n int64)) (hash string, err error) {
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
		ClientOptions: clientOptions(accessToken, received),
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

func clientOptions(accessToken string, received func(n int64)) []client.Option {
	var options []client.Option
	if accessToken != "" {
		options = append(options, client.WithHTTPAuth(&githttp.BasicAuth{
			Username: "gcs",
			Password: accessToken,
		}))
	}
	if received != nil {
		options = append(options, client.WithHTTPClient(&http.Client{
			Transport: &countingTransport{base: defaultHTTPTransport(), received: received},
		}))
	}
	return options
}

// countingTransport reports the data arriving in the responses it hands back. go-git has no progress reporting of its
// own for what it receives: the messages it writes to CloneOptions.Progress come from the server and stop before the
// pack is sent, so counting what turns up here is the only measure of how far a clone has gotten.
type countingTransport struct {
	base     http.RoundTripper
	received func(n int64)
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rsp, err := t.base.RoundTrip(req)
	if err != nil || rsp == nil || rsp.Body == nil || rsp.Body == http.NoBody {
		return rsp, err
	}
	rsp.Body = newCountingReadCloser(rsp.Body, t.received)
	return rsp, nil
}

// defaultHTTPTransport returns a copy of the transport go-git would have used had it not been handed a client of its
// own, so that supplying one to count bytes doesn't quietly discard the proxy and TLS behavior that came with it.
func defaultHTTPTransport() http.RoundTripper {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		return transport.Clone()
	}
	return http.DefaultTransport
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
