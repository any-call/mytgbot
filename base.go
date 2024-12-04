package mytgbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func sendMessage(bot *tgbotapi.BotAPI, chattable tgbotapi.Chattable) (*tgbotapi.Message, error) {
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

func SendMessage(bot *tgbotapi.BotAPI, replyId int, chatId int64, message string) (*tgbotapi.Message, error) {
	sendMsg := tgbotapi.NewMessage(chatId, message)
	sendMsg.ReplyToMessageID = replyId
	return sendMessage(bot, sendMsg)
}

func SendMessageEx(bot *tgbotapi.BotAPI, chatId int64, message string, configCb func(messageCfg *tgbotapi.MessageConfig)) (*tgbotapi.Message, error) {
	sendMsg := tgbotapi.NewMessage(chatId, message)
	if configCb != nil {
		configCb(&sendMsg)
	}

	return sendMessage(bot, sendMsg)
}

func SendMessageWithRMarkup(bot *tgbotapi.BotAPI, replyId int, chatId int64, message string, markUpFun func() any) (*tgbotapi.Message, error) {
	sendMsg := tgbotapi.NewMessage(chatId, message)
	sendMsg.ReplyToMessageID = replyId
	if markUpFun != nil {
		sendMsg.ReplyMarkup = markUpFun()
	}
	return sendMessage(bot, sendMsg)
}

func SendMessageWithAutoDelete(bot *tgbotapi.BotAPI, replyId int, chatId int64, message string, autoDele time.Duration) error {
	sendMsg := tgbotapi.NewMessage(chatId, message)
	sendMsg.ReplyToMessageID = replyId
	retMsg, err := sendMessage(bot, sendMsg)
	if err != nil {
		return err
	}

	go func(bt *tgbotapi.BotAPI, msgID int, groupID int64) {
		time.Sleep(autoDele)
		_, _ = bt.Request(tgbotapi.NewDeleteMessage(groupID, msgID))
	}(bot, retMsg.MessageID, chatId)

	return nil
}

func SendMessageByToken(token string, toChatId int64, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", toChatId))
	data.Set("text", message)
	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("消息发送失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func SendPhoto(bot *tgbotapi.BotAPI, chatId int64, imageFileFn func() tgbotapi.RequestFileData, configCb func(photoConfig *tgbotapi.PhotoConfig)) error {
	var fileData tgbotapi.RequestFileData = nil
	if imageFileFn != nil {
		fileData = imageFileFn()
	}

	if fileData == nil {
		return fmt.Errorf("RequestFileData is nil")
	}

	photo := tgbotapi.NewPhoto(chatId, fileData)
	if configCb != nil {
		configCb(&photo)
	}

	// 发送图片消息
	if _, err := bot.Send(photo); err != nil {
		return err
	}
	return nil
}

func SendAnimation(bot *tgbotapi.BotAPI, chatID int64, animationFileFn func() tgbotapi.RequestFileData, configCb func(photoConfig *tgbotapi.AnimationConfig)) error {
	var fileData tgbotapi.RequestFileData = nil
	if animationFileFn != nil {
		fileData = animationFileFn()
	}

	if fileData == nil {
		return fmt.Errorf("RequestFileData is nil")
	}

	animationMsg := tgbotapi.NewAnimation(chatID, fileData)
	if configCb != nil {
		configCb(&animationMsg)
	}

	// 发送图片消息
	if _, err := bot.Send(animationMsg); err != nil {
		return err
	}

	return nil
}
