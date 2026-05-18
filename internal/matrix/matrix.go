package matrix

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Tool    string
	Version string
	Status  string
	Source  string
}

func LookupIOS(version string) ([]Entry, string) {
	entries, err := fetchAppleWiki(version)
	if err == nil && len(entries) > 0 {
		return entries, sourceURL(version)
	}
	return fallback(version), "embedded fallback"
}

func fetchAppleWiki(version string) ([]Entry, error) {
	url := sourceURL(version)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "jailbreakit")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("Apple Wiki HTTP %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	text := normalizeHTML(string(body))
	if strings.Contains(strings.ToLower(text), "security check") {
		return nil, fmt.Errorf("Apple Wiki security challenge")
	}
	return parseText(version, text), nil
}

func parseText(version, text string) []Entry {
	idx := strings.Index(text, "## iOS")
	if idx >= 0 {
		text = text[idx:]
	}
	next := strings.Index(text, "## tvOS")
	if next >= 0 {
		text = text[:next]
	}

	versionPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(version) + `\b`)
	loc := versionPattern.FindStringIndex(text)
	if loc == nil {
		return nil
	}
	row := text[loc[0]:]
	nextVersion := regexp.MustCompile(`\n\s*\d+\.\d+(?:\.\d+)?\b`).FindStringIndex(row[1:])
	if nextVersion != nil {
		row = row[:nextVersion[0]+1]
	}

	var entries []Entry
	for _, tool := range []string{"palera1n", "Dopamine"} {
		entry, ok := parseTool(row, tool)
		if ok {
			entry.Source = sourceURL(version)
			entries = append(entries, entry)
		}
	}
	return entries
}

func parseTool(row, tool string) (Entry, bool) {
	idx := strings.Index(strings.ToLower(row), strings.ToLower(tool))
	if idx < 0 {
		return Entry{}, false
	}
	rest := row[idx+len(tool):]
	versionMatch := regexp.MustCompile(`\d+(?:\.\d+)*(?:\s+Beta\s+\d+)?(?:-[A-Za-z0-9.]+)?`).FindString(rest)
	if versionMatch == "" {
		return Entry{}, false
	}
	status := "Yes"
	if strings.Contains(rest, "Semi-Tethered") {
		status = "Semi-Tethered"
	}
	return Entry{Tool: tool, Version: strings.TrimSpace(versionMatch), Status: status}, true
}

func fallback(version string) []Entry {
	major := majorVersion(version)
	var entries []Entry
	switch major {
	case 15:
		entries = append(entries,
			Entry{Tool: "palera1n", Version: "2.2.1", Status: "Semi-Tethered", Source: "embedded fallback"},
			Entry{Tool: "Dopamine", Version: "2.5 Beta 3", Status: "Yes", Source: "embedded fallback"},
		)
	case 16:
		entries = append(entries,
			Entry{Tool: "palera1n", Version: "2.2.1", Status: "Semi-Tethered", Source: "embedded fallback"},
			Entry{Tool: "Dopamine", Version: "2.5 Beta 3", Status: "Yes", Source: "embedded fallback"},
		)
	}
	return entries
}

func sourceURL(version string) string {
	major := majorVersion(version)
	if major < 12 || major > 18 {
		major = 15
	}
	return fmt.Sprintf("https://theapplewiki.com/wiki/Jailbreak/%d.x#iOS", major)
}

func normalizeHTML(html string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	text := re.ReplaceAllString(html, "\n")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.Join(strings.Fields(line), " ")
	}
	return strings.Join(lines, "\n")
}

func majorVersion(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return -1
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	return n
}
