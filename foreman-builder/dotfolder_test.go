package foremanbuilder_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
)

func TestDoesFileOrDirectoryExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Existing file
	filePath := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(filePath, []byte("hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := foremanbuilder.DoesFileOrDirectoryExist(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatalf("expected file to exist")
	}

	missingPath := filepath.Join(tmpDir, "missing.txt")
	exists, err = foremanbuilder.DoesFileOrDirectoryExist(missingPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("expected file NOT to exist")
	}
}
func TestAppendToFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "append.txt")

	err := foremanbuilder.AppendToFile(filePath, "line1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = foremanbuilder.AppendToFile(filePath, "line2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	expected := "line1\nline2\n"
	if string(content) != expected {
		t.Fatalf("expected %q, got %q", expected, string(content))
	}
}
func TestDeleteLineInFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "delete.txt")

	initial := "apple\nbanana\norange\n"
	err := os.WriteFile(filePath, []byte(initial), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = foremanbuilder.DeleteLineInFile(filePath, "banana")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	expected := "apple\norange\n"
	if string(content) != expected {
		t.Fatalf("expected %q, got %q", expected, string(content))
	}
}

func TestGetAllLinesInFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "delete.txt")
	lines := "container-1::orb\ncontainer2::lima\ncontainer-1-3::3182321ujd12"
	err := os.WriteFile(filePath, []byte(lines), 0644)
	if err != nil {
		t.Fatal(err)
	}
	// test no split
	noSplit, err := foremanbuilder.GetAllLines(filePath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(noSplit) != 3 {
		t.Errorf("Expected 3 got %d", len(noSplit))
	}
	if noSplit[0] != "container-1::orb" {
		t.Errorf("Expected container1-orb got %s", noSplit[0])
	}

	split, err := foremanbuilder.GetAllLines(filePath, "::")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(split) != 3 {
		t.Errorf("Expected 3 got %d", len(split))
	}
	fmt.Printf("string 1: %s", split[0])
	if split[0] != "container-1" {
		t.Errorf("Expected container1-orb got %s", split[0])
	}
	if split[2] != "container-1-3" {
		t.Errorf("Expected container-1-3 got %s", split[2])
	}

}
