package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const githubLatestRelease = "https://api.github.com/repos/opa334/Dopamine/releases/latest"
const githubReleaseByTag = "https://api.github.com/repos/opa334/Dopamine/releases/tags/"

type release struct {
	Assets []asset `json:"assets"`
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func DownloadDopamine(explicitURL, outDir string) (string, error) {
	return DownloadDopamineVersion(explicitURL, outDir, "")
}

func DownloadDopamineVersion(explicitURL, outDir, version string) (string, error) {
	url := explicitURL
	if url == "" {
		resolved, err := resolveIPA(version)
		if err != nil {
			return "", err
		}
		url = resolved
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	name := filepath.Base(url)
	if name == "." || name == "/" || !strings.HasSuffix(strings.ToLower(name), ".ipa") {
		name = "Dopamine.ipa"
	}
	target := filepath.Join(outDir, name)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	file, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}
	return target, nil
}

func resolveIPA(version string) (string, error) {
	if tag := tagFromVersion(version); tag != "" {
		return resolveReleaseIPA(githubReleaseByTag + tag)
	}
	return resolveLatestIPA()
}

func resolveLatestIPA() (string, error) {
	return resolveReleaseIPA(githubLatestRelease)
}

func resolveReleaseIPA(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "jailbreakit")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("GitHub release lookup failed: %s", resp.Status)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	for _, asset := range rel.Assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), ".ipa") {
			return asset.URL, nil
		}
	}
	return "", fmt.Errorf("latest Dopamine release has no IPA asset")
}

func tagFromVersion(version string) string {
	normalized := strings.ToLower(strings.Join(strings.Fields(version), ""))
	normalized = strings.ReplaceAll(normalized, "beta", "b")
	switch normalized {
	case "2.5b3":
		return "2.5b3"
	case "2.5b2":
		return "2.5b2"
	case "2.5b1":
		return "2.5b1"
	}
	if normalized != "" && strings.ContainsAny(normalized, "0123456789") {
		return normalized
	}
	return ""
}
