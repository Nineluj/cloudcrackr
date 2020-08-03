package cmd

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	log "github.com/visionmedia/go-cli-log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var globalCfg config
var awsSession *session.Session

var configFileName = ".cloudcrackr"
var cfgFile string
var defaultCfgPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudcrackr",
	Short: "Facilitate password cracking using AWS",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: preRun,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

func preRun(cmd *cobra.Command, args []string) error {
	err := setupAwsSession()
	if err != nil {
		return err
	}

	err = unmarshalConfig()
	if err != nil {
		return err
	}

	return nil
}

func unmarshalConfig() error {
	return viper.Unmarshal(&globalCfg)
}

func setupAwsSession() error {
	var err error
	awsSession, err = session.NewSessionWithOptions(session.Options{
		Profile: viper.GetString("ProfileName"),
		Config:  aws.Config{Region: aws.String(viper.GetString("Region"))},
	})

	if err != nil {
		return err
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err == nil {
		os.Exit(0)
	} else {
		log.Error(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is $HOME/.cloudcrackr.yaml)",
	)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	changesMade := false

	cfgPath := ""

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		cfgPath = cfgFile
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cloudcrackr" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(configFileName)

		cfgPath = fmt.Sprintf("%s/%s.yaml", home, configFileName)
		defaultCfgPath = cfgPath
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; need to create one
			fmt.Println("Config was not found. " +
				"If you do not want to create it, rerun the program with --config " +
				"to use a config file in a custom location")
			err = generateConfig()

			if err != nil {
				log.Error(err)
				os.Exit(-1)
			} else {
				changesMade = true
			}
		} else {
			// Config file was found but another error was produced
			// TODO: handle this case
			fmt.Println("Cannot handle error", err.Error())
			os.Exit(-1)
		}
	}

	// Config was found or created
	// Verify that we have the necessary information

	// Missing information
	if getMissingConf() {
		changesMade = true
	}
	// changesMade = true

	// Write changes made
	if changesMade {
		_, err := os.Stat(cfgPath)
		if !os.IsExist(err) {
			if _, err := os.Create(cfgPath); err != nil { // perm 0666
				log.Error(err)
			}
		}
		if err := viper.WriteConfig(); err != nil {
			// handle failed write
			log.Error(err)
		}
	}
}
