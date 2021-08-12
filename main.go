package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"time"

	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/go-shared-hosting/gsh"
	yaml "gopkg.in/yaml.v2"
)

func main() {

	var configPath = flag.String("Config", "config", "the config file path")
	flag.Parse()

	var config = new(gsh.Config)
	home, _ := os.UserHomeDir()
	b, err := ioutil.ReadFile(path.Join(home, *configPath))
	checkErr(err)

	err = yaml.Unmarshal(b, config)
	checkErr(err)

	b, err = ioutil.ReadFile(path.Join(home, config.DomainConfigPath))

	checkErr(err)

	var routes = gocache.NewStringCache()
	err = json.Unmarshal(b, &routes)
	checkErr(err)

	router, err := gsh.NewRouter(config, routes)
	checkErr(err)

	go func() {
		err = router.Run()
		if err != nil {
			router.Error("FATAL", err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("exiting...")
	router.Info("exiting", time.Now().Format(time.RFC3339))
	time.Sleep(time.Second)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
