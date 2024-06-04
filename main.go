package main

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"issue_pr_board/config"
	"issue_pr_board/controllers"
	"issue_pr_board/models"
	_ "issue_pr_board/models"
	_ "issue_pr_board/routers"
	"issue_pr_board/utils"
	"os"
)

func init() {
	if err := config.InitAppConfig(os.Getenv("CONFIG_PATH")); err != nil {
		logs.Error("[init] Fail to init app config, err:", err)
		os.Exit(1)
	}
	dataSource := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&loc=Local", config.AppConfig.DBUsername,
		config.AppConfig.DBPassword, config.AppConfig.DBHost, config.AppConfig.DBPort, config.AppConfig.DBName,
		config.AppConfig.DBChar)
	if err := orm.RegisterDataBase("default", "mysql", dataSource); err != nil {
		logs.Error("[init] Fail to register database, err:", err)
		return
	}
	orm.RegisterModel(new(models.Pull), new(models.Issue), new(models.Repo), new(models.Verify), new(models.Label),
		new(models.IssueType))
	if err := orm.RunSyncdb("default", false, true); err != nil {
		logs.Error("[init] Fail to sync databases, err:", err)
		return
	}
}

func main() {
	go runTasks()
	go models.InitIssueType()
	go models.InitLabels()
	go utils.InitCaptcha()
	beego.ErrorController(&controllers.ErrorController{})
	beego.RunWithMiddleWares("", utils.LimitMiddleware)
}
