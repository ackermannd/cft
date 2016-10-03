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

var tag string

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:   "tag <image pattern> [<image pattern> <image pattern>...]",
	Short: "Changes tags on images in docker-compose files",
	Long:  `Changes tags of images a docker-compose file. `,
	RunE: func(cmd *cobra.Command, args []string) error {
		if composeFile == "" {
			return errors.New("No docker-compose file set, either set CFT_COMPOSE environment variable or supply via flag")
		}

		if tag == "" && len(args) == 0 && force == false {
			if !Confirm("No tag nor image pattern given, really remove all tags from all images? [y/n]") {
				os.Exit(0)
			}
		}
		cf, err := os.Open(composeFile)
		if err != nil {
			return err
		}
		defer cf.Close()

		cfd, _ := ioutil.ReadAll(cf)
		res := string(cfd)
		if len(args) == 0 {
			re := regexp.MustCompilePOSIX("(image:[^:]*[^:]*)(:.*)?")
			rp := "$1"
			if tag != "" {
				rp = rp + ":" + tag
			}
			res = re.ReplaceAllString(res, rp)
		}
		for _, val := range args {
			re := regexp.MustCompilePOSIX("(image:[^:]*" + val + "[^:]*)(:.*)?")
			rp := "$1"
			if tag != "" {
				rp = rp + ":" + tag
			}
			res = re.ReplaceAllString(res, rp)
		}

		diff := difflib.Diff(strings.Split(string(cfd), "\n"), strings.Split(string(res), "\n"))
		fmt.Println("Changes: ")
		for _, val := range diff {
			if val.Delta.String() != " " {
				fmt.Println(val)
			}
		}

		out := []byte(res)
		if err := ioutil.WriteFile(composeFile, out, 0666); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(tagCmd)
	tagCmd.Flags().StringVarP(&tag, "tag", "t", "", "set this tag for the image(s), if no tag is set, existing tags will be removed")
}
