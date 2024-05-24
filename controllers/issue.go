package controllers

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/go-playground/validator/v10"
	"io"
	"issue_pr_board/config"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"mime/multipart"
	"net/http"
	"strings"
)

type IssuesController struct {
	BaseController
}

type QueryIssueParam struct {
	Org        string `validate:"max=20"`
	Repo       string `validate:"max=100"`
	Sig        string `validate:"max=100"`
	State      string `validate:"max=20"`
	Number     string `validate:"max=10"`
	Author     string `validate:"max=50"`
	Assignee   string `validate:"max=50"`
	Branch     string `validate:"max=100"`
	Label      string
	Exclusion  string
	IssueState string
	IssueType  string `validate:"max=20"`
	Priority   string `validate:"max=10"`
	Search     string `validate:"max=255"`
	Sort       string `validate:"max=10"`
	Direction  string `validate:"max=4"`
	Milestone  string `validate:"max=255"`
	Page       int    `validate:"min=1"`
	PerPage    int    `validate:"max=100"`
}

func formQueryIssueSql(q QueryIssueParam) (int64, string) {
	rawSql := "select * from issue where sig != 'Private'"
	org := q.Org
	repo := q.Repo
	sig := q.Sig
	state := q.State
	number := q.Number
	author := q.Author
	assignee := q.Assignee
	branch := q.Branch
	label := q.Label
	exclusion := q.Exclusion
	issueState := q.IssueState
	issueType := q.IssueType
	priority := q.Priority
	search := q.Search
	sort := q.Sort
	direction := q.Direction
	milestone := q.Milestone
	page := q.Page
	perPage := q.PerPage
	org = utils.CheckParams(org)
	repo = utils.CheckParams(repo)
	sig = utils.CheckParams(sig)
	state = utils.CheckParams(state)
	number = utils.CheckParams(number)
	author = utils.CheckParams(author)
	assignee = utils.CheckParams(assignee)
	branch = utils.CheckParams(branch)
	label = utils.CheckParams(label)
	exclusion = utils.CheckParams(exclusion)
	issueState = utils.CheckParams(issueState)
	issueType = utils.CheckParams(issueType)
	priority = utils.CheckParams(priority)
	search = utils.CheckParams(search)
	milestone = utils.CheckMilestonesParams(milestone)
	if issueState != "" {
		issueStateSql := ""
		for index, issueStateStr := range strings.Split(issueState, ",") {
			if index == 0 {
				issueStateSql += fmt.Sprintf("issue_state='%s'", issueStateStr)
			} else {
				issueStateSql += fmt.Sprintf(" or issue_state='%s'", issueStateStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", issueStateSql)
	}
	if milestone != "" {
		milestoneSql := ""
		for index, msStr := range strings.Split(milestone, ",") {
			if index == 0 {
				milestoneSql += fmt.Sprintf("milestone='%s'", msStr)
			} else {
				milestoneSql += fmt.Sprintf(" or milestone='%s'", msStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", milestoneSql)
	}
	if assignee != "" {
		assigneeSql := ""
		for index, asStr := range strings.Split(assignee, ",") {
			if index == 0 {
				assigneeSql += fmt.Sprintf("assignee='%s'", asStr)
			} else {
				assigneeSql += fmt.Sprintf(" or assignee='%s'", asStr)
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", assigneeSql)
	}
	if author != "" {
		authorSql := ""
		for index, atStr := range strings.Split(author, ",") {
			if index == 0 {
				if strings.Contains(atStr, "@") {
					newAuthor := strings.Split(atStr, "@")[0]
					if newAuthor != "" {
						authorSql += fmt.Sprintf("reporter regexp '^%s'", newAuthor)
					}
				} else {
					authorSql += fmt.Sprintf("author='%s'", atStr)
				}
			} else {
				if strings.Contains(atStr, "@") {
					newAuthor := strings.Split(atStr, "@")[0]
					if newAuthor != "" {
						authorSql += fmt.Sprintf(" or reporter regexp '^%s'", newAuthor)
					}
				} else {
					authorSql += fmt.Sprintf(" or author='%s'", atStr)
				}
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", authorSql)
	}
	if state != "" {
		rawSql += fmt.Sprintf(" and state='%s'", state)
	}
	if org != "" {
		rawSql += fmt.Sprintf(" and org='%s'", org)
	}
	if repo != "" {
		rawSql += fmt.Sprintf(" and repo='%s'", repo)
	}
	if sig != "" {
		rawSql += fmt.Sprintf(" and sig='%s'", sig)
	}
	if number != "" {
		rawSql += fmt.Sprintf(" and number='%s'", number)
	}
	if branch != "" {
		rawSql += fmt.Sprintf(" and branch='%s'", branch)
	}
	if label != "" {
		label = strings.Replace(label, "，", ",", -1)
		for _, labelStr := range strings.Split(label, ",") {
			rawSql += fmt.Sprintf(" and find_in_set('%s', labels)", labelStr)
		}
	}
	if exclusion != "" {
		exclusion = strings.Replace(exclusion, "，", ",", -1)
		for _, exclusionStr := range strings.Split(exclusion, ",") {
			rawSql += fmt.Sprintf(" and !find_in_set('%s', labels)", exclusionStr)
		}
	}
	if issueType != "" {
		rawSql += fmt.Sprintf(" and issue_type='%s'", issueType)
	}
	if priority != "" {
		rawSql += fmt.Sprintf(" and priority='%s'", priority)
	}
	if search != "" {
		searchSql := " and concat (repo, title, number) like '%{search}%'"
		rawSql += strings.Replace(searchSql, "{search}", search, -1)
	}
	if sort != "updated_at" {
		sort = "created_at"
	}
	if direction == "asc" {
		rawSql += fmt.Sprintf(" order by %s asc", sort)
	} else {
		rawSql += fmt.Sprintf(" order by %s desc", sort)
	}
	o := orm.NewOrm()
	countSql := strings.Replace(rawSql, "*", "count(*)", -1)
	var sqlCount int
	_ = o.Raw(countSql).QueryRow(&sqlCount)
	offset := perPage * (page - 1)
	rawSql += fmt.Sprintf(" limit %v offset %v", perPage, offset)
	return int64(sqlCount), rawSql
}

func (c *IssuesController) Get() {
	var issue []models.Issue
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 10)
	qp := QueryIssueParam{
		Org:        c.GetString("org", ""),
		Repo:       c.GetString("repo", ""),
		Sig:        c.GetString("sig", ""),
		State:      c.GetString("state", ""),
		Number:     c.GetString("number", ""),
		Author:     c.GetString("author", ""),
		Assignee:   c.GetString("assignee", ""),
		Branch:     c.GetString("branch", ""),
		Label:      c.GetString("label", ""),
		Exclusion:  c.GetString("exclusion", ""),
		IssueState: c.GetString("issue_state", ""),
		IssueType:  c.GetString("issue_type", ""),
		Priority:   c.GetString("priority", ""),
		Sort:       c.GetString("sort", ""),
		Direction:  c.GetString("direction", ""),
		Search:     c.GetString("search", ""),
		Milestone:  c.GetString("milestone", ""),
		Page:       page,
		PerPage:    perPage,
	}
	validate := validator.New()
	validateErr := validate.Struct(qp)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	count, sql := formQueryIssueSql(qp)
	o := orm.NewOrm()
	_, err := o.Raw(sql).QueryRows(&issue)
	res := make([]models.Issue, 0)
	if err == nil {
		for _, i := range issue {
			reporter := i.Reporter
			if reporter != "" {
				tail := reporter[len(reporter)-1:]
				reporter = strings.Split(reporter, "@")[0] + "@***" + tail
				i.Reporter = reporter
			}
			description := i.Description
			rawDescription, err := base64.StdEncoding.DecodeString(description)
			if err != nil {
				logs.Error("Fail to decode raw description, err:", err)
				c.ApiJsonReturn("解码异常", 400, err)
			}
			i.Description = string(rawDescription)
			res = append(res, i)
		}
		c.ApiDataReturn(count, page, perPage, res)
	} else {
		c.ApiJsonReturn("查询错误", 400, err)
	}
}

type IssueNewController struct {
	BaseController
}

type NewIssueParams struct {
	Email       string `json:"email" validate:"email"`
	Code        string `json:"code" validate:"len=6"`
	ProjectId   int    `json:"project_id" validate:"max=100000000"`
	Title       string `json:"title" validate:"max=255"`
	Description string `json:"description" validate:"max=65535"`
	IssueTypeId int    `json:"issue_type_id" validate:"max=100000000"`
}

type NewIssueResponse struct {
	Id       float64 `json:"id"`
	Ident    string  `json:"ident"`
	IssueUrl string  `json:"issue_url"`
}

type NewIssueRequestBody struct {
	AccessToken string `json:"access_token"`
	ProjectID   int    `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IssueTypeId int    `json:"issue_type_id"`
}

func (c *IssueNewController) Post() {
	logs.Info("Receive a request of creating an issue")
	reqBody, err := io.ReadAll(c.Ctx.Request.Body)
	err = c.Ctx.Request.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of creating a issue, err:", err)
		c.ApiJsonReturn("Body关闭异常", 400, err)
	}
	var params NewIssueParams
	err = json.Unmarshal(reqBody, &params)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		c.ApiJsonReturn("反解析JSON异常", 400, err)
	}
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	if params.ProjectId != config.AppConfig.TestProjectId {
		c.ApiJsonReturn("禁止向非测试仓库提交issue", 400, validateErr)
	}
	addr := params.Email
	code := params.Code
	if !checkCode(addr, code) {
		c.ApiJsonReturn("验证码错误", 400, "")
	}
	var newIssueRequestBody NewIssueRequestBody
	newIssueRequestBody.AccessToken = config.AppConfig.V8Token
	newIssueRequestBody.ProjectID = params.ProjectId
	newIssueRequestBody.Title = params.Title
	annoyAddr := strings.Split(addr, "@")[0] + "@***" +
	    strings.Split(strings.Split(addr, "@")[1], "")[len(strings.Split(addr, "@")[1])-1]
	newIssueRequestBody.Description = params.Description + fmt.Sprintf("\n\n-- submitted by %v", annoyAddr)
	newIssueRequestBody.IssueTypeId = params.IssueTypeId
	fmt.Println("origin desc:", params.Description)
	fmt.Println("new desc:", newIssueRequestBody.Description)
	requestBodyByte, err := json.Marshal(newIssueRequestBody)
	if err != nil {
		logs.Error("Fail to marshal request body, err:", err)
		c.ApiJsonReturn("解析JSON异常", 400, err)
	}
	payload := strings.NewReader(string(requestBodyByte))
	enterpriseId := config.AppConfig.EnterpriseId
	url := fmt.Sprintf("%v/enterprises/%v/issues", config.AppConfig.GiteeV8ApiPrefix, enterpriseId)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Error("Fail to send post request, err:", err)
		c.ApiJsonReturn("发送请求异常", 400, err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Fail to create quick issue, err:", err)
		c.ApiJsonReturn("创建issue失败", 400, err)
	}
	if resp.StatusCode != 201 {
		c.ApiJsonReturn("创建issue失败", resp.StatusCode, resp.Body)
	}
	content, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of creating enterprise issues, err:", err)
		c.ApiJsonReturn("无法关闭创建issue的响应", 400, err)
	}
	var res NewIssueResponse
	err = json.Unmarshal(content, &res)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		c.ApiJsonReturn("反解析JSON异常", 400, err)
	}
	logs.Info("An issue has been created, ready to save the info")
	issueId := res.Id
	number := res.Ident
	issueUrl := res.IssueUrl
	result := make(map[string]interface{})
	result["issue_id"] = issueId
	result["number"] = number
	if !SearchIssueRecord(number) {
		o := orm.NewOrm()
		insertSql := fmt.Sprintf("insert into issue (state, number, reporter) values('open', '%s', '%s')",
			number, addr)
		_, err = o.Raw(insertSql).Exec()
		if err != nil {
			logs.Error("Fail to create issue with reporter:", err)
			c.ApiJsonReturn("创建issue失败", 400, err)
		} else {
			logs.Info("Save issue successfully:", number)
		}
	} else {
		o := orm.NewOrm()
		updateSql := fmt.Sprintf("update issue set reporter='%s' where number='%s'", addr, number)
		_, err = o.Raw(updateSql).Exec()
		if err != nil {
			logs.Error("Fail to update issue reporter:", err)
			c.ApiJsonReturn("更新issue失败", 400, err)
		} else {
			logs.Info("Update issue successfully:", number)
		}
	}
	cleanCode(addr, code)
	go NewIssueNotify(params.ProjectId, number, issueUrl, params.Title)
	c.ApiJsonReturn("创建成功", 200, result)
}

type AuthorsController struct {
	BaseController
}

type CommonParams struct {
	KeyWord string `validate:"max=100"`
	Page    int    `validate:"min=1"`
	PerPage int    `validate:"max=100"`
}

func (c *AuthorsController) Get() {
	var issue []models.Issue
	var issue2 []models.Issue
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct author from issue union select distinct reporter from issue order by author"
	count, err := o.Raw(sql).QueryRows(&issue)
	if err != nil {
		c.ApiJsonReturn("查询错误", 400, err)
	}
	separateSql := sql + fmt.Sprintf(" limit %v offset %v", perPage, offset)
	_, err = o.Raw(separateSql).QueryRows(&issue2)
	if err != nil {
		c.ApiJsonReturn("分页查询错误", 400, err)
	}
	res := make([]string, 0)
	for _, i := range issue2 {
		author := i.Author
		if strings.Contains(author, "@") {
			author = strings.Split(author, "@")[0] + "@***" + author[len(author)-1:]
		}
		res = append(res, author)
	}
	if keyWord == "" {
		c.ApiDataReturn(count, page, perPage, res)
	} else {
		newRes := make([]string, 0)
		for _, j := range issue {
			author := j.Author
			if strings.Contains(author, "@") {
				author = strings.Split(author, "@")[0] + "@***" + author[len(author)-1:]
			}
			if strings.Contains(strings.ToLower(author), strings.ToLower(keyWord)) {
				newRes = append(newRes, author)
			}
		}
		count = int64(len(newRes))
		finalRes := make([]string, 0)
		if offset > int(count) {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		if int(count) > offset && int(count) < perPage+offset {
			c.ApiDataReturn(count, page, perPage, newRes[offset:])
		}
		if int(count) == 0 {
			c.ApiDataReturn(count, page, perPage, finalRes)
		}
		c.ApiDataReturn(count, page, perPage, newRes[offset:offset+perPage])
	}
}

type AssigneesController struct {
	BaseController
}

func (c *AssigneesController) Get() {
	var issue []models.Issue
	var issue2 []models.Issue
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct assignee from issue order by assignee"
	count, err := o.Raw(sql).QueryRows(&issue)
	if err != nil {
		c.ApiJsonReturn("查询错误", 400, err)
	}
	separateSql := sql + fmt.Sprintf(" limit %v offset %v", perPage, offset)
	_, err = o.Raw(separateSql).QueryRows(&issue2)
	if err == nil {
		res := make([]string, 0)
		for _, i := range issue2 {
			if i.Assignee != "" {
				res = append(res, i.Assignee)
			}
		}
		if keyWord == "" {
			c.ApiDataReturn(count, page, perPage, res)
		} else {
			newRes := make([]string, 0)
			for _, j := range issue {
				if strings.Contains(strings.ToLower(j.Assignee), strings.ToLower(keyWord)) {
					newRes = append(newRes, j.Assignee)
				}
			}
			count = int64(len(newRes))
			finalRes := make([]string, 0)
			if offset > int(count) {
				c.ApiDataReturn(count, page, perPage, finalRes)
			}
			if int(count) > offset && int(count) < perPage+offset {
				c.ApiDataReturn(count, page, perPage, newRes[offset:])
			}
			if int(count) == 0 {
				c.ApiDataReturn(count, page, perPage, finalRes)
			}
			c.ApiDataReturn(count, page, perPage, newRes[offset:offset+perPage])
		}
	}
}

type BranchesController struct {
	BaseController
}

func (c *BranchesController) Get() {
	var issue []models.Issue
	var issue2 []models.Issue
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct branch from issue order by branch"
	count, err := o.Raw(sql).QueryRows(&issue)
	if err != nil {
		c.ApiJsonReturn("查询错误", 400, err)
	}
	separateSql := sql + fmt.Sprintf(" limit %v offset %v", perPage, offset)
	_, err = o.Raw(separateSql).QueryRows(&issue2)
	if err == nil {
		res := make([]string, 0)
		for _, i := range issue2 {
			if i.Branch == "" {
				count -= 1
			} else {
				res = append(res, i.Branch)
			}
		}
		if keyWord == "" {
			c.ApiDataReturn(count, page, perPage, res)
		} else {
			newRes := make([]string, 0)
			for _, j := range issue {
				if strings.Contains(strings.ToLower(j.Branch), strings.ToLower(keyWord)) {
					newRes = append(newRes, j.Branch)
				}
			}
			count = int64(len(newRes))
			finalRes := make([]string, 0)
			if offset > int(count) {
				c.ApiDataReturn(count, page, perPage, finalRes)
			}
			if int(count) > offset && int(count) < perPage+offset {
				c.ApiDataReturn(count, page, perPage, newRes[offset:])
			}
			if int(count) == 0 {
				c.ApiDataReturn(count, page, perPage, finalRes)
			}
			c.ApiDataReturn(count, page, perPage, newRes[offset:offset+perPage])
		}
	}
}

type MilestonesController struct {
	BaseController
}

func (c *MilestonesController) Get() {
	var issue []models.Issue
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckMilestonesParams(keyWord)
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	o := orm.NewOrm()
	var sql string
	sql = "select distinct milestone from issue order by milestone"
	_, err := o.Raw(sql).QueryRows(&issue)
	if err == nil {
		res := make([]string, 0)
		for _, i := range issue {
			if i.Milestone == "" {
				continue
			}
			for _, j := range strings.Split(i.Milestone, ",") {
				if utils.InMap(utils.ConvertStrSlice2Map(res), j) {
					continue
				}
				if keyWord == "" {
					res = append(res, j)
				} else {
					if strings.Contains(strings.ToLower(j), strings.ToLower(keyWord)) {
						res = append(res, j)
					}
				}
			}
		}
		count := int64(len(res))
		offset := perPage * (page - 1)
		resp := make([]string, 0)
		if offset > int(count) {
			c.ApiDataReturn(count, page, perPage, resp)
		}
		if int(count) > offset && int(count) < perPage+offset {
			c.ApiDataReturn(count, page, perPage, res[offset:])
		}
		if int(count) == 0 {
			c.ApiDataReturn(count, page, perPage, resp)
		}
		c.ApiDataReturn(count, page, perPage, res[offset:offset+perPage])
	}
}

type LabelsController struct {
	BaseController
}

func (c *LabelsController) Get() {
	var issue []models.Issue
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	keyWord = utils.CheckParams(keyWord)
	params := CommonParams{
		KeyWord: keyWord,
		Page:    page,
		PerPage: perPage,
	}
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	o := orm.NewOrm()
	var sql string
	sql = "select distinct labels from issue order by labels"
	_, err := o.Raw(sql).QueryRows(&issue)
	if err == nil {
		res := make([]string, 0)
		for _, i := range issue {
			if i.Labels == "" {
				continue
			}
			for _, j := range strings.Split(i.Labels, ",") {
				if utils.InMap(utils.ConvertStrSlice2Map(res), j) {
					continue
				}
				if keyWord == "" {
					res = append(res, j)
				} else {
					if strings.Contains(strings.ToLower(j), strings.ToLower(keyWord)) {
						res = append(res, j)
					}
				}

			}
		}
		count := int64(len(res))
		offset := perPage * (page - 1)
		resp := make([]string, 0)
		if offset > int(count) {
			c.ApiDataReturn(count, page, perPage, resp)
		}
		if int(count) > offset && int(count) < perPage+offset {
			c.ApiDataReturn(count, page, perPage, res[offset:])
		}
		if int(count) == 0 {
			c.ApiDataReturn(count, page, perPage, resp)
		}
		c.ApiDataReturn(count, page, perPage, res[offset:offset+perPage])
	}
}

type TypesController struct {
	BaseController
}

type QueryIssueTypesParam struct {
	Name         string `validate:"max=50"`
	Platform     string `validate:"max=50"`
	Organization string `validate:"max=50"`
}

func formQueryIssueTypesSql(q QueryIssueTypesParam) string {
	rawSql := "select * from issue_type"
	name := q.Name
	platform := q.Platform
	organization := q.Organization
	name = utils.CheckParams(name)
	platform = utils.CheckParams(platform)
	organization = utils.CheckParams(organization)
	if name != "" {
		if len(rawSql) == 24 {
			rawSql += fmt.Sprintf(" where name='%v'", name)
		} else {
			rawSql += fmt.Sprintf(" and name='%v'", name)
		}
	}
	if platform != "" {
		if len(rawSql) == 24 {
			rawSql += fmt.Sprintf(" where platform='%v'", platform)
		} else {
			rawSql += fmt.Sprintf(" and platform='%v'", platform)
		}
	}
	if organization != "" {
		if len(rawSql) == 24 {
			rawSql += fmt.Sprintf(" where organization='%v'", organization)
		} else {
			rawSql += fmt.Sprintf(" and organization='%v'", organization)
		}
	}
	return rawSql
}

func (c *TypesController) Get() {
	var issueTypes []models.IssueType
	qp := QueryIssueTypesParam{
		Name:         c.GetString("name", ""),
		Platform:     c.GetString("platform", ""),
		Organization: c.GetString("organization", ""),
	}
	validate := validator.New()
	validateErr := validate.Struct(qp)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	o := orm.NewOrm()
	sql := formQueryIssueTypesSql(qp)
	_, err := o.Raw(sql).QueryRows(&issueTypes)
	if err != nil {
		logs.Error("Fail to query issue types, err:", err)
		c.ApiJsonReturn("请求失败", 400, err)
	}
	c.ApiJsonReturn("请求成功", 200, issueTypes)
}

type UploadImageController struct {
	BaseController
}

type UploadImageRequestBody struct {
	Base64 string `json:"base64"`
}

type UploadImageResponse struct {
	Success bool         `json:"success"`
	File    FileResponse `json:"file"`
	Message string       `json:"message"`
}

type FileResponse struct {
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Url      string `json:"url"`
}

func (c *UploadImageController) Post() {
	logs.Info("Ready to upload a image")
	file, _, err := c.GetFile("file")
	defer func(file multipart.File) {
		err = file.Close()
		if err != nil {
			logs.Error(err)
			c.ApiJsonReturn("关闭File失败", 400, err)
		}
	}(file)
	if err != nil {
		c.ApiJsonReturn("读取File失败", 400, err)
	}
	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		return
	}
	bufWriter := bufio.NewWriter(buf)
	content, err := io.ReadAll(buf)
	if err != nil {
		logs.Error("Cannot read buf of the uploaded file")
		c.ApiJsonReturn("上传失败", 400, err)
	}
	encoder := base64.NewEncoder(base64.StdEncoding, bufWriter)
	_, err = encoder.Write(content)
	if err != nil {
		c.ApiJsonReturn("base64解码失败", 400, err)
	}
	encodedString := string(buf.Bytes())
	var uploadImageRequestBody UploadImageRequestBody
	uploadImageRequestBody.Base64 = fmt.Sprintf("data:image/png;base64,%s", encodedString)
	requestBodyByte, err := json.Marshal(uploadImageRequestBody)
	if err != nil {
		logs.Error("Fail to marshal request body, err:", err)
		c.ApiJsonReturn("解析JSON异常", 400, err)
	}
	payload := strings.NewReader(string(requestBodyByte))
	token := config.AppConfig.V8Token
	if token == "" {
		logs.Warn("Cannot get a valid V8 access token")
		c.ApiJsonReturn("认证失败", 401, "")
	}
	enterpriseId := config.AppConfig.EnterpriseId
	url := fmt.Sprintf("%v/enterprises/%v/attach_files/upload_with_base_64?access_token=%s",
		config.AppConfig.GiteeV8ApiPrefix, enterpriseId, token)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Error("Fail to send post request, err:", err)
		c.ApiJsonReturn("发送请求异常", 400, err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Fail to upload file, err:", err)
		c.ApiJsonReturn("发送Post请求失败", 400, err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logs.Error("Fail to close response body of uploading file")
			c.ApiJsonReturn("Body关闭异常", 400, err)
		}
	}(resp.Body)
	result, _ := io.ReadAll(resp.Body)
	var uploadImageResponse UploadImageResponse
	err = json.Unmarshal(result, &uploadImageResponse)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		c.ApiJsonReturn("反解析JSON异常", 400, err)
	}
	if uploadImageResponse.Success == true {
		c.ApiJsonReturn("上传成功", 200, uploadImageResponse.File)
	}
	c.ApiJsonReturn("上传失败", 400, uploadImageResponse.Message)
}

func SearchIssueRecord(number string) bool {
	o := orm.NewOrm()
	searchSql := fmt.Sprintf("select * from issue where number='%s'", number)
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}

func GetIssuePriority(priorityNum float64) string {
	switch priorityNum {
	case 0:
		return "不指定"
	case 1:
		return "不重要"
	case 2:
		return "次要"
	case 3:
		return "主要"
	case 4:
		return "严重"
	default:
		return "不指定"
	}
}

type NotifyConf struct {
	Sigs []struct {
		Name      string   `json:"name"`
		Receivers []string `json:"receivers"`
	}
	Repos []struct {
		Name      string   `json:"name"`
		Receivers []string `json:"receivers"`
	}
}

func NewIssueNotify(enterpriseNumber int, number, link, title string) {
	sigName, repoName := SearchRepoByNumber(enterpriseNumber)
	if sigName == "" || repoName == "" {
		return
	}

	var notifyConf = &NotifyConf{}
	err := config.LoadFromYaml("conf/new_issue_notify.yaml", notifyConf)
	if err != nil {
		logs.Error(err)
		return
	}
	sigs := notifyConf.Sigs
	repos := notifyConf.Repos

	for _, sig := range sigs {
		name := sig.Name
		if name != sigName {
			continue
		} else {
			receivers := sig.Receivers
			for _, receiver := range receivers {
				ep := utils.EmailParams{Receiver: receiver, Repo: repoName, Number: number, Link: link}
				err = utils.SendNewIssueNotifyEmail(ep)
				if err != nil {
					logs.Error(err)
				}
			}
			break
		}
	}
	for _, repo := range repos {
		name := repo.Name
		if name != repoName {
			continue
		} else {
			receivers := repo.Receivers
			for _, receiver := range receivers {
				ep := utils.EmailParams{Receiver: receiver, Repo: repoName, Number: number, Link: link}
				err = utils.SendNewIssueNotifyEmail(ep)
				if err != nil {
					logs.Error(err)
				}
			}
			break
		}
	}
}
