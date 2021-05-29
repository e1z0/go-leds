package main

import (
        "io/ioutil"
        "fmt"
        "encoding/json"
        "crypto/tls"
)

var Global = &Config{}

type Config struct {
        Port             int             `json:"port"`
        AdminSecret      string          `json:"admin_secret"`
        LedsLength	 int		 `json:"ledslength"`
        LedsRotation     Rotation        `json:"ledsrotation"`
        ClockAtIdle      bool		 `json:"clockatidle"`
        Debug		 bool            `json:"debug"`
	DisplayErorMessages bool         `json:"displayerrormessages"`
        ServerOptions    ServerOptions   `json:"server_options"`
}

func (c *Config) Load(filePath string) {
        configuration, err := ioutil.ReadFile(filePath)
        if err != nil {
                //logrus.WithError(err).Error("Couldn't read config file")
                fmt.Printf("Couldn't read config file")
        }

        err = json.Unmarshal(configuration, &c)
        if err != nil {
                //logrus.WithError(err).Error("Couldn't unmarshal configuration")
		fmt.Printf("Couldn't unmarshal configuration")
        }
}

type ServerOptions struct {
        EnableTLS bool       `json:"enable_tls"`
        CertFile  string     `json:"cert_file"`
        KeyFile   string     `json:"key_file"`
        TLSConfig tls.Config `json:"tls_config"`
}
