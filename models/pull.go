package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"issue_pr_board/utils"
	"net/http"
	"os"
)

type Pull struct {
	Id          int    `json:"-"`
	Org         string `json:"org" orm:"size(20);null" description:"组织"`
	Repo        string `json:"repo" orm:"size(100);null" description:"仓库"`
	Ref         string `json:"ref" orm:"size(100);null" description:"目标分支"`
	Sig         string `json:"sig" orm:"size(100);null" description:"所属sig组"`
	Link        string `json:"link" orm:"size(255);null;unique" description:"链接"`
	State       string `json:"state" orm:"size(20);null" description:"状态"`
	Author      string `json:"author" orm:"size(50);null" description:"提交者"`
	Assignees   string `json:"assignees" orm:"size(255);null" description:"指派者"`
	CreatedAt   string `json:"created_at" orm:"size(20);null" description:"PR创建时间"`
	UpdatedAt   string `json:"updated_at" orm:"size(20);null" description:"PR更新时间"`
	Title       string `json:"title" orm:"type(text);null" description:"标题"`
	Description string `json:"-" orm:"type(text);null" description:"描述"`
	Labels      string `json:"labels" orm:"type(text);null" description:"标签"`
}

func init() {
	dbHost := beego.AppConfig.String("dbhost")
	dbPort := beego.AppConfig.String("dbport")
	dbUser := beego.AppConfig.String("dbuser")
	dbPassword := beego.AppConfig.String("dbpassword")
	dbName := beego.AppConfig.String("dbname")
	dbChar := beego.AppConfig.String("dbchar")
	dataSource := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName, dbChar)
	err := orm.RegisterDataBase("default", "mysql", dataSource)
	if err != nil {
		logs.Error("Fail to register database, err:", err)
		return
	}
	orm.RegisterModel(new(Pull), new(Issue), new(Repo), new(Secret), new(Verify), new(Label), new(IssueType))
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		logs.Error("Fail to sync databases, err:", err)
		return
	}
	url := fmt.Sprintf("https://gitee.com/api/v5/enterprises/open_euler/labels?access_token=%v", os.Getenv("AccessToken"))
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get enterprise labels colors, err：", err)
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting enterprise labels colors, status:", resp.Status)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of enterprise issues, err:", err)
	}
	if len(string(body)) == 2 {
	}
	labels := utils.JsonToSlice(string(body))
	for _, i := range labels {
		var lb Label
		lb.Name = i["name"].(string)
		lb.Color = i["color"].(string)
		lb.UniqueId = i["id"].(float64)
		if SearchLabel(lb.Name) {
			o := orm.NewOrm()
			qs := o.QueryTable("label")
			_, err = qs.Filter("name", lb.Name).Update(orm.Params{
				"color":     lb.Color,
				"unique_id": lb.UniqueId,
			})
			if err != nil {
				logs.Error("Update label failed, err:", err)
			}
		} else {
			o := orm.NewOrm()
			_, err = o.Insert(&lb)
			if err != nil {
				logs.Error("Insert label failed, err:", err)
			}
		}
	}
}
