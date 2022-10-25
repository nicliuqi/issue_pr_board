package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/chenhg5/collection"
	"io/ioutil"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"net/http"
	"os"
	"strings"
)

type HooksController struct {
	BaseController
}

func HandleIssueEvent(reqBody map[string]interface{}) {
	action := reqBody["action"].(string)
	number := reqBody["issue"].(map[string]interface{})["number"].(string)
	if action == "delete" {
		o := orm.NewOrm()
		deleteSql := fmt.Sprintf("delete from issue where number='%s'", number)
		_, err := o.Raw(deleteSql).Exec()
		if err != nil {
			logs.Error("Fail to delete the non existed issue, err:", err)
		} else {
			logs.Info("Success to clean non existed issue:", number)
		}
		return
	}
	_, repos := utils.GetSigsMapping()
	if repos == nil {
		logs.Error("Fail to get sigs mapping.")
		return
	}
	url := fmt.Sprintf("https://gitee.com/api/v5/enterprises/open_euler/issues/%v?access_token=%v", number, os.Getenv("AccessToken"))
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get repo members, err：", err)
		return
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting repo members, status:", resp.Status)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of repo members, err:", err)
		return
	}
	issue := utils.JsonToMap(string(body))
	repository := issue["repository"]
	if repository == nil {
		return
	}
	htmlUrl := issue["html_url"].(string)
	fullName := issue["repository"].(map[string]interface{})["full_name"].(string)
	org := strings.Split(fullName, "/")[0]
	if org != "src-openeuler" && org != "openeuler" {
		return
	}
	author := issue["user"].(map[string]interface{})["login"].(string)
	state := issue["state"].(string)
	issueType := issue["issue_type"].(string)
	issueState := issue["issue_state_detail"].(map[string]interface{})["title"].(string)
	createdAt := issue["created_at"].(string)
	updatedAt := issue["updated_at"].(string)
	sig := utils.GetSigByRepo(repos, fullName)
	assignee := issue["assignee"]
	assigneeLogin := ""
	if assignee != nil {
		assigneeLogin = assignee.(map[string]interface{})["login"].(string)
	}
	title := issue["title"].(string)
	description := issue["body"]
	if description == nil {
		description = ""
	}
	description = base64.StdEncoding.EncodeToString([]byte(description.(string)))
	labels := issue["labels"]
	priorityNum := issue["priority"]
	priority := GetIssuePriority(priorityNum.(float64))
	branch := issue["branch"]
	if branch == nil {
		branch = ""
	}
	tags := make([]string, 0)
	if labels != nil {
		for _, label := range labels.([]interface{}) {
			tags = append(tags, label.(map[string]interface{})["name"].(string))
		}
	}
	var ti models.Issue
	ti.Org = org
	ti.Repo = fullName
	ti.Sig = sig
	ti.Link = htmlUrl
	ti.Number = number
	ti.State = state
	ti.IssueType = issueType
	ti.IssueState = issueState
	ti.Author = author
	ti.Assignee = assigneeLogin
	ti.CreatedAt = utils.FormatTime(createdAt)
	ti.UpdatedAt = utils.FormatTime(updatedAt)
	ti.Title = title
	ti.Description = description.(string)
	ti.Priority = priority
	ti.Labels = strings.Join(tags, ",")
	ti.Branch = branch.(string)
	issueExists := SearchIssueRecord(number)
	if issueExists == true {
		o := orm.NewOrm()
		qs := o.QueryTable("issue")
		_, err := qs.Filter("number", ti.Number).Update(orm.Params{
			"org":         ti.Org,
			"repo":        ti.Repo,
			"sig":         ti.Sig,
			"link":        ti.Link,
			"state":       ti.State,
			"issue_type":  ti.IssueType,
			"issue_state": ti.IssueState,
			"author":      ti.Author,
			"assignee":    ti.Assignee,
			"created_at":  ti.CreatedAt,
			"updated_at":  ti.UpdatedAt,
			"title":       ti.Title,
			"description": ti.Description,
			"priority":    ti.Priority,
			"labels":      ti.Labels,
			"branch":      ti.Branch,
		})
		if err != nil {
			logs.Error("Update issue event failed, err:", err)
		}
	} else {
		o := orm.NewOrm()
		_, err := o.Insert(&ti)
		if err != nil {
			logs.Error("Insert issue event failed, err:", err)
		}
	}
}

