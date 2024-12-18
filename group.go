package mytgbot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type group struct {
}

func ImpGroup() group {
	return group{}
}

func (self group) GetChatMember(bot *tgbotapi.BotAPI, chatId int64, userId int64) (tgbotapi.ChatMember, error) {
	chatMemberConfig := tgbotapi.ChatConfigWithUser{
		ChatID: chatId,
		UserID: userId,
	}

	return bot.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: chatMemberConfig})
}

func (self group) ListAdminChatMember(bot *tgbotapi.BotAPI, chatId int64) ([]tgbotapi.ChatMember, error) {
	chatConfig := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatId,
		},
	}

	return bot.GetChatAdministrators(chatConfig)
}

func (self group) GetChatMembersCount(bot *tgbotapi.BotAPI, chatId int64) (int, error) {
	return bot.GetChatMembersCount(tgbotapi.ChatMemberCountConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatId,
		},
	})
}

func (self group) GetChatUser(bot *tgbotapi.BotAPI, chatId int64, username string) (tgbotapi.Chat, error) {
	return bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID:             chatId,
			SuperGroupUsername: username,
		},
	})
}
