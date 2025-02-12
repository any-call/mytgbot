package mytgbot

import (
	"testing"
	"time"
)

func TestCreateInviteLink(t *testing.T) {
	if ret, err := CreateTempInviteLink("71330734343434",
		-1002493972175, "test", time.Now().Add(time.Second*100), 2, false); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("invite link is :", ret)
	}
}

func TestGetBotUserName(t *testing.T) {
	user, err := GetBotUserName("")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("user is :", user.Result.Username)
}
