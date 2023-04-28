package main

import (
	"context"
	"fmt"
	"os"

	trans "github.com/hekmon/transmissionrpc/v2"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port     string `yaml:"port" envconfig:"TRANSMISSIONPORT"`
		IP       string `yaml:"ip" envconfig:"TRANSMISSIONIP"`
		Username string `yaml:"username" envconfig:"TRANSMISSIONUSERNAME"`
		Password string `yaml:"password" envconfig:"TRANSMISSIONPASSWORD"`
	} `yaml:"server"`

	Bookmarks struct {
		Bookmarks []string `yaml:"paths"`
	} `yaml:"bookmarks"`

	UI struct {
		Columns []string `yaml:"columns"`
	} `yaml:"ui"`
}

func readConfig(cfg *Config) {
	f, err := os.Open("settings.yml")

	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Println(err)
	}
}

func readEnvironmentalVariables(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		fmt.Printf("Error Reading Environmental Variables: %s", err)
	}
}

var Cfg Config
var TransmissionClient *trans.Client

func transmissionClientInit() {
	TransmissionClient = setupCheck()
}

func setupCheck() *trans.Client {
	readConfig(&Cfg)
	readEnvironmentalVariables(&Cfg)
	transmissionPassword := Cfg.Server.Password
	transmissionUserName := Cfg.Server.Username
	transmissionIP := Cfg.Server.IP
	client, err := trans.New(transmissionIP, transmissionUserName, transmissionPassword, nil)

	if err != nil {
		fmt.Println("Unable to create transmission client.")
		panic(err)
	}

	if transmissionPassword == "" || transmissionIP == "" || transmissionUserName == "" {
		panic(`Credentials error. Are the environmental variables set?
'TRANSMISSIONPASSWORD',
'TRANSMISSIONUSERNAME',
'TRANSMISSIONIP'`)
	}

	ok, serverVersion, serverMinimumVersion, err := client.RPCVersion(context.TODO())
	if err != nil {
		panic(err)
	}
	if !ok {
		panic(fmt.Sprintf("Remote transmission RPC version (v%d) is incompatible with the transmission library (v%d): remote needs at least v%d",
			serverVersion, trans.RPCVersion, serverMinimumVersion))
	}
	return client

}

func MoveTorrent(id int64, location string) error {
	return TransmissionClient.TorrentSetLocation(context.TODO(), id, location, true)
}
