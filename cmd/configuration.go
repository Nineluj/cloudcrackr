package cmd

import (
	"cloudcrackr/cmd/utility"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	log "github.com/visionmedia/go-cli-log"
	"os"
	"reflect"
)

const (
	UserCreateConfigurationDeniedError = "user did not want to create config"
	FixedConfigCommandError            = "command doesn't work with custom config path"
)

// Config declared this way to force the presence of these values at runtime
type config struct {
	// Could extend this with "optional" fields
	Region          string `instr:"The AWS region to use"`
	ProfileName     string `instr:"The name of the profile to use (see ~/.aws/credentials)"`
	S3BucketName    string `instr:"The name of the S3 bucket to use."`
	ClusterName     string `instr:"The name of the cluster to use for ECS"`
	IAMRoleNamePath string `instr:"Prefix for IAM role names"`
}

var cfgFormat = config{}

// configureCmd represents the configure command
var configurationCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"config", "cfg", "conf", "c"},
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
	RunE:  configClean,
}

var configWhereCmd = &cobra.Command{
	Use:   "where",
	Short: "Shows the default configuration location",
	Args:  cobra.ExactArgs(0),
	RunE:  configWhere,
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
		return errors.New(UserCreateConfigurationDeniedError)
	}

	// Necessary evil of reflect to make the config logic more elegant
	v := reflect.TypeOf(cfgFormat)

	fmt.Println("The resources will be created by running init, please do use already existing" +
		" resource names")

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fmt.Println("> " + field.Tag.Get("instr"))
		input := utility.GetInput(fmt.Sprintf("%s", field.Name))
		viper.Set(field.Name, input)
	}

	return nil
}

func getMissingConf() bool {
	cfg := config{}
	//reflectConfElem := reflect.ValueOf(cfg).Elem()

	fixedMissingConf := false
	v := reflect.TypeOf(cfg)

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

func configWhere(c *cobra.Command, _ []string) error {
	if c.Flag("config").Value.String() != "" {
		return errors.New(FixedConfigCommandError)
	}

	fmt.Println(defaultCfgPath)
	return nil
}

func configClean(c *cobra.Command, _ []string) error {
	if c.Flag("config").Value.String() != "" {
		return errors.New(FixedConfigCommandError)
	}

	err := os.Remove(defaultCfgPath)
	return err
}
