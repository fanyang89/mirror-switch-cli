package switcher

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type RepoType string

const (
	Rocky     RepoType = "rocky"
	EPEL      RepoType = "epel"
	Fedora    RepoType = "fedora"
	RPMFusion RepoType = "rpmfusion"
)

type Config struct {
	RepoDir         string
	MirrorHost      string
	DisableOpenH264 bool
}

func Switch(repoType RepoType, config Config) ([]string, error) {
	var pattern string
	switch repoType {
	case Rocky:
		pattern = "rocky*.repo"
	case EPEL:
		pattern = "epel*.repo"
	case Fedora:
		pattern = "fedora*.repo"
	case RPMFusion:
		pattern = "rpmfusion*.repo"
	default:
		return nil, fmt.Errorf("unsupported repo type: %s", repoType)
	}

	files, err := filepath.Glob(filepath.Join(config.RepoDir, pattern))
	if err != nil {
		return nil, err
	}

	var switched []string
	for _, file := range files {
		if repoType == Fedora && isFedoraOpenH264RepoFile(file) && !config.DisableOpenH264 {
			continue
		}

		if err := switchFile(file, repoType, config); err != nil {
			return switched, fmt.Errorf("failed to switch %s: %w", file, err)
		}
		switched = append(switched, file)
	}

	return switched, nil
}

func switchFile(filePath string, repoType RepoType, config Config) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines:    true,
		AllowShadows:               true,
		AllowPythonMultilineValues: true,
	}, filePath)
	if err != nil {
		return err
	}

	isOpenH264Repo := repoType == Fedora && isFedoraOpenH264RepoFile(filePath)

	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}

		if isOpenH264Repo {
			if config.DisableOpenH264 {
				section.Key("enabled").SetValue("0")
			}
			continue
		}

		if repoType == Fedora && updateFedoraBaseURL(section, config.MirrorHost) {
			continue
		}

		// Disable mirrorlist/metalink
		disableKey(section, "mirrorlist")
		disableKey(section, "metalink")

		// Try to find baseurl, or even a commented out one
		var baseURL string
		if section.HasKey("baseurl") {
			baseURL = section.Key("baseurl").Value()
		} else if section.HasKey("#baseurl") {
			baseURL = section.Key("#baseurl").Value()
			section.DeleteKey("#baseurl")
		}

		if baseURL != "" {
			newURL := normalizeURL(baseURL, config.MirrorHost, repoType)
			section.Key("baseurl").SetValue(newURL)
		}
	}

	return cfg.SaveTo(filePath)
}

func disableKey(section *ini.Section, keyName string) {
	if !section.HasKey(keyName) {
		return
	}

	key := section.Key(keyName)
	commentedKeyName := "#" + keyName
	if !section.HasKey(commentedKeyName) {
		_, _ = section.NewKey(commentedKeyName, key.Value())
	}
	section.DeleteKey(keyName)
}

func isFedoraOpenH264RepoFile(filePath string) bool {
	return strings.HasPrefix(filepath.Base(filePath), "fedora-cisco-openh264")
}

func updateFedoraBaseURL(section *ini.Section, mirrorHost string) bool {
	repoID := fedoraRepoID(section)
	subPath, ok := fedoraRepoSubPath(repoID)
	if !ok {
		subPath, ok = fedoraSectionSubPath(section.Name())
	}
	if !ok {
		return false
	}

	disableKey(section, "mirrorlist")
	disableKey(section, "metalink")
	if section.HasKey("#baseurl") {
		section.DeleteKey("#baseurl")
	}

	baseURL := (&url.URL{
		Scheme: "https",
		Host:   mirrorHost,
		Path:   path.Join("/fedora", subPath),
	}).String()
	section.Key("baseurl").SetValue(baseURL)
	return true
}

func fedoraRepoID(section *ini.Section) string {
	for _, keyName := range []string{"metalink", "#metalink"} {
		if !section.HasKey(keyName) {
			continue
		}

		value := section.Key(keyName).Value()
		parsed, err := url.Parse(value)
		if err != nil {
			continue
		}

		repoID := parsed.Query().Get("repo")
		if repoID != "" {
			return repoID
		}
	}

	return ""
}

