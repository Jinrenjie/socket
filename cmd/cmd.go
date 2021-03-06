package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/Jinrenjie/socket/api"
	"github.com/Jinrenjie/socket/database"
	"github.com/Jinrenjie/socket/internal/debug"
	"github.com/Jinrenjie/socket/internal/im"
	"github.com/Jinrenjie/socket/internal/logs"
	"github.com/Jinrenjie/socket/internal/service"
	"github.com/garyburd/redigo/redis"
	"github.com/naoina/denco"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "socket",
	Short: "Web Socket service",
	Long:  "Instant Messaging service based on Golang implementation",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display this service version",
	Long:  "Display this service version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version 1.0.1")
	},
}

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(versionCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().StringVarP(&confFile, "config", "c", ".socket.yaml", "defien runtime config file path")
	startCmd.Flags().BoolP("web-ui", "u", false, "Enable managerment web ui")
	startCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "Start the service as a daemon")
	startCmd.Flags().BoolVarP(&ssl, "ssl", "s", false, "Start service on ssl connection")
}

var (
	ssl      bool
	daemon   bool
	confFile string
)

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
			command := exec.Command(os.Args[0], "start")
			if err := command.Start(); err != nil {
				fmt.Println("Service startup failed.", err)
				os.Exit(0)
			}
			fmt.Println("Service started successfully.")
			runtime := viper.GetStringMapString("runtime")
			pidPath := runtime["pid"]
			if err := ioutil.WriteFile(pidPath, []byte(fmt.Sprintf("%d", command.Process.Pid)), 0666); err != nil {
				panic(err)
			}
			daemon = false
			os.Exit(0)
		}
		startService()
	},
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if confFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(confFile)
	} else {
		// Find home directory.
		path, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".socket" (without extension).
		viper.AddConfigPath(path)
		viper.SetConfigName(".socket")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("ERR:", err)
		os.Exit(2)
	}
}

// Start Service
func startService() {
	bootstrap()
	apiConf := viper.GetStringMapString("api")
	apiPrefix := apiConf["prefix"]
	web := viper.GetStringMapString("web")
	socket := viper.GetStringMapString("socket")
	mux := denco.NewMux()
	handler, err := mux.Build([]denco.Handler{
		mux.GET(socket["prefix"], im.Handle),
		mux.GET(apiPrefix+"/check/:id", api.CheckOnline),
		mux.GET("/debug/uid/all", api.Connections),
		mux.POST(apiPrefix+"/deliver/:id", api.Deliver),
		mux.GET("/health", api.Health),
	})
	if err != nil {
		panic(err)
	}

	//http.HandleFunc(socket["prefix"], im.Handle)
	//http.HandleFunc(apiPrefix + "/check/:id", api.CheckOnline)
	//http.HandleFunc(apiPrefix + "/deliver/:id", api.Deliver)
	//http.HandleFunc(apiPrefix, im.Handle)
	//http.HandleFunc("/debug/uid/all", api.Connections)

	if viper.GetBool("debug") {
		go debug.StartDebug()
	}

	port, err := strconv.Atoi(web["port"])
	if err != nil {
		log.Println(err)
	}
	if viper.GetBool("consul.enable") {
		service.Registration(web["host"], port, ssl)
	}

	//http.Handle("/", http.FileServer(http.Dir("web")))
	socketAddr := fmt.Sprintf("%v:%v", web["host"], web["port"])
	fmt.Printf("Web service listen on %v \n", socketAddr)
	if ssl {
		sslConf := viper.GetStringMapString("ssl")
		if err := http.ListenAndServeTLS(socketAddr, sslConf["cert"], sslConf["key"], handler); err != nil {
			fmt.Println(err)
		}
	} else {
		if err := http.ListenAndServe(socketAddr, handler); err != nil {
			fmt.Println(err)
		}
	}
}

func bootstrap() {
	redisConf := viper.GetStringMapString("redis")
	redisAddr := fmt.Sprintf("%v:%v", redisConf["host"], redisConf["port"])
	db, err := strconv.Atoi(redisConf["db"])
	if err != nil {
		db = 1
	}
	database.CreateRedisPool(redisAddr, redisConf["pass"], db)
	// clear()
}

func clear() {
	connection := database.Pool.Get()
	defer func() {
		if err := connection.Close(); err != nil {
			logs.OutPut("ERROR", "REDIS-CLEAR", err.Error())
		}
	}()
	if _, err := redis.String(connection.Do("FLUSHDB")); err != nil {
		panic(err)
	}
}

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}
