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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ackermannd/clifmt"
	"github.com/spf13/cobra"
)

var branch string
var remoteOnly bool

// gitCoCmd represents the git-co command
var gitCoCmd = &cobra.Command{
	Use:   "git-co <service name> [<service name> <service name> ...]",
	Short: "Checkout specific branches for the given services",
	Long:  `Takes information from Buildpaths of the given services and checks out the given branch. If local changes are represent, they'll be stashed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if composeFile == "" {
			return errors.New("No docker-compose file set, either set CFT_COMPOSE environment variable or supply via flag")
		}
		if branch == "" {
			return errors.New("No branch name given")
		}

		cf, err := os.Open(composeFile)
		if err != nil {
			return err
		}
		defer cf.Close()

		if len(args) == 0 {
			if force == false {
				if !confirm("No service name given, this will iterate through all services and tries to check out the remote branch if it exists. Continue? [y/n]") {
					os.Exit(0)
				}
			}
			tmpComposeFolder := strings.Split(composeFile, "/")
			cmd := exec.Command("docker-compose", "config", "--services")
			cmd.Dir = "/" + strings.Join(tmpComposeFolder[1:len(tmpComposeFolder)-1], "/")
			tmp, _ := cmd.Output()
			args = strings.Split(string(tmp), "\n")
			args = args[0 : len(args)-1]
		}

		cfd, _ := ioutil.ReadAll(cf)
		origData := string(cfd)
		clifmt.Settings.Intendation = " "
		for _, sv := range args {
			service := extractService(sv, origData)

			checkReg := regexp.MustCompile("build:(.*)")
			folder := strings.TrimSpace(checkReg.ReplaceAllString(checkReg.FindString(service), "$1"))

			var stderr bytes.Buffer

			fmt.Println("Stashing changes in " + folder)
			cmd := exec.Command("git", "stash")
			cmd.Dir = folder
			cmd.Stderr = &stderr
			_, err := cmd.Output()
			if err != nil {
				return errors.New(err.Error() + ": " + stderr.String())
			}

			stderr.Reset()
			clifmt.Println("Checking if remote origin exists")
			cmd = exec.Command("git", "remote", "show", "origin")
			cmd.Dir = folder
			cmd.Stderr = &stderr
			_, err = cmd.Output()
			if err != nil && err.Error() != "exit status 128" {
				return errors.New(err.Error() + ": " + stderr.String())
			}
			if err != nil && err.Error() == "exit status 128" {
				if remoteOnly {
					clifmt.Println("No remote origin available")
					continue
				}
				clifmt.Println("No remote origin available, creating local branch")
				cmd = exec.Command("git", "checkout", "-B", branch)
				var stdout bytes.Buffer
				stderr.Reset()
				cmd.Dir = folder
				cmd.Stderr = &stderr
				cmd.Stdout = &stdout

				err = cmd.Run()
				if err != nil {
					return errors.New(err.Error() + ": " + stderr.String())
				}
				if stdout.String() != "" {
					clifmt.Println(strings.Replace(stdout.String(), "\n", "\n    ", -1))
				}

				if stderr.String() != "" {
					clifmt.Println(strings.Replace(stderr.String(), "\n", "\n    ", -1))
				}
				continue
			}

			stderr.Reset()
			clifmt.Println("Fetching remote")
			cmd = exec.Command("git", "fetch", "--all")
			cmd.Dir = folder
			cmd.Stderr = &stderr
			_, err = cmd.Output()
			if err != nil {
				return errors.New(err.Error() + ": " + stderr.String())
			}
			stderr.Reset()
			clifmt.Println("Checking if branch exists in remote")
			cmd = exec.Command("git", "ls-remote", "--heads", "--exit-code", "origin", branch)
			cmd.Dir = folder
			cmd.Stderr = &stderr
			_, err = cmd.Output()
			if err != nil {
				if err.Error() != "exit status 2" {
					return errors.New(err.Error() + ": " + stderr.String())
				}
				if remoteOnly {
					clifmt.Println("Branch not available on remote")
					continue
				}
				clifmt.Println("Branch not available on remote, switchting to local branch")
				cmd = exec.Command("git", "checkout", "-B", branch, "develop")
				var stdout bytes.Buffer
				stderr.Reset()
				cmd.Dir = folder
				cmd.Stderr = &stderr
				cmd.Stdout = &stdout

				err = cmd.Run()
				if err != nil {
					return errors.New(err.Error() + ": " + stderr.String())
				}
				if stdout.String() != "" {
					clifmt.Println(strings.Replace(stdout.String(), "\n", "\n    ", -1))
				}

				if stderr.String() != "" {
					clifmt.Println(strings.Replace(stderr.String(), "\n", "\n    ", -1))
				}
				continue
			}
			clifmt.Println("Checking out branch origin/" + branch)
			cmd = exec.Command("git", "checkout", "-B", branch, "--track", "origin/"+branch)
			var stdout bytes.Buffer
			stderr.Reset()
			cmd.Dir = folder
			cmd.Stderr = &stderr
			cmd.Stdout = &stdout

			err = cmd.Run()
			if err != nil {
				return errors.New(err.Error() + ": " + stderr.String())
			}
			if stdout.String() != "" {
				clifmt.Println(strings.Replace(stdout.String(), "\n", "\n    ", -1))
			}

			if stderr.String() != "" {
				clifmt.Println(strings.Replace(stderr.String(), "\n", "\n    ", -1))
			}
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(gitCoCmd)
	gitCoCmd.Flags().StringVarP(&branch, "branch", "b", "", "the branch which should be checked out from the remote origin")
	gitCoCmd.Flags().BoolVarP(&remoteOnly, "remoteOnly", "r", false, "when no service names are given, only check out given branch if it exists in remote origin ")
}
