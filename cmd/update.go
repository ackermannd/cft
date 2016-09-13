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
	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/kardianos/osext"
	"github.com/spf13/cobra"
)

// gitCoCmd represents the git-co command
var uCmd = &cobra.Command{
	Use:   "update",
	Short: "updates if a newer version exists",
	Long:  `updates if a newer version exists`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if force == false {
			for {
				fmt.Println("Really update? [yes/no]")
				reader := bufio.NewReader(os.Stdin)
				conf, _ := reader.ReadString('\n')
				conf = strings.ToLower(strings.TrimSpace(conf))
				if conf == "y" || conf == "yes" {
					break
				} else if conf == "n" || conf == "no" {
					os.Exit(0)
				}
			}
		}

		resp, err := http.Get("https://raw.githubusercontent.com/ackermannd/docker-compose-file-tool/master/VERSION")
		if err != nil {
			return errors.New("Couldn't fetch current version number from server :(")
		}
		body, _ := ioutil.ReadAll(resp.Body)
		nv := strings.TrimSpace(string(body))
		resp.Body.Close()

		if nv == VERSION {
			fmt.Println("Already newest version installed!")
			os.Exit(0)
		}
		tarname := nv + "-version.tar.gz"
		out, err := os.Create("./" + tarname)
		defer func(fname string) {
			os.Remove(fname)
		}("./" + tarname)

		if err != nil {
			return errors.New("Couldn't create temporary file for download")
		}
		fmt.Println("Downloading newer Version " + nv)
		dl, _ := http.Get("https://github.com/ackermannd/docker-compose-file-tool/releases/download/" + nv + "/cft-darwin-amd64.tar.gz")
		defer dl.Body.Close()
		io.Copy(out, dl.Body)
		out.Close()

		//unzip
		tarf, err := os.OpenFile("./"+tarname, os.O_RDONLY, 0444)
		if err != nil {
			return err
		}
		defer tarf.Close()
		gr, err := gzip.NewReader(tarf)
		if err != nil {
			return err
		}
		defer gr.Close()
		tr := tar.NewReader(gr)
		fmt.Println("Unpacking...")
		path := ""
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				// end of tar archive
				break
			}
			if err != nil {
				return err
			}
			path = hdr.Name
			switch hdr.Typeflag {
			case tar.TypeReg:
				ow, err := os.Create(path + ".new")
				defer ow.Close()
				if err != nil {
					return (err)
				}
				if _, err := io.Copy(ow, tr); err != nil {
					return (err)
				}
				ow.Chmod(0555)
			}
		}
		exec, _ := osext.Executable()
		os.Rename(path+".new", exec)
		fmt.Println("newer version " + nv + " is now usable!")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(uCmd)
}
