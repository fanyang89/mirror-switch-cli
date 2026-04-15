package switcher

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RepoType string

const (
	Rocky     RepoType = "rocky"
	EPEL      RepoType = "epel"
	Fedora    RepoType = "fedora"
	RPMFusion RepoType = "rpmfusion"
)

type Config struct {
	RepoDir string
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
		if err := switchFile(file, repoType); err != nil {
			return switched, fmt.Errorf("failed to switch %s: %w", file, err)
		}
		switched = append(switched, file)
	}

	return switched, nil
}

func switchFile(filePath string, repoType RepoType) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var result []string

	for _, line := range lines {
		newLine := applyReplacements(line, repoType)
		result = append(result, newLine)
	}

	return os.WriteFile(filePath, []byte(strings.Join(result, "\n")), 0644)
}

func applyReplacements(line string, repoType RepoType) string {
	line = strings.TrimRight(line, "\r")

	// 1. Comment out mirrorlist/metalink
	if strings.HasPrefix(strings.TrimSpace(line), "mirrorlist=") || strings.HasPrefix(strings.TrimSpace(line), "metalink=") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			return "#" + line
		}
	}

	// 2. Enable and replace baseurl
	switch repoType {
	case Rocky:
		re := regexp.MustCompile(`^#?baseurl=http://dl.rockylinux.org/\$contentdir`)
		if re.MatchString(line) {
			return "baseurl=https://mirrors.cernet.edu.cn/rocky"
		}
	case EPEL:
		re := regexp.MustCompile(`^#?baseurl=http://download.fedoraproject.org/pub/epel/`)
		if re.MatchString(line) {
			return "baseurl=https://mirrors.cernet.edu.cn/epel/"
		}
	case Fedora:
		re := regexp.MustCompile(`^#?baseurl=http://download.example/pub/fedora/linux`)
		if re.MatchString(line) {
			return "baseurl=https://mirrors.cernet.edu.cn/fedora"
		}
		// Some fedora repos use a different baseurl pattern or might already have a commented out one
		if strings.Contains(line, "download.fedoraproject.org/pub/fedora/linux") {
			line = strings.Replace(line, "https://mirrors.fedoraproject.org/metalink?repo=fedora-$releasever&arch=$basearch", "https://mirrors.cernet.edu.cn/fedora/releases/$releasever/Everything/$basearch/os/", 1)
			line = strings.Replace(line, "https://mirrors.fedoraproject.org/metalink?repo=updates-released-f$releasever&arch=$basearch", "https://mirrors.cernet.edu.cn/fedora/updates/$releasever/Everything/$basearch/", 1)
		}

	case RPMFusion:
		if strings.Contains(line, "download.viva-mp.com/pub/rpmfusion/") {
			line = strings.Replace(line, "http://download.viva-mp.com/pub/rpmfusion/", "https://mirrors.cernet.edu.cn/rpmfusion/", 1)
			line = strings.TrimPrefix(line, "#")
		}
		if strings.Contains(line, "download.rpmfusion.org/") {
			line = strings.Replace(line, "http://download.rpmfusion.org/", "https://mirrors.cernet.edu.cn/rpmfusion/", 1)
			line = strings.TrimPrefix(line, "#")
		}
	}

	return line
}
