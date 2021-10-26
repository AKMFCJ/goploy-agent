package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/joho/godotenv"
	"github.com/zhenorzz/goploy-agent/core"
	"github.com/zhenorzz/goploy-agent/route"
	"github.com/zhenorzz/goploy-agent/task"
	"github.com/zhenorzz/goploy-agent/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"
)

var (
	help bool
	v    bool
	s    string
)

const appVersion = "1.0.0"

func init() {
	flag.StringVar(&core.AssetDir, "asset-dir", "", "default: ./")
	flag.StringVar(&s, "s", "", "stop")
	flag.BoolVar(&help, "help", false, "list available subcommands and some concept guides")
	flag.BoolVar(&v, "version", false, "show goploy-agent version")
	// 改变默认的 Usage
	flag.Usage = usage
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	if v {
		println(appVersion)
		return
	}
	handleClientSignal()
	println(`
   ______            __           
  / ____/___  ____  / /___  __  __
 / / __/ __ \/ __ \/ / __ \/ / / /
/ /_/ / /_/ / /_/ / / /_/ / /_/ / 
\____/\____/ .___/_/\____/\__, /  
          /_/            /____/   ` + appVersion + "\n")
	install()
	_ = godotenv.Load(core.GetEnvFile())
	pid := strconv.Itoa(os.Getpid())
	_ = ioutil.WriteFile(path.Join(core.GetAssetDir(), "goploy-agent.pid"), []byte(pid), 0755)
	println("Start at " + time.Now().String())
	println("goploy-agent -h for more help")
	println("Current pid:   " + pid)
	println("Config Loaded: " + core.GetEnvFile())
	println("Log:           " + os.Getenv("LOG_PATH"))
	println("Listen:        " + os.Getenv("PORT"))
	println("Running...")
	core.CreateValidator()
	route.Init()
	task.Init()
	// server
	srv := http.Server{
		Addr: ":" + os.Getenv("PORT"),
	}
	go checkUpdate()
	core.Gwg.Add(1)
	go func() {
		defer core.Gwg.Done()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		println("Received the signal: " + (<-c).String())
		println("Server is trying to shutdown, wait for a minute")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			println("Server shutdown failed, err: %v\n", err)
		}
		println("Server shutdown gracefully")

		println("Task is trying to shutdown, wait for a minute")
		if err := task.Shutdown(ctx); err != nil {
			println("Task shutdown failed, err: %v\n", err)
		}
		println("Task shutdown gracefully")
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("ListenAndServe: ", err.Error())
	}
	_ = os.Remove(path.Join(core.GetAssetDir(), "goploy.pid"))
	println("Goroutine is trying to shutdown, wait for a minute")
	core.Gwg.Wait()
	println("Goroutine shutdown gracefully")
	println("Success")
	return
}

func install() {
	_, err := os.Stat(core.GetEnvFile())
	if err == nil || os.IsExist(err) {
		println("The configuration file already exists, no need to reinstall (if you need to reinstall, please back up the database `goploy` first, delete the .env file, then restart.)")
		return
	}
	println("Installation guide ↓")
	inputReader := bufio.NewReader(os.Stdin)
	println("Installation guidelines (Enter to confirm input)")
	println("Please enter the mysql user:")
	mysqlUser, err := inputReader.ReadString('\n')
	if err != nil {
		panic("There were errors reading, exiting program.")
	}
	mysqlUser = utils.ClearNewline(mysqlUser)
	println("Please enter the mysql password:")
	mysqlPassword, err := inputReader.ReadString('\n')
	if err != nil {
		panic("There were errors reading, exiting program.")
	}
	mysqlPassword = utils.ClearNewline(mysqlPassword)
	if len(mysqlPassword) != 0 {
		mysqlPassword = ":" + mysqlPassword
	}
	println("Please enter the mysql host(default 127.0.0.1, without port):")
	mysqlHost, err := inputReader.ReadString('\n')
	if err != nil {
		panic("There were errors reading, exiting program.")
	}
	mysqlHost = utils.ClearNewline(mysqlHost)
	if len(mysqlHost) == 0 {
		mysqlHost = "127.0.0.1"
	}
	println("Please enter the mysql port(default 3306):")
	mysqlPort, err := inputReader.ReadString('\n')
	if err != nil {
		panic("There were errors reading, exiting program.")
	}
	mysqlPort = utils.ClearNewline(mysqlPort)
	if len(mysqlPort) == 0 {
		mysqlPort = "3306"
	}
	println("Please enter the absolute path of the log directory(default stdout):")
	logPath, err := inputReader.ReadString('\n')
	if err != nil {
		panic("There were errors reading, exiting program.")
	}
	logPath = utils.ClearNewline(logPath)
	if len(logPath) == 0 {
		logPath = "stdout"
	}
	println("Please enter the listening port(default 80):")
	port, err := inputReader.ReadString('\n')
	if err != nil {
		panic("There were errors reading, exiting program.")
	}
	port = utils.ClearNewline(port)
	if len(port) == 0 {
		port = "80"
	}
	println("Write configuration file completed")
}

func handleClientSignal() {
	switch s {
	case "stop":
		pidStr, err := ioutil.ReadFile(path.Join(core.GetAssetDir(), "goploy.pid"))
		if err != nil {
			log.Fatal("handle stop, ", err.Error(), ", may be the server not start")
		}
		pid, _ := strconv.Atoi(string(pidStr))
		process, err := os.FindProcess(pid)
		if err != nil {
			log.Fatal("handle stop, ", err.Error(), ", may be the server not start")
		}
		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Fatal("handle stop, ", err.Error())
		}
		os.Exit(1)
	}
}

func checkUpdate() {
	resp, err := http.Get("https://api.github.com/repos/zhenorzz/goploy-agent/releases/latest")
	if err != nil {
		println("Check failed")
		println(err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Check failed")
		println(err.Error())
		return
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		println("Check failed")
		println(err.Error())
		return
	}

	if _, ok := result["tag_name"]; ok {
		tagName := result["tag_name"].(string)
		tagVer, err := version.NewVersion(tagName)
		if err != nil {
			println("Check version error")
			println(err.Error())
			return
		}
		currentVer, _ := version.NewVersion(appVersion)
		if tagVer.GreaterThan(currentVer) {
			println("New release available")
			println(result["html_url"].(string))
		}
	}
}
