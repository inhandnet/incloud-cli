package update

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/creativeprojects/go-selfupdate"
)

const (
	githubTimeout = 8 * time.Second
	mirrorBaseURL = "https://incloud-cli-releases.s3.cn-north-1.amazonaws.com.cn"
)

// newSource creates a fallbackSource: GitHub (5s timeout) → S3 mirror.
func newSource(errOut io.Writer) (selfupdate.Source, error) {
	direct, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return nil, err
	}

	mirror, err := selfupdate.NewHttpSource(selfupdate.HttpConfig{
		BaseURL: mirrorBaseURL,
	})
	if err != nil {
		return nil, err
	}

	return &fallbackSource{
		primary:        direct,
		fallback:       mirror,
		primaryTimeout: githubTimeout,
		errOut:         errOut,
	}, nil
}

// fallbackSource tries the primary source first (with timeout).
// If it fails, falls back to the secondary source.
// The source that succeeded for ListReleases is reused for DownloadReleaseAsset.
type fallbackSource struct {
	primary        selfupdate.Source
	fallback       selfupdate.Source
	primaryTimeout time.Duration
	errOut         io.Writer

	chosen selfupdate.Source
}

func (f *fallbackSource) ListReleases(ctx context.Context, repo selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, f.primaryTimeout)
	releases, err := f.primary.ListReleases(timeoutCtx, repo)
	cancel()
	if err == nil {
		f.chosen = f.primary
		return releases, nil
	}

	fmt.Fprintf(f.errOut, "Using alternate source...\n")
	releases, err = f.fallback.ListReleases(ctx, repo)
	if err == nil {
		f.chosen = f.fallback
	}
	return releases, err
}

func (f *fallbackSource) DownloadReleaseAsset(ctx context.Context, rel *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	if f.chosen == nil {
		return nil, fmt.Errorf("no source selected: ListReleases must be called first")
	}
	return f.chosen.DownloadReleaseAsset(ctx, rel, assetID)
}

var _ selfupdate.Source = (*fallbackSource)(nil)
