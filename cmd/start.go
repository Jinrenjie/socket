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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"socket/database"
	"socket/internal/im"
)

var daemon bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Web Socket service",
	Long: "Start Web Socket service on this host",
	Run: func(cmd *cobra.Command, args []string) {
		if daemon {
			command := exec.Command("socket", "start")
			if err := command.Start(); err != nil {
				panic(err)
			}
			fmt.Printf("socket start, [PID] %d running...\n", command.Process.Pid)
			if err := ioutil.WriteFile("socket.lock", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0666); err != nil {
				panic(err)
			}
			daemon = false
			os.Exit(0)
		} else {

		}
		startService()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().BoolP("web-ui", "u", false, "Enable managerment web ui")
	startCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "Start the service as a daemon")
}

func startService()  {
	web := viper.GetStringMapString("web")
	api := viper.GetStringMapString("api")
	redis := viper.GetStringMapString("redis")
	socket := viper.GetStringMapString("socket")
	prefix := api["prefix"]

	if prefix == "" {
		prefix = "/api"
	}

	type Response struct {
		Status int `json:"status"`
		Message string `json:"message"`
		Data interface{}
	}

	http.HandleFunc(prefix, func(writer http.ResponseWriter, request *http.Request) {
		data := make([]string, 0)
		response, err := json.Marshal(&Response{
			Status: 200,
			Message: "Success",
			Data: data,
		})
		if err != nil {
			log.Fatal(err)
		}
		if _, err := writer.Write(response); err != nil {
			log.Fatal(err)
		}
	})

	redisAddr := fmt.Sprintf("%v:%v", redis["host"], redis["port"])
	database.CreateRedisPool(redisAddr, redis["pass"])
	http.HandleFunc(socket["prefix"], im.Handle)
	http.Handle("/", http.FileServer(http.Dir("web")))
	socketAddr := fmt.Sprintf("%v:%v", web["host"], web["port"])

	fmt.Printf("Web service listen on %v", socketAddr)
	if err := http.ListenAndServe(socketAddr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

