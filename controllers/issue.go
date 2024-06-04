package controllers

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/go-playground/validator/v10"

	"issue_pr_board/config"
	"issue_pr_board/models"
	"issue_pr_board/utils"
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
	Page       int    `validate:"min=1,max=1000000"`
	PerPage    int    `validate:"min=1,max=100"`
}

func formQueryIssueSql(q QueryIssueParam) (int64, string, []string) {
	sqlParams := make([]string, 0, 0)
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
				issueStateSql += fmt.Sprintf("issue_state=?")
			} else {
				issueStateSql += fmt.Sprintf(" or issue_state=?")
			}
			sqlParams = append(sqlParams, issueStateStr)
		}
		rawSql += fmt.Sprintf(" and (%s)", issueStateSql)
	}
	if milestone != "" {
		milestoneSql := ""
		for index, msStr := range strings.Split(milestone, ",") {
			if index == 0 {
				milestoneSql += fmt.Sprintf("milestone=?")
			} else {
				milestoneSql += fmt.Sprintf(" or milestone=?")
			}
			sqlParams = append(sqlParams, msStr)
		}
		rawSql += fmt.Sprintf(" and (%s)", milestoneSql)
	}
	if assignee != "" {
		assigneeSql := ""
		for index, asStr := range strings.Split(assignee, ",") {
			if index == 0 {
				assigneeSql += fmt.Sprintf("assignee=?")
			} else {
				assigneeSql += fmt.Sprintf(" or assignee=?")
			}
			sqlParams = append(sqlParams, asStr)
		}
		rawSql += fmt.Sprintf(" and (%s)", assigneeSql)
	}
	if author != "" {
		authorSql := ""
		for index, atStr := range strings.Split(author, ",") {
			if index == 0 {
				if strings.Contains(atStr, "***@") {
					newAuthor := strings.Replace(atStr, "***", "%", 1)
					authorSql += fmt.Sprintf("reporter like ? ")
					sqlParams = append(sqlParams, newAuthor)
				} else {
					authorSql += fmt.Sprintf("author=?")
					sqlParams = append(sqlParams, atStr)
				}
			} else {
				if strings.Contains(atStr, "***@") {
					newAuthor := strings.Replace(atStr, "***", "%", 1)
					authorSql += fmt.Sprintf(" or reporter like ?")
					sqlParams = append(sqlParams, newAuthor)
				} else {
					authorSql += fmt.Sprintf(" or author=?")
					sqlParams = append(sqlParams, atStr)
				}
			}
		}
		rawSql += fmt.Sprintf(" and (%s)", authorSql)
	}
	if state != "" {
		rawSql += fmt.Sprintf(" and state=?")
		sqlParams = append(sqlParams, state)
	}
	if org != "" {
		rawSql += fmt.Sprintf(" and org=?")
		sqlParams = append(sqlParams, org)
	}
	if repo != "" {
		rawSql += fmt.Sprintf(" and repo=?")
		sqlParams = append(sqlParams, repo)
	}
	if sig != "" {
		rawSql += fmt.Sprintf(" and sig=?")
		sqlParams = append(sqlParams, sig)
	}
	if number != "" {
		rawSql += fmt.Sprintf(" and number=?")
		sqlParams = append(sqlParams, number)
	}
	if branch != "" {
		rawSql += fmt.Sprintf(" and branch=?")
		sqlParams = append(sqlParams, branch)
	}
	if label != "" {
		label = strings.Replace(label, "，", ",", -1)
		for _, labelStr := range strings.Split(label, ",") {
			rawSql += fmt.Sprintf(" and find_in_set(?, labels)")
			sqlParams = append(sqlParams, labelStr)
		}
	}
	if exclusion != "" {
		exclusion = strings.Replace(exclusion, "，", ",", -1)
		for _, exclusionStr := range strings.Split(exclusion, ",") {
			rawSql += fmt.Sprintf(" and !find_in_set(?, labels)")
			sqlParams = append(sqlParams, exclusionStr)
		}
	}
	if issueType != "" {
		rawSql += fmt.Sprintf(" and issue_type=?")
		sqlParams = append(sqlParams, issueType)
	}
	if priority != "" {
		rawSql += fmt.Sprintf(" and priority=?")
		sqlParams = append(sqlParams, priority)
	}
	if search != "" {
		rawSql += " and concat (repo, title, number) like ?"
		search = "%" + search + "%"
		sqlParams = append(sqlParams, search)
	}
	if sort != "updated_at" {
		if direction == "asc" {
			rawSql += " order by created_at asc"
		} else {
			rawSql += " order by created_at desc"
		}
	} else {
		if direction == "asc" {
			rawSql += " order by updated_at asc"
		} else {
			rawSql += " order by updated_at desc"
		}
	}
	o := orm.NewOrm()
	countSql := strings.Replace(rawSql, "*", "count(*)", -1)
	var sqlCount int
	_ = o.Raw(countSql, sqlParams).QueryRow(&sqlCount)
	offset := perPage * (page - 1)
	rawSql += " limit ? offset ?"
	sqlParams = append(sqlParams, strconv.Itoa(perPage), strconv.Itoa(offset))
	return int64(sqlCount), rawSql, sqlParams
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
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	count, sql, sqlParams := formQueryIssueSql(qp)
	o := orm.NewOrm()
	if _, err := o.Raw(sql, sqlParams).QueryRows(&issue); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	res := make([]models.Issue, 0)
	for _, i := range issue {
		reporter := i.Reporter
		if reporter != "" {
			reporter = strings.Split(strings.Split(reporter, "@")[0], "")[0] + "***@" +
				strings.Split(reporter, "@")[1]
			i.Reporter = reporter
		}
		description := i.Description
		rawDescription, err := base64.StdEncoding.DecodeString(description)
		if err != nil {
			logs.Error("Fail to decode raw description, err:", err)
			continue
		}
		i.Description = string(rawDescription)
		res = append(res, i)
	}
	c.ApiDataReturn(count, page, perPage, res)
}

