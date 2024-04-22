package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"issue_pr_board/config"
	_ "issue_pr_board/config"
	"issue_pr_board/controllers"
	"issue_pr_board/models"
	_ "issue_pr_board/models"
	"issue_pr_board/utils"
	"net/http"
	"strings"
)

var token = config.AppConfig.AccessToken

func SyncEnterprisePulls() error {
	logs.Info("Starting to sync pulls...")
	_, repos := utils.GetSigsMapping()
	if repos == nil {
		logs.Error("Fail to get sigs mapping.")
		return nil
	}
	page := 1
	for {
		url := fmt.Sprintf("https://gitee.com/api/v5/enterprise/open_euler/pull_requests?state=all&sort=created"+
			"&direction=asc&page=%v&per_page=100&access_token=%v", page, token)
		resp, err := http.Get(url)
		if err != nil {
			logs.Error("Fail to get enterprise pull requests, err：", err)
			return err
		}
		if resp.StatusCode != 200 {
			logs.Error("Get unexpected response when getting enterprise pulls, status:", resp.Status)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		err = resp.Body.Close()
		if err != nil {
			logs.Error("Fail to close response body of enterprise pull requests, err:", err)
			return err
		}
		if len(string(body)) == 2 {
			break
		}
		var rps []utils.ResponsePull
		err = json.Unmarshal(body, &rps)
		if err != nil {
			logs.Error("Fail to unmarshal response to json, err:", err)
			return err
		}
		for _, pull := range rps {
			htmlUrl := pull.HtmlUrl
			org := strings.Split(htmlUrl, "/")[3]
			if org != "src-openeuler" && org != "openeuler" {
				continue
			}
			repo := strings.Split(htmlUrl, "/")[4]
			fullName := org + "/" + repo
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
			if controllers.SearchPullRecord(htmlUrl) {
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
					logs.Error("Update pull failed, err:", err)
				}
			} else {
				o := orm.NewOrm()
				_, err := o.Insert(&tp)
				if err != nil {
					logs.Error("Insert pull failed, err:", err)
				}
			}
		}
		page += 1
	}
	logs.Info("Ends of pulls sync, wait the next time...")
	return nil
}

func SyncEnterpriseIssues() error {
	logs.Info("Starting to sync issues...")
	_, repos := utils.GetSigsMapping()
	if repos == nil {
		logs.Error("Fail to get sigs mapping.")
		return nil
	}
	page := 1
	for {
		url := fmt.Sprintf("https://gitee.com/api/v5/enterprises/open_euler/issues?state=all&sort=created"+
			"&direction=asc&page=%v&per_page=100&access_token=%v", page, token)
		resp, err := http.Get(url)
		if err != nil {
			logs.Error("Fail to get enterprise issues, err：", err)
			return err
		}
		if resp.StatusCode != 200 {
			logs.Error("Get unexpected response when getting enterprise issues, status:", resp.Status)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		err = resp.Body.Close()
		if err != nil {
			logs.Error("Fail to close response body of enterprise issues, err:", err)
			return err
		}
		if len(string(body)) == 2 {
			break
		}
		var ris []utils.ResponseIssue
		err = json.Unmarshal(body, &ris)
		if err != nil {
			logs.Error("Fail to unmarshal response to json, err:", err)
			return err
		}
		for _, issue := range ris {
			htmlUrl := issue.HtmlUrl
			fullName := issue.Repository.FullName
			org := strings.Split(fullName, "/")[0]
			if org != "src-openeuler" && org != "openeuler" {
				continue
			}
			author := issue.User.Login
			number := issue.Number
			state := issue.State
			issueType := issue.IssueType
			issueState := issue.IssueStateDetail.Title
			createdAt := issue.CreatedAt
			updatedAt := issue.UpdatedAt
			sig := utils.GetSigByRepo(repos, fullName)
			milestone := issue.Milestone
			assignee := issue.Assignee
			assigneeLogin := assignee.Login
			title := issue.Title
			description := issue.Description
			description = base64.StdEncoding.EncodeToString([]byte(description))
			labels := issue.Labels
			priorityNum := issue.Priority
			priority := controllers.GetIssuePriority(priorityNum)
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
			ti.Assignee = assigneeLogin
			ti.CreatedAt = utils.FormatTime(createdAt)
			ti.UpdatedAt = utils.FormatTime(updatedAt)
			ti.Title = title
			ti.Description = description
			ti.Priority = priority
			ti.Labels = strings.Join(tags, ",")
			ti.Branch = branch
			ti.Milestone = milestone.Title
			issueExists := controllers.SearchIssueRecord(number)
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
					logs.Error("Update issue failed, err:", err)
				}
			} else {
				o := orm.NewOrm()
				_, err := o.Insert(&ti)
				if err != nil {
					logs.Error("Insert issue failed, err:", err)
				}
			}
		}
		page += 1
	}
	logs.Info("Ends of issues sync, wait the next time...")
	return nil
}

func SyncEnterpriseRepos() error {
	logs.Info("Starting to sync repos...")
	_, repos := utils.GetSigsMapping()
	if repos == nil {
		logs.Error("Fail to get sigs mapping.")
		return nil
	}
	for repo, sig := range repos {
		var tr models.Repo
		tr.Name = repo
		tr.Sig = sig
		tr.Branches = getRepoBranches(repo)
		tr.Reviewers = getRepoReviewers(repo)
		repoExists := searchRepoRecord(repo)
		if repoExists {
			o := orm.NewOrm()
			qs := o.QueryTable("repo")
			_, err := qs.Filter("name", tr.Name).Update(orm.Params{
				"sig":       tr.Sig,
				"branches":  tr.Branches,
				"reviewers": tr.Reviewers,
			})
			if err != nil {
				logs.Error("Update repo failed, err:", err)
			}
		} else {
			o := orm.NewOrm()
			_, err := o.Insert(&tr)
			if err != nil {
				logs.Error("Insert repo failed, err:", err)
			}
		}
	}
	err := controllers.SyncRepoNumber()
	if err != nil {
		return err
	}
	logs.Info("Ends of repos sync, wait the next time...")
	return nil
}

type ResponsePullBranch struct {
	Name	string	`json:"name"`
}

func getRepoBranches(repo string) string {
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/branches?access_token=%v", repo, token)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get repo branches, err：", err)
		return ""
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting repo branches, status:", resp.Status)
		return ""
	}
	body, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of repo branches, err:", err)
		return ""
	}
	var rbs []ResponsePullBranch
	err = json.Unmarshal(body, &rbs)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return ""
	}
	res := make([]string, 0)
	for _, branchItem := range rbs {
		branch := branchItem.Name
		res = append(res, branch)
	}
	return strings.Join(res, ",")
}

type ResponsePullCollaborator struct {
	Login	string	`json:"login"`
}

func getRepoReviewers(repo string) string {
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/collaborators?access_token=%v&page=1&per_page=100",
	    repo, token)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get repo members, err：", err)
		return ""
	}
	if resp.StatusCode != 200 {
		logs.Error("Get unexpected response when getting repo members, status:", resp.Status)
		return ""
	}
	body, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of repo members, err:", err)
		return ""
	}
	var rcs []ResponsePullCollaborator
	err = json.Unmarshal(body, &rcs)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return ""
	}
	res := make([]string, 0)
	for _, memberItem := range rcs {
		member := memberItem.Login
		if member != "openeuler-ci-bot" {
			res = append(res, member)
		}
	}
	return strings.Join(res, ",")
}

func searchRepoRecord(repo string) bool {
	o := orm.NewOrm()
	searchSql := fmt.Sprintf("select * from repo where name='%s'", repo)
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}
