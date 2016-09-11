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
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/aryann/difflib"
	"github.com/spf13/cobra"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch <service name> [<service name> <service name> ...]",
	Short: "Switches comments on image and build commands",
	Long:  `If for a given service, build commands are commented out, these comments will be removed while image will be commented out and vice versa`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if composeFile == "" {
			return errors.New("No docker-compose file set, either set CFT_COMPOSE environment variable or supply via flag")
		}
		if len(args) == 0 {
			return errors.New("No service name given")
		}
		cf, err := os.Open(composeFile)
		if err != nil {
			return err
		}
		defer cf.Close()

		cfd, _ := ioutil.ReadAll(cf)
		origData := string(cfd)

		replaceData := origData

		for _, sv := range args {
			svReg := regexp.MustCompilePOSIX(".*" + sv + ":")
			found := svReg.FindString(origData)
			whitespace := strings.Split(found, sv+":")[0]
			
			services := regexp.MustCompilePOSIX("^"+whitespace+"[a-zA-Z-]*:( *|\t*)?").FindAllString(origData, -1)

			nxtService := ""
			if len(services) > 1 {
				for key, val := range services {
					if strings.Contains(val, whitespace+sv+":") {
						if (key+1 < len(services)) {
							nxtService = services[key+1]
						}
						break
					}
				}
			}

			allReg := regexp.MustCompile(whitespace + sv + ":\\s([\\w\\s\\W]*)" + nxtService)
			found = allReg.FindString(origData)

			replReg := regexp.MustCompile("(" + whitespace + sv + ":\\s|" + nxtService + ")")
			toReplace := replReg.ReplaceAllString(found, "")

			checkReg := regexp.MustCompile("#(\\s)*image")
			found = checkReg.FindString(toReplace)

			replaced := ""
			//we need a workaround placeholder because replaceallString
			//somehow can handle to replace $1 when it contains whitespaces?
			//WTF
			workaround := " __workaround__unique__blablabla__ "
			if found == "" {
				//switch from image to build
				replReg = regexp.MustCompilePOSIX("^( *|\t*)image:")
				replaced = replReg.ReplaceAllString(toReplace, "#$1"+workaround+"image:")

				replReg = regexp.MustCompilePOSIX("#( *|\t*)?build:")
				replaced = replReg.ReplaceAllString(replaced, "$1"+workaround+"build:")

				replReg = regexp.MustCompilePOSIX("#( *|\t*)?volumes:")
				replaced = replReg.ReplaceAllString(replaced, "$1"+workaround+"volumes:")

				replReg = regexp.MustCompilePOSIX("#( *|\t*)?-( *.*/.*)")
				replaced = replReg.ReplaceAllString(replaced, "$1"+workaround+"-$2")

			} else {
				// switch to build to image
				replReg = regexp.MustCompilePOSIX("#( *|\t*)?image:")
				replaced = replReg.ReplaceAllString(toReplace, "$1"+workaround+"image:")

				replReg = regexp.MustCompilePOSIX("^( *|\t*)?build:")
				replaced = replReg.ReplaceAllString(replaced, "#$1"+workaround+"build:")

				replReg = regexp.MustCompilePOSIX("^( *|\t*)?volumes:")
				replaced = replReg.ReplaceAllString(replaced, "#$1"+workaround+"volumes:")

				replReg = regexp.MustCompilePOSIX("^( *|\t*)?-( *.*/.*)")
				replaced = replReg.ReplaceAllString(replaced, "#$1"+workaround+"-$2")
			}
			replReg = regexp.MustCompile(workaround)
			replaced = replReg.ReplaceAllString(replaced, "")

			toWrite := strings.Replace(replaceData, toReplace, replaced, 1)
			replaceData = toWrite
		}

		diff := difflib.Diff(strings.Split(origData, "\n"), strings.Split(replaceData, "\n"))
		fmt.Println("Changes: ")
		for _, val := range diff {
			if val.Delta.String() != " " {
				fmt.Println(val)
			}
		}

		out := []byte(replaceData)
		if err := ioutil.WriteFile(composeFile, out, 0666); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(switchCmd)
}
