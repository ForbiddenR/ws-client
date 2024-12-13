package main

import (
	"fmt"
	"os"

	"github.com/ForbiddenR/ws-client/pkg/hserver"
	"github.com/ForbiddenR/ws-client/pkg/ws"
	"github.com/ForbiddenR/ws-client/wait"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	config string
)

type runConfig struct {
	Addresss string `yaml:"address"`
	CertPath string `yaml:"cert_path"`
	KeyPath  string `yaml:"key_path"`
	CaPath   string `yaml:"ca_path"`
	MTLS     bool   `yaml:"mtls"`
	HttpPort string `yaml:"http_port"`
	SN       string `yaml:"sn"`
}

func newRunConfig() *runConfig {
	return new(runConfig)
}

func (c *runConfig) parse(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = yaml.NewDecoder(file).Decode(c)
	return err
}

func init() {
	c := newRunConfig()
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the weboscket client",
		PreRun: func(cmd *cobra.Command, args []string) {
			err := c.parse(config)
			if err != nil {
				fmt.Println("failed to read config file", err)
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			sender := make(chan []byte, 1000)
			hs := hserver.NewServer(c.HttpPort, sender)
			wss := ws.NewServer(c.Addresss, c.CertPath, c.KeyPath, c.CaPath, c.MTLS, c.SN, sender)
			err := wait.Start(hs, wss)
			if err != nil {
				fmt.Println("sever stopped", err)
			}
		},
	}
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&config, "config", "c", "config.yaml", "configuration of the websocket client")
}
