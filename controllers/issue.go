package controllers

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/chenhg5/collection"
	"io"
	"io/ioutil"
	"issue_pr_board/models"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type IssuesController struct {
	BaseController
}

type QueryIssueParam struct {
	Org        string
	Repo       string
	Sig        string
	State      string
	Number     string
	Author     string
	Assignee   string
	Branch     string
	Label      string
	Exclusion  string
	IssueState string
	IssueType  string
	Priority   string
	Search     string
	Sort       string
	Direction  string
	Page       int
	PerPage    int
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
	page := q.Page
	perPage := q.PerPage
	if issueState != "" {
		issueState = strings.Replace(issueState, "，", ",", -1)
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
	if author != "" {
		if strings.Contains(author, "@") {
			newAuthor := strings.Split(author, "@")[0]
			if newAuthor != "" {
				rawSql += fmt.Sprintf(" and reporter regexp '^%s'", newAuthor)
			}
		} else {
			rawSql += fmt.Sprintf(" and author='%s'", author)
		}
	}
	if assignee != "" {
		rawSql += fmt.Sprintf(" and assignee='%s'", assignee)
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
	if perPage >= 100 {
		perPage = 100
	}
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
		Page:       page,
		PerPage:    perPage,
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
			rawDescription, _ := base64.StdEncoding.DecodeString(description)
			i.Description = string(rawDescription)
			res = append(res, i)
		}
		c.ApiDataReturn(count, page, perPage, res)
	}
}

func (c *IssuesController) Post() {
	logs.Info("Receive a request of creating an issue")
	body := c.Ctx.Input.RequestBody
	if body == nil {
		return
	}
	reqBody := collection.Collect(string(body)).ToMap()
	addr := reqBody["email"].(string)
	code := reqBody["code"].(string)
	if !checkCode(addr, code) {
		c.ApiJsonReturn("验证码错误", 400, "")
	}
	payloadMap := make(map[string]interface{})
	payloadMap["access_token"] = models.GetV8Token(3)
	payloadMap["project_id"] = reqBody["project_id"]
	payloadMap["title"] = reqBody["title"]
	payloadMap["description"] = reqBody["description"]
	payloadMap["issue_type_id"] = reqBody["issue_type_id"]
	enterpriseId := os.Getenv("EnterpriseId")
	url := fmt.Sprintf("https://api.gitee.com/enterprises/%v/issues", enterpriseId)
	payload := strings.NewReader(collection.Collect(payloadMap).ToJson())
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Fail to create quick issue, err:", err)
		c.ApiJsonReturn("创建issue失败", 400, err)
	}
	if resp.StatusCode != 201 {
		c.ApiJsonReturn("创建issue失败", resp.StatusCode, resp.Body)
	}
	content, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of creating enterprise issues, err:", err)
		c.ApiJsonReturn("无法关闭创建issue的响应", 400, err)
	}
	logs.Info("An issue has been created, ready to save the info")
	res := collection.Collect(string(content)).ToMap()
	issueId := res["id"]
	number := res["ident"]
	result := make(map[string]interface{})
	result["issue_id"] = issueId
	result["number"] = number
	if !SearchIssueRecord(number.(string)) {
		o := orm.NewOrm()
		insertSql := fmt.Sprintf("insert into issue (state, number, reporter) values('open', '%s', '%s')", number, addr)
		_, err = o.Raw(insertSql).Exec()
		if err != nil {
			logs.Error("Fail to create issue with reporter:", err)
		} else {
			logs.Info("Save issue successfully:", number)
		}
	} else {
		o := orm.NewOrm()
		updateSql := fmt.Sprintf("update issue set reporter='%s' where number='%s'", addr, number)
		_, err = o.Raw(updateSql).Exec()
		if err != nil {
			logs.Error("Fail to update issue reporter:", err)
		} else {
			logs.Info("Update issue successfully:", number)
		}
	}
	cleanCode(addr, code)
	c.ApiJsonReturn("创建成功", 201, result)
}

type AuthorsController struct {
	BaseController
}

