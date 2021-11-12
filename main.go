package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/mrod502/go-shared-hosting/gsh"
	yaml "gopkg.in/yaml.v2"
)

func main() {

	var config = new(gsh.Config)

	b, err := ioutil.ReadFile(os.Args[1])
	checkErr(err)

	err = yaml.Unmarshal(b, config)
	checkErr(err)

	checkErr(err)

	b, err = os.ReadFile(config.RoutesFile)

	var routes = make(map[string]*gsh.Route)
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
