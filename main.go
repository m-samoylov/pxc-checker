package main

import (
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/labstack/gommon/log"
	"github.com/namsral/flag"
	"github.com/valyala/fasthttp"
	"os"
	"time"
)

type NodeStatus struct {
	WSRepStatus   int
	RWEnabled     bool
	NodeAvailable bool
	Timestamp     int64
}

type Config struct {
	WebListen        string
	WebReadTimeout   int
	WebWriteTimeout  int
	CheckROEnabled   bool
	CheckInterval    int
	CheckFailTimeout int64
	CheckForceEnable bool
	MysqlHost        string
	MysqlPort        int
	MysqlUser        string
	MysqlPass        string
	MysqlTimeout     int
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	status  = &NodeStatus{}
	config  *Config
)

func main() {
	config = parseFlags()

	go checker(status)

	router := getRouter()
	server := &fasthttp.Server{
		Handler:          router.Handler,
		DisableKeepalive: true,
		Concurrency:      2048,
		ReadTimeout:      time.Duration(config.WebReadTimeout) * time.Millisecond,
		WriteTimeout:     time.Duration(config.WebWriteTimeout) * time.Millisecond,
	}

	log.Printf("Server starting on %s", config.WebListen)
	if err := server.ListenAndServe(config.WebListen); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func getRouter() *fasthttprouter.Router {
	router := fasthttprouter.New()
	router.GET("/", checkerHandler)
	router.HEAD("/", checkerHandler)
	return router
}

func parseFlags() *Config {
	var versionFlag bool
	config := Config{}

	flag.StringVar(&config.WebListen, "WEB_LISTEN", ":9200", "Web server listening interface and port")
	flag.IntVar(&config.WebReadTimeout, "WEB_READ_TIMEOUT", 30000, "Web server request read timeout, ms")
	flag.IntVar(&config.WebWriteTimeout, "WEB_WRITE_TIMEOUT", 30000, "Web server request write timeout, ms")
	flag.BoolVar(&config.CheckROEnabled, "CHECK_RO_ENABLED", false, "Mark 'read_only' node as available")
	flag.BoolVar(&config.CheckForceEnable, "CHECK_FORCE_ENABLE", false, "Ignoring the status of the checks and always marking the node as available")
	flag.IntVar(&config.CheckInterval, "CHECK_INTERVAL", 500, "Mysql checks interval, ms")
	flag.Int64Var(&config.CheckFailTimeout, "CHECK_FAIL_TIMEOUT", 3000, "Mark the node inaccessible if for the specified time there were no successful checks, ms")
	flag.StringVar(&config.MysqlHost, "MYSQL_HOST", "127.0.0.1", "MySQL host addr")
	flag.IntVar(&config.MysqlPort, "MYSQL_PORT", 3306, "MySQL port")
	flag.StringVar(&config.MysqlUser, "MYSQL_USER", "pxc_checker", "MySQL username")
	flag.StringVar(&config.MysqlPass, "MYSQL_PASS", "", "MySQL password")

	flag.BoolVar(&versionFlag, "version", false, "Show program version")
	if versionFlag {
		fmt.Printf("Version: %s\nGit commit: %s\nBuilding date: %s\n", version, commit, date)
		os.Exit(0)
	}

	flag.Parse()
	return &config
}
