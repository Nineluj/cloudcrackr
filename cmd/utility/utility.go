package utility

import (
	"bufio"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	log "github.com/visionmedia/go-cli-log"
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

func Pluralize(word string, n int) string {
	if n == 1 {
		return word
	}

	return word + "s"
}

func IgnoreAWSError(err error, ignoreErr string) error {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ignoreErr:
				log.Info("Ignoring", ignoreErr)
				return nil
			default:
				return err
			}
		} else {
			return err
		}
	}
	return err
}
