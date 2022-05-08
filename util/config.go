package util

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

type Config struct {
	PCE           string `mapstructure:"PCE"`
	OrgId         string `mapstructure:"ORG_ID"`
	ApiUser       string `mapstructure:"API_USER"`
	ApiKey        string `mapstructure:"API_KEY"`
	Report        string `mapstructure:"REPORT"`
	FixedLocLabel string `mapstructure:"LOC_LABEL"`
}

const (
	ColorRed    = "\u001b[31m"
	ColorGreen  = "\u001b[32m"
	ColorYellow = "\u001b[33m"
	ColorBlue   = "\u001b[34m"
	ColorReset  = "\u001b[0m"
)

// ingest config parameter from the config.env file
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

// ask for the user for confirmation before proceed for the next
func ShallProceed(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)

	for {
		boldRed.Printf("%s [y/n]: ", s)

		resp, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		resp = strings.ToLower(strings.TrimSpace(resp))

		if resp == "y" {
			return true
		} else if resp == "n" {
			return false
		}
	}
}
