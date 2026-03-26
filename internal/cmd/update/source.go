package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/google/go-github/v74/github"
)

const (
	directTimeout  = 10 * time.Second
	envUpdateProxy = "INCLOUD_UPDATE_PROXY"
)

// newSource creates the appropriate selfupdate.Source based on proxy configuration.
// If a proxy is configured, it returns a fallbackSource that tries direct GitHub first
// (with a short timeout), then falls back to the proxied source.
func newSource(proxyURL string, errOut io.Writer) (src selfupdate.Source, err error) {
	if proxyURL == "" {
		proxyURL = os.Getenv(envUpdateProxy)
	}

	direct, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return nil, err
	}

	if proxyURL == "" {
		src, err = direct, nil
		return
	}

	proxied, err := newProxiedGitHubSource(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("creating proxied source: %w", err)
	}

	return &fallbackSource{
		sources: []namedSource{
			{name: "GitHub (direct)", source: direct, timeout: directTimeout},
			{name: "GitHub (proxy)", source: proxied},
		},
		errOut: errOut,
	}, nil
}

// namedSource pairs a Source with a display name and optional timeout.
type namedSource struct {
	name    string
	source  selfupdate.Source
	timeout time.Duration // 0 means use parent context
}

// fallbackSource tries multiple sources in order. ListReleases probes connectivity
// and remembers the working source; DownloadReleaseAsset reuses that source.
type fallbackSource struct {
	sources      []namedSource
	preferredIdx int
	errOut       io.Writer
}

func (f *fallbackSource) ListReleases(ctx context.Context, repo selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	var lastErr error
	for i, src := range f.sources {
		attemptCtx := ctx
		var cancel context.CancelFunc
		if src.timeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, src.timeout)
		}
		releases, err := src.source.ListReleases(attemptCtx, repo)
		if cancel != nil {
			cancel()
		}
		if err == nil {
			f.preferredIdx = i
			if i > 0 {
				fmt.Fprintf(f.errOut, "Connected via %s\n", src.name)
			}
			return releases, nil
		}
		lastErr = err
		if i < len(f.sources)-1 {
			fmt.Fprintf(f.errOut, "%s unavailable, trying %s...\n", src.name, f.sources[i+1].name)
		}
	}
	return nil, lastErr
}

func (f *fallbackSource) DownloadReleaseAsset(ctx context.Context, rel *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	return f.sources[f.preferredIdx].source.DownloadReleaseAsset(ctx, rel, assetID)
}

// proxiedGitHubSource accesses GitHub API and downloads through an HTTP proxy.
type proxiedGitHubSource struct {
	api       *github.Client
	transport *http.Transport
}

func newProxiedGitHubSource(proxyURL string) (*proxiedGitHubSource, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL %q: %w", proxyURL, err)
	}
	transport := &http.Transport{Proxy: http.ProxyURL(u)}
	client := &http.Client{Transport: transport}
	return &proxiedGitHubSource{
		api:       github.NewClient(client),
		transport: transport,
	}, nil
}

func (s *proxiedGitHubSource) ListReleases(ctx context.Context, repository selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	owner, repo, err := repository.GetSlug()
	if err != nil {
		return nil, err
	}
	rels, res, err := s.api.Repositories.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	releases := make([]selfupdate.SourceRelease, len(rels))
	for i, rel := range rels {
		releases[i] = selfupdate.NewGitHubRelease(rel)
	}
	return releases, nil
}

// DownloadReleaseAsset downloads using the asset's browser download URL through the proxy.
// This avoids needing access to the unexported Release.repository field.
func (s *proxiedGitHubSource) DownloadReleaseAsset(ctx context.Context, rel *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	if rel == nil {
		return nil, selfupdate.ErrInvalidRelease
	}

	var downloadURL string
	if rel.AssetID == assetID {
		downloadURL = rel.AssetURL
	} else if rel.ValidationAssetID == assetID {
		downloadURL = rel.ValidationAssetURL
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("asset ID %d: %w", assetID, selfupdate.ErrAssetNotFound)
	}

	client := &http.Client{Transport: s.transport}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download via proxy failed: HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}

var (
	_ selfupdate.Source = (*fallbackSource)(nil)
	_ selfupdate.Source = (*proxiedGitHubSource)(nil)
)
