package mytgbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

type (
	ErrStatus int
)

const (
	ErrStatusNil          ErrStatus = 0
	ErrStatusKicked       ErrStatus = 1   //对话已被踢出群聊
	ErrStatusGroupDeleted ErrStatus = 2   //群组被删除
	ErrStatusUnknown      ErrStatus = 100 //未知的错误
)

func CheckErrStatus(err error) ErrStatus {
	if err == nil {
		return ErrStatusNil
	}

	if strings.Contains(err.Error(), "Forbidden: bot was kicked") {
		// 对话已被踢出群聊 修改群配置
		return ErrStatusKicked
	} else if strings.Contains(err.Error(), "Forbidden: the group chat was deleted") {
		//群组被删除 修改群配置
		return ErrStatusGroupDeleted
	}

	return ErrStatusUnknown
}

func SendMessage(bot *tgbotapi.BotAPI, chattable tgbotapi.Chattable) (*tgbotapi.Message, error) {
	if bot == nil {
		return nil, fmt.Errorf("bot api is nil")
	}

	if chattable == nil {
		return nil, fmt.Errorf("chattable is nil")
	}

	sendMsg, err := bot.Send(chattable)
	if err != nil {
		return nil, err
	}

	return &sendMsg, nil
}

func SetBotCommand(bot *tgbotapi.BotAPI, list []tgbotapi.BotCommand) error {
	if bot == nil {
		return fmt.Errorf("bot api is nil")
	}

	if list == nil {
		return fmt.Errorf("command list is nil ")
	}

	if _, err := bot.Request(tgbotapi.NewSetMyCommands(list...)); err != nil {
		return err
	}

	return nil
}
