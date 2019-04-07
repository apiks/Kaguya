package features

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/Kaguya/config"
	"github.com/r-anime/Kaguya/misc"
	//"../config"
	//"../misc"
)

var (
	reactChannelJoinMap = make(map[string]*reactChannelJoinStruct)
	EmojiRoleMap        = make(map[string][]string)
)

type reactChannelJoinStruct struct {
	RoleEmojiMap []map[string][]string `json:"roleEmoji"`
}

// Gives a specific role to a user if they react
func ReactJoinHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	misc.GlobalMutex.Lock()
	if reactChannelJoinMap[r.MessageID] == nil {
		misc.GlobalMutex.Unlock()
		return
	}
	misc.GlobalMutex.Unlock()

	// Pulls all of the server roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	misc.GlobalMutex.Lock()
	for _, roleEmojiMap := range reactChannelJoinMap[r.MessageID].RoleEmojiMap {
		for role, emojiSlice := range roleEmojiMap {
			for _, emoji := range emojiSlice {
				if reactLowercase != emoji {
					continue
				}

				// If the role is over 17 in characters it checks if it's a valid role ID and gives the role if so
				// Otherwise it iterates through all roles to find the proper one
				if len(role) >= 17 {
					if _, err := strconv.ParseInt(role, 10, 64); err == nil {
						// Gives the role
						err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, role)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.GlobalMutex.Unlock()
								return
							}
							misc.GlobalMutex.Unlock()
							return
						}
						misc.GlobalMutex.Unlock()
						return
					}
				}
				for _, serverRole := range roles {
					if serverRole.Name == role {
						// Gives the role
						err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, serverRole.ID)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
							if err != nil {
								misc.GlobalMutex.Unlock()
								return
							}
							misc.GlobalMutex.Unlock()
							return
						}
					}
				}
			}
		}
	}
	misc.GlobalMutex.Unlock()
}

// Removes a role from user if they unreact
func ReactRemoveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	misc.GlobalMutex.Lock()
	if reactChannelJoinMap[r.MessageID] == nil {
		misc.GlobalMutex.Unlock()
		return
	}
	misc.GlobalMutex.Unlock()

	// Pulls all of the server roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	misc.GlobalMutex.Lock()
	for _, roleEmojiMap := range reactChannelJoinMap[r.MessageID].RoleEmojiMap {
		for role, emojiSlice := range roleEmojiMap {
			for _, emoji := range emojiSlice {
				if reactLowercase != emoji {
					continue
				}

				// If the role is over 17 in characters it checks if it's a valid role ID and removes the role if so
				// Otherwise it iterates through all roles to find the proper one
				if len(role) >= 17 {
					if _, err := strconv.ParseInt(role, 10, 64); err == nil {
						// Removes the role
						err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, role)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.GlobalMutex.Unlock()
								return
							}
							misc.GlobalMutex.Unlock()
							return
						}
						misc.GlobalMutex.Unlock()
						return
					}
				}
				for _, serverRole := range roles {
					if serverRole.Name == role {
						// Removes the role
						err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, serverRole.ID)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
							if err != nil {
								misc.GlobalMutex.Unlock()
								return
							}
							misc.GlobalMutex.Unlock()
							return
						}
					}
				}
			}
		}
	}
	misc.GlobalMutex.Unlock()
}