func (c *AuthorsController) Get() {
	var issue []models.Issue
	var issue2 []models.Issue
	page, _ := c.GetInt("page", 1)
	perPage, _ := c.GetInt("per_page", 20)
	keyWord := c.GetString("keyword", "")
	if perPage > 100 {
		perPage = 100
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
	if perPage > 100 {
		perPage = 100
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
	if perPage > 100 {
		perPage = 100
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

type LabelsController struct {
	BaseController
}

func (c *LabelsController) Get() {
	var issue []models.Issue
	keyWord := c.GetString("keyword", "")
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
				if collection.Collect(res).Contains(j) {
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
		page, _ := c.GetInt("page", 1)
		perPage, _ := c.GetInt("per_page", 20)
		if perPage > 100 {
			perPage = 100
		}
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

type IssueTypesResponse struct {
	TotalCount int                 `json:"total_count"`
	Data       []IssueTypeResponse `json:"data"`
}

type IssueTypeResponse struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

func (c *TypesController) Get() {
	mode := c.GetString("mode", "")
	if mode == "local" {
		var issue []models.Issue
		o := orm.NewOrm()
		searchSql := "select distinct issue_type from issue"
		_, err := o.Raw(searchSql).QueryRows(&issue)
		if err != nil {
			logs.Error("Fail to search issue types, err:", err)
		}
		res := make([]string, 0)
		for _, i := range issue {
			res = append(res, i.IssueType)
		}
		c.ApiJsonReturn("请求成功", 200, res)
	}
	token := models.GetV8Token(3)
	enterpriseId := os.Getenv("EnterpriseId")
	url := fmt.Sprintf("https://api.gitee.com/enterprises/%s/issue_types?page=1&per_page=100&access_token=%s", enterpriseId, token)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get enterprise issue types, err:", err)
		c.ApiJsonReturn("获取任务类型列表失败", 400, err)
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting enterprise issue types, status:", resp.Status)
		c.ApiJsonReturn("获取任务类型列表响应未知", resp.StatusCode, resp.Body)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logs.Error("Fail to close response body of enterprise issue types, err：", err)
		}
	}(resp.Body)
	res := make([]map[string]interface{}, 0)
	var issuesTypesResponse IssueTypesResponse
	err = json.Unmarshal(body, &issuesTypesResponse)
	issuesTypes := issuesTypesResponse.Data
	for _, issueType := range issuesTypes {
		id := issueType.Id
		title := issueType.Title
		tmpMap := make(map[string]interface{})
		tmpMap["id"] = id
		tmpMap["title"] = title
		res = append(res, tmpMap)
	}
	c.ApiJsonReturn("请求成功", 200, res)
}

type UploadAttachmentController struct {
	BaseController
}

type Attachment struct {
	AttachId string `form:"attach_id"`
}

func (c *UploadAttachmentController) Post() {
	// save attachment
	tmpDir, _ := os.MkdirTemp("", "")
	var attachment Attachment
	err := c.ParseForm(&attachment)
	if err != nil {
		c.ApiJsonReturn("解析表单出错", 400, err)
	}
	logs.Info("Ready to upload a attachment")
	file, h, err := c.GetFile("file")
	defer func(file multipart.File) {
		err = file.Close()
		if err != nil {
			logs.Error("Fail to close uploaded file, err:", err)
		}
	}(file)
	if err != nil {
		logs.Error("Fail to get uploaded file")
	}
	tmpPath := fmt.Sprintf(tmpDir + "/" + h.Filename)
	err = c.SaveToFile("file", tmpPath)
	if err != nil {
		logs.Error("Fail to save file")
	}
	// upload attachment
	enterpriseId := os.Getenv("EnterpriseId")
	token := models.GetV8Token(3)
	url := fmt.Sprintf("https://api.gitee.com/enterprises/%v/attach_files/upload", enterpriseId)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	uploadFile, err := os.Open(tmpPath)
	if err != nil {
		logs.Error("Fail to open uploadFile")
	}
	defer func(uploadFile multipart.File) {
		err = uploadFile.Close()
		if err != nil {
			logs.Error("Fail to close file")
		}
	}(uploadFile)
	part1, _ := writer.CreateFormFile("file", filepath.Base(tmpPath))
	_, err = io.Copy(part1, file)
	if err != nil {
		logs.Error("Fail to add file to form data")
	}
	_ = writer.WriteField("attach_type", "issue")
	_ = writer.WriteField("attach_id", attachment.AttachId)
	_ = writer.WriteField("access_token", token)
	err = writer.Close()
	if err != nil {
		logs.Error("Fail to close writer")
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Error("Fail to send request")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		logs.Error("Fail to get response")
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logs.Error("Fail to close body")
		}
	}(res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logs.Error("Fail to read Body of res, err:", err)
	}
	logs.Info(string(body))
	// remove temp file
	err = os.Remove(tmpPath)
	if err != nil {
		logs.Error("Fail to remove temp file, err:", err)
	}
	if res.StatusCode != 201 {
		c.ApiJsonReturn("上传失败", res.StatusCode, collection.Collect(string(body)).ToMap())
	}
	c.ApiJsonReturn("成功上传附件", 200, collection.Collect(string(body)).ToMap())
}

type UploadImageController struct {
	BaseController
}

func (c *UploadImageController) Post() {
	logs.Info("Ready to upload a image")
	file, _, err := c.GetFile("file")
	defer func(file multipart.File) {
		err = file.Close()
		if err != nil {
			logs.Error(err)
		}
	}(file)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		return
	}
	bufWriter := bufio.NewWriter(buf)
	content, err := ioutil.ReadAll(buf)
	if err != nil {
		logs.Error("Cannot read buf of the uploaded file")
		c.ApiJsonReturn("上传失败", 400, err)
	}
	encoder := base64.NewEncoder(base64.StdEncoding, bufWriter)
	_, err = encoder.Write(content)
	if err != nil {
		return
	}
	encodedString := string(buf.Bytes())
	payloadMap := make(map[string]interface{})
	payloadMap["base64"] = fmt.Sprintf("data:image/png;base64,%s", encodedString)
	payload := strings.NewReader(collection.Collect(payloadMap).ToJson())
	token := models.GetV8Token(3)
	if token == "" {
		logs.Warn("Cannot get a valid V8 access token")
		c.ApiJsonReturn("认证失败", 401, "")
	}
	enterpriseId := os.Getenv("EnterpriseId")
	url := fmt.Sprintf("https://api.gitee.com/enterprises/%v/attach_files/upload_with_base_64?access_token=%s", enterpriseId, token)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("Fail to upload file, err:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logs.Error("Fail to close response body of uploading file")
		}
	}(resp.Body)
	result, _ := ioutil.ReadAll(resp.Body)
	res := collection.Collect(string(result)).ToMap()
	if res["success"] == true {
		c.ApiJsonReturn("上传成功", 200, res["file"])
	}
	c.ApiJsonReturn("上传失败", 400, res["message"])
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
