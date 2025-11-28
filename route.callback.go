package mytgbot

import (
	"fmt"
	"github.com/any-call/gobase/util/myconv"
	"strings"
)

type CBData struct {
	path string
	data string
}

// Encode
func (c CBData) Encode() string {
	parts := []string{
		"p:" + c.path,
		"d:" + c.data,
	}

	return strings.Join(parts, ";")
}

func (c CBData) GetCopy() CBData {
	return c
}

func (c CBData) PushPathAndEncode(p string, value string) string {
	c.PushPath(p, value)
	return c.Encode()
}

func (c *CBData) PushPath(p string, value string) {
	if c.path == "" {
		c.path = p
	} else {
		c.path = c.path + "," + p
	}
	c.data = value
}

// PopPath 弹出 Path 栈的最后一个元素，返回该元素，同时更新 Path
func (c *CBData) PopPath() string {
	if c.path == "" {
		return ""
	}
	parts := strings.Split(c.path, ",")
	last := parts[len(parts)-1]
	if len(parts) > 1 {
		c.path = strings.Join(parts[:len(parts)-1], ",")
	} else {
		c.path = ""
	}
	return last
}

func (c *CBData) SetData(v string) {
	c.data = v
}

func (c *CBData) SetListData(v ...string) {
	if v != nil || len(v) > 0 {
		c.data = strings.Join(v, ",")
	}
}

func (c *CBData) IsEmptyData() bool {
	return c.data == ""
}

func (c *CBData) GetData() string {
	return c.data
}

func (c *CBData) GetListData() []string {
	if c.data != "" {
		return strings.Split(c.data, ",")
	}
	return []string{}
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

func BackToRoot(data CBData) string {
	data.path = data.Root()
	data.data = ""
	return data.Encode()
}

func Back(c CBData) string {
	c.PopPath()
	c.data = ""
	return c.Encode()
}

func BackWithData(c CBData, value string) string {
	c.PopPath()
	c.data = value
	return c.Encode()
}

func NextPage(c CBData) string {
	intV, _ := myconv.StrToNum[int](c.data)
	c.data = fmt.Sprintf("%d", intV+1)
	return c.Encode()
}

func NextPageWithData(c CBData, v string) string {
	intV, _ := myconv.StrToNum[int](strings.Split(c.data, ",")[0])
	c.data = fmt.Sprintf("%d,%s", intV+1, v)
	return c.Encode()
}

func PreviousPage(c CBData) string {
	intV, _ := myconv.StrToNum[int](c.data)
	if intV > 1 {
		intV = intV - 1
	} else if intV <= 0 {
		intV = 1
	}
	c.data = fmt.Sprintf("%d", intV)
	return c.Encode()
}

func PreviousPageWithData(c CBData, v string) string {
	intV, _ := myconv.StrToNum[int](strings.Split(c.data, ",")[0])
	if intV > 1 {
		intV = intV - 1
	} else if intV <= 0 {
		intV = 1
	}
	c.data = fmt.Sprintf("%d,%s", intV, v)
	return c.Encode()
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
