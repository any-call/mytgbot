package mytgbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"mime/multipart"
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

type (
	InviteTempData struct {
		InviteLink         string `json:"invite_link"`
		Name               string `json:"name"`
		ExpireDate         int    `json:"expire_date"`
		MemberLimit        int    `json:"member_limit"`
		CreatesJoinRequest bool   `json:"creates_join_request"`
	}

	UserData struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"result"`
	}

	ChatResp[T any] struct {
		Ok     bool `json:"ok"`
		Result T    `json:"result"`
	}

	// 用户信息
	TelegramUser struct {
		ID        int64  `json:"id"`
		Type      string `json:"type"`     // 必须是 "private"
		Username  string `json:"username"` // 可能为空
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	// 群组/频道信息
	TelegramChat struct {
		ID          int64  `json:"id"`
		Type        string `json:"type"`        // "group", "supergroup", "channel"
		Title       string `json:"title"`       // 群组或频道的名称
		Username    string `json:"username"`    // 可能为空
		Description string `json:"description"` // 仅频道有
		InviteLink  string `json:"invite_link"`
	}
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

func GetChatDesc(bot *tgbotapi.BotAPI, chatID int64) (tgbotapi.Chat, error) {
	return bot.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{
		ChatID: chatID,
	}})
}

func GetChatByToken(token string, groupID int64) (*TelegramChat, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%d", token, groupID)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var result ChatResp[TelegramChat]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Ok {
		return nil, fmt.Errorf("failed to get chat info")
	}

	return &result.Result, nil
}

func GetUserByToken(token string, userID int64) (*TelegramUser, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%d", token, userID)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var result ChatResp[TelegramUser]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Ok {
		return nil, fmt.Errorf("failed to get chat info")
	}

	return &result.Result, nil
}

func GenUserNameLink(userName string) string {
	return fmt.Sprintf("https://t.me/%s", userName)
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

func SendPhotoByToken(token string, toChatId int64, photoName string, photoData []byte, caption string, configFn func(values *multipart.Writer)) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", token)
	// 构造表单数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// 添加 chat_id 字段
	if err := writer.WriteField("chat_id", fmt.Sprintf("%d", toChatId)); err != nil {
		return fmt.Errorf("failed to add chat_id: %w", err)
	}

	// 添加 caption 字段（可选）
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return fmt.Errorf("failed to add caption: %w", err)
		}
	}

	if configFn != nil {
		configFn(writer) //此处可以在外转增加一些参数 ：parse_mode / reply_markup
	}

	// 添加图片数据
	part, err := writer.CreateFormFile("photo", photoName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// 写入图片字节数据
	if _, err := io.Copy(part, bytes.NewReader(photoData)); err != nil {
		return fmt.Errorf("failed to copy photo data: %w", err)
	}

	// 关闭表单
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 执行请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API returned error: %s", string(respBody))
	}

	return nil
}

func EditMessageCaption(bot *tgbotapi.BotAPI, chatId int64, editMessageID int, caption string, configFn func(editMsgConfig *tgbotapi.EditMessageCaptionConfig)) (tgbotapi.Message, error) {
	editMsg := tgbotapi.NewEditMessageCaption(chatId, editMessageID, caption)
	if configFn != nil {
		configFn(&editMsg)
	}

	return bot.Send(editMsg)
}

func EditMessage(bot *tgbotapi.BotAPI, chatId int64, editMessageID int, text string, configFn func(editMsgConfig *tgbotapi.EditMessageTextConfig)) (tgbotapi.Message, error) {
	editMsg := tgbotapi.NewEditMessageText(chatId, editMessageID, text)
	if configFn != nil {
		configFn(&editMsg)
	}

	return bot.Send(editMsg)
}

func AlertCallback(bot *tgbotapi.BotAPI, cbId string, message string, configFn func(callbackConfig *tgbotapi.CallbackConfig)) (tgbotapi.Message, error) {
	alert := tgbotapi.NewCallbackWithAlert(cbId, message)
	if configFn != nil {
		configFn(&alert)
	}

	return bot.Send(alert)
}

func PingMessage(bot *tgbotapi.BotAPI, chatID int64, messagaId int, notifaicationFlag bool) error {
	pinMsg := tgbotapi.PinChatMessageConfig{
		ChatID:              chatID,
		MessageID:           messagaId,
		DisableNotification: notifaicationFlag,
	}

	_, err := bot.Request(pinMsg)
	return err
}

func GetBotUserName(token string) (*UserData, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var result UserData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Ok {
		return nil, fmt.Errorf("failed to get bot info")
	}

	return &result, nil
}

// 生成群永久链接，这个链接中永久的，唯一的，多次生成，则新的替换成旧的 。
func CreatePermanentInviteLink(token string, chatId int64) (ret string, err error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/exportChatInviteLink", token)
	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatId))

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("消息发送失败，状态码: %d", resp.StatusCode)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 解析响应
	var result struct {
		Ok     bool   `json:"ok"`
		Result string `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// 返回邀请链接
	if !result.Ok {
		return "", fmt.Errorf("failed to get invite link")
	}

	return result.Result, nil

}

// 生成临时邀请链接 ，可为不同人生成，场景更丰富
func CreateTempInviteLink(token string, chatId int64, name string, expireDate time.Time, maxLimit int, joinCheck bool) (ret *InviteTempData, err error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/createChatInviteLink", token)
	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatId))
	data.Set("name", name)
	data.Set("expire_date", fmt.Sprintf("%d", expireDate.Unix()))
	data.Set("member_limit", fmt.Sprintf("%d", maxLimit))
	data.Set("creates_join_request", fmt.Sprintf("%t", joinCheck))

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("消息发送失败，状态码: %d", resp.StatusCode)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 解析响应
	var result struct {
		Ok     bool           `json:"ok"`
		Result InviteTempData `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// 返回邀请链接
	if !result.Ok {
		return nil, fmt.Errorf("failed to get invite link")
	}

	return &result.Result, nil

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

func WebhookHandler(w http.ResponseWriter, r *http.Request, cbFun func(update tgbotapi.Update)) {
	var update tgbotapi.Update
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "Cannot decode update", http.StatusBadRequest)
		return
	}

	if cbFun != nil {
		cbFun(update)
	}
}
