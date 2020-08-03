package cmd

import (
	"cloudcrackr/utility"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	log "github.com/visionmedia/go-cli-log"
	"os"
	"reflect"
)

// Config declared this way to force the presence of these values at runtime
type config struct {
	Region       string `instr:"The AWS region to use"`
	ProfileName  string `instr:"The name of the profile to use (see ~/.aws/credentials)"`
	S3BucketName string `instr:"The name of the S3 bucket to use."`
	//JobTableName  string `instr:"The name of the table to use for active jobs. Avoid existing table names.'"`
}

var confFormat = config{}

// configureCmd represents the configure command
var configurationCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"config", "conf", "c"},
	Short:   "Handle the configuration of the program",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration settings",
	Args:  cobra.ExactArgs(0),
	Run:   showConfig,
}

var configCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes the current configuration file",
	Args:  cobra.ExactArgs(0),
	Run:   configClean,
}

var configWhereCmd = &cobra.Command{
	Use:   "where",
	Short: "Shows the default configuration location",
	Args:  cobra.ExactArgs(0),
	Run:   configWhere,
}

func init() {
	configurationCmd.AddCommand(configShowCmd)
	configurationCmd.AddCommand(configCleanCmd)
	configurationCmd.AddCommand(configWhereCmd)

	rootCmd.AddCommand(configurationCmd)
}

func showConfig(_ *cobra.Command, _ []string) {
	for _, key := range viper.AllKeys() {
		log.Info(key, "%s", viper.Get(key))
	}
}

func generateConfig() error {
	confirm := utility.GetBoolean("Do you want to create a config now?")

	if !confirm {
		return errors.New("user did not want to create config")
	}

	// Necessary evil of reflect to make the config logic more elegant
	v := reflect.TypeOf(confFormat)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fmt.Println("> " + field.Tag.Get("instr"))
		input := utility.GetInput(fmt.Sprintf("%s", field.Name))
		viper.Set(field.Name, input)
	}

	return nil
}

func getMissingConf() bool {
	conf := config{}
	//reflectConfElem := reflect.ValueOf(conf).Elem()

	fixedMissingConf := false
	v := reflect.TypeOf(conf)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if viper.Get(field.Name) == nil {
			if !fixedMissingConf {
				log.Info("Configuration", "%s", "Found missing configuration, please set these values")
			}

			fixedMissingConf = true

			fmt.Println("> " + field.Tag.Get("instr"))
			input := utility.GetInput(fmt.Sprintf("%s", field.Name))
			viper.Set(field.Name, input)

			// Set struct
			//reflectConfElem.Field(i).Set(reflect.ValueOf(input))
		}
	}

	return fixedMissingConf
}

func configWhere(c *cobra.Command, _ []string) {
	if c.Flag("config").Value.String() != "" {
		log.Error(errors.New("command doesn't work with custom config path"))
		return
	}

	fmt.Println(defaultCfgPath)
}

func configClean(c *cobra.Command, _ []string) {
	if c.Flag("config").Value.String() != "" {
		log.Error(errors.New("command doesn't work with custom config path"))
		return
	}

	err := os.Remove(defaultCfgPath)
	if err != nil {
		log.Error(err)
	}

}
