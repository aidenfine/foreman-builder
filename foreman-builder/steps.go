package foremanbuilder

import (
	"fmt"
	"io"
)

// Takes a string input and will ask a user that text with [y/n] appended to the end
// if user responses with "y" we return true.
// if user responses with anything other than "y" or "n" it loops back.
func ConfirmStep(in io.Reader, out io.Writer, text string) bool {
	var input string
	for input != "y" && input != "n" {
		fmt.Fprintln(out, text, "\n [y/n]")
		fmt.Fscanln(in, &input)
	}
	if input == "n" {
		return false
	}
	return true
}
