package models

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/chenhg5/collection"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type Secret struct {
	Id           int    `json:"-"`
	Token        string `json:"token" orm:"size(100);null" description:"token"`
	RefreshToken string `json:"refresh_token" orm:"size(100);null" description:"refresh_token"`
	State        string `json:"state" orm:"size(10);null" description:"状态"`
}

func existToken() bool {
	o := orm.NewOrm()
	searchSql := "select * from secret"
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}

func createToken() {
	var secret Secret
	tmpToken := os.Getenv("V8Token")
	tmpRefreshToken := os.Getenv("V8RefreshToken")
	secret.Token = tmpToken
	secret.RefreshToken = tmpRefreshToken
	secret.State = "normal"
	o := orm.NewOrm()
	_, err := o.Insert(&secret)
	if err != nil {
		logs.Error("Create secret failed, err:", err)
	}
}

func updateToken() {
	var secret Secret
	o := orm.NewOrm()
	searchSql := "select * from secret where id=1"
	err := o.Raw(searchSql).QueryRow(&secret)
	if err != nil {
		logs.Error("Fail to search secret, err:", err)
		return
	}
	secret.State = "update"
	qs := o.QueryTable("secret")
	_, err = qs.Filter("id", 1).Update(orm.Params{
		"Id":           1,
		"Token":        secret.Token,
		"RefreshToken": secret.RefreshToken,
		"State":        "update",
	})
	if err != nil {
		logs.Error("Update secret state failed, err:", err)
		return
	}
	refreshToken := secret.RefreshToken
	url := fmt.Sprintf("https://gitee.com/oauth/token?grant_type=refresh_token&refresh_token=%s", refreshToken)
	payloadMap := make(map[string]interface{})
	payload := strings.NewReader(collection.Collect(payloadMap).ToJson())
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Fail to refresh token, err:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Error("Fail to close response body of refreshing token")
		}
	}(resp.Body)
	result, _ := ioutil.ReadAll(resp.Body)
	content := string(result)
	newToken := collection.Collect(content).ToMap()["access_token"]
	newRefreshToken := collection.Collect(content).ToMap()["refresh_token"]
	qs2 := o.QueryTable("secret")
	_, err = qs2.Filter("id", 1).Update(orm.Params{
		"Id":           1,
		"Token":        newToken,
		"RefreshToken": newRefreshToken,
		"State":        "normal",
	})
	if err != nil {
		logs.Error("Fail to update secret")
	}
}

func SyncSecret() error {
	if existToken() == false {
		createToken()
		logs.Info("Create token")
	} else {
		updateToken()
		logs.Info("Update token")
	}
	return nil
}

func GetV8Token(retry int) string {
	var secret Secret
	o := orm.NewOrm()
	searchSql := "select * from secret where id=1"
	err := o.Raw(searchSql).QueryRow(&secret)
	if err != nil {
		logs.Error("Fail to search secret, err:", err)
		return ""
	}
	token := secret.Token
	state := secret.State
	if state != "normal" {
		if retry == 0 {
			return ""
		}
		time.Sleep(1)
		retry -= 1
		GetV8Token(retry)
	}
	return token
}
