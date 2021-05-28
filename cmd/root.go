/*
Copyright © 2021 willbenica

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/pgetal/lf-cli/internal"
	"go.uber.org/zap"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// Variable to use for logging
	logger    *zap.Logger
	logConfig *zap.Config

	// Use this to enusre we always know where the script is running
	DirWhereScriptStarted string

	cfgFile string
	// BaseURL is used to form all requests
	BaseURL string
	// Token is used to provide authentication for users - created in the lf UI
	Token string
	// AccountID let's leadfeeder know which account the user would like to access
	AccountID string
	// Verbose increases the log level
	Verbose bool

	// The variables below are used in sub commands!

	// All determines if we should loop through to the last page automatically
	All bool
	// StartDate is the start the date to retrun leads - YYYY-MM-DD
	StartDate string
	// EndDate is the end the date to retrun leads - YYYY-MM-DD
	EndDate string
	// PageSize is the number of results to retrun per call (needs to be between 1-100)
	PageSize int
	// PageNumber is based off the number of results (default is 1)
	PageNumber int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lf-cli",
	Short: "Dump leadfeeder data to a local file",
	Long: `Get leadfeeder data from a specific API endpoint and push to a local file (JSON).
For ease of use create a config file under $HOME/.config/lf-cli/.lf-cli.yaml
or under $HOME/.lf-cli.yaml with the following
  account: "myAccountID"
  token:   "myApiToken"
	`,
	Version: "2021.01",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// cobra.CheckErr(rootCmd.Execute())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// These prvent the sorting fo the flags defined below, they should be sorted in the code
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to a config file (default is $HOME/.config/lf-cli/.lf-cli.yaml)")
	rootCmd.PersistentFlags().StringVarP(&BaseURL, "lf-url", "", "https://api.leadfeeder.com", "leadfeeder URL")
	rootCmd.PersistentFlags().StringVarP(&AccountID, "accountID", "", "", "Account for which data should be accessed")
	rootCmd.PersistentFlags().StringVarP(&Token, "token", "", "", "API token used to access lf")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Increases loglevel to DEBUG for trouble shooting.")

	cobra.OnInitialize(initConfig)

	DirWhereScriptStarted, _ = os.Getwd()

	// Initalize logging and apply loglevel, etc
	logger, logConfig = internal.InitLogger()
	rootLogger := zap.ReplaceGlobals(logger)
	defer rootLogger()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)
		// Optionally use $HOME/.config/lf-cli instead of just the home folder
		viper.AddConfigPath(home + "/.config/lf-cli")
		viper.AddConfigPath(home)
		viper.SetConfigName(".lf-cli")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		if flagNotSet(BaseURL) {
			BaseURL = viper.GetString("lf-url")
		}
		if flagNotSet(Token) {
			Token = viper.GetString("token")
		}
		if flagNotSet(AccountID) {
			AccountID = viper.GetString("account")
		}
	}
}

func flagNotSet(flag string) bool {
	return flag == ""
}