func fedoraRepoSubPath(repoID string) (string, bool) {
	switch repoID {
	case "fedora-$releasever":
		return "releases/$releasever/Everything/$basearch/os", true
	case "fedora-debug-$releasever":
		return "releases/$releasever/Everything/$basearch/debug/tree", true
	case "fedora-source-$releasever":
		return "releases/$releasever/Everything/source/tree", true
	case "updates-released-f$releasever":
		return "updates/$releasever/Everything/$basearch", true
	case "updates-released-debug-f$releasever":
		return "updates/$releasever/Everything/$basearch/debug", true
	case "updates-released-source-f$releasever":
		return "updates/$releasever/Everything/source/tree", true
	case "updates-testing-f$releasever":
		return "updates/testing/$releasever/Everything/$basearch", true
	case "updates-testing-debug-f$releasever":
		return "updates/testing/$releasever/Everything/$basearch/debug", true
	case "updates-testing-source-f$releasever":
		return "updates/testing/$releasever/Everything/source/tree", true
	case "rawhide":
		return "development/rawhide/Everything/$basearch/os", true
	case "rawhide-debug":
		return "development/rawhide/Everything/$basearch/debug/tree", true
	case "rawhide-source":
		return "development/rawhide/Everything/source/tree", true
	default:
		return "", false
	}
}

func fedoraSectionSubPath(sectionName string) (string, bool) {
	switch sectionName {
	case "fedora":
		return "releases/$releasever/Everything/$basearch/os", true
	case "fedora-debuginfo":
		return "releases/$releasever/Everything/$basearch/debug/tree", true
	case "fedora-source":
		return "releases/$releasever/Everything/source/tree", true
	case "updates":
		return "updates/$releasever/Everything/$basearch", true
	case "updates-debuginfo":
		return "updates/$releasever/Everything/$basearch/debug", true
	case "updates-source":
		return "updates/$releasever/Everything/source/tree", true
	case "updates-testing":
		return "updates/testing/$releasever/Everything/$basearch", true
	case "updates-testing-debuginfo":
		return "updates/testing/$releasever/Everything/$basearch/debug", true
	case "updates-testing-source":
		return "updates/testing/$releasever/Everything/source/tree", true
	case "rawhide":
		return "development/rawhide/Everything/$basearch/os", true
	case "rawhide-debuginfo":
		return "development/rawhide/Everything/$basearch/debug/tree", true
	case "rawhide-source":
		return "development/rawhide/Everything/source/tree", true
	default:
		return "", false
	}
}

func normalizeURL(originalURL string, mirrorHost string, repoType RepoType) string {
	u, err := url.Parse(originalURL)
	if err != nil || u.Host == "" {
		return originalURL
	}

	u.Scheme = "https"
	u.Host = mirrorHost

	path := u.Path
	// Strip known public prefixes and redundant segments
	path = strings.TrimPrefix(path, "/pub/rocky")
	path = strings.TrimPrefix(path, "/pub/epel")
	path = strings.TrimPrefix(path, "/pub/fedora/linux")
	path = strings.TrimPrefix(path, "/pub/rpmfusion")
	path = strings.TrimPrefix(path, "/$contentdir")

	// Special case: if it already starts with the repo name but has 'linux' in it, strip it
	// e.g., /fedora/linux/releases/... -> /fedora/releases/...
	if strings.HasPrefix(path, "/fedora/linux/") {
		path = "/fedora/" + strings.TrimPrefix(path, "/fedora/linux/")
	}

	// Ensure the path starts with the repo name
	repoPrefix := "/" + string(repoType)
	if !strings.HasPrefix(path, repoPrefix) {
		path = repoPrefix + "/" + strings.TrimPrefix(path, "/")
	}

	// Clean up any double slashes
	u.Path = strings.ReplaceAll(path, "//", "/")
	return u.String()
}
