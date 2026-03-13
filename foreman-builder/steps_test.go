package foremanbuilder_test

import (
	"bytes"
	"strings"
	"testing"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
)

func TestStepsValidString(t *testing.T) {
	input := strings.NewReader("y\n")
	output := &bytes.Buffer{}

	result := foremanbuilder.ConfirmStep(input, output, "hello my name is x")

	if !result {
		t.Errorf("expected true, got false")
	}

	expected := "hello my name is x \n [y/n]\n"
	if output.String() != expected {
		t.Errorf("unexpected output: %q", output.String())
	}
}

func TestConfirmStepYes(t *testing.T) {
	input := strings.NewReader("y\n")
	output := &bytes.Buffer{}

	result := foremanbuilder.ConfirmStep(input, output, "continue?")

	if !result {
		t.Fatal("expected true")
	}
}
func TestConfirmStepNo(t *testing.T) {
	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}

	result := foremanbuilder.ConfirmStep(input, output, "continue?")

	if result {
		t.Fatal("expected false")
	}
}

func TestConfirmStepInvalid(t *testing.T) {
	input := strings.NewReader("123\ny\n")
	output := &bytes.Buffer{}

	foremanbuilder.ConfirmStep(input, output, "continue?")

	expected := "continue? \n [y/n]\ncontinue? \n [y/n]\n"
	if output.String() != expected {
		t.Errorf("unexpected output: %q", output.String())
	}
}
