package cmd

import (
	"cloudcrackr/utility"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	log "github.com/visionmedia/go-cli-log"
	"reflect"
)

// configureCmd represents the configure command
var configurationCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"config", "conf", "c"},
	Short:   "Handle the configuration of the program",
}

var configShowCmd = &cobra.Command{
	Use:  "show",
	Args: cobra.ExactArgs(0),
	Run:  showConfig,
}

func init() {
	configurationCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configurationCmd)
}

func showConfig(_ *cobra.Command, _ []string) {
	for _, key := range viper.AllKeys() {
		log.Info(key, "%s", viper.Get(key))
	}
}

type Config struct {
	Region            string `instr:"The AWS region to use"`
	ProfileName       string `instr:"The name of the profile to use (see ~/.aws/credentials)"`
	PasswordTableName string `instr:"The name of the table to use for the password file list. Avoid existing table names.'"`
	HashTableName     string `instr:"The name of the table to use for the hash file list. Avoid existing table names.'"`
	JobTableName      string `instr:"The name of the table to use for active jobs. Avoid existing table names.'"`
}

func generateConfig() error {
	confirm := utility.GetBoolean("Do you want to create a config now?")

	if !confirm {
		return errors.New("user did not want to create config")
	}

	// Necessary evil of reflect to make the config logic more elegant
	newConfig := Config{}
	v := reflect.TypeOf(newConfig)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fmt.Println("> " + field.Tag.Get("instr"))
		input := utility.GetInput(fmt.Sprintf("%s", field.Name))
		viper.Set(field.Name, input)
	}

	return nil
}
