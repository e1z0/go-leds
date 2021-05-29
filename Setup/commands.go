package main

import (
	"os/exec"
)

// Command for device network commands.
type Command struct {
	Runner   CmdRunner
	SetupCfg *SetupCfg
}

// RemoveApInterface removes the AP interface.
func (c *Command) RemoveApInterface() {
	cmd := exec.Command("iw", "dev", "wlan0", "del")
	cmd.Start()
	cmd.Wait()
}

func (c *Command) Reboot() {
     cmd := exec.Command("reboot")
     cmd.Start()
     cmd.Wait()
}

// ConfigureApInterface configured the AP interface.
func (c *Command) ConfigureApInterface() {
	cmd := exec.Command("ifconfig", "wlan0", c.SetupCfg.HostApdCfg.Ip)
	cmd.Start()
	cmd.Wait()
}

func (c *Command) killdhcp() {
       cmd := exec.Command("pkill","dhclient")
       cmd.Start()
       cmd.Wait()
}

func (c *Command) dhclient() {
        cmd := exec.Command("dhclient","wlan0")
        cmd.Start()
        cmd.Wait()
}

func (c *Command) PkillDnsMasq() {
       cmd := exec.Command("pkill","dnsmasq")
       cmd.Start()
       cmd.Wait()
}

func (c *Command) PkillHostapd() {
       cmd := exec.Command("pkill","hostapd")
       cmd.Start()
       cmd.Wait()
}


func (c *Command) KillWpa() {
        cmd := exec.Command("pkill","wpa_supplicant")
        cmd.Start()
        cmd.Wait()
}

// UpApInterface ups the AP Interface.
func (c *Command) UpApInterface() {
//	cmd := exec.Command("ifconfig", "wlan0", "up")
        cmd := exec.Command("ifup","wlan0")
	cmd.Start()
	cmd.Wait()
}

func (c *Command) DownInterface() {
//        cmd := exec.Command("ifconfig", "wlan0", "down")
        cmd := exec.Command("ifdown","wlan0")
        cmd.Start()
        cmd.Wait()
}

// AddApInterface adds the AP interface.
func (c *Command) AddApInterface() {
	cmd := exec.Command("iw", "phy", "phy0", "interface", "add", "wlan0", "type", "__ap")
	cmd.Start()
	cmd.Wait()
}

// CheckInterface checks the AP interface.
func (c *Command) CheckApInterface() {
	cmd := exec.Command("ifconfig", "wlan0")
	go c.Runner.ProcessCmd("ifconfig_wlan0", cmd)
}

// StartWpaSupplicant starts wpa_supplicant.
func (c *Command) StartWpaSupplicant() {

	args := []string{
		"-d",
		"-Dnl80211",
		"-iwlan0",
		"-c/etc/wpa_supplicant/wpa_supplicant.conf",
	}

	cmd := exec.Command("wpa_supplicant", args...)
	go c.Runner.ProcessCmd("wpa_supplicant", cmd)
}

// StartDnsmasq starts dnsmasq.
func (c *Command) StartDnsmasq() {
	// hostapd is enabled, fire up dnsmasq
	args := []string{
		"--no-hosts", // Don't read the hostnames in /etc/hosts.
		"--keep-in-foreground",
		"--log-queries",
		"--no-resolv",
		"--address=" + c.SetupCfg.DnsmasqCfg.Address,
		"--dhcp-range=" + c.SetupCfg.DnsmasqCfg.DhcpRange,
		"--dhcp-vendorclass=" + c.SetupCfg.DnsmasqCfg.VendorClass,
		"--dhcp-authoritative",
		"--log-facility=-",
	}

	cmd := exec.Command("dnsmasq", args...)
	go c.Runner.ProcessCmd("dnsmasq", cmd)
}
