package utility

import (
	"bufio"
	"fmt"
	"os"
)

var reader = bufio.NewReader(os.Stdin)

func GetInput(prompt string) string {
	fmt.Print(prompt + ": ")
	input, err := reader.ReadString('\n')
	text := ""

	if err == nil {
		text = input[:len(input)-1]
	} else {
		os.Exit(1)
	}

	return text
}

func GetBoolean(prompt string) bool {
	return GetInput(fmt.Sprintf("%s [y/n] ", prompt)) == "y"
}
