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

var (
	red = color.New(color.FgHiRed)
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

	for {
		red.Printf("%s [y/n]: ", s)

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

// remove fqdn from the hostname such as hostname.abc.com, and uppercase while comparing
func Normalise(str string) string {
	res := strings.Split(str, ".")
	return strings.ToUpper(res[0])
}