// Sets react joins per specific message and emote
func setReactJoinCommand (s *discordgo.Session, m *discordgo.Message) {

	var roleExists bool

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 4)

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"setreact [messageID] [emoji] [role]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if it's a valid messageID
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil || num < 17 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches all server roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if the role exists in the server roles
	for _, role := range roles {
		if strings.ToLower(role.Name) == commandStrings[3] {
			roleExists = true
			break
		}
	}
	if !roleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid role.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parses if it's custom emoji or unicode emoji
	re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(messageLowercase, 1)
	if emojiRegex != nil {

		// Fetches emoji API name
		re = regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
		emojiName := re.FindAllString(emojiRegex[0], 1)

		// Sets the data in memory to be ready for writing
		SaveReactJoin(commandStrings[1], commandStrings[3], emojiName[0])

		// Writes the data to storage
		ReactChannelJoinWrite(reactChannelJoinMap)

		// Reacts with the set emote if possible and gives success
		_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], emojiName[0])
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	// If the above is false, it's a non-valid emoji or an unicode emoji (the latter preferably) and saves that

	// Sets the data in memory to be ready for writing
	SaveReactJoin(commandStrings[1], commandStrings[3], commandStrings[2])

	// Writes the data to storage
	ReactChannelJoinWrite(reactChannelJoinMap)

	// Reacts with the set emote if possible
	_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], commandStrings[2])
	_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func removeReactJoinCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		messageExists bool
		validEmoji =  false

		messageID     string
		emojiRegexAPI []string
		emojiAPI	  []string
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 && len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"removereact [messageID] Optional[emoji]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if it's a valid messageID
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil || num < 17 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.GlobalMutex.Lock()
	if len(reactChannelJoinMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.GlobalMutex.Unlock()
				return
			}
			misc.GlobalMutex.Unlock()
			return
		}
		misc.GlobalMutex.Unlock()
		return
	}
	// Checks if the messageID already exists in the map
	for k := range reactChannelJoinMap {
		if commandStrings[1] == k {
			messageExists = true
			messageID = k
			break
		}
	}
	misc.GlobalMutex.Unlock()
	if messageExists == false {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such messageID is set in storage")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes the entire message from the map and writes to storage
	misc.GlobalMutex.Lock()
	if len(commandStrings) == 2 {
		delete(reactChannelJoinMap, commandStrings[1])
		ReactChannelJoinWrite(reactChannelJoinMap)
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed entire message emoji react join.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.GlobalMutex.Unlock()
				return
			}
			misc.GlobalMutex.Unlock()
			return
		}
		misc.GlobalMutex.Unlock()
		return
	}

	if reactChannelJoinMap[messageID].RoleEmojiMap == nil {
		misc.GlobalMutex.Unlock()
		return
	}
	misc.GlobalMutex.Unlock()

	// Parses if it's custom emoji or unicode
	re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(commandStrings[2], 1)
	if emojiRegex == nil {
		// Second parser if it's custom emoji or unicode but for emoji API name instead
		reAPI := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
		emojiRegexAPI = reAPI.FindAllString(commandStrings[2], 1)
	}

	misc.GlobalMutex.Lock()
	for storageMessageID := range reactChannelJoinMap[messageID].RoleEmojiMap {
		for role, emojiSlice := range reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID] {
			for index, emoji := range emojiSlice {

				// Checks for unicode emoji
				if len(emojiRegex) == 0 && len(emojiRegexAPI) == 0 {
					if commandStrings[2] == emoji {
						validEmoji = true
					}
					// Checks for non-unicode emoji
				} else {
					// Trims non-unicode emoji name to fit API emoji name
					re = regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
					if len(emojiRegex) == 0 {
						if len(emojiRegexAPI) != 0 {
							emojiAPI = re.FindAllString(emojiRegexAPI[0], 1)
							if emoji == emojiAPI[0] {
								validEmoji = true
							}
						}
					} else {
						emojiAPI = re.FindAllString(emojiRegex[0], 1)
						if emoji == emojiAPI[0] {
							validEmoji = true
						}
					}
				}

				// Delete only if it's a valid emoji in map
				if validEmoji {
					// Delete the entire message from map if it's the only set emoji react join
					if len(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID]) == 1 && len(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID][role]) == 1 {
						delete(reactChannelJoinMap, commandStrings[1])
						ReactChannelJoinWrite(reactChannelJoinMap)
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.GlobalMutex.Unlock()
								return
							}
							misc.GlobalMutex.Unlock()
							return
						}
						// Delete only the role from map if other set react join roles exist in the map
					} else if len(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID][role]) == 1 {
						delete(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID], role)
						ReactChannelJoinWrite(reactChannelJoinMap)
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.GlobalMutex.Unlock()
								return
							}
							misc.GlobalMutex.Unlock()
							return
						}
						// Delete only that specific emoji for that specific role
					} else {
						a := reactChannelJoinMap[commandStrings[1]].RoleEmojiMap[storageMessageID][role]
						a = append(a[:index], a[index+1:]...)
						reactChannelJoinMap[commandStrings[1]].RoleEmojiMap[storageMessageID][role] = a
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.GlobalMutex.Unlock()
								return
							}
							misc.GlobalMutex.Unlock()
							return
						}
					}
					misc.GlobalMutex.Unlock()
					return
				}

			}
		}
	}
	misc.GlobalMutex.Unlock()

	// If it comes this far it means it's an invalid emoji
	if emojiRegex == nil && emojiRegexAPI == nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid emoji. Please input a valid emoji or emoji API name.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
}

