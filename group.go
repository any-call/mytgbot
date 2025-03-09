package mytgbot

import (
	"encoding/json"
	"fmt"
	"github.com/any-call/gobase/frame/myctrl"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"net/url"
	"time"
)

type group struct {
}

func ImpGroup() group {
	return group{}
}

func (self group) LeaveChat(bot *tgbotapi.BotAPI, chatId int64) (*tgbotapi.APIResponse, error) {
	// 让机器人主动退出群聊
	leaveChat := tgbotapi.LeaveChatConfig{
		ChatID: chatId,
	}

	return bot.Request(leaveChat)
}

func (self group) LeaveChatByToken(token string, chatId int64) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/leaveChat?chat_id=%d", token, chatId)

	resp, err := http.Get(apiURL)
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
	defer func() {
		_ = resp.Body.Close()
	}()

	// 解析响应
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Println("result is :", result)
	if result["ok"].(bool) {
		return nil
	}

	return fmt.Errorf("invalid data")
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

func (self group) MuteUser(bot *tgbotapi.BotAPI, chatID int64, tgUserID int64, t time.Duration) error {
	// 设置禁言权限
	restrictConfig := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: tgUserID,
		},
		UntilDate: myctrl.ObjFun(func() int64 {
			if t == 0 {
				return 0 //永久禁言
			}
			return time.Now().Add(t).Unix()
		}),
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

func (self group) MuteUserByToken(token string, chatID int64, tgUserID int64, t time.Duration) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/restrictChatMember", token)

	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
	data.Set("user_id", fmt.Sprintf("%d", tgUserID))
	data.Set("can_send_messages", "false")
	if t != 0 {
		data.Set("until_date", fmt.Sprintf("%d", time.Now().Add(t).Unix()))
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

func (self group) UnmuteUser(bot *tgbotapi.BotAPI, chatID int64, tgUserID int64) error {
	// 设置禁言权限
	restrictConfig := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: tgUserID,
		},
		Permissions: &tgbotapi.ChatPermissions{
			CanSendMessages:       true,
			CanSendMediaMessages:  true,
			CanSendOtherMessages:  true,
			CanAddWebPagePreviews: true,
		},
	}

	// 调用 API 禁言
	_, err := bot.Request(restrictConfig)
	if err != nil {
		return err
	}
	return nil
}

func (self group) UnmuteUserByToken(token string, chatID int64, tgUserID int64) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/restrictChatMember", token)

	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
	data.Set("user_id", fmt.Sprintf("%d", tgUserID))
	data.Set("can_send_messages", "true")

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

// 临时踢出 封禁一段时间，过期后自动解封
func (self group) KickUserTemporarily(bot *tgbotapi.BotAPI, chatID int64, userID int64, duration time.Duration) error {
	untilDate := time.Now().Add(duration).Unix() // 计算解封时间

	kickConfig := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		UntilDate: untilDate, // 过期后自动解除封禁
	}

	_, err := bot.Request(kickConfig)
	if err != nil {
		return err
	}

	return nil
}

func (self group) KickUserTemporarilyByToken(token string, chatID int64, tgUserID int64, duration time.Duration) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/kickChatMember", token)

	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
	data.Set("user_id", fmt.Sprintf("%d", tgUserID))
	data.Set("until_date", fmt.Sprintf("%d", time.Now().Add(duration).Unix()))

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

// 永久踢出 永久封禁，不能再加入
func (self group) KickUserPermanently(bot *tgbotapi.BotAPI, chatID int64, userID int64) error {
	kickConfig := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		UntilDate: 0, // 0 表示永久封禁
	}

	_, err := bot.Request(kickConfig)
	if err != nil {
		return err
	}

	return nil
}
func (self group) KickUserPermanentlyByToken(token string, chatID int64, tgUserID int64) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/kickChatMember", token)

	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
	data.Set("user_id", fmt.Sprintf("%d", tgUserID))

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

// 仅踢出但可重新加入,仅踢出，用户可手动重新加入
func (self group) KickUserAllowRejoin(bot *tgbotapi.BotAPI, chatID int64, userID int64) error {
	// 先踢出用户
	kickConfig := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
	}
	_, err := bot.Request(kickConfig)
	if err != nil {
		return err
	}

	// 立即解除封禁，允许重新加入
	unbanConfig := tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		OnlyIfBanned: false, // 解除封禁，允许重新加入
	}

	_, err = bot.Request(unbanConfig)
	if err != nil {
		return err
	}

	return nil
}
func (self group) KickUserAllowRejoinByToken(token string, chatID int64, tgUserID int64) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/unbanChatMember", token)

	// 构造请求参数
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
	data.Set("user_id", fmt.Sprintf("%d", tgUserID))

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
