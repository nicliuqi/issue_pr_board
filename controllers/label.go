package controllers

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"issue_pr_board/utils"
	"net/http"
	"os"
)

type LabelsColorsController struct {
	BaseController
}

func (c *LabelsColorsController) Get() {
	url := fmt.Sprintf("https://gitee.com/api/v5/enterprises/open_euler/labels?access_token=%v", os.Getenv("AccessToken"))
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get enterprise labels colors, err：", err)
		c.ApiJsonReturn("获取企业标签颜色失败", 400, err)
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting enterprise labels colors, status:", resp.Status)
		c.ApiJsonReturn("获取企业标签颜色响应异常", 400, err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of enterprise issues, err:", err)
		c.ApiJsonReturn("关闭响应内容失败", 400, err)
	}
	if len(string(body)) == 2 {
		c.ApiJsonReturn("请求成功", 200, nil)
	}
	labels := utils.JsonToSlice(string(body))
	res := make([]map[string]interface{}, 0)
	for _, label := range labels {
		labelMap := make(map[string]interface{})
		labelMap["name"] = label["name"]
		labelMap["color"] = label["color"]
		res = append(res, labelMap)
	}
	c.ApiJsonReturn("请求成功", 200, res)
}