// Prints all currently set React Joins in memory
func viewReactJoinsCommand(s *discordgo.Session, m *discordgo.Message) {

	var line string

	misc.GlobalMutex.Lock()
	if len(reactChannelJoinMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.GlobalMutex.Unlock()
				return
			}
			misc.GlobalMutex.Unlock()
			return
		}
		misc.GlobalMutex.Unlock()
		return
	}

	// Iterates through all of the set channel joins and assigns them to a string
	for messageID, value := range reactChannelJoinMap {

		// Formats message
		line = "——————\n`MessageID: " + (messageID + "`\n")
		for i := 0; i < len(value.RoleEmojiMap); i++ {
			for role, emoji := range value.RoleEmojiMap[i] {
				line = line + "`" + role + "` — "
				for j := 0; j < len(emoji); j++ {
					if j != len(emoji)-1 {
						line = line + emoji[j] + ", "
					} else {
						line = line + emoji[j] + "\n"
					}
				}
			}
		}

		_, err := s.ChannelMessageSend(m.ChannelID, line)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.GlobalMutex.Unlock()
				return
			}
			misc.GlobalMutex.Unlock()
			return
		}
	}
	misc.GlobalMutex.Unlock()
}

// Reads set message react join info from reactChannelJoin.json
func ReactInfoRead() {

	// Reads all the set react joins from the reactChannelJoin.json file and puts them in reactChannelJoinMap as bytes
	reactChannelJoinByte, err := ioutil.ReadFile("database/reactChannelJoin.json")
	if err != nil {
		return
	}

	// Takes all the set react join from reactChannelJoin.json from byte and puts them into the reactChannelJoinMap map
	misc.GlobalMutex.Lock()
	err = json.Unmarshal(reactChannelJoinByte, &reactChannelJoinMap)
	if err != nil {
		misc.GlobalMutex.Unlock()
		return
	}
	misc.GlobalMutex.Unlock()
}

// Writes react channel join info to ReactChannelJoin.json
func ReactChannelJoinWrite(info map[string]*reactChannelJoinStruct) {

	// Turns info slice into byte ready to be pushed to file
	marshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return
	}

	// Writes to file
	err = ioutil.WriteFile("database/reactChannelJoin.json", marshaledStruct, 0644)
	if err != nil {
		return
	}
}

// Saves the react channel join and parses if it already exists
func SaveReactJoin(messageID string, role string, emoji string) {

	var (
		temp		  reactChannelJoinStruct
		emojiExists = false
	)

	// Uses this if the message already has a set emoji react
	misc.GlobalMutex.Lock()
	if reactChannelJoinMap[messageID] != nil {
		temp = *reactChannelJoinMap[messageID]

		if temp.RoleEmojiMap == nil {
			temp.RoleEmojiMap = append(temp.RoleEmojiMap, EmojiRoleMap)
		}

		for i := 0; i < len(temp.RoleEmojiMap); i++ {
			if temp.RoleEmojiMap[i][role] == nil {
				temp.RoleEmojiMap[i][role] = append(temp.RoleEmojiMap[i][role], emoji)
			}

			for j := 0; j < len(temp.RoleEmojiMap[i][role]); j++ {
				if temp.RoleEmojiMap[i][role][j] == emoji {
					emojiExists = true
					break
				}
			}
			if !emojiExists {
				temp.RoleEmojiMap[i][role] = append(temp.RoleEmojiMap[i][role], emoji)
			}
		}

		reactChannelJoinMap[messageID] = &temp
		misc.GlobalMutex.Unlock()
		return
	}

	// Initializes temp.RoleEmoji if the message doesn't have a set emoji react
	EmojiRoleMapDummy := make(map[string][]string)
	if temp.RoleEmojiMap == nil {
		temp.RoleEmojiMap = append(temp.RoleEmojiMap, EmojiRoleMapDummy)
	}

	for i := 0; i < len(temp.RoleEmojiMap); i++ {
		if temp.RoleEmojiMap[i][role] == nil {
			temp.RoleEmojiMap[i][role] = append(temp.RoleEmojiMap[i][role], emoji)
		}
	}

	reactChannelJoinMap[messageID] = &temp
	misc.GlobalMutex.Unlock()
}

