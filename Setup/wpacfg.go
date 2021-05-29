package main

import (
        "fmt"
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// WpaCfg for configuring wpa
type WpaCfg struct {
	WpaCmd []string
	WpaCfg *SetupCfg
}

// WpaNetwork defines a wifi network to connect to.
type WpaNetwork struct {
	Bssid       string `json:"bssid"`
	Frequency   string `json:"frequency"`
	SignalLevel string `json:"signal_level"`
	Flags       string `json:"flags"`
	Ssid        string `json:"ssid"`
}

// WpaCredentials defines wifi network credentials.
type WpaCredentials struct {
	Ssid string `json:"ssid"`
	Psk  string `json:"psk"`
}

// WpaConnection defines a WPA connection.
type WpaConnection struct {
	Ssid    string `json:"ssid"`
	State   string `json:"state"`
	Ip      string `json:"ip"`
	Message string `json:"message"`
}

// NewWpaCfg produces WpaCfg configuration types.
func NewWpaCfg(cfgLocation string) *WpaCfg {

	setupCfg, err := loadCfg(cfgLocation)
	if err != nil {
		fmt.Printf("Could not load config: %s\n", err.Error())
		panic(err)
	}

	return &WpaCfg{
		WpaCfg: setupCfg,
	}
}

// StartAP starts AP mode.
func (wpa *WpaCfg) StartAP() {
	fmt.Printf("Starting Hostapd.\n")

	command := &Command{
		SetupCfg: wpa.WpaCfg,
	}

        command.DownInterface()
	command.ConfigureApInterface()

	cmd := exec.Command("hostapd", "-d", "/dev/stdin")

	// pipes
	hostapdPipe, _ := cmd.StdinPipe()
	cmdStdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	messages2 := make(chan string, 1)

	stdOutScanner := bufio.NewScanner(cmdStdoutReader)
	go func() {
		for stdOutScanner.Scan() {
			fmt.Printf("HOSTAPD GOT: %s\n", stdOutScanner.Text())
			messages2 <- stdOutScanner.Text()
		}
	}()

	cfg := `interface=wlan0
        ssid=` + wpa.WpaCfg.HostApdCfg.Ssid + `
        driver=nl80211
        hw_mode=g
        channel=` + wpa.WpaCfg.HostApdCfg.Channel + `
        macaddr_acl=0`

	fmt.Printf("Hostapd CFG: %s\n", cfg)
	hostapdPipe.Write([]byte(cfg))

	cmd.Start()
	hostapdPipe.Close()
        fmt.Printf("Phase 3\n")

	for {
		out := <-messages2 // Block until we receive a message on the channel
		if strings.Contains(out, "wlan0: AP-DISABLED") {
			fmt.Printf("Hostapd DISABLED")
			return

		}
		if strings.Contains(out, "wlan0: AP-ENABLED") {
			fmt.Printf("Hostapd ENABLED")
			return
		}
	}
        fmt.Printf("Phase 4\n")
       fmt.Printf("Pasibaige wpacfg ciklas\n")
}

// ConfiguredNetworks returns a list of configured wifi networks.
func (wpa *WpaCfg) ConfiguredNetworks() string {
	netOut, err := exec.Command("wpa_cli", "-i", "wlan0", "scan").Output()
	if err != nil {
		fmt.Printf("err: %s\n",err)
	}

	return string(netOut)
}

// ConnectNetwork connects to a wifi network
func (wpa *WpaCfg) ConnectNetwork(creds WpaCredentials) (WpaConnection, error) {
	connection := WpaConnection{}

	// 1. Add a network
	addNetOut, err := exec.Command("wpa_cli", "-i", "wlan0", "add_network").Output()
	if err != nil {
		fmt.Printf("err: %s\n",err)
		return connection, err
	}
	net := strings.TrimSpace(string(addNetOut))
	fmt.Printf("WPA add network got: %s\n", net)

	// 2. Set the ssid for the new network
	addSsidOut, err := exec.Command("wpa_cli", "-i", "wlan0", "set_network", net, "ssid", "\""+creds.Ssid+"\"").Output()
	if err != nil {
		fmt.Printf("err: %s\n",err)
		return connection, err
	}
	ssidStatus := strings.TrimSpace(string(addSsidOut))
	fmt.Printf("WPA add ssid got: %s\n", ssidStatus)

	// 3. Set the psk for the new network
	addPskOut, err := exec.Command("wpa_cli", "-i", "wlan0", "set_network", net, "psk", "\""+creds.Psk+"\"").Output()
	if err != nil {
		fmt.Printf(err.Error())
		return connection, err
	}
	pskStatus := strings.TrimSpace(string(addPskOut))
	fmt.Printf("WPA psk got: %s\n", pskStatus)

	// 4. Enable the new network
	enableOut, err := exec.Command("wpa_cli", "-i", "wlan0", "enable_network", net).Output()
	if err != nil {
		fmt.Printf("err: %s\n",err.Error())
		return connection, err
	}
	enableStatus := strings.TrimSpace(string(enableOut))
	fmt.Printf("WPA enable got: %s\n", enableStatus)

 	// 5. Select the new network
 	selectOut, err := exec.Command("wpa_cli", "-i", "wlan0", "select_network", net).Output()
 	if err != nil {
 		fmt.Printf("wpa select: %s\n",err.Error())
 		return connection, err
 	}
 	selectStatus := strings.TrimSpace(string(selectOut))
 	fmt.Printf("WPA select got\n: %s", selectStatus)

	// regex for state
	rState := regexp.MustCompile("(?m)wpa_state=(.*)\n")

	// loop for status every second
	for i := 0; i < 5; i++ {
		fmt.Printf("WPA Checking wifi state\n")

		stateOut, err := exec.Command("wpa_cli", "-i", "wlan0", "status").Output()
		if err != nil {
			fmt.Printf("Got error checking state: %s\n", err.Error())
			return connection, err
		}
		ms := rState.FindSubmatch(stateOut)

		if len(ms) > 0 {
			state := string(ms[1])
			fmt.Printf("WPA Enable state: %s\n", state)
			// see https://developer.android.com/reference/android/net/wifi/SupplicantState.html
			if state == "COMPLETED" {
				// save the config
				saveOut, err := exec.Command("wpa_cli", "-i", "wlan0", "save_config").Output()
				if err != nil {
					fmt.Printf("err: %s\n",err.Error())
					return connection, err
				}
				saveStatus := strings.TrimSpace(string(saveOut))
				fmt.Printf("WPA save got: %s\n", saveStatus)

				connection.Ssid = creds.Ssid
				connection.State = state

				return connection, nil
			}
		}

		time.Sleep(3 * time.Second)
	}

	connection.State = "FAIL"
	connection.Message = "Unable to connection to " + creds.Ssid
	return connection, nil
}

// Status returns the WPA wireless status.
func (wpa *WpaCfg) Status() (map[string]string, error) {
	cfgMap := make(map[string]string, 0)

	stateOut, err := exec.Command("wpa_cli", "-i", "wlan0", "status").Output()
	if err != nil {
		fmt.Printf("Got error checking state: %s\n", err.Error())
		return cfgMap, err
	}

	cfgMap = cfgMapper(stateOut)

	return cfgMap, nil
}

// cfgMapper takes a byte array and splits by \n and then by = and puts it all in a map.
func cfgMapper(data []byte) map[string]string {
	cfgMap := make(map[string]string, 0)

	lines := bytes.Split(data, []byte("\n"))

	for _, line := range lines {
		kv := bytes.Split(line, []byte("="))
		if len(kv) > 1 {
			cfgMap[string(kv[0])] = string(kv[1])
		}
	}

	return cfgMap
}

// ScanNetworks returns a map of WpaNetwork data structures.
func (wpa *WpaCfg) ScanNetworks() (map[string]WpaNetwork, error) {
	wpaNetworks := make(map[string]WpaNetwork, 0)

	scanOut, err := exec.Command("wpa_cli", "-i", "wlan0", "scan").Output()
	if err != nil {
		fmt.Printf("fatal: %s\n",err)
		return wpaNetworks, err
	}
	scanOutClean := strings.TrimSpace(string(scanOut))

	// wait one second for results
	time.Sleep(1 * time.Second)

	if scanOutClean == "OK" {
		networkListOut, err := exec.Command("wpa_cli", "-i", "wlan0", "scan_results").Output()
		if err != nil {
			fmt.Printf("err: %s\n",err)
			return wpaNetworks, err
		}

		networkListOutArr := strings.Split(string(networkListOut), "\n")
		for _, netRecord := range networkListOutArr[1:] {
			if strings.Contains(netRecord, "[P2P]") {
				continue
			}

			fields := strings.Fields(netRecord)

			if len(fields) > 4 {
				ssid := strings.Join(fields[4:], " ")
				wpaNetworks[ssid] = WpaNetwork{
					Bssid:       fields[0],
					Frequency:   fields[1],
					SignalLevel: fields[2],
					Flags:       fields[3],
					Ssid:        ssid,
				}
			}
		}

	}

	return wpaNetworks, nil
}
