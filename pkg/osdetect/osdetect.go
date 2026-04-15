package osdetect

import (
	"bufio"
	"os"
	"strings"
)

type OSInfo struct {
	ID        string
	VersionID string
}

var OSReleasePath = "/etc/os-release"

func Detect() (*OSInfo, error) {
	f, err := os.Open(OSReleasePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info := &OSInfo{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			info.ID = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			info.VersionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return info, nil
}