// Adds role to the user that uses this command if the role is between opt-in dummy roles
func joinCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleID         string
		name           string
		chanMention    string
		topic		   string

		hasRoleAlready bool
		roleExists	   bool
	)

	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "join [channel]`")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls the role name from strings after "joinchannel " or "join "
	if strings.HasPrefix(messageLowercase, config.BotPrefix+"joinchannel ") {
		name = strings.Replace(messageLowercase, config.BotPrefix+"joinchannel ", "", -1)
	} else {
		name = strings.Replace(messageLowercase, config.BotPrefix+"join ", "", -1)
	}

	// Pulls info on server roles
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Pulls info on server channels
	cha, err := s.GuildChannels(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if there's a # before the channel name and removes it if so
	if strings.Contains(name, "#") {
		name = strings.Replace(name, "#", "", -1)

		// Checks if it's in a mention format. If so then user already has access to channel
		if strings.Contains(name, "<") {

			// Fetches mention
			name = strings.Replace(name, ">", "", -1)
			name = strings.Replace(name, "<", "", -1)
			name = misc.ChMentionID(name)

			// Sends error message to user in DMs
			dm, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {
				return
			}
			_, _ = s.ChannelMessageSend(dm.ID, "You're already in "+name)
			return
		}
	}

	// Checks if the role exists on the server, sends error message if not
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name {
			roleID = deb[i].ID
			if strings.Contains(deb[i].ID, roleID) {
				roleExists = true
				break
			}
		}
	}
	if !roleExists {

		// Sends error message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "There's no #"+name)
		return
	}

	// Sets role ID
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name && roleID != "" {
			roleID = deb[i].ID
			break
		}
	}

	// Checks if the user already has the role. Sends error message if he does
	for i := 0; i < len(mem.Roles); i++ {
		if strings.Contains(mem.Roles[i], roleID) {
			hasRoleAlready = true
			break
		}
	}
	if hasRoleAlready {
		// Sets the channel mention to the variable chanMention
		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		// Sends error message to user in DMs
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "You're already in "+chanMention)
		return
	}

	// Updates the position of opt-in-under and opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == config.OptInUnder {
			misc.OptinUnderPosition = deb[i].Position
		} else if deb[i].Name == config.OptInAbove {
			misc.OptinAbovePosition = deb[i].Position
		}
	}

	// Sets role
	role, err := s.State.Role(config.ServerID, roleID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Gives role to user if the role is between dummy opt-ins
	if role.Position < misc.OptinUnderPosition &&
		role.Position > misc.OptinAbovePosition {
		err = s.GuildMemberRoleAdd(config.ServerID, m.Author.ID, roleID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}

		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				topic = cha[j].Topic
				// Sets the channel mention to the variable chanMention
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		success := "You have joined " + chanMention
		if topic != "" {
			success = success + "\n **Topic:** " + topic
		}

		// Sends success message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, success)
	}
}

// Removes a role from the user that uses this command if the role is between opt-in dummy roles
func leaveCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleID         string
		name           string
		chanMention    string

		hasRoleAlready bool
		roleExists	   bool
	)

	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "leave [channel]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls the role name from strings after "leavechannel " or "leave "
	if strings.HasPrefix(messageLowercase, config.BotPrefix+"leavechannel ") {
		name = strings.Replace(messageLowercase, config.BotPrefix+"leavechannel ", "", -1)
	} else {
		name = strings.Replace(messageLowercase, config.BotPrefix+"leave ", "", -1)
	}

	// Pulls info on server roles
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Pulls info on server channels
	cha, err := s.GuildChannels(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if there's a # before the channel name and removes it if so
	if strings.Contains(name, "#") {
		name = strings.Replace(name, "#", "", -1)
		// Checks if it's in a mention format. If so then user already has access to channel
		if strings.Contains(name, "<") {

			// Fetches mention
			name = strings.Replace(name, ">", "", -1)
			name = strings.Replace(name, "<", "", -1)
			name = misc.ChMentionID(name)

			// Sends error message to user in DMs
			dm, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {
				return
			}
			_, _ = s.ChannelMessageSend(dm.ID, "You cannot leave "+name + " using this command.")
			return
		}
	}

	// Checks if the role exists on the server, sends error message if not
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name {
			roleID = deb[i].ID
			if strings.Contains(deb[i].ID, roleID) {
				roleExists = true
				break
			}
		}
	}
	if !roleExists  {
		// Sends error message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "There's no #"+name+"")
		return
	}

	// Sets role ID
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name && roleID != "" {
			roleID = deb[i].ID
			break
		}
	}

	// Checks if the user already has the role. Sends error message if he does
	for i := 0; i < len(mem.Roles); i++ {
		if strings.Contains(mem.Roles[i], roleID) {
			hasRoleAlready = true
			break
		}
	}
	if !hasRoleAlready {

		// Sets the channel mention to the variable chanMention
		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		// Sends error message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "You're already out of " + chanMention + "")
		return
	}

	// Updates the position of opt-in-under and opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == config.OptInUnder {
			misc.OptinUnderPosition = deb[i].Position
		} else if deb[i].Name == config.OptInAbove {
			misc.OptinAbovePosition = deb[i].Position
		}
	}

	// Sets role
	role, err := s.State.Role(config.ServerID, roleID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Removes role from user if the role is between dummy opt-ins
	if role.Position < misc.OptinUnderPosition &&
		role.Position > misc.OptinAbovePosition {

		var (
			chanMention string
		)

		err = s.GuildMemberRoleRemove(config.ServerID, m.Author.ID, roleID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}

		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				// Sets the channel mention to the variable chanMention
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		// Sends success message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "You have left " + chanMention)
	}
}

