package models

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"issue_pr_board/config"
	"issue_pr_board/utils"
	"os"
	"path"
	"reflect"
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
	searchSql := "select * from issue_type where name=? and platform=? and organization=?"
	err := o.Raw(searchSql, name, platform, organization).QueryRow(&issueType)
	if err == orm.ErrNoRows {
		return false, 0
	}
	return true, issueType.Id
}

func InitIssueType() {
	organizations, err := os.ReadDir(path.Join("templates", "issues"))
	if err != nil {
		logs.Error("Fail to get organization directory list, err:", err)
	}
	for _, organization := range organizations {
		files, err := os.ReadDir(path.Join("templates", "issues", organization.Name()))
		if err != nil {
			logs.Error("Fail to get templates directory list, err:", err)
		}
		orgFiles := make([]string, 0)
		for _, file := range files {
			orgFiles = append(orgFiles, file.Name())
		}
		if !utils.InMap(utils.ConvertStrSlice2Map(orgFiles), "issue_types.yaml") {
			continue
		}
		var info = &[]IssueTypeInfo{}
		err = config.LoadFromYaml(path.Join("templates", "issues", organization.Name(), "issue_types.yaml"), info)
		if err != nil {
			logs.Error(err)
			return
		}
		confIssueTypes := make([]map[string]interface{}, 0)
		var issueType IssueType
		for _, i := range *info {
			issueType.Organization = organization.Name()
			issueType.Name = i.Name
			if utils.InMap(utils.ConvertStrSlice2Map(orgFiles), i.Name+".md") {
				templateFile := path.Join("templates", "issues", organization.Name(), i.Name+".md")
				data, err := os.ReadFile(templateFile)
				if err != nil {
					logs.Error("Fail to read issue type template, err:", err)
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
				if exist {
					o := orm.NewOrm()
					qs := o.QueryTable("issue_type")
					_, err = qs.Filter("id", issueTypeId).Update(orm.Params{
						"unique_id": issueType.UniqueId,
						"template":  issueType.Template,
					})
					if err != nil {
						logs.Error("Update issue_type failed, err:", err)
					}
				} else {
					o := orm.NewOrm()
					newIssueType := IssueType{
						Name:         issueType.Name,
						UniqueId:     issueType.UniqueId,
						Platform:     issueType.Platform,
						Organization: issueType.Organization,
						Template:     issueType.Template,
					}
					_, err = o.Insert(&newIssueType)
					if err != nil {
						logs.Error("Insert issue_type failed, err:", err)
					}
				}
			}
		}
		var dbIssueTypes []IssueType
		sql := "select * from issue_type where organization=?"
		o := orm.NewOrm()
		_, err = o.Raw(sql, organization.Name()).QueryRows(&dbIssueTypes)
		if err != nil {
			logs.Error("Fail to query issue types, err:", err)
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
				o := orm.NewOrm()
				qs := o.QueryTable("issue_type")
				_, err = qs.Filter("id", id).Delete()
				if err != nil {
					logs.Error("Clean redundant issue type failed, err:", err)
				}
			}
		}
	}
}
