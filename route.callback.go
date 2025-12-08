package mytgbot

import (
	"fmt"
	"github.com/any-call/gobase/util/myconv"
	"github.com/any-call/gobase/util/mylog"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

type PaginatedResult struct {
	TextItems  []string
	ButtonRows [][]tgbotapi.InlineKeyboardButton
}

type CBData struct {
	path string
	data string
}

func NewCBData(path string, value ...string) *CBData {
	m := &CBData{
		path: path,
	}
	m.SetData(value...)
	return m
}

// Encode
func (c CBData) Encode() string {
	parts := []string{
		"p:" + c.path,
		"d:" + c.data,
	}

	ret := strings.Join(parts, ";")
	mylog.Info("encode is :", ret)
	return ret
}

func (c CBData) PushByData(p string, value ...string) CBData {
	if p != "" {
		if c.path == "" {
			c.path = p
		} else {
			c.path = c.path + "," + p
		}
	}

	c.SetData(value...)
	return c
}

// 回退到根节点
func (c CBData) PopRootByData(v ...string) CBData {
	c.path = c.Root()
	c.SetData(v...)
	return c
}

func (c CBData) PopByData(v ...string) CBData {
	list := strings.Split(c.path, ",")
	if len(list) > 1 {
		c.path = strings.Join(list[:len(list)-1], ",")
	}
	c.SetData(v...)
	return c
}

func (c CBData) PopSpecPathByData(path string, v ...string) CBData {
	list := strings.Split(c.path, ",")
	for i := len(list) - 1; i >= 0; i-- {
		if list[i] == path {
			c.path = strings.Join(list[:i+1], ",")
			break
		}
	}

	c.SetData(v...)
	return c
}

func (c CBData) NextPageByData(v ...string) CBData {
	intV, _ := myconv.StrToNum[int](c.GetData()[0])
	param := []string{fmt.Sprintf("%d", intV+1)}
	param = append(param, v...)
	c.SetData(param...)
	return c
}

func (c CBData) PrevPageByData(v ...string) CBData {
	intV, _ := myconv.StrToNum[int](c.GetData()[0])
	if intV > 1 {
		intV = intV - 1
	} else if intV <= 0 {
		intV = 1
	}
	param := []string{fmt.Sprintf("%d", intV)}
	param = append(param, v...)
	c.SetData(param...)
	return c
}

func (c *CBData) SetData(v ...string) {
	c.data = strings.Join(v, ",")
}

func (c *CBData) GetData() []string {
	if c.data != "" {
		return strings.Split(c.data, ",")
	}
	return []string{}
}

func (c *CBData) IsEmptyData() bool {
	return c.data == ""
}

func (c *CBData) PathLen() int {
	return len(strings.Split(c.path, ","))
}

func (c *CBData) BackButton(name string, value ...string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(name, c.PopByData(value...).Encode())
}

func (c *CBData) BackToRootButton(name string, value ...string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(name, c.PopRootByData(value...).Encode())
}

// 当前 Action（当前页面）= path 末尾
func (c *CBData) Action() string {
	if c.path == "" {
		return ""
	}
	parts := strings.Split(c.path, ",")
	return parts[len(parts)-1]
}

// 主页入口（第一个节点）
func (c *CBData) Root() string {
	if c.path == "" {
		return ""
	}
	parts := strings.Split(c.path, ",")
	return parts[0]
}

// Decode 将 base64 字符串解码并解 gzip，还原为 CBData
func Decode(str string) (*CBData, error) {
	c := &CBData{}
	for _, part := range strings.Split(str, ";") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "p":
			c.path = kv[1]
		case "d":
			c.data = kv[1]
		}
	}
	return c, nil
}

func BuildBackList(data CBData,
	backFun func() (name string, value []string),
	bRootFun func() (name string, value []string)) []tgbotapi.InlineKeyboardButton {
	var backName, backRootName string
	var backValue, backRootValue []string
	if backFun != nil {
		backName, backValue = backFun()
	}

	if bRootFun != nil {
		backRootName, backRootValue = bRootFun()
	}

	var ret []tgbotapi.InlineKeyboardButton
	if data.PathLen() > 2 {
		ret = append(ret, tgbotapi.NewInlineKeyboardButtonData(backName, data.PopByData(backValue...).Encode()),
			tgbotapi.NewInlineKeyboardButtonData(backRootName, data.PopRootByData(backRootValue...).Encode()))
	} else {
		ret = append(ret, tgbotapi.NewInlineKeyboardButtonData(backRootName, data.PopRootByData(backRootValue...).Encode()))
	}
	return ret
}

func BuildBackListRow(data CBData,
	backFun func() (name string, value []string),
	bRootFun func() (name string, value []string)) [][]tgbotapi.InlineKeyboardButton {
	var backName, backRootName string
	var backValue, backRootValue []string
	if backFun != nil {
		backName, backValue = backFun()
	}

	if bRootFun != nil {
		backRootName, backRootValue = bRootFun()
	}

	var ret [][]tgbotapi.InlineKeyboardButton
	if data.PathLen() > 2 {
		ret = append(ret, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(backName, data.PopByData(backValue...).Encode())),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(backRootName, data.PopRootByData(backRootValue...).Encode())))
	} else {
		ret = append(ret, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(backRootName, data.PopRootByData(backRootValue...).Encode())))
	}
	return ret
}

func BuildPaginatedList[T any, V string | tgbotapi.InlineKeyboardButton](
	list []T,
	currPage, totalPage int,
	cbItemFn func(index int, item T) V,
	cbPageBtn func() (prev, next tgbotapi.InlineKeyboardButton),
) PaginatedResult {
	var res PaginatedResult

	for i, item := range list {
		v := cbItemFn(i, item)

		switch val := any(v).(type) {
		case string:
			res.TextItems = append(res.TextItems, val)
		case tgbotapi.InlineKeyboardButton:
			// 按钮放一行
			res.ButtonRows = append(res.ButtonRows, tgbotapi.NewInlineKeyboardRow(val))
		}
	}

	// 分页按钮
	if cbPageBtn != nil {
		prevBtn, nextBtn := cbPageBtn()
		if totalPage > 1 {
			//有多页情况
			if currPage == 1 { //第一页
				res.ButtonRows = append(res.ButtonRows, tgbotapi.NewInlineKeyboardRow(nextBtn))
			} else if currPage == totalPage { //最后一页
				res.ButtonRows = append(res.ButtonRows, tgbotapi.NewInlineKeyboardRow(prevBtn))
			} else { //中间页
				res.ButtonRows = append(res.ButtonRows, tgbotapi.NewInlineKeyboardRow(prevBtn, nextBtn))
			}
		}
	}

	return res
}
