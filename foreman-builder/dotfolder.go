package foremanbuilder

import (
	"errors"
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
		if line != name {
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

func GetAllLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil

}
