package controllers

import (
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

type HooksController struct {
	BaseController
}

func HandleIssueEvent(r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	err = r.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of handling issue events, err:", err)
		return
	}
	var req utils.WebhookRequest
	err = json.Unmarshal(reqBody, &req)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return
	}
	action := req.Action
	number := req.Issue.Number
	if action == "delete" {
		o := orm.NewOrm()
		if _, err := o.Delete(&models.Issue{Number: number}); err != nil {
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
	url := fmt.Sprintf("%v/enterprises/open_euler/issues/%v?access_token=%v", config.AppConfig.GiteeV5ApiPrefix,
		number, config.AppConfig.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get the issue, err：", err)
		return
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting the issue, status:", resp.Status)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of the issue, err:", err)
		return
	}
	var issue utils.ResponseIssue
	err = json.Unmarshal(body, &issue)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return
	}
	htmlUrl := issue.HtmlUrl
	fullName := issue.Repository.FullName
	org := strings.Split(fullName, "/")[0]
	if org != "src-openeuler" && org != "openeuler" {
		return
	}
	author := issue.User.Login
	state := issue.State
	issueType := issue.IssueType
	issueState := issue.IssueStateDetail.Title
	createdAt := issue.CreatedAt
	updatedAt := issue.UpdatedAt
	sig := utils.GetSigByRepo(repos, fullName)
	milestone := issue.Milestone
	assigneeLogin := issue.Assignee
	title := issue.Title
	description := issue.Description
	description = base64.StdEncoding.EncodeToString([]byte(description))
	labels := issue.Labels
	priorityNum := issue.Priority
	priority := GetIssuePriority(priorityNum)
	branch := issue.Branch
	tags := make([]string, 0)
	if labels != nil {
		for _, label := range labels {
			var lb models.Label
			lb.Name = label.Name
			lb.Color = label.Color
			lb.UniqueId = label.Id
			if models.SearchLabel(lb.Name) {
				o := orm.NewOrm()
				qs := o.QueryTable("label")
				_, err = qs.Filter("name", lb.Name).Update(orm.Params{
					"color":     lb.Color,
					"unique_id": lb.UniqueId,
				})
				if err != nil {
					logs.Error("Update label failed, err:", err)
				}
			} else {
				o := orm.NewOrm()
				_, err = o.Insert(&lb)
				if err != nil {
					logs.Error("Insert label failed, err:", err)
				}
			}
			tags = append(tags, label.Name)
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
	ti.Assignee = assigneeLogin.Login
	ti.CreatedAt = utils.FormatTime(createdAt)
	ti.UpdatedAt = utils.FormatTime(updatedAt)
	ti.Title = title
	ti.Description = description
	ti.Priority = priority
	ti.Labels = strings.Join(tags, ",")
	ti.Branch = branch
	ti.Milestone = milestone.Title
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
			"milestone":   ti.Milestone,
		})
		if err != nil {
			logs.Error("Update issue event failed, err:", err)
		}
		var item models.Issue
		_ = qs.Filter("number", ti.Number).One(&item)
		if item.Reporter == "" {
			return
		} else {
			if action == "comment" {
				commenterId := req.Author.Login
				if commenterId == "openeuler-ci-bot" {
					return
				}
				commentBody := req.Comment.Body
				ep := utils.EmailParams{Receiver: item.Reporter, Commenter: commenterId, Number: number, Title: title,
					Link: htmlUrl, Body: commentBody}
				err = utils.SendCommentAttentionEmail(ep)
				if err != nil {
					logs.Error("Fail to send issue comment attention email, err:", err)
				}
			}
			if action == "state_change" {
				ep := utils.EmailParams{Receiver: item.Reporter, State: issueState, Number: number, Title: title,
					Link: htmlUrl}
				err = utils.SendStateChangeAttentionEmail(ep)
				if err != nil {
					logs.Error("Fail to send issue state change attention email, err:", err)
				}
			}
		}
	} else {
		o := orm.NewOrm()
		_, err := o.Insert(&ti)
		if err != nil {
			logs.Error("Insert issue event failed, err:", err)
		}
	}
}

