package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/Kaguya/config"
	"github.com/r-anime/Kaguya/features"
	"github.com/r-anime/Kaguya/misc"
//	"./config"
//	"./features"
//	"./misc"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	Start()

	<-make(chan struct{})
	return
}

// Starts BOT and its Handlers
func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		panic(err)
	}

	err = goBot.Open()
	if err != nil {
		panic(err)
	}

	// Reads set react joins from reactChannelJoin.json
	features.ReactInfoRead()

	// Abstraction of a command handler
	goBot.AddHandler(features.HandleCommand)

	// Discord Status
	goBot.AddHandler(misc.StatusReady)

	// React Channel Join Handler
	goBot.AddHandler(features.ReactJoinHandler)

	// React Channel Leave Handler
	goBot.AddHandler(features.ReactRemoveHandler)

	// BOT Fluff on BOT Ping
	goBot.AddHandler(misc.OnBotPing)

	fmt.Println("Bot is running!")
}