// Deletes all reacts linked to a specific channel
func deleteChannelReacts(s *discordgo.Session, m *discordgo.Message) {
	var (
		channelID 		string
		channelName 	string
		roleName		string

		message 		discordgo.Message
		author  		discordgo.User
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "killchannelreacts` [channel]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches channel ID
	channelID, channelName = misc.ChannelParser(s, commandStrings[1])
	if channelID == "" && channelName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fixes role name bug by hyphenating the channel name
	roleName = strings.Replace(strings.TrimSpace(channelName), " ", "-", -1)
	roleName = strings.Replace(roleName, "--", "-", -1)

	// Deletes all set reacts that link to the role ID if not using Kaguya
	misc.GlobalMutex.Lock()
	for messageID, roleMapMap := range reactChannelJoinMap {
		for _, roleEmojiMap := range roleMapMap.RoleEmojiMap {
			for role, emojiSlice := range roleEmojiMap {
				if strings.ToLower(role) == strings.ToLower(roleName) {
					for _, emoji := range emojiSlice {
						// Remove React Join command
						author.ID = s.State.User.ID
						message.ID = messageID
						message.Author = &author
						message.Content = fmt.Sprintf("%vremovereact %v %v", config.BotPrefix, messageID, emoji)
						misc.GlobalMutex.Unlock()
						removeReactJoinCommand(s, &message)
						misc.GlobalMutex.Lock()
					}
				}
			}
		}
	}
	misc.GlobalMutex.Unlock()

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err := s.ChannelMessageSend(m.ChannelID, "Success: Channel `" + channelName + "`'s set react joins were removed!")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Deletes all reacts linked to the channels of a specific category
func deleteCategoryReacts(s *discordgo.Session, m *discordgo.Message) {
	var (
		categoryID 		string
		categoryName	string

		message 		discordgo.Message
		author  		discordgo.User
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "killcategoryreacts` [category]")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches category ID
	categoryID, categoryName = misc.ChannelParser(s, commandStrings[1])
	if categoryID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Starting channel react deletion. For categories with a lot of channels you will have to wait more. A message will be sent when it is done.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	channels, err := s.GuildChannels(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
	for _, channel := range channels {
		if channel.ParentID == categoryID {
			// Delete channel reacts Command
			author.ID = s.State.User.ID
			message.Author = &author
			message.ChannelID = m.ChannelID
			message.Content = fmt.Sprintf("%vkillchannelreacts %v", config.BotPrefix, channel.ID)
			deleteChannelReacts(s, &message)
		}
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Category `" + categoryName + "`'s set react joins were removed!")
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
		execute:  setReactJoinCommand,
		trigger:  "setreact",
		aliases:  []string{"setreactjoin", "addreact"},
		desc:     "Sets a react join on a specific message, role and emote.",
		elevated: true,
	})
	add(&command{
		execute:  removeReactJoinCommand,
		trigger:  "removereact",
		aliases:  []string{"removereactjoin", "deletereact"},
		desc:     "Removes a set react join.",
		elevated: true,
	})
	add(&command{
		execute:  viewReactJoinsCommand,
		trigger:  "viewreacts",
		aliases:  []string{"viewreactjoins", "viewreact", "viewreacts", "reacts", "react"},
		desc:     "Views all set react joins.",
		elevated: true,
	})
	add(&command{
		execute:  joinCommand,
		trigger:  "join",
		aliases:  []string{"joinchannel"},
		desc:     "Join a spoiler channel.",
		deleteAfter: true,
	})
	add(&command{
		execute:  leaveCommand,
		trigger:  "leave",
		aliases:  []string{"leavechannel"},
		desc:     "Leave a spoiler channel.",
		deleteAfter: true,
	})
	add(&command{
		execute:  deleteChannelReacts,
		trigger:  "killchannelreacts",
		aliases:  []string{"removechannelreacts", "removechannelreact", "killchannelreact", "deletechannelreact", "deletechannelreacts"},
		desc:     "Removes all reacts linked to a specific channel.",
		elevated: true,
	})
	add(&command{
		execute:  deleteCategoryReacts,
		trigger:  "killcategoryreacts",
		aliases:  []string{"removecategoryreacts", "removecategoryreact", "killcategoryreact", "deletecategoryreact", "deletecategoryreacts"},
		desc:     "Removes all reacts linked to a specific category.",
		elevated: true,
	})
}