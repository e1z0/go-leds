package main

import (
	"time"
	"context"
	"net/http"
	"html/template"
	"fmt"
	"github.com/gorilla/mux"
	"os/exec"
)

var ShutHttpChan chan int
var messages chan CmdMessage

type Network struct {
   Id string
   Name string
}

func WpaNets() []Network {
    Networks := []Network{}
    wpaNetworks, err := wpcl.ScanNetworks()
    if err != nil {
    fmt.Printf("We got some error when tried to scan for the wifi networks: %s\n", err)
	return Networks
    }
    for i, item := range wpaNetworks {
    Networks = append(Networks,Network{Name: "Name: "+item.Ssid+" Bssid: "+item.Bssid+" Freq: "+item.Frequency+" Signal: "+item.SignalLevel, Id: i})
    }
    return Networks
}


func WifiRoot(w http.ResponseWriter, r *http.Request) {
    Networks:= WpaNets()
    parsedTemplate, _ := template.ParseFiles("forms/setup.html")
    err := parsedTemplate.Execute(w, Networks)
    if err != nil {
        fmt.Println("Error executing template :", err)
        return
    }
}



func WifiSet(w http.ResponseWriter, r *http.Request) {
        network := r.FormValue("network")
        password := r.FormValue("password")
        checkbox := r.FormValue("checkbox")
        fmt.Printf("Got text from form:\nAP:%s\nPass: %s\nCheckbox: %s\n",network,password,checkbox)
        if checkbox == "on" {
        fmt.Printf("Ready to be saved!\n")
        w.Write([]byte("Setting parameters... The LED Display will now connecting to your specified Wireless Access Point!"))
        fmt.Printf("Sending shutdown command to hotspot...\n")
        messages <- CmdMessage{Id: "kill"}
        // here we should write wpa supplicant config and disable access point
         fmt.Printf("Configuring submitted network...\n")
         setupCfg, err := loadCfg("cfg/wificfg.json")
         if err != nil {
                fmt.Printf("Could not load config: %s\n", err.Error())
                return
         }
         cmdRunner := CmdRunner{
                Messages: messages,
                Handlers: make(map[string]func(cmsg CmdMessage), 0),
                Commands: make(map[string]*exec.Cmd, 0),
         }
         command := &Command{
                Runner:   cmdRunner,
                SetupCfg: setupCfg,
         }
         fmt.Printf("Killing dhclient\n")
         command.killdhcp()
         fmt.Printf("Killing wpa client\n")
         command.KillWpa()
         fmt.Printf("Bringing down interface\n")
         command.DownInterface()
         fmt.Printf("Starting wpa supplicant...\n")
         command.StartWpaSupplicant()
         time.Sleep(5*time.Second)
         net := WpaCredentials{Ssid: network, Psk: password}
         connection, err := wpcl.ConnectNetwork(net)
         fmt.Printf("rebooting...\n")
         command.Reboot()
         if err != nil {
            fmt.Printf("There was an error when trying to connect to wifi, restoring last state...\n")
            time.Sleep(10*time.Second)
//            WifiSetup()
         }
         ShutHttpChan <- 1
         fmt.Printf("Successfully connected to %s!\n",connection)
         return
        }
}

func GenerateRouter() *mux.Router {
        r := mux.NewRouter()
        r.HandleFunc("/", WifiRoot).Methods(http.MethodGet)
        r.HandleFunc("/set", WifiSet).Methods(http.MethodPost)
        return r
}

func TurnInternalHttpd() {
        fmt.Printf("Turning on internal http server for remote configuration of access points!\n")
        router := GenerateRouter()
        addr := fmt.Sprintf(":%d", 80)
        ShutHttpChan = make(chan int)
        srv := &http.Server{
                        Addr:         addr,
                        Handler:      router,
                }
        go func() {
            if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
               fmt.Printf("Fatal http server error: %s\n",err)
             }
        }()
        select {
           case code := <-ShutHttpChan:
            // Post process after shutdown here
            fmt.Printf("Got code from http shutdown channel: %d\n", code)
            srv.Shutdown(context.Background())
        }
        fmt.Printf("Finished serving httpd\n")
}

func WifiSetup() {
// here we should switch to AP mode and run webserver
	fmt.Printf("Starting Hotspot...\n")
	messages = make(chan CmdMessage, 1)
	go RunWifi(messages, "cfg/wificfg.json")
	TurnInternalHttpd()
	fmt.Printf("We have finished ap setup\n")
	// here we should stop the internal access point
	SETUP_AVAILABLE=true
}

func SwitchToNormal() {
	fmt.Printf("Switch to normal wifi, not implemented yet!")
}



