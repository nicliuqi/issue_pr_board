// Package routers @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"github.com/astaxie/beego"
	"issue_pr_board/controllers"
)

func init() {
	beego.Router("/attachment", &controllers.UploadAttachmentController{})
	beego.Router("/colors", &controllers.LabelsColorsController{})
	beego.Router("/hooks", &controllers.HooksController{})
	beego.Router("/image", &controllers.UploadImageController{})
	beego.Router("/issues", &controllers.IssuesController{})
	beego.Router("/issues/assignees", &controllers.AssigneesController{})
	beego.Router("/issues/authors", &controllers.AuthorsController{})
	beego.Router("/issues/branches", &controllers.BranchesController{})
	beego.Router("/issues/labels", &controllers.LabelsController{})
	beego.Router("/issues/types", &controllers.TypesController{})
	beego.Router("/repos", &controllers.ReposController{})
	beego.Router("/pulls", &controllers.PullsController{})
	beego.Router("/pulls/authors", &controllers.PullsAuthorsController{})
	beego.Router("/pulls/assignees", &controllers.PullsAssigneesController{})
	beego.Router("/pulls/labels", &controllers.PullsLabelsController{})
	beego.Router("/pulls/refs", &controllers.PullsRefsController{})
	beego.Router("/pulls/sigs", &controllers.PullsSigsController{})
	beego.Router("/pulls/repos", &controllers.PullsReposController{})
	beego.Router("/verify", &controllers.VerifyController{})
}
