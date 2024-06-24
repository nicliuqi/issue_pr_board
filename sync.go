package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/task"
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

const (
	defaultPage = 1
	perPageMax  = 100
)

func syncEnterprisePulls() error {
	logs.Info("Starting to sync pulls...")
	var repos []models.Repo
	o := orm.NewOrm()
	if _, err := o.QueryTable("repo").All(&repos); err != nil {
		logs.Error("Fail to query repo")
		return err
	}
	PRLinkMap := make(map[string]bool)
	for _, repo := range repos {
		page := defaultPage
		for {
			url := fmt.Sprintf("%v/repos/%v/pulls?state=all&sort=created&direction=desc&page=%v&per_page=%v&"+
				"access_token=%v", config.AppConfig.GiteeV5ApiPrefix, repo.Name, page, perPageMax,
				config.AppConfig.AccessToken)
			resp, err := http.Get(url)
			if err != nil {
				page += 1
				continue
			}
			if resp.StatusCode != http.StatusOK {
				page += 1
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			if err = resp.Body.Close(); err != nil {
				logs.Error("Fail to close response body of enterprise pull requests, err:", err)
				page += 1
				continue
			}

			var rps []utils.ResponsePull
			if err = json.Unmarshal(body, &rps); err != nil {
				logs.Error("Fail to unmarshal response to json, err:", err)
				page += 1
				continue
			}
			if len(rps) == 0 {
				break
			}
			for _, pull := range rps {
				labels := pull.Labels
				assignees := pull.Assignees
				labelsSlice := make([]string, 0)
				assigneesSlice := make([]string, 0)
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
				tp.Org = strings.Split(repo.Name, "/")[0]
				tp.Repo = repo.Name
				tp.Ref = pull.Base.Ref
				tp.Sig = repo.Sig
				tp.Link = pull.HtmlUrl
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
				if !models.SearchPullRecord(pull.HtmlUrl) {
					if _, err = o.Insert(&tp); err != nil {
						logs.Error("Insert pull failed, err:", err)
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
						logs.Error("Update pull failed, err:", err)
					}
				}
				PRLinkMap[pull.HtmlUrl] = true
			}
			page += 1
		}
	}

	var prs []models.Pull
	if _, err := o.QueryTable("pull").All(&prs); err != nil {
		logs.Error("Fail to query pull requests")
		return err
	}
	for _, pr := range prs {
		if _, ok := PRLinkMap[pr.Link]; !ok {
			if _, err := o.QueryTable("pull").Filter("link", pr.Link).Delete(); err != nil {
				logs.Error("Fail to clean pull record:", pr.Link)
			}
		}
	}
	logs.Info("Ends of pulls sync, wait the next time...")
	return nil
}

func syncEnterpriseIssues() error {
	logs.Info("Starting to sync issues...")
	var repos []models.Repo
	o := orm.NewOrm()
	if _, err := o.QueryTable("repo").All(&repos); err != nil {
		logs.Error("Fail to query repo")
		return err
	}
	issueNumberMap := make(map[string]bool)
	for _, repo := range repos {
		page := defaultPage
		for {
			url := fmt.Sprintf("%v/repos/%v/issues?state=all&sort=created&direction=desc&page=%v&per_page=%v&"+
				"access_token=%v", config.AppConfig.GiteeV5ApiPrefix, repo.Name, page, perPageMax,
				config.AppConfig.AccessToken)
			resp, err := http.Get(url)
			if err != nil {
				page += 1
				continue
			}
			if resp.StatusCode != http.StatusOK {
				page += 1
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			if err = resp.Body.Close(); err != nil {
				logs.Error("Fail to close response body of enterprise issues, err:", err)
				page += 1
				continue
			}

			var ris []utils.ResponseIssue
			if err = json.Unmarshal(body, &ris); err != nil {
				logs.Error("Fail to unmarshal response to json, err:", err)
				page += 1
				continue
			}
			if len(ris) == 0 {
				break
			}
			for _, issue := range ris {
				labels := issue.Labels
				tags := make([]string, 0)
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
						tags = append(tags, label.Name)
					}
				}

				var ti models.Issue
				ti.Org = strings.Split(repo.Name, "/")[0]
				ti.Repo = repo.Name
				ti.Sig = repo.Sig
				ti.Link = issue.HtmlUrl
				ti.Number = issue.Number
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
				if !models.SearchIssueRecord(issue.Number) {
					if _, err = o.Insert(&ti); err != nil {
						logs.Error("Insert issue failed, err:", err)
					}
				} else {
					if _, err = o.QueryTable("issue").Filter("number", ti.Number).Update(orm.Params{
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
						logs.Error("Update issue failed, err:", err)
					}
				}
				issueNumberMap[ti.Number] = true
			}
			page += 1
		}
	}

	var issues []models.Issue
	if _, err := o.QueryTable("issue").All(&issues); err != nil {
		logs.Error("Fail to query issues")
		return err
	}
	for _, issue := range issues {
		if _, ok := issueNumberMap[issue.Number]; !ok {
			if _, err := o.QueryTable("issue").Filter("number", issue.Number).Delete(); err != nil {
				logs.Error("Fail to clean issue record:", issue.Number)
			}
		}
	}
	logs.Info("Ends of issues sync, wait the next time...")
	return nil
}

func syncEnterpriseRepos() error {
	logs.Info("Starting to sync repos...")
	repos := utils.GetSigsMapping()
	if repos == nil {
		logs.Error("Fail to get sigs mapping.")
		return nil
	}
	o := orm.NewOrm()
	for repo, sig := range repos {
		var tr models.Repo
		tr.Name = repo
		tr.Sig = sig
		if !models.SearchRepoRecord(repo) {
			if _, err := o.Insert(&tr); err != nil {
				logs.Error("Insert repo failed, err:", err)
			}
		} else {
			if _, err := o.QueryTable("repo").Filter("name", tr.Name).Update(orm.Params{
				"sig": tr.Sig,
			}); err != nil {
				logs.Error("Update repo failed, err:", err)
			}
		}
	}

	reposMap := syncRepoNumber()

	var reposData []models.Repo
	if _, err := o.QueryTable("repo").All(&reposData); err != nil {
		return err
	}
	for _, repoData := range reposData {
		if _, ok := reposMap[repoData.Name]; !ok {
			if _, err := o.QueryTable("repo").Filter("name", repoData.Name).Delete(); err != nil {
				logs.Error("Fail to remove repo:", repoData.Name)
			}
		}
	}
	logs.Info("Ends of repos sync, wait the next time...")
	return nil
}

type repoResponse struct {
	Id       int    `json:"id"`
	FullName string `json:"full_name"`
	Public   bool   `json:"public"`
	Status   string `json:"status"`
}

func syncRepoNumber() (reposMap map[string]bool) {
	logs.Info("Starting to sync repos numbers...")
	reposMap = make(map[string]bool)
	page := defaultPage
	for {
		url := fmt.Sprintf("%v/enterprises/open_euler/repos?type=all&page=%v&per_page=%v&access_token=%v",
			config.AppConfig.GiteeV5ApiPrefix, page, perPageMax, config.AppConfig.AccessToken)
		resp, err := http.Get(url)
		if err != nil {
			page += 1
			continue
		}
		if resp.StatusCode != http.StatusOK {
			page += 1
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		if err = resp.Body.Close(); err != nil {
			logs.Error("Fail to close response body of enterprise repos, err：", err)
			page += 1
			continue
		}

		var repos []repoResponse
		if err = json.Unmarshal(body, &repos); err != nil {
			logs.Error("Fail to unmarshal response to json, err:", err)
			page += 1
			continue
		}
		if len(repos) == 0 {
			break
		}
		for _, repo := range repos {
			name := repo.FullName
			number := repo.Id
			public := repo.Public
			status := repo.Status
			if public != true && status != "开始" {
				continue
			}
			var r models.Repo
			r.Name = name
			r.EnterpriseNumber = number
			if models.SearchRepoRecord(name) {
				o := orm.NewOrm()
				if _, err = o.QueryTable("repo").Filter("name", name).Update(orm.Params{
					"enterprise_number": number,
				}); err != nil {
					logs.Error("Update repo enterprise number failed, err:", err)
				} else {
					reposMap[name] = true
				}
			}
		}
		page += 1
	}
	logs.Info("Ends of repos numbers sync, wait the next time...")
	return reposMap
}

func runTasks() {
	tk1 := task.NewTask("syncEnterprisePulls", "0 0 3 * * ?", func(ctx context.Context) error {
		return syncEnterprisePulls()
	})
	tk2 := task.NewTask("syncEnterpriseIssues", "0 0 1 * * ?", func(ctx context.Context) error {
		return syncEnterpriseIssues()
	})
	tk3 := task.NewTask("syncEnterpriseRepos", "0 0 0 * * ?", func(ctx context.Context) error {
		return syncEnterpriseRepos()
	})
	tk4 := task.NewTask("cleanVerification", "0 * * * * *", func(ctx context.Context) error {
		return controllers.CleanVerification()
	})
	task.AddTask("syncEnterprisePulls", tk1)
	task.AddTask("syncEnterpriseIssues", tk2)
	task.AddTask("syncEnterpriseRepos", tk3)
	task.AddTask("cleanVerification", tk4)
	task.StartTask()
	defer task.StopTask()
}