package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// File for BOT, server and role info

var (
	Token        			string
	BotPrefix    			string
	BotID        			string
	ServerID     			string
	BotLogID     			string
	CommandRoles 			[]string
	OptInUnder   			string
	OptInAbove   			string

	config 					*configStruct
)

type configStruct struct {
	Token        			string   		`json:"-"`
	BotPrefix    			string   		`json:"BotPrefix"`
	BotID        			string   		`json:"BotID"`
	ServerID     			string   		`json:"ServerID"`
	BotLogID     			string   		`json:"BotLogID"`
	CommandRoles 			[]string 		`json:"CommandRoles"`
	OptInUnder   			string   		`json:"OptInUnder"`
	OptInAbove   			string   		`json:"OptInAbove"`
}

// Loads config.json values
func ReadConfig() error {

	fmt.Println("Reading from config file...")

	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	BotPrefix = config.BotPrefix
	BotID = config.BotID
	ServerID = config.ServerID
	BotLogID = config.BotLogID
	CommandRoles = config.CommandRoles
	OptInUnder = config.OptInUnder
	OptInAbove = config.OptInAbove

	// Takes the bot token from the environment variable. Reason is to avoid pushing token to github
	if os.Getenv("KaguyaToken") == "" {
		panic("No token set in your environment variables for key \"KaguyaToken\"")
	}
	Token = os.Getenv("KaguyaToken")
	return nil
}