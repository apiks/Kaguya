# Kaguya is a Discord Channel React Join/Leave BOT. It's functionalities are the following:

<br/>

* Show avatar for a target user. Works for people not in the server

* Full spoiler/opt-in/hidden channel support with reaction based role-giving or just join/leave commands. Tracks hidden channels between two dummy roles

* BOT say/edit commands that any mod can use to send or edit important messages with the BOT, or pretend they're a ROBOT

<br/>

How to install:
1. Download in a folder.
2. Edit config.json with your own values. Use only one for each, except for CommandRoles. Everything is required unless stated otherwise:

       BotPrefix is the character that needs to be used before every command

       BotID is the ID of the BOT you are using

       ServerID is the ID of the server the BOT is going to be managing

       BotLogID is the ID of the channel in which the bot will dump errors, timed events, punishments and other things

       CommandRoles are the admin/mod/bot role IDs

       OptInUnder is the name of the top dummy role for spoiler/opt-in/hidden channels

       OptInAbove is the name of the bottom dummy role for spoiler/opt-in/hidden channels

3. Set your "KaguyaToken" environment variable to the BOT token (either hidden on the system env or in config.go ReadConfig func with os.Setenv("KaguyaToken", "TOKEN"))
4. Compile in your favorite IDE or compiler with "go build" (or type "set GOOS=linux" to change OS first and then "go build".)
5. Invite BOT to server and give it an admin role
6. Start the BOT and use

<br/>

If you have discovered any bugs or have questions, please message Apiks or raise an issue.

If you use the BOT successfuly, please also let Apiks know