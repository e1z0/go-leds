package main

import (
      . "github.com/cyoung/rpi"
        "fmt"
        "time"
)

var (
SETUP_AVAILABLE=true
)

func GpioEvents() {
    btn_pushed := 0
    last_time := time.Now().UnixNano() / 1000000
    for pinas := range WiringPiISR(1, INT_EDGE_FALLING) {
              if (pinas > -1 && SETUP_AVAILABLE) {
                 n := time.Now().UnixNano() / 1000000
                 delta := n - last_time
                 if delta > 1000 {
                 // reset counter
                 fmt.Printf("Reseting counter\n")
                 btn_pushed=0
                 }
                 if delta > 300 { //software debouncing
                        fmt.Printf("Button pressed: %d times\n",btn_pushed)
                        if (btn_pushed > 7 && SETUP_AVAILABLE) {
                        SETUP_AVAILABLE=false
                        WifiSetup()
                        btn_pushed=0
                        }
                        last_time = n
                        btn_pushed++
                 }
               }
   }

}
