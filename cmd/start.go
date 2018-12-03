package cmd

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"socket/api"
	"socket/database"
	"socket/internal/im"
)

var daemon bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Web Socket service",
	Long:  "Start Web Socket service on this host",
	Run: func(cmd *cobra.Command, args []string) {
		if daemon {
			exist := isExist("socket.lock")
			if exist {
				fmt.Println("Service is started.")
				os.Exit(0)
			}
			command := exec.Command("socket", "start")
			if err := command.Start(); err != nil {
				fmt.Println("Service startup failed.")
				os.Exit(0)
			}
			fmt.Println("Service started successfully.")
			if err := ioutil.WriteFile("socket.lock", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0666); err != nil {
				panic(err)
			}
			daemon = false
			os.Exit(0)
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

func startService() {
	bootstrap()
	api.Register()
	web := viper.GetStringMapString("web")
	socket := viper.GetStringMapString("socket")
	http.HandleFunc(socket["prefix"], im.Handle)
	http.Handle("/", http.FileServer(http.Dir("web")))
	socketAddr := fmt.Sprintf("%v:%v", web["host"], web["port"])

	fmt.Printf("Web service listen on %v \n", socketAddr)
	if err := http.ListenAndServe(socketAddr, nil); err != nil {
		fmt.Println(err)
	}
}

func bootstrap() {
	redisConf := viper.GetStringMapString("redis")
	redisAddr := fmt.Sprintf("%v:%v", redisConf["host"], redisConf["port"])
	database.CreateRedisPool(redisAddr, redisConf["pass"])
	clear()
}

func clear() {
	connection := database.Pool.Get()
	if _, err := redis.String(connection.Do("FLUSHDB")); err != nil {
		panic(err)
	}
	if err := connection.Close(); err != nil {
		fmt.Println(err)
	}
}

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}
