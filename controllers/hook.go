package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"

	"issue_pr_board/config"
	"issue_pr_board/models"
	"issue_pr_board/utils"
)

const (
	issueHookName = "Issue Hook"
	PRHookName    = "Merge Request Hook"
)

func handleIssueEvent(request utils.WebhookRequest) {
	action := request.Action
	number := request.Issue.Number

	if action == "delete" {
		if !models.SearchIssueRecord(number) {
			return
		}
		o := orm.NewOrm()
		qt := o.QueryTable("issue")
		num, delErr := qt.Filter("number", number).Delete()
		if delErr != nil {
			logs.Error("Fail to delete the non existed issue, err:", delErr)
		}
		if num != 0 {
			logs.Info("Success to clean non existed issue:", number)
		}
		return
	}

	url := fmt.Sprintf("%v/enterprises/open_euler/issues/%v?access_token=%v", config.AppConfig.GiteeV5ApiPrefix,
		number, config.AppConfig.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get the issue, err：", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		logs.Error("Get unexpected response when getting the issue, status:", resp.Status)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	if err = resp.Body.Close(); err != nil {
		logs.Error("Fail to close response body of the issue, err:", err)
		return
	}

	var issue utils.ResponseIssue
	if err = json.Unmarshal(body, &issue); err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return
	}
	sig := models.GetSigByRepo(issue.Repository.FullName)
	if sig == "" {
		return
	}
	labels := issue.Labels
	tags := make([]string, 0)
	o := orm.NewOrm()
	if labels != nil {
		for _, label := range labels {
			var lb models.Label
			lb.Name = label.Name
			lb.Color = label.Color
			lb.UniqueId = label.Id
			if models.SearchLabel(lb.Name) {
				qs := o.QueryTable("label")
				if _, err = qs.Filter("name", lb.Name).Update(orm.Params{
					"color":     lb.Color,
					"unique_id": lb.UniqueId,
				}); err != nil {
					logs.Error("Update label failed, err:", err)
				}
			} else {
				if _, err = o.Insert(&lb); err != nil {
					logs.Error("Insert label failed, err:", err)
				}
			}
			tags = append(tags, label.Name)
		}
	}

	var ti models.Issue
	ti.Repo = issue.Repository.FullName
	ti.Org = strings.Split(ti.Repo, "/")[0]
	ti.Sig = sig
	ti.Link = issue.HtmlUrl
	ti.Number = number
	ti.State = issue.State
	ti.IssueType = issue.IssueType
	ti.IssueState = issue.IssueStateDetail.Title
	ti.Author = issue.User.Login
	ti.Assignee = issue.Assignee.Login
	ti.CreatedAt = utils.FormatTime(issue.CreatedAt)
	ti.UpdatedAt = utils.FormatTime(issue.UpdatedAt)
	ti.Title = issue.Title
	ti.Description = base64.StdEncoding.EncodeToString([]byte(issue.Description))
	ti.Priority = models.GetIssuePriority(issue.Priority)
	ti.Labels = strings.Join(tags, ",")
	ti.Branch = issue.Branch
	ti.Milestone = issue.Milestone.Title

	if !models.SearchIssueRecord(number) {
		if _, err = o.Insert(&ti); err != nil {
			logs.Error("Insert issue event failed, err:", err)
			return
		}
	}
	qs := o.QueryTable("issue")
	if _, err = qs.Filter("number", ti.Number).Update(orm.Params{
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
		"milestone":   ti.Milestone,
	}); err != nil {
		logs.Error("Update issue event failed, err:", err)
	}

	var item models.Issue
	_ = qs.Filter("number", ti.Number).One(&item)
	if item.Reporter == "" {
		return
	}
	if action == "comment" {
		commenterId := request.Author.Login
		if commenterId == "openeuler-ci-bot" {
			return
		}
		commentBody := request.Comment.Body
		ep := utils.EmailParams{Receiver: item.Reporter, Commenter: commenterId, Number: number, Title: item.Title,
			Link: ti.Link, Body: commentBody}
		if err = utils.SendCommentAttentionEmail(ep); err != nil {
			logs.Error("Fail to send issue comment attention email, err:", err)
		}
	}
	if action == "state_change" {
		ep := utils.EmailParams{Receiver: item.Reporter, State: ti.IssueState, Number: number, Title: item.Title,
			Link: ti.Link}
		if err = utils.SendStateChangeAttentionEmail(ep); err != nil {
			logs.Error("Fail to send issue state change attention email, err:", err)
		}
	}
}

