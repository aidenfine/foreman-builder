package foremanbuilder

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func DoesFileOrDirectoryExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func AppendToFile(path, line string) error {
	f, err := os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(line + "\n")
	return err
}
func DeleteLineInFile(path, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")

	var kept []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Split(line, "::")[0] != name {
			kept = append(kept, line)
		}
	}

	newData := strings.Join(kept, "\n") + "\n"

	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, []byte(newData), 0644); err != nil {
		return err
	}

	return os.Rename(tempFile, path)
}
func GetAllLines(path, splitBy string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	if splitBy == "" {
		return lines, nil
	}

	var splitLines []string
	for _, line := range lines {
		parts := strings.SplitN(line, splitBy, 2)
		splitLines = append(splitLines, parts[0])
	}

	return splitLines, nil
}

// Can also be used to check if line exists in file, due to return error "not found" if line does not exist
// using containerType == "" will just check for the containerName itself and not care what `-<container-type>`
// example: We want to find a container named container1
func GetLineInFile(path, line, containerType string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var lines []string
	for s := range strings.SplitSeq(string(data), "\n") {
		lines = append(lines, s)
	}
	// if containerType == "" we can just ignore container postfix
	for _, fileLine := range lines {
		if strings.TrimSpace(fileLine) == "" {
			continue
		}
		filePrefix := strings.SplitN(fileLine, "::", 2)[0]
		searchPrefix := strings.SplitN(line, "::", 2)[0]
		if containerType != "" {
			containerMatch := fmt.Sprintf("%s::%s", searchPrefix, containerType)
			if containerMatch == fileLine {
				return fileLine, nil
			}
		}
		if filePrefix == searchPrefix {
			return fileLine, nil
		}
	}
	return "", errors.New("not_found")
}