type IssueNewController struct {
	BaseController
}

type NewIssueParams struct {
	Email       string `json:"email" validate:"email"`
	Code        string `json:"code" validate:"len=6"`
	Repo        string `json:"repo" validate:"max=100"`
	Title       string `json:"title" validate:"max=191"`
	Description string `json:"description" validate:"max=65535"`
	IssueTypeId int    `json:"issue_type_id" validate:"min=1,max=100000000"`
	Privacy     bool   `json:"privacy"`
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
	if err = c.Ctx.Request.Body.Close(); err != nil {
		logs.Error("Fail to close request body of creating a issue, err:", err)
		c.ApiJsonReturn("Fail to release content of request body", http.StatusBadRequest, nil)
	}

	var params NewIssueParams
	if err = json.Unmarshal(reqBody, &params); err != nil {
		logs.Error("Fail to unmarshal request to json, err:", err)
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	projectId := models.GetProjectIdByRepoName(params.Repo)
	if projectId != config.AppConfig.TestProjectId {
		c.ApiJsonReturn("Forbidden to submit issues to non test repository", http.StatusBadRequest, nil)
	}
	addr := strings.ToLower(params.Email)
	annoyAddr := strings.Split(strings.Split(addr, "@")[0], "")[0] + "***@" + strings.Split(addr, "@")[1]
	code := params.Code
	privacy := params.Privacy
	if privacy != true {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	logs.Info(fmt.Sprintf("User whose email address is %v had agreed the privacy to submit an issue", annoyAddr))
	if !models.CheckCode(addr, code) {
		c.ApiJsonReturn("Invalid verification", http.StatusBadRequest, nil)
	}

	var newIssueRequestBody NewIssueRequestBody
	newIssueRequestBody.AccessToken = config.AppConfig.V8Token
	newIssueRequestBody.ProjectID = projectId
	newIssueRequestBody.Title = params.Title
	newIssueRequestBody.Description = params.Description + fmt.Sprintf("\n\n`-- submitted by %v`", annoyAddr)
	newIssueRequestBody.IssueTypeId = params.IssueTypeId
	requestBodyByte, err := json.Marshal(newIssueRequestBody)
	if err != nil {
		logs.Error("Fail to marshal request body, err:", err)
		c.ApiJsonReturn("Fail to marshal request body", http.StatusBadRequest, nil)
	}
	payload := strings.NewReader(string(requestBodyByte))

	enterpriseId := config.AppConfig.EnterpriseId
	url := fmt.Sprintf("%v/enterprises/%v/issues", config.AppConfig.GiteeV8ApiPrefix, enterpriseId)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Error("Fail to send post request, err:", err)
		c.ApiJsonReturn("Fail to create an issue", http.StatusBadRequest, nil)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Fail to create quick issue, err:", err)
		c.ApiJsonReturn("Fail to create an issue", http.StatusBadRequest, nil)
	}
	if resp.StatusCode != http.StatusCreated {
		logs.Error("Get unexpected status when creating an issue, status:", resp.Status)
		c.ApiJsonReturn("Fail to create an issue", http.StatusBadRequest, nil)
	}
	content, _ := io.ReadAll(resp.Body)
	if err = resp.Body.Close(); err != nil {
		logs.Error("Fail to close response body of creating enterprise issues, err:", err)
		c.ApiJsonReturn("Fail to close response body of creating issues", http.StatusBadRequest, nil)
	}

	var res NewIssueResponse
	if err = json.Unmarshal(content, &res); err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		c.ApiJsonReturn("Fail to solve response of creating an issue", http.StatusBadRequest, nil)
	}
	logs.Info("An issue has been created, ready to save the info")
	issueId := res.Id
	number := res.Ident
	issueUrl := res.IssueUrl
	result := make(map[string]interface{})
	result["issue_id"] = issueId
	result["number"] = number

	o := orm.NewOrm()
	if !models.SearchIssueRecord(number) {
		issue := models.Issue{
			State:    "open",
			Number:   number,
			Reporter: addr,
		}
		if _, err = o.Insert(&issue); err != nil {
			logs.Error("Fail to create issue with reporter:", err)
			c.ApiJsonReturn("Fail to create an issue", http.StatusBadRequest, nil)
		} else {
			logs.Info(fmt.Sprintf("An issue had been created by user whose email address is %v, issue number:"+
				"%v", annoyAddr, number))
		}
	} else {
		if _, err = o.QueryTable("issue").Filter("number", number).Update(orm.Params{
			"reporter": addr,
		}); err != nil {
			logs.Error("Fail to update issue reporter:", err)
			c.ApiJsonReturn("Fail to update issue", http.StatusBadRequest, nil)
		} else {
			logs.Info("Update issue successfully:", number)
		}
	}

	cleanCode(addr, code)

	go NewIssueNotify(projectId, number, issueUrl)

	c.ApiJsonReturn("Success to create an issue", http.StatusOK, result)
}