func HandlePullEvent(reqBody map[string]interface{}) {
	_, repos := utils.GetSigsMapping()
	if repos == nil {
		logs.Error("Fail to get sigs mapping.")
		return
	}
	htmlUrl := reqBody["pull_request"].(map[string]interface{})["html_url"].(string)
	org := strings.Split(htmlUrl, "/")[3]
	if org != "src-openeuler" && org != "openeuler" {
		return
	}
	repo := strings.Split(htmlUrl, "/")[4]
	fullName := org + "/" + repo
	number := strings.Split(htmlUrl, "/")[6]
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/pulls/%v?access_token=%v", fullName, number, os.Getenv("AccessToken"))
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get enterprise pull requests, err：", err)
		return
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting enterprise pulls, status:", resp.Status)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of enterprise pull requests, err：", err)
		return
	}
	pull := utils.JsonToMap(string(body))
	state := pull["state"].(string)
	ref := pull["base"].(map[string]interface{})["ref"].(string)
	author := pull["user"].(map[string]interface{})["login"].(string)
	createdAt := pull["created_at"].(string)
	updatedAt := pull["updated_at"].(string)
	sig := utils.GetSigByRepo(repos, fullName)
	labels := pull["labels"]
	assignees := pull["assignees"]
	labelsSlice := make([]string, 0)
	assigneesSlice := make([]string, 0)
	if labels != nil {
		for _, label := range labels.([]interface{}) {
			labelsSlice = append(labelsSlice, label.(map[string]interface{})["name"].(string))
		}
	}
	if assignees != nil {
		for _, assignee := range assignees.([]interface{}) {
			assigneesSlice = append(assigneesSlice, assignee.(map[string]interface{})["login"].(string))
		}
	}
	var tp models.Pull
	tp.Org = org
	tp.Repo = fullName
	tp.Ref = ref
	tp.Sig = sig
	tp.Link = htmlUrl
	tp.State = state
	tp.Author = author
	tp.Assignees = strings.Join(assigneesSlice, ",")
	tp.CreatedAt = utils.FormatTime(createdAt)
	tp.UpdatedAt = utils.FormatTime(updatedAt)
	pullExists := SearchPullRecord(htmlUrl)
	if pullExists == true {
		o := orm.NewOrm()
		qs := o.QueryTable("pull")
		_, err := qs.Filter("link", tp.Link).Update(orm.Params{
			"ref":        tp.Ref,
			"sig":        tp.Sig,
			"state":      tp.State,
			"author":     tp.Author,
			"assignees":  tp.Assignees,
			"updated_at": tp.UpdatedAt,
		})
		if err != nil {
			logs.Error("Update pull event failed, err:", err)
		}
	} else {
		o := orm.NewOrm()
		_, err := o.Insert(&tp)
		if err != nil {
			logs.Error("Insert pull event failed, err:", err)
		}
	}
}

func (c *HooksController) Post() {
	headers := c.Ctx.Request.Header
	_, ok := headers["X-Gitee-Event"]
	if !ok {
		logs.Warn("Notice a fake WebHook and ignore.")
		return
	}
	action := headers["X-Gitee-Event"]
	logs.Info("Receive a", action)
	body := c.Ctx.Input.RequestBody
	logs.Info(collection.Collect(string(body)).ToJson())
	reqBody := collection.Collect(string(body)).ToMap()
	switch {
	case collection.Collect(action).Contains("Issue Hook"):
		HandleIssueEvent(reqBody)
	case collection.Collect(action).Contains("Merge Request Hook"):
		HandlePullEvent(reqBody)
	default:
		issue, ok := reqBody["issue"]
		if ok && issue != nil {
			HandleIssueEvent(reqBody)
		}
		pr, ok2 := reqBody["pull_request"]
		if ok2 && pr != nil {
			HandlePullEvent(reqBody)
		}
		return
	}
}
