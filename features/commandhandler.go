package features

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/Kaguya/config"
	"github.com/r-anime/Kaguya/misc"
	//"../config"
	//"../misc"
)

var (
	commandMap = make(map[string]*command)
	aliasMap   = make(map[string]string)
	l          = log.New(os.Stderr, "cmds: ", log.LstdFlags|log.Lshortfile)
)

type command struct {
	execute      func(*discordgo.Session, *discordgo.Message)
	trigger      string
	aliases      []string
	desc         string
	commandCount int
	deleteAfter  bool
	elevated     bool
	category	 string
}

func add(c *command) {
	commandMap[c.trigger] = c
	for _, alias := range c.aliases {
		aliasMap[alias] = c.trigger
	}
	l.Printf("Added command %s | %d aliases | %v category", c.trigger, len(c.aliases), c.category)
}

// HandleCommand handles the incoming message
func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				l.Println(err.Error())
				l.Println(rec)
			}
		}
	}()

	s.State.RWMutex.RLock()
	if m.Author.ID == s.State.User.ID {
		s.State.RWMutex.RUnlock()
		return
	}
	if m.Author.Bot {
		s.State.RWMutex.RUnlock()
		return
	}
	if len(m.Message.Content) == 0 {
		s.State.RWMutex.RUnlock()
		return
	}
	if m.Message.Content[0:len(config.BotPrefix)] != config.BotPrefix {
		s.State.RWMutex.RUnlock()
		return
	}

	cmdTrigger := strings.Split(m.Content, " ")[0][len(config.BotPrefix):]
	cmdTrigger = strings.ToLower(cmdTrigger)
	cmd, ok := commandMap[cmdTrigger]
	if !ok {
		cmd, ok = commandMap[aliasMap[cmdTrigger]]
		if !ok {
			s.State.RWMutex.RUnlock()
			return
		}
	}
	if cmd.elevated && !hasElevatedPermissions(s, m.Author) {
		s.State.RWMutex.RUnlock()
		return
	}
	s.State.RWMutex.RUnlock()
	cmd.execute(s, m.Message)
	misc.GlobalMutex.Lock()
	cmd.commandCount++
	misc.GlobalMutex.Unlock()
	if cmd.deleteAfter {
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			return
		}
	}
}

func hasElevatedPermissions(s *discordgo.Session, u *discordgo.User) bool {
	mem, err := s.State.Member(config.ServerID, u.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, u.ID)
		if err != nil {
			fmt.Println(err)
		}
	}
	return misc.HasPermissions(mem)
}