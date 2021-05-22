package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) handlerMessageCreate(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	guild := bot.guild(messageCreate.GuildID)

	if strings.HasPrefix(messageCreate.Content, guild.prefix) {
		_ = bot.Router.Dispatch(session, guild.prefix, session.State.User.ID, messageCreate.Message)
	}
}

func (bot *Bot) handlerMessageUpdate(session *discordgo.Session, messageUpdate *discordgo.MessageUpdate) {
	msg, err := session.ChannelMessage(messageUpdate.ChannelID, messageUpdate.ID)
	if err != nil {
		bot.Log.WithError(err).Error("Getting message", messageUpdate.ID)
		return
	}

	for _, r := range msg.Reactions {
		if r.Me {
			return
		}
	}

	guild := bot.guild(messageUpdate.GuildID)

	if strings.HasPrefix(messageUpdate.Content, guild.prefix) {
		_ = bot.Router.Dispatch(session, guild.prefix, session.State.User.ID, messageUpdate.Message)
	}
}

func (bot *Bot) handlerGuildCreate(_ *discordgo.Session, guildCreate *discordgo.GuildCreate) {
	for _, m := range bot.Modules {
		m.Configure(&bot.Configuration, guildCreate.Guild)
	}

	s := bot.guild(guildCreate.ID)

	bot.m.Lock()
	bot.configure(s, guildCreate.Guild)
	bot.m.Unlock()

	err := bot.Discord.RequestGuildMembers(guildCreate.ID, "", 0, false)
	if err != nil {
		bot.Log.WithError(err).Error("requesting members", guildCreate)
	}
}
