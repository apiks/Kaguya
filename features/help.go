package features

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/Kaguya/config"
	"github.com/r-anime/Kaguya/misc"
	//"../config"
	//"../misc"
)

// Prints pretty help
func helpEmbedCommand(s *discordgo.Session, m *discordgo.Message) {

	var admin bool

	// Checks if it's within the config server
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			return
		}
	}
	if ch.GuildID != config.ServerID {
		return
	}

	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks for mod perms and handles accordingly
	s.State.RWMutex.RLock()
	if misc.HasPermissions(mem) {
		admin = true
	}
	s.State.RWMutex.RUnlock()

	err = helpEmbed(s, m, admin)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Embed message for general all-purpose help message
func helpEmbed(s *discordgo.Session, m *discordgo.Message, admin bool) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed              []*discordgo.MessageEmbedField
		user               discordgo.MessageEmbedField
		permission         discordgo.MessageEmbedField
		userCommands       discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets user field
	user.Name = "User:"
	user.Value = m.Author.Mention()
	user.Inline = true

	// Sets permission field
	permission.Name = "Permission Level:"
	if admin {
		permission.Value = "_Admin_"
	} else {
		permission.Value = "_User_"
	}
	permission.Inline = true

	// Sets user commands field
	userCommands.Name = "Command:"
	userCommands.Inline = true

	// Iterates through non-mod commands and adds them to the embed sorted
	misc.GlobalMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if !commandMap[commands[i]].elevated && !admin {
			userCommands.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		} else if admin {
			userCommands.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.GlobalMutex.Unlock()

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", config.BotPrefix)
	embedMess.Footer = &embedFooter

	// Adds the fields to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &user)
	embed = append(embed, &permission)
	embed = append(embed, &userCommands)

	// Adds everything together
	embedMess.Fields = embed

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Prints plaintext help
func helpPlaintextCommand(s *discordgo.Session, m *discordgo.Message) {
	plainHelp := "`" + config.BotPrefix + "about` | Shows information about me.\n" +
		"`" + config.BotPrefix + "avatar [@user or userID]` | Show user avatar. Add [@mention] or [userID] to specify a user.\n" +
		"`" + config.BotPrefix + "help` | Print all available commands in embed form.\n" +
		"`" + config.BotPrefix + "helpplain` | Print all available commands in plain text.\n" +
		"`" + config.BotPrefix + "join` | Join a spoiler channel.\n" +
		"`" + config.BotPrefix + "leave` | Leave a spoiler channel.\n"

	_, err := s.ChannelMessageSend(m.ChannelID, plainHelp)
	if err != nil {
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute:  helpEmbedCommand,
		trigger:  "help",
		aliases:  []string{"h"},
		desc:     "Print all available commands in embed form.",
		category: "normal",
	})
	add(&command{
		execute:  helpPlaintextCommand,
		trigger:  "helpplain",
		desc:     "Prints all non-admin commands in plain text.",
		category: "normal",
	})
}