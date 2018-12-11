// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Web Socket service",
	Long:  `Stop Web Socket service.`,
	Run: func(cmd *cobra.Command, args []string) {
		runtime := viper.GetStringMapString("runtime")
		path := runtime["pid"]
		origin, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("ERR read pid file", err)
		}
		pid, err := strconv.Atoi(string(origin))
		if err != nil {
			log.Fatal(err)
		}

		if err := syscall.Kill(int(pid), syscall.SIGKILL); err != nil {
			log.Fatal("ERR stop service error", err)
		}
		if err := os.Remove(path); err != nil {
			log.Fatal("ERR remove pid file", err)
		}
		fmt.Println("The service has stopped.")
	},
}

func byte2Int(data []byte) int {
	var ret = 0
	var count = len(data)
	var i uint = 0
	for i = 0; i < uint(count); i++ {
		ret = ret | (int(data[i]) << (i * 8))
	}
	return ret
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