type AuthorsController struct {
	BaseController
}

type CommonParams struct {
	KeyWord string `validate:"max=100"`
	Page    int    `validate:"min=1,max=1000000"`
	PerPage int    `validate:"min=1,max=100"`
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
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct author from issue union select distinct reporter from issue order by author"
	count, err := o.Raw(sql).QueryRows(&issue)
	if err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	separateSql := sql + fmt.Sprintf(" limit ? offset ?")
	if _, err = o.Raw(separateSql, perPage, offset).QueryRows(&issue2); err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	res := make([]string, 0)
	for _, i := range issue2 {
		author := i.Author
		if strings.Contains(author, "@") {
			author = strings.Split(strings.Split(author, "@")[0], "")[0] + "***@" +
				strings.Split(author, "@")[1]
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
				author = strings.Split(strings.Split(author, "@")[0], "")[0] + "***@" +
					strings.Split(author, "@")[1]
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
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct assignee from issue order by assignee"
	count, err := o.Raw(sql).QueryRows(&issue)
	if err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	separateSql := sql + fmt.Sprintf(" limit ? offset ?")
	_, err = o.Raw(separateSql, perPage, offset).QueryRows(&issue2)
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
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	offset := perPage * (page - 1)
	o := orm.NewOrm()
	sql := "select distinct branch from issue order by branch"
	count, err := o.Raw(sql).QueryRows(&issue)
	if err != nil {
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	separateSql := sql + fmt.Sprintf(" limit ? offset ?")
	if _, err = o.Raw(separateSql, perPage, offset).QueryRows(&issue2); err != nil {
		logs.Error("Fail to query issue branches")
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
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
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	o := orm.NewOrm()
	var sql string
	sql = "select distinct milestone from issue order by milestone"
	if _, err := o.Raw(sql).QueryRows(&issue); err != nil {
		logs.Error("Fail to query milestones")
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
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
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	o := orm.NewOrm()
	var sql string
	sql = "select distinct labels from issue order by labels"
	if _, err := o.Raw(sql).QueryRows(&issue); err != nil {
		logs.Error("Fail to query issue labels")
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
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

type TypesController struct {
	BaseController
}

type QueryIssueTypesParam struct {
	Name         string `validate:"max=50"`
	Platform     string `validate:"max=50"`
	Organization string `validate:"max=50"`
}

func formQueryIssueTypesSql(q QueryIssueTypesParam) (string, []string) {
	sqlParams := make([]string, 0, 0)
	rawSql := "select * from issue_type"
	const originSqlLen = 24
	name := q.Name
	platform := q.Platform
	organization := q.Organization
	name = utils.CheckParams(name)
	platform = utils.CheckParams(platform)
	organization = utils.CheckParams(organization)
	if name != "" {
		rawSql += fmt.Sprintf(" where name=?")
		sqlParams = append(sqlParams, name)
	}
	if platform != "" {
		if len(rawSql) == originSqlLen {
			rawSql += fmt.Sprintf(" where platform=?")
			sqlParams = append(sqlParams, platform)
		} else {
			rawSql += fmt.Sprintf(" and platform=?")
			sqlParams = append(sqlParams, platform)
		}
	}
	if organization != "" {
		if len(rawSql) == originSqlLen {
			rawSql += fmt.Sprintf(" where organization=?")
			sqlParams = append(sqlParams, organization)
		} else {
			rawSql += fmt.Sprintf(" and organization=?")
			sqlParams = append(sqlParams, organization)
		}
	}
	return rawSql, sqlParams
}

func (c *TypesController) Get() {
	var issueTypes []models.IssueType
	qp := QueryIssueTypesParam{
		Name:         c.GetString("name", ""),
		Platform:     c.GetString("platform", ""),
		Organization: c.GetString("organization", ""),
	}
	validate := validator.New()
	if validateErr := validate.Struct(qp); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	o := orm.NewOrm()
	sql, sqlParams := formQueryIssueTypesSql(qp)
	if _, err := o.Raw(sql, sqlParams).QueryRows(&issueTypes); err != nil {
		logs.Error("Fail to query issue types, err:", err)
		c.ApiJsonReturn("Query error", http.StatusBadRequest, nil)
	}
	c.ApiJsonReturn("Success", http.StatusOK, issueTypes)
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

func matchUploadFileSuffix(suffix string) bool {
	allowSuffixSlice := []string{".tiff", ".jfif", ".bmp", ".gif", ".svg", ".png", ".jpeg", ".svgz", ".jpg",
		".webp", ".ico", ".xbm", ".pjp", ".apng", ".tif", ".pjpeg", ".avif"}
	allowSuffixMap := make(map[string]bool, len(allowSuffixSlice))
	for _, v := range allowSuffixSlice {
		allowSuffixMap[v] = true
	}
	if _, ok := allowSuffixMap[suffix]; !ok {
		return false
	}
	return true
}

func matchUploadFileSize(size int64) bool {
	// limit the max size of the uploaded image
	if size > 2*1024*1024 {
		return false
	}
	return true
}

func (c *UploadImageController) Post() {
	logs.Info("Ready to upload a image")
	file, h, err := c.GetFile("file")
	defer func(file multipart.File) {
		if err = file.Close(); err != nil {
			logs.Error(err)
			c.ApiJsonReturn("Fail to release object of the upload file", http.StatusBadRequest, nil)
		}
	}(file)
	if err != nil {
		c.ApiJsonReturn("Fail to read content of the upload file", http.StatusBadRequest, nil)
	}
	if !matchUploadFileSuffix(path.Ext(h.Filename)) {
		c.ApiJsonReturn("Invalid type of upload file", http.StatusBadRequest, nil)
	}
	if !matchUploadFileSize(h.Size) {
		c.ApiJsonReturn("Beyond upload file size", http.StatusBadRequest, nil)
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		return
	}
	bufWriter := bufio.NewWriter(buf)
	content, err := io.ReadAll(buf)
	if err != nil {
		logs.Error("Cannot read buf of the uploaded file:", err)
		c.ApiJsonReturn("Fail to upload image", http.StatusBadRequest, nil)
	}
	encoder := base64.NewEncoder(base64.StdEncoding, bufWriter)
	if _, err = encoder.Write(content); err != nil {
		logs.Error("Cannot encode the uploaded file:", err)
		c.ApiJsonReturn("Fail to encode the upload image", http.StatusBadRequest, nil)
	}
	encodedString := string(buf.Bytes())

	var uploadImageRequestBody UploadImageRequestBody
	uploadImageRequestBody.Base64 = fmt.Sprintf("data:image/png;base64,%s", encodedString)
	requestBodyByte, err := json.Marshal(uploadImageRequestBody)
	if err != nil {
		logs.Error("Fail to marshal request body, err:", err)
		c.ApiJsonReturn("Fail to marshal request body", http.StatusBadRequest, nil)
	}
	payload := strings.NewReader(string(requestBodyByte))
	token := config.AppConfig.V8Token
	if token == "" {
		logs.Warn("Cannot get a valid V8 access token")
		c.ApiJsonReturn("Unauthorized", http.StatusUnauthorized, nil)
	}
	enterpriseId := config.AppConfig.EnterpriseId
	url := fmt.Sprintf("%v/enterprises/%v/attach_files/upload_with_base_64?access_token=%s",
		config.AppConfig.GiteeV8ApiPrefix, enterpriseId, token)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Error("Fail to send post request, err:", err)
		c.ApiJsonReturn("Fail to request", http.StatusBadRequest, nil)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Fail to upload file, err:", err)
		c.ApiJsonReturn("Fail to upload image file", http.StatusBadRequest, nil)
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			logs.Error("Fail to close response body of uploading file:", err)
			c.ApiJsonReturn("Fail to release response body", http.StatusBadRequest, nil)
		}
	}(resp.Body)
	result, _ := io.ReadAll(resp.Body)

	var uploadImageResponse UploadImageResponse
	if err = json.Unmarshal(result, &uploadImageResponse); err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		c.ApiJsonReturn("Fail to unmarshal response to json", http.StatusBadRequest, nil)
	}
	if uploadImageResponse.Success == true {
		c.ApiJsonReturn("Success to upload image", http.StatusOK, uploadImageResponse.File)
	}
	c.ApiJsonReturn("Fail to upload image", http.StatusBadRequest, uploadImageResponse.Message)
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

func NewIssueNotify(enterpriseNumber int, number, link string) {
	sigName, repoName := models.SearchRepoByNumber(enterpriseNumber)
	if sigName == "" || repoName == "" {
		return
	}

	var notifyConf = &NotifyConf{}
	if err := config.LoadFromYaml("conf/new_issue_notify.yaml", notifyConf); err != nil {
		logs.Error("Fail to load notify yaml:", err)
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
				if err := utils.SendNewIssueNotifyEmail(ep); err != nil {
					logs.Error("Fail to send new issue notify:", err)
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
				if err := utils.SendNewIssueNotifyEmail(ep); err != nil {
					logs.Error("Fail to send new issue notify:", err)
				}
			}
			break
		}
	}
}
