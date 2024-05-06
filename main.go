package main

import (
	"context"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/task"
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
	tk1 := task.NewTask("syncEnterprisePulls", "0 0 5 * * ?", func(ctx context.Context) error {
		return SyncEnterprisePulls()
	})
	tk2 := task.NewTask("syncEnterpriseIssues", "0 0 3 * * ?", func(ctx context.Context) error {
		return SyncEnterpriseIssues()
	})
	tk3 := task.NewTask("syncEnterpriseRepos", "0 0 1 * * ?", func(ctx context.Context) error {
		return SyncEnterpriseRepos()
	})
	tk4 := task.NewTask("cleanVerification", "0 */10 * * * *", func(ctx context.Context) error {
		return controllers.CleanVerification()
	})
	task.AddTask("syncEnterprisePulls", tk1)
	task.AddTask("syncEnterpriseIssues", tk2)
	task.AddTask("syncEnterpriseRepos", tk3)
	task.AddTask("cleanVerification", tk4)
	task.StartTask()
	defer task.StopTask()
	go models.InitIssueType()
	go utils.InitCaptchaFactory()
	beego.Run()
}
