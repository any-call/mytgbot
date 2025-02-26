package mytgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type inLine struct {
}

func ImpInline() inLine {
	return inLine{}
}

func (self inLine) SendQueryResultArticle(bot *tgbotapi.BotAPI, queryID string, messageText string, configFn func(inlineCfg *tgbotapi.InlineQueryResultArticle)) (*tgbotapi.APIResponse, error) {
	//返回查询的卡片
	article := tgbotapi.NewInlineQueryResultArticle(queryID, "", messageText)
	if configFn != nil {
		configFn(&article)
	}

	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: queryID,
		Results:       []interface{}{article},
		CacheTime:     1,
	}

	return bot.Request(inlineConfig)
}