func HandlePullEvent(r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	err = r.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of repo members, err:", err)
		return
	}
	var req utils.WebhookRequest
	err = json.Unmarshal(reqBody, &req)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return
	}
	_, repos := utils.GetSigsMapping()
	if repos == nil {
		logs.Error("Fail to get sigs mapping.")
		return
	}
	htmlUrl := req.PullRequest.HtmlUrl
	org := strings.Split(htmlUrl, "/")[3]
	if org != "src-openeuler" && org != "openeuler" {
		return
	}
	repo := strings.Split(htmlUrl, "/")[4]
	fullName := org + "/" + repo
	number := strings.Split(htmlUrl, "/")[6]
	url := fmt.Sprintf("%v/repos/%v/pulls/%v?access_token=%v", config.AppConfig.GiteeV5ApiPrefix, fullName,
		number, config.AppConfig.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get the pull request, err：", err)
		return
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting the pull request, status:", resp.Status)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of the pull request, err：", err)
		return
	}
	var pull utils.ResponsePull
	err = json.Unmarshal(body, &pull)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return
	}
	state := pull.State
	ref := pull.Base.Ref
	author := pull.User.Login
	createdAt := pull.CreatedAt
	updatedAt := pull.UpdatedAt
	sig := utils.GetSigByRepo(repos, fullName)
	title := pull.Title
	description := pull.Body
	description = base64.StdEncoding.EncodeToString([]byte(description))
	labels := pull.Labels
	assignees := pull.Assignees
	draft := pull.Draft
	mergeable := pull.MergeAble
	labelsSlice := make([]string, 0)
	assigneesSlice := make([]string, 0)
	if labels != nil {
		for _, label := range labels {
			var lb models.Label
			lb.Name = label.Name
			lb.Color = label.Color
			lb.UniqueId = label.Id
			if models.SearchLabel(lb.Name) {
				o := orm.NewOrm()
				qs := o.QueryTable("label")
				_, err = qs.Filter("name", lb.Name).Update(orm.Params{
					"color":     lb.Color,
					"unique_id": lb.UniqueId,
				})
				if err != nil {
					logs.Error("Update label failed, err:", err)
				}
			} else {
				o := orm.NewOrm()
				_, err = o.Insert(&lb)
				if err != nil {
					logs.Error("Insert label failed, err:", err)
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
	tp.Ref = ref
	tp.Sig = sig
	tp.Link = htmlUrl
	tp.State = state
	tp.Author = author
	tp.Assignees = strings.Join(assigneesSlice, ",")
	tp.CreatedAt = utils.FormatTime(createdAt)
	tp.UpdatedAt = utils.FormatTime(updatedAt)
	tp.Title = title
	tp.Description = description
	tp.Labels = strings.Join(labelsSlice, ",")
	tp.Draft = draft
	tp.Mergeable = mergeable
	if SearchPullRecord(htmlUrl) {
		o := orm.NewOrm()
		qs := o.QueryTable("pull")
		_, err := qs.Filter("link", tp.Link).Update(orm.Params{
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
		c.ApiJsonReturn("Bad Request", 400, nil)
	}
	action := headers["X-Gitee-Event"]
	_, ok2 := headers["X-Gitee-Token"]
	if !ok2 {
		logs.Warn("Notice a fake WebHook and ignore.")
		c.ApiJsonReturn("Bad Request", 400, nil)
	}
	token := headers["X-Gitee-Token"]
	if token[0] != config.AppConfig.WebhookToken {
		logs.Warn("Notice a fake WebHook and ignore.")
		c.ApiJsonReturn("Bad Request", 400, nil)
	}
	body := c.Ctx.Input.RequestBody
	var webhookRequest utils.WebhookRequest
	err := json.Unmarshal(body, &webhookRequest)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		c.ApiJsonReturn("Bad Request", 400, err)
	}
	switch {
	case action[0] == "Issue Hook":
		HandleIssueEvent(c.Ctx.Request)
	case action[0] == "Merge Request Hook":
		HandlePullEvent(c.Ctx.Request)
	default:
		if webhookRequest.Issue.HtmlUrl != "" {
			HandleIssueEvent(c.Ctx.Request)
		}
		if webhookRequest.PullRequest.HtmlUrl != "" {
			HandlePullEvent(c.Ctx.Request)
		}
		c.ApiJsonReturn("OK", 200, nil)
	}
}
