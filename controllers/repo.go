package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/go-playground/validator/v10"
	"io"
	"issue_pr_board/config"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"net/http"
	"strings"
)

type ReposController struct {
	BaseController
}

type QueryRepoParam struct {
	Keyword   string `validate:"max=100"`
	Sig       string `validate:"max=100"`
	Page      int    `validate:"min=1"`
	PerPage   int    `validate:"max=100"`
	Direction string `validate:"max=4"`
	Public    string `validate:"max=5"`
	Status    string `validate:"max=20"`
}

func formQueryRepoSql(q QueryRepoParam) (int64, string) {
	rawSql := "select * from repo"
	keyword := q.Keyword
	sig := q.Sig
	page := q.Page
	perPage := q.PerPage
	direction := q.Direction
	public := q.Public
	status := q.Status
	sig = utils.CheckParams(sig)
	keyword = utils.CheckParams(keyword)
	public = utils.CheckParams(public)
	status = utils.CheckParams(status)
	if keyword != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where instr (name, '%s')", strings.ToLower(keyword))
		} else {
			rawSql += fmt.Sprintf(" where instr (name, '%s')", strings.ToLower(keyword))
		}
	}
	if sig != "" {
		if len(rawSql) == 18 {
			rawSql += fmt.Sprintf(" where sig='%s'", sig)
		} else {
			rawSql += fmt.Sprintf(" and sig='%s'", sig)
		}
	}
	if public != "" {
		if len(rawSql) == 18 {
			if public == "true" {
				rawSql += fmt.Sprintf(" where public=true")
			}
			if public == "false" {
				rawSql += fmt.Sprintf(" where public=false")
			}
		} else {
			if public == "true" {
				rawSql += fmt.Sprintf(" and public=true")
			}
			if public == "false" {
				rawSql += fmt.Sprintf(" and public=false")
			}
		}
	}
	if status != "" {
		if len(rawSql) == 18 {
			if status == "开始" || status == "关闭" {
				rawSql += fmt.Sprintf(" where status='%s'", status)
			}
		} else {
			if status == "开始" || status == "关闭" {
				rawSql += fmt.Sprintf(" and status='%s'", status)
			}
		}
	}
	if direction != "desc" {
		rawSql += " order by name"
	} else {
		rawSql += " order by name desc"
	}
	var repo []models.Repo
	o := orm.NewOrm()
	count, err := o.Raw(rawSql).QueryRows(&repo)
	if err != nil {
		return 0, "select * from repo"
	}
	offset := perPage * (page - 1)
	rawSql += fmt.Sprintf(" limit %v offset %v", perPage, offset)
	return count, rawSql
}

func (c *ReposController) Get() {
	var repo []models.Repo
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	qp := QueryRepoParam{
		Keyword:   c.GetString("keyword", ""),
		Sig:       c.GetString("sig", ""),
		Page:      page,
		PerPage:   perPage,
		Direction: c.GetString("direction", ""),
		Public:    c.GetString("public", ""),
		Status:    c.GetString("status", ""),
	}
	validate := validator.New()
	validateErr := validate.Struct(qp)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	count, sql := formQueryRepoSql(qp)
	o := orm.NewOrm()
	_, err := o.Raw(sql).QueryRows(&repo)
	if err == nil {
		c.ApiDataReturn(count, page, perPage, repo)
	} else {
		c.ApiJsonReturn("查询错误", 400, err)
	}
}

func SearchRepo(name string) bool {
	o := orm.NewOrm()
	searchSql := fmt.Sprintf("select * from repo where name='%s'", name)
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}

func SearchRepoByNumber(number int) (sig string, repo string) {
	var repos []models.Repo
	o := orm.NewOrm()
	searchSql := fmt.Sprintf("select * from repo where enterprise_number=%v", number)
	_, err := o.Raw(searchSql).QueryRows(&repos)
	if err != nil {
		logs.Error(err)
		return "", ""
	} else {
		for _, r := range repos {
			return r.Sig, r.Name
		}
	}
	return
}

type RepoResponse struct {
	Id        int    `json:"id"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Public    bool   `json:"public"`
	Status    string `json:"status"`
}

func SyncRepoNumber() error {
	logs.Info("Starting to sync repos numbers...")
	page := 1
	for {
		logs.Info("Sync repos: Page", page)
		url := fmt.Sprintf("%v/enterprises/open_euler/repos?type=all&page=%v&per_page=100&access_token=%v",
			config.AppConfig.GiteeV5ApiPrefix, page, config.AppConfig.AccessToken)
		resp, err := http.Get(url)
		if err != nil {
			logs.Error("Fail to get enterprise pull requests, err：", err)
			return err
		}
		if resp.StatusCode != 200 {
			logs.Error("Get unexpected response when getting V8 enterprise repos, status:", resp.Status)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		err = resp.Body.Close()
		if err != nil {
			logs.Error("Fail to close response body of V8 enterprise repos, err：", err)
			return err
		}
		if len(string(body)) == 2 {
			break
		}
		var repos []RepoResponse
		err = json.Unmarshal(body, &repos)
		if err != nil {
			logs.Error(err)
			return nil
		}
		if len(repos) == 0 {
			break
		}
		for _, repo := range repos {
			var r models.Repo
			name := repo.FullName
			number := repo.Id
			createdAt := repo.CreatedAt
			updatedAt := repo.UpdatedAt
			public := repo.Public
			status := repo.Status
			r.Name = name
			r.EnterpriseNumber = number
			r.CreatedAt = utils.FormatTime(createdAt)
			r.UpdatedAt = utils.FormatTime(updatedAt)
			r.Public = public
			r.Status = status
			if SearchRepo(name) {
				o := orm.NewOrm()
				qs := o.QueryTable("repo")
				_, err := qs.Filter("name", name).Update(orm.Params{
					"enterprise_number": number,
					"created_at":        r.CreatedAt,
					"updated_at":        r.UpdatedAt,
					"public":            r.Public,
					"status":            r.Status,
				})
				if err != nil {
					logs.Error("Update repo enterprise number failed, err:", err)
				} else {
					logs.Info("更新仓库", name)
				}
			}
		}
		page += 1
	}
	logs.Info("Ends of repos numbers sync, wait the next time...")
	return nil
}
