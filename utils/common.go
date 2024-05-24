package utils

import (
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"io"
	"issue_pr_board/config"
	"net/http"
	"strings"
)

type WebhookRequest struct {
	Action      string       `json:"action"`
	Issue       RequestIssue `json:"issue"`
	PullRequest RequestPull  `json:"pull_request"`
	Author      Author       `json:"author"`
	Comment     Comment      `json:"comment"`
}

type RequestIssue struct {
	ResponseIssue
}

type RequestPull struct {
	ResponsePull
}

type ResponsePull struct {
	HtmlUrl   string                 `json:"html_url"`
	State     string                 `json:"state"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Labels    []ResponsePullLabel    `json:"labels"`
	Assignees []ResponsePullAssignee `json:"assignees"`
	Draft     bool                   `json:"draft"`
	MergeAble bool                   `json:"mergeable"`
	User      ResponsePullUser       `json:"user"`
	Base      ResponsePullBase       `json:"base"`
}

type ResponsePullLabel struct {
	Name  string  `json:"name"`
	Color string  `json:"color"`
	Id    float64 `json:"id"`
}

type ResponsePullAssignee struct {
	Login string `json:"login"`
}

type ResponsePullUser struct {
	Login string `json:"login"`
}

type ResponsePullBase struct {
	Ref string `json:"ref"`
}

type ResponseIssue struct {
	Repository       ResponseIssueRepository  `json:"repository"`
	HtmlUrl          string                   `json:"html_url"`
	User             ResponseIssueUser        `json:"user"`
	Number           string                   `json:"number"`
	State            string                   `json:"state"`
	IssueType        string                   `json:"issue_type"`
	IssueStateDetail ResponseIssueStateDetail `json:"issue_state_detail"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
	Milestone        ResponseIssueMilestone   `json:"milestone"`
	Assignee         ResponseIssueAssignee    `json:"assignee"`
	Title            string                   `json:"title"`
	Description      string                   `json:"body"`
	Labels           []ResponseIssueLabel     `json:"labels"`
	Priority         float64                  `json:"priority"`
	Branch           string                   `json:"branch"`
}

type ResponseIssueRepository struct {
	FullName string `json:"full_name"`
}

type ResponseIssueUser struct {
	Login string `json:"login"`
}

type ResponseIssueStateDetail struct {
	Title string `json:"title"`
}

type ResponseIssueMilestone struct {
	Title string `json:"title"`
}

type ResponseIssueAssignee struct {
	Login string `json:"login"`
}

type ResponseIssueLabel struct {
	Name  string  `json:"name"`
	Color string  `json:"color"`
	Id    float64 `json:"id"`
}

type Author struct {
	Login string `json:"login"`
}

type Comment struct {
	Body string `json:"body"`
}

type ResponseRepoDir struct {
	Tree []ResponseRepoTree `json:"tree"`
}

type ResponseRepoTree struct {
	Path string `json:"path"`
}

func JsonToMap(str string) map[string]interface{} {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		logs.Error(err)
		logs.Error("Parse string to map error, the string is:", str)
		return nil
	}
	return tempMap
}

func FormatTime(createdAt string) string {
	createdStr := strings.Replace(createdAt[:len(createdAt)-6], "T", " ", -1)
	return createdStr
}

func GetSigsMapping() (map[string][]string, map[string]string) {
	url := fmt.Sprintf("%v/repos/openeuler/community/git/trees/master?access_token=%s"+
		"&recursive=1", config.AppConfig.GiteeV5ApiPrefix, config.AppConfig.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get sigs mapping, err: %v", err)
		return nil, nil
	}
	body, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of getting sigs mapping, err:", err)
		return nil, nil
	}
	var rrd ResponseRepoDir
	err = json.Unmarshal(body, &rrd)
	if err != nil {
		logs.Error("Fail to unmarshal response to json, err:", err)
		return nil, nil
	}
	sigs := map[string][]string{}
	repos := map[string]string{}
	for _, tree := range rrd.Tree {
		path := tree.Path
		pathSlices := strings.Split(path, "/")
		if len(pathSlices) == 5 && strings.HasPrefix(path, "sig") &&
			strings.HasSuffix(path, ".yaml") {
			sigName := pathSlices[1]
			repoName := pathSlices[2] + "/" + pathSlices[4][:len(pathSlices[4])-5]
			repos[repoName] = sigName
			_, ok := sigs[sigName]
			if !ok {
				sigs[sigName] = []string{repoName}
			} else {
				sigs[sigName] = append(sigs[sigName], repoName)
			}
		}
	}
	return sigs, repos
}

func GetSigByRepo(repos map[string]string, repo string) string {
	sig, ok := repos[repo]
	if !ok {
		return ""
	}
	return sig
}

func CheckParams(param string) string {
	warningWords := []string{" ", "'", "\"", "<", ">", "=", "&", "\\", "#", ";", "(", ")", "%", "!"}
	for _, warningWord := range warningWords {
		if strings.Contains(param, warningWord) {
			return ""
		}
	}
	return param
}

func CheckMilestonesParams(param string) string {
	warningWords := []string{"'", "\"", "<", ">", "=", "&", "\\", "#", ";", "(", ")", "%", "!"}
	for _, warningWord := range warningWords {
		if strings.Contains(param, warningWord) {
			return ""
		}
	}
	return param
}

func ConvertStrSlice2Map(sl []string) map[string]struct{} {
	set := make(map[string]struct{}, len(sl))
	for _, v := range sl {
		set[v] = struct{}{}
	}
	return set
}

func InMap(m map[string]struct{}, s string) bool {
	_, ok := m[s]
	return ok
}
