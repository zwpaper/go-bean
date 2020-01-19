package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/zwpaper/go-bean/cmd/chopstick/importcmd"
	"github.com/zwpaper/go-bean/pkg/config"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chopstick",
	Short: "chopstick for beancount",
	Long:  ``,
}

func init() {
	// Debug
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return fmt.Sprintf("%s()", path.Base(f.Function)), fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
	})

	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "~/.config/chopstick.yaml",
		"config file")
	rootCmd.PersistentFlags().StringVarP(&beanPath, "set-bean", "b", "",
		"beancount file")

	rootCmd.AddCommand(importcmd.NewCommand(&c))

	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	cfgFile, err = homedir.Expand(cfgFile)
	if err != nil {
		fmt.Println("expand path file error, ", err)
	}
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		// not matter what, read config first
		viper.Unmarshal(&c)
	}
	if beanPath != "" {
		// user specified beanPath
		// try and record in config
		_, err := ioutil.ReadFile(beanPath)
		if err != nil {
			fmt.Printf("can not read beancount file, %v\n", err)
			os.Exit(1)
		}
		c.BeanFile = beanPath
		added, err := yaml.Marshal(c)
		if err != nil {
			fmt.Printf("can not marshal config file, will try next time...\n(%v)\n", err)
			return
		}
		err = ioutil.WriteFile(cfgFile, added, 0755)
		if err != nil {
			fmt.Printf("can not save config file, will try next time...\n(%v)\n", err)
		}
	}

	if c.BeanFile == "" {
		logrus.Error("must use --set-bean or update config for the first run")
		os.Exit(1)
	}
}

// Specifically, find will return the influxd run command if no sub-command
func find(args []string) *cobra.Command {
	cmd, _, err := rootCmd.Find(args)
	if err == nil && cmd == rootCmd {
		// Execute the run command if no sub-command is specified
		return importcmd.NewCommand(&c)
	}

	return rootCmd
}

func main() {
	// cmd := find(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var cfgFile string
var c config.Config

var beanPath string
