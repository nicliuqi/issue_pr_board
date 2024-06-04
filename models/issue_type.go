package models

import (
	"os"
	"path"
	"reflect"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"

	"issue_pr_board/config"
	"issue_pr_board/utils"
)

type IssueType struct {
	Id           int    `json:"-"`
	Name         string `json:"name" orm:"size(50);null" description:"issue类型名称"`
	UniqueId     int    `json:"id" orm:"unique;null" description:"issue类型唯一id"`
	Platform     string `json:"platform" orm:"size(50);null" description:"平台"`
	Organization string `json:"organization" orm:"size(50);null" description:"组织"`
	Template     string `json:"template" orm:"type(text);null" description:"模板"`
}

type IssueTypeInfo struct {
	Name      string
	Platforms []IssueTypePlatform
}

type IssueTypePlatform struct {
	UniqueId int    `json:"unique_id"`
	Platform string `json:"platform"`
}

func searchIssueType(name string, platform string, organization string) (bool, int) {
	var issueType IssueType
	o := orm.NewOrm()
	if err := o.QueryTable("issue_type").Filter("name", name).Filter("platform", platform).
		Filter("organization", organization).One(&issueType); err != nil {
		logs.Error("[searchIssueType] Fail to search issue type, err:", err)
		return false, 0
	}
	return true, issueType.Id
}

func InitIssueType() {
	organizations, err := os.ReadDir(path.Join("templates", "issues"))
	if err != nil {
		logs.Error("[InitIssueType] Fail to get organization directory list, err:", err)
	}
	o := orm.NewOrm()
	for _, organization := range organizations {
		files, fileErr := os.ReadDir(path.Join("templates", "issues", organization.Name()))
		if fileErr != nil {
			logs.Error("[InitIssueType] Fail to get templates directory list, err:", fileErr)
		}
		orgFiles := make([]string, 0)
		for _, file := range files {
			orgFiles = append(orgFiles, file.Name())
		}
		if !utils.InMap(utils.ConvertStrSlice2Map(orgFiles), "issue_types.yaml") {
			continue
		}
		var info = &[]IssueTypeInfo{}
		if err = config.LoadFromYaml(path.Join("templates", "issues", organization.Name(), "issue_types.yaml"),
			info); err != nil {
			logs.Error("[InitIssueType] Fail to load from yaml file, err:", err)
			return
		}
		confIssueTypes := make([]map[string]interface{}, 0)
		var issueType IssueType
		for _, i := range *info {
			issueType.Organization = organization.Name()
			issueType.Name = i.Name
			if utils.InMap(utils.ConvertStrSlice2Map(orgFiles), i.Name+".md") {
				templateFile := path.Join("templates", "issues", organization.Name(), i.Name+".md")
				data, dataErr := os.ReadFile(templateFile)
				if dataErr != nil {
					logs.Error("[InitIssueType] Fail to read issue type template, err:", dataErr)
				}
				issueType.Template = string(data)
			} else {
				issueType.Template = ""
			}
			for _, platform := range i.Platforms {
				issueType.UniqueId = platform.UniqueId
				issueType.Platform = platform.Platform
				issueTypeMap := make(map[string]interface{})
				issueTypeMap["name"] = issueType.Name
				issueTypeMap["organization"] = issueType.Organization
				issueTypeMap["platform"] = issueType.Platform
				issueTypeMap["unique_id"] = issueType.UniqueId
				confIssueTypes = append(confIssueTypes, issueTypeMap)
				exist, issueTypeId := searchIssueType(issueType.Name, issueType.Platform, issueType.Organization)
				if !exist {
					newIssueType := IssueType{
						Name:         issueType.Name,
						UniqueId:     issueType.UniqueId,
						Platform:     issueType.Platform,
						Organization: issueType.Organization,
						Template:     issueType.Template,
					}
					if _, err = o.Insert(&newIssueType); err != nil {
						logs.Error("[InitIssueType] Fail to create issue type, err:", err)
					}
				} else {
					if _, err = o.QueryTable("issue_type").Filter("id", issueTypeId).Update(orm.Params{
						"unique_id": issueType.UniqueId,
						"template":  issueType.Template,
					}); err != nil {
						logs.Error("[InitIssueType] Fail to update issue type, err:", err)
					}
				}
			}
		}
		var dbIssueTypes []IssueType
		if _, err = o.QueryTable("issue_type").Filter("organization", organization.Name()).
			All(&dbIssueTypes); err != nil {
			logs.Error("[InitIssueType] Fail to query issue types, err:", err)
		}
		for _, dbIssueType := range dbIssueTypes {
			id := dbIssueType.Id
			dbIssueTypeMap := make(map[string]interface{})
			dbIssueTypeMap["name"] = dbIssueType.Name
			dbIssueTypeMap["organization"] = dbIssueType.Organization
			dbIssueTypeMap["platform"] = dbIssueType.Platform
			dbIssueTypeMap["unique_id"] = dbIssueType.UniqueId
			equal := false
			for _, confIssueType := range confIssueTypes {
				if reflect.DeepEqual(confIssueType, dbIssueTypeMap) {
					equal = true
					break
				}
			}
			if !equal {
				if _, err = o.QueryTable("issue_type").Filter("id", id).Delete(); err != nil {
					logs.Error("[InitIssueType] Fail to clean redundant issue type, err:", err)
				}
			}
		}
	}
}
