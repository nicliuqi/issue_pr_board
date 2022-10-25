package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"issue_pr_board/controllers"
	"issue_pr_board/models"
	_ "issue_pr_board/models"
	_ "issue_pr_board/routers"
	"issue_pr_board/utils"
)

func init() {
	utils.LogInit()
}

func main() {
	tk1 := toolbox.NewTask("syncEnterprisePulls", "0 0 3 * * 6", SyncEnterprisePulls)
	tk2 := toolbox.NewTask("syncEnterpriseIssues", "0 0 3 * * ?", SyncEnterpriseIssues)
	tk3 := toolbox.NewTask("syncEnterpriseRepos", "0 0 1 * * ?", SyncEnterpriseRepos)
	tk4 := toolbox.NewTask("syncSecret", "0 0 0 * * ?", models.SyncSecret)
	tk5 := toolbox.NewTask("cleanVerification", "0 */10 * * * *", controllers.CleanVerification)
	toolbox.AddTask("syncEnterprisePulls", tk1)
	toolbox.AddTask("syncEnterpriseIssues", tk2)
	toolbox.AddTask("syncEnterpriseRepos", tk3)
	toolbox.AddTask("syncSecret", tk4)
	toolbox.AddTask("cleanVerification", tk5)
	toolbox.StartTask()
	defer toolbox.StopTask()
	beego.Run()
}