func handlePullEvent(request utils.WebhookRequest) {
	htmlUrl := request.PullRequest.HtmlUrl
	if len(strings.Split(htmlUrl, "/")) != 7 {
		return
	}
	org := strings.Split(htmlUrl, "/")[3]
	repo := strings.Split(htmlUrl, "/")[4]
	fullName := org + "/" + repo
	sig := models.GetSigByRepo(fullName)
	if sig == "" {
		return
	}
	number := strings.Split(htmlUrl, "/")[6]

	url := fmt.Sprintf("%v/repos/%v/pulls/%v?access_token=%v", config.AppConfig.GiteeV5ApiPrefix, fullName,
		number, config.AppConfig.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get the pull request, err：", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		logs.Error("Get unexpected response when getting the pull request, status:", resp.Status)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	if err = resp.Body.Close(); err != nil {
		logs.Error("Fail to close response body of the pull request, err：", err)
		return
	}

	var pull utils.ResponsePull
	if err = json.Unmarshal(body, &pull); err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return
	}
	labels := pull.Labels
	assignees := pull.Assignees
	labelsSlice := make([]string, 0)
	assigneesSlice := make([]string, 0)
	o := orm.NewOrm()
	if labels != nil {
		for _, label := range labels {
			var lb models.Label
			lb.Name = label.Name
			lb.Color = label.Color
			lb.UniqueId = label.Id
			if !models.SearchLabel(lb.Name) {
				if _, err = o.Insert(&lb); err != nil {
					logs.Error("Insert label failed, err:", err)
				}
			} else {
				if _, err = o.QueryTable("label").Filter("name", lb.Name).Update(orm.Params{
					"color":     lb.Color,
					"unique_id": lb.UniqueId,
				}); err != nil {
					logs.Error("Update label failed, err:", err)
				}
			}
			labelsSlice = append(labelsSlice, label.Name)
		}
	}
	if assignees != nil {
		for _, assignee := range assignees {
			assigneesSlice = append(assigneesSlice, assignee.Login)
		}
	}

	var tp models.Pull
	tp.Org = org
	tp.Repo = fullName
	tp.Ref = pull.Base.Ref
	tp.Sig = sig
	tp.Link = htmlUrl
	tp.State = pull.State
	tp.Author = pull.User.Login
	tp.Assignees = strings.Join(assigneesSlice, ",")
	tp.CreatedAt = utils.FormatTime(pull.CreatedAt)
	tp.UpdatedAt = utils.FormatTime(pull.UpdatedAt)
	tp.Title = pull.Title
	tp.Description = base64.StdEncoding.EncodeToString([]byte(pull.Body))
	tp.Labels = strings.Join(labelsSlice, ",")
	tp.Draft = pull.Draft
	tp.Mergeable = pull.MergeAble
	if !models.SearchPullRecord(htmlUrl) {
		if _, err = o.Insert(&tp); err != nil {
			logs.Error("Insert pull event failed, err:", err)
		}
	} else {
		if _, err = o.QueryTable("pull").Filter("link", tp.Link).Update(orm.Params{
			"org":         tp.Org,
			"repo":        tp.Repo,
			"ref":         tp.Ref,
			"sig":         tp.Sig,
			"state":       tp.State,
			"author":      tp.Author,
			"assignees":   tp.Assignees,
			"created_at":  tp.CreatedAt,
			"updated_at":  tp.UpdatedAt,
			"title":       tp.Title,
			"description": tp.Description,
			"labels":      tp.Labels,
			"draft":       tp.Draft,
			"mergeable":   tp.Mergeable,
		}); err != nil {
			logs.Error("Update pull event failed, err:", err)
		}
	}
}

func payloadSignature(timestamp, key string) string {
	mac := hmac.New(sha256.New, []byte(key))

	c := fmt.Sprintf("%s\n%s", timestamp, key)
	mac.Write([]byte(c))

	h := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(h)
}

type HooksController struct {
	BaseController
}

func (c *HooksController) Post() {
	headers := c.Ctx.Request.Header
	action, ok := headers["X-Gitee-Event"]
	if !ok {
		c.ApiJsonReturn("Bad Request", http.StatusBadRequest, nil)
	}
	token, ok2 := headers["X-Gitee-Token"]
	if !ok2 {
		c.ApiJsonReturn("Bad Request", http.StatusBadRequest, nil)
	}
	timestamp, ok3 := headers["X-Gitee-Timestamp"]
	if !ok3 {
		c.ApiJsonReturn("Bad Request", http.StatusBadRequest, nil)
	}
	if token[0] != payloadSignature(timestamp[0], config.AppConfig.WebhookToken) {
		c.ApiJsonReturn("Bad Request", http.StatusBadRequest, nil)
	}

	body := c.Ctx.Input.RequestBody
	var webhookRequest utils.WebhookRequest
	if err := json.Unmarshal(body, &webhookRequest); err != nil {
		logs.Error("Fail to unmarshal request body to json, err:", err)
		c.ApiJsonReturn("Bad Request", http.StatusBadRequest, nil)
	}
	switch {
	case action[0] == issueHookName:
		handleIssueEvent(webhookRequest)
	case action[0] == PRHookName:
		handlePullEvent(webhookRequest)
	default:
		c.ApiJsonReturn("OK", http.StatusOK, nil)
	}
}
