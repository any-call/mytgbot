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

func SendMessage(bot *tgbotapi.BotAPI, chatId int64, message string, configCb func(messageCfg *tgbotapi.MessageConfig)) (*tgbotapi.Message, error) {
	sendMsg := tgbotapi.NewMessage(chatId, message)
	if configCb != nil {
		configCb(&sendMsg)
	}

	return sendMessage(bot, sendMsg)
}

func SendMessageByAutoDel(bot *tgbotapi.BotAPI, chatId int64, message string, configCb func(messageCfg *tgbotapi.MessageConfig), autoDele time.Duration) error {
	msg, err := SendMessage(bot, chatId, message, configCb)
	if err != nil {
		return err
	}

	go func(bt *tgbotapi.BotAPI, msgID int, groupID int64) {
		time.Sleep(autoDele)
		_, _ = bt.Request(tgbotapi.NewDeleteMessage(groupID, msgID))
	}(bot, msg.MessageID, chatId)

	return nil
}

func SendPhoto(bot *tgbotapi.BotAPI, chatId int64, imageFileFn func() tgbotapi.RequestFileData, configCb func(photoConfig *tgbotapi.PhotoConfig)) (messageID int, imageFileId string, err error) {
	var fileData tgbotapi.RequestFileData = nil
	if imageFileFn != nil {
		fileData = imageFileFn()
	}

	if fileData == nil {
		return 0, "", fmt.Errorf("RequestFileData is nil")
	}

	photo := tgbotapi.NewPhoto(chatId, fileData)
	if configCb != nil {
		configCb(&photo)
	}

	// 发送图片消息
	uploadResp, err := bot.Send(photo)
	if err != nil {
		return 0, "", err
	}

	return uploadResp.MessageID, uploadResp.Photo[0].FileID, nil
}

func SendAnimation(bot *tgbotapi.BotAPI, chatID int64, animationFileFn func() tgbotapi.RequestFileData, configCb func(photoConfig *tgbotapi.AnimationConfig)) (messageID int, animationFileId string, err error) {
	var fileData tgbotapi.RequestFileData = nil
	if animationFileFn != nil {
		fileData = animationFileFn()
	}

	if fileData == nil {
		return 0, "", fmt.Errorf("RequestFileData is nil")
	}

	animationMsg := tgbotapi.NewAnimation(chatID, fileData)
	if configCb != nil {
		configCb(&animationMsg)
	}

	// 发送图片消息
	uploadResp, err := bot.Send(animationMsg)
	if err != nil {
		return 0, "", err
	}

	return uploadResp.MessageID, uploadResp.Animation.FileID, nil
}

// 禁言
func ForbidSpeaking(bot *tgbotapi.BotAPI, chatID int64, tgUserID int64, t time.Duration) error {
	// 禁言时长（例如禁言 1 小时）
	untilDate := time.Now().Add(t).Unix()
	// 设置禁言权限
	restrictConfig := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: tgUserID,
		},
		UntilDate: untilDate,
		Permissions: &tgbotapi.ChatPermissions{
			CanSendMessages:       false,
			CanSendMediaMessages:  false,
			CanSendOtherMessages:  false,
			CanAddWebPagePreviews: false,
		},
	}

	// 调用 API 禁言
	_, err := bot.Request(restrictConfig)
	if err != nil {
		return err
	}
	return nil
}

// 踢出群
func RemoveUser(bot *tgbotapi.BotAPI, chatID int64, tgUserID int64, t time.Duration) error {
	// 禁言时长（例如禁言 1 小时）
	var untilDate int64 = 0 //表示永久禁止加入
	if t > 0 {
		untilDate = time.Now().Add(t).Unix()
	}

	// 配置踢出用户
	kickConfig := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: tgUserID,
		},
		UntilDate: untilDate,
	}

	// 调用 API 踢出用户
	_, err := bot.Request(kickConfig)
	if err != nil {
		return err
	}

	return nil
}

func SendMessageByToken(token string, toChatId int64, message string, configFn func(values url.Values)) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", toChatId))
	data.Set("text", message)
	if configFn != nil {
		configFn(data) //此处可以在外转增加一些参数 ：parse_mode / reply_markup
	}
	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("消息发送失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// 关于markdown 格式的特别显示
func MarkdownLinkText(showText, linkText string) string {
	return fmt.Sprintf("[%s](%s)", showText, linkText)
}

func MarkdownCopyText(showText string) string {
	return fmt.Sprintf("`%s`", showText)
}

func HtmlLinkText(showText, linkText string) string {
	return fmt.Sprintf("<a href='%s'>%s</a>", linkText, showText)
}

func GenUserLink(userID int64) string {
	return fmt.Sprintf("tg://user?id=%d", userID)
}
