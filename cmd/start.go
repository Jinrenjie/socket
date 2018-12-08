package cmd

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/naoina/denco"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"socket/api"
	"socket/database"
	"socket/internal/im"
	"socket/internal/service"
	"strconv"
)

var (
	ssl bool
	daemon bool
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

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(startCmd)

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
		mux.GET(apiPrefix + "/check/:id", api.CheckOnline),
		mux.GET("/debug/uid/all", api.Connections),
		mux.POST(apiPrefix + "/deliver/:id", api.Deliver),
		mux.GET("/health", api.Health),
	})
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(web["port"])
	go service.Registration(web["host"], port, ssl)

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
