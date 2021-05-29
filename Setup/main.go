package main

import (
	. "github.com/cyoung/rpi"
	"time"
	"fmt"
	"flag"
)

var (
	wpcl    *WpaCfg
)


func init() {
	WiringPiSetup()
	wpcl = NewWpaCfg("cfg/wificfg.json")
}

func main() {
	fmt.Printf("Launched!\n")
	ap := flag.Bool("ap", false, "Start in ap mode")
	flag.Parse()
	if *ap {
		WifiSetup()
	}
	go GpioEvents()
	for {
		// just silly loop
		time.Sleep(10*time.Second)
	}
}
