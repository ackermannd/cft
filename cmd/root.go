// Copyright © 2016 Daniel Ackermann <ackermann.d@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ackermannd/clifmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var composeFile string
var force bool

// RootCmd is the main command that holds all subcommands
var RootCmd = &cobra.Command{
	Use:   "cft",
	Short: "compose file tool",
	Long:  `Tool for modifying docker-compose files via CLI and some additional neat automations`,
}

// Execute calls RootCmdExecute and prints errors if some occurs
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&composeFile, "compose-file", "c", os.Getenv("CFT_COMPOSE"), "docker-compose file to change, if none set $CFT_COMPOSE will be used")
	RootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Skips security confirmation prompts")
	if composeFile == "" {
		clifmt.Settings.Color = clifmt.Red
		clifmt.Println("Neither -c flag nor CFT_COMPOSE ENV given, trying to use docker-compose.yml in current directoy")
		clifmt.Settings.Color = ""
		if _, err := os.Stat("./docker-compose.yml"); err == nil {
			composeFile = "./docker-compose.yml"
		}
	}
}

func initConfig() {
	viper.SetConfigName(".cft")  // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Confirm will ask the given string as yes/no confirmation on the CLI
func confirm(q string) bool {
	for {
		fmt.Println(q)
		reader := bufio.NewReader(os.Stdin)
		conf, _ := reader.ReadString('\n')
		conf = strings.ToLower(strings.TrimSpace(conf))
		if conf == "y" || conf == "yes" {
			return true
		} else if conf == "n" || conf == "no" {
			return false
		}
	}
}

func extractService(sv, origData string) string {
	svReg := regexp.MustCompilePOSIX(".*" + sv + ":")
	found := svReg.FindString(origData)
	whitespace := strings.Split(found, sv+":")[0]

	services := regexp.MustCompilePOSIX("^"+whitespace+"[a-zA-Z-]*:( *|\t*)?").FindAllString(origData, -1)

	nxtService := ""
	if len(services) > 1 {
		for key, val := range services {
			if strings.Contains(val, whitespace+sv+":") {
				if key+1 < len(services) {
					nxtService = services[key+1]
				}
				break
			}
		}
	}

	allReg := regexp.MustCompile(whitespace + sv + ":\\s([\\w\\s\\W]*)" + nxtService)
	found = allReg.FindString(origData)

	replReg := regexp.MustCompile("(" + whitespace + sv + ":\\s|" + nxtService + ")")
	return replReg.ReplaceAllString(found, "")
}
