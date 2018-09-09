package cmd

import (
	"fmt"
	"os"

	cmd "github.com/adriamb/gotoma/commands"
	cfg "github.com/adriamb/gotoma/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile is the configuration file path.
	cfgFile string
	// verbose is the verbosity level used in logrus.
	verbose string
)

var RootCmd = &cobra.Command{
	Use:   "gotoma",
	Short: "EthBerlin tomahawk!",
	Long:  "EthBerlin tomahawk!",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Looping sync forever",
	Long:  "Looping sync forever",
	Run:   cmd.Serve,
}

// ExecuteCmd adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func ExecuteCmd() {

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)

	}
}

// init is called when the package loads and initializes cobra.
func init() {

	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	RootCmd.PersistentFlags().StringVar(&verbose, "verbose", "INFO", "verbose level")

	RootCmd.AddCommand(serveCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	cfg.Valid = false

	if logLevel, err := log.ParseLevel(verbose); err == nil {
		log.SetLevel(logLevel)
	} else {
		panic(err)
	}

	viper.SetConfigType("yaml")
	viper.SetConfigName("config") // name ofconfig file (without extension)
	viper.AddConfigPath(".")      // adding current directory as first search path
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.SetEnvPrefix("GOTOMA")  // so viper.AutomaticEnv will get matching envvars starting with O2M_
	viper.AutomaticEnv()          // read in environment variables that match

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Error(err)
	} else {
		log.WithField("file", viper.ConfigFileUsed()).Debug("Using config file")

		if err := viper.Unmarshal(&cfg.C); err != nil {
			log.Error(err)
		} else {
			cfg.Valid = true
		}
	}
}
