package models

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io"
	"issue_pr_board/config"
	"net/http"
)

type V8Token struct {
	AccessToken string `json:"access_token"`
}

func GetV8Token() string {
	url := fmt.Sprintf("%v?token=%v", config.AppConfig.V8Url, config.AppConfig.QueryToken)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get v8 token, err：", err)
		return ""
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting v8 token, status:", resp.Status)
		return ""
	}
	body, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of the issue, err:", err)
	}
	var token V8Token
	err = json.Unmarshal(body, &token)
	return token.AccessToken
}
