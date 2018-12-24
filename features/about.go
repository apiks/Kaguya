package features

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/Kaguya/config"
	"github.com/r-anime/Kaguya/misc"
	//"../config"
	//"../misc"
)

// Returns a message on "about" for bot information
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Good day. I'm Kaguya from the series _Kaguya-sama: Love Is War_." +
		"I was made by Apiks for /r/anime as a react BOT. I'm written in Go. Use `" + config.BotPrefix +
		"help` to list what commands you have access to. I wish you an excellent future, for all is fair in love and war.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute: aboutCommand,
		trigger: "about",
		desc:    "Get info about me.",
		category:"normal",
	})
}