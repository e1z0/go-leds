package main


import (
"github.com/grandcat/zeroconf"
"os"
"syscall"
"fmt"
"os/signal"
)

func RegisterZeroConf() {
server, err := zeroconf.Register("LedMatrix", "_leds._tcp", "local.", Global.Port, []string{"txtv=0", "lo=1", "la=2"}, nil)
if err != nil {
    panic(err)
}
defer server.Shutdown()

// Clean exit.
sig := make(chan os.Signal, 1)
signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
select {
case <-sig:
    // Exit by user
//case <-time.After(time.Second * 120):
    // Exit by timeout
}

fmt.Println("Shutting down.")
server.Shutdown()
os.Exit(0)
}
