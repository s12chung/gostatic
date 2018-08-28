package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func promptStdIn(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	s := strings.TrimSuffix(strings.TrimSpace(prompt), ":") + ": "
	fmt.Print(s)
	return reader.ReadString('\n')
}
