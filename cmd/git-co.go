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
	"errors"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var branch string

// gitCoCmd represents the git-co command
var gitCoCmd = &cobra.Command{
	Use:   "git-co <service name> [<service name> <service name> ...]",
	Short: "Checkout specific branches for the given services",
	Long:  `Takes information from Buildpaths of the given services and checks out the given branch. If local changes are represent, they'll be stashed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if composeFile == "" {
			return errors.New("No docker-compose file set, either set CFT_COMPOSE environment variable or supply via flag")
		}
		if len(args) == 0 {
			return errors.New("No service name given")
		}
		if branch == "" {
			return errors.New("No branch name given")
		}
		cf, err := os.Open(composeFile)
		if err != nil {
			return err
		}
		defer cf.Close()

		cfd, _ := ioutil.ReadAll(cf)
		origData := string(cfd)

		for _, sv := range args {
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
			service := replReg.ReplaceAllString(found, "")


			checkReg := regexp.MustCompile("build:(.*)")
			folder := strings.TrimSpace(checkReg.ReplaceAllString(checkReg.FindString(service), "$1"))

			var stderr bytes.Buffer

			fmt.Println("Stashing changes in " + folder)
			cmd := exec.Command("git", "stash")
			cmd.Dir = folder
			cmd.Stderr = &stderr
			output, err := cmd.Output()
			if err != nil {
				return errors.New(err.Error() + ": " + stderr.String())
			}

			fmt.Println("Fetching remote")
			cmd = exec.Command("git", "fetch", "--all")
			cmd.Dir = folder
			cmd.Stderr = &stderr
			output, err = cmd.Output()
			if err != nil {
				return errors.New(err.Error() + ": " + stderr.String())
			}

			fmt.Println("Checking out branch origin/" + branch)
			cmd = exec.Command("git", "checkout", "-B", branch, "--track", "origin/"+branch)
			cmd.Dir = folder
			cmd.Stderr = &stderr
			output, err = cmd.Output()
			if err != nil {
				return errors.New(err.Error() + ": " + stderr.String())
			}
			fmt.Println(string(output))
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(gitCoCmd)
	gitCoCmd.Flags().StringVarP(&branch, "branch", "b", "", "the branch which should be checked out from the remote origin")
}
