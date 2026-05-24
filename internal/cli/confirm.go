package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// confirmAction prompts the user with a yes/no question and returns true
// if they confirm. Defaults to "no" on empty input. Respects --yes flag
// to skip the prompt in non-interactive or scripted contexts.
func confirmAction(prompt string) bool {
	if autoYes {
		return true
	}

	fmt.Printf("%s [y/N]: ", prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
