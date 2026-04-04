package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	GithubRepo = "welcomehaichao/Cloud189CLI"
	GithubAPI  = "https://api.github.com"
	GithubRaw  = "https://raw.githubusercontent.com"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", GithubAPI, GithubRepo)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release json: %w", err)
	}

	return &release, nil
}

func FindAsset(release *GitHubRelease, os, arch string) (string, string, error) {
	var expectedName string

	switch os {
	case "windows":
		expectedName = fmt.Sprintf("cloud189-windows-%s.zip", arch)
	case "linux":
		expectedName = fmt.Sprintf("cloud189-linux-%s.tar.gz", arch)
	case "darwin":
		expectedName = fmt.Sprintf("cloud189-darwin-%s.tar.gz", arch)
	default:
		return "", "", fmt.Errorf("unsupported os: %s", os)
	}

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset.Name, asset.URL, nil
		}
	}

	return "", "", fmt.Errorf("asset not found: %s", expectedName)
}
