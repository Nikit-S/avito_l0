package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("wb_orders", "pusher")
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	} else {
		log.Println("STAN connection established")
	}

	dataFromFile, _ := ioutil.ReadFile(os.Args[1])
	if err = sc.Publish("nw", dataFromFile); err != nil {
		log.Println(err.Error())
	} else {
		println("Success")
	}

}
