### API参考

**endpoint**: https://ipb.osinfra.cn

- issue列表
  - 功能介绍

    获取issue列表
  - URI

    GET /issues
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | org | 否 | string | issue所属组织
    | repo | 否 | string | issue所属仓库
    | sig | 否 | string | issue所属SIG
    | state | 否 | string | issue的状态
    | number | 否 | string | issue编号
    | author | 否 | string | 提交人
    | assignee | 否 | string | 指派者
    | branch | 否 | string | 指定分支
    | label | 否 | string | issue的标签
    | exclusion | 否 | string | 过滤的issue标签
    | issue_state | 否 | string | issue的具体状态
    | issue_type | 否 | string | issue的类型
    | priority | 否 | int | 优先级
    | sort | 否 | string | 排序方式，默认created_at
    | direction | 否 | string | 排序的顺序，默认desc
    | search | 否 | string | 模糊搜索字段，匹配issue编号、仓库和标题
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/issues?number=I5LKZQ
    ```

    输出示例
    ```
    {
      "total": 1,
      "page": 1,
      "per_page": 10,
      "data": [
        {
          "org": "src-openeuler",
          "repo": "src-openeuler/rubygem-redis",
          "sig": "sig-ruby",
          "link": "https://gitee.com/src-openeuler/rubygem-redis/issues/I5LKZQ",
          "number": "I5LKZQ",
          "state": "closed",
          "issue_type": "缺陷",
          "issue_state": "已完成",
          "author": "zhouxiaxiang",
          "reporter": "",
          "assignee": "small_leek",
          "created_at": "2022-08-10 11:50:10",
          "updated_at": "2022-08-12 16:13:49",
          "title": "rubygem-redis build problem in openEuler:22.03:LTS",
          "priority": "次要",
          "labels": "sig/sig-ruby",
          "branch": "",
          "milestone": "openEuler-22.03-LTS-Dailybuild"
        }
      ]
    }
    ```

- issue指派者列表
  - 功能介绍

    获取issue所有指派者的列表
  - URI

    GET /issues/assignees
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配issue的指派者
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/issues/assignees?keyword=al
    ```

    输出示例
    ```
    {
      "total": 12,
      "page": 1,
      "per_page": 20,
      "data": [
        "alec-z",
        "AlexZ11",
        "algorithmofdish",
        "allesgute",
        "calvinyu",
        "qiegewala",
        "rabbitali",
        "small_leek",
        "walkingwalk",
        "XWwalker",
        "yjt-dali",
        "zhangjialin11"
      ]
    }
    ```

- issue提交人列表
  - 功能介绍

    获取issue所有提交人的列表
  - URI

    GET /issues/authors
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配issue的提交人，keyword中包含"@"时匹配reporter
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/issues/authors?keyword=469
    ```

    输出示例
    ```
    {
      "total": 2,
      "page": 1,
      "per_page": 20,
      "data": [
        "469227928@***m",
        "liuqi469227928"
      ]
    }
    ```

- issue分支列表
  - 功能介绍

    获取issue所有指定分支的列表
  - URI

    GET /issues/branches
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配issue的指定分支
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/issues/branches?keyword=22.03
    ```

    输出示例
    ```
    {
      "total": 4,
      "page": 1,
      "per_page": 20,
      "data": [
        "openEuler-22.03-LTS",
        "openEuler-22.03-LTS-Next",
        "openEuler-22.03-LTS-round2",
        "openEuler-22.03-LTS-SP1"
      ]
    }
    ```

- issue标签列表
  - 功能介绍

    获取issue所有标签的列表
  - URI

    GET /issues/labels
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配issue的标签
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/issues/labels?page=2&per_page=10
    ```

    输出示例
    ```
    {
      "total": 167,
      "page": 2,
      "per_page": 10,
      "data": [
        "sig/sig-RaspberryPi",
        "issue_feature",
        "kind/design",
        "priority/high",
        "sig/A-Tune",
        "sig/Application",
        "sig/Base-service",
        "sig/DB",
        "sig/iSulad",
        "sig/Networking"
      ]
    }
    ```

- issue类型列表
  - 功能介绍

    获取issue所有类型的列表
  - URI

    GET /issues/types
  - 请求参数

    无
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/issues/types
    ```

    输出示例
    ```
    {
      "code": 200,
      "msg": "请求成功",
      "data": [
        {
        "name": "任务",
        "id": 118736,
        "platform": "gitee",
        "organization": "openEuler",
        "template": "<!-- #请根据issue的类型在标题右侧下拉框中选择对应的选项（需求、缺陷或CVE等）-->"
        },
        {
        "name": "开源之夏2022",
        "id": 205654,
        "platform": "gitee",
        "organization": "openEuler",
        "template": "项目标题：（填在issue标题处即可）\n【项目难度】：（基础、进阶）\n【项目描述】：（清楚描述本项目背景以及要做什么）\n【产出标准】：1、2、（分条列出需要产出哪些东西、合格标准）\n【技术要求】：1、2、（分条列出承接本项目需要初步具备哪些开发技能、编程语言要求）\n【导师姓名/导师邮箱】：\n【成果提交仓库】：\n【相关参考资料】：（包括项目的代码仓库链接、学习资料等）\n\n本次活动将面向全球学生，题目内容请同时提交中英文版本。\n\nproject name：\n【Difficulty】：\n【Description】：\n【Output Requirements】：1、2、\n【Technical Requirements】：1、2、\n【Mentor/Email】：\n【Project Repository】：\n【Notes】：\n"
        },
        ...
      ]
    }
    ```

- 仓库列表
  - 功能介绍

    获取所有openeuler/src-openeuler组织下仓库的列表
  - URI

    GET /repos
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配仓库名
    | sig | 否 | string | 仓库所属SIG
    | direction | 否 | string | 排序，默认按仓库名称排序
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/repos?sig=Infrastructure
    ```

    输出示例
    ```
    {
      "total": 8,
      "page": 1,
      "per_page": 10,
      "data": [
        {
          "repo": "openeuler/cve-ease",
          "sig": "Infrastructure"
        },
        {
          "repo": "openeuler/cve-manager",
          "sig": "Infrastructure"
        },
        ...
      ]
    }
    ```

- Pull Request列表
  - 功能介绍

    获取所有openeuler/src-openeuler组织下仓库的Pull Request列表
  - URI

    GET /pulls
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | org | 否 | string | Pull Request所属组织
    | repo | 否 | string | Pull Request所属仓库
    | sig | 否 | string | Pull Request所属SIG
    | state | 否 | string | Pull Request的状态
    | ref | 否 | string | Pull Request指定的分支
    | author | 否 | string | 提交人
    | assignee | 否 | string | 指派者
    | sort | 否 | string | 排序方式，默认created_at
    | direction | 否 | string | 排序的顺序，默认desc
    | label | 否 | string | Pull Request的标签
    | exclusion | 否 | string | 过滤的Pull Request标签
    | search | 否 | string | 模糊搜索字段，匹配SIG、仓库和标题
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/pulls?state=open&ref=openEuler-22.09
    ```

    输出示例
    ```
    {
      "total": 32,
      "page": 1,
      "per_page": 10,
      "data": [
        {
          "org": "openeuler",
          "repo": "openeuler/yocto-meta-openeuler",
          "ref": "openEuler-22.09",
          "sig": "sig-Yocto",
          "link": "https://gitee.com/openeuler/yocto-meta-openeuler/pulls/790",
          "state": "open",
          "author": "saarloos",
          "assignees": "vonhust,beilingxie,ilisimin,fanglinxu",
          "created_at": "2022-12-19 19:44:36",
          "updated_at": "2022-12-20 10:55:36",
          "title": "lcr: update lcr version",
          "labels": "openeuler-cla/yes,ci_failed,sig/sig-Yocto",
          "draft": false,
          "mergeable": true
        },
        {
          "org": "openeuler",
          "repo": "openeuler/yocto-meta-openeuler",
          "ref": "openEuler-22.09",
          "sig": "sig-Yocto",
          "link": "https://gitee.com/openeuler/yocto-meta-openeuler/pulls/787",
          "state": "open",
          "author": "fanglinxu",
          "assignees": "vonhust,beilingxie,ilisimin,fanglinxu",
          "created_at": "2022-12-19 14:22:50",
          "updated_at": "2022-12-19 14:30:45",
          "title": " mcs: fix url error of mcs demo of 22.09 ",
          "labels": "openeuler-cla/yes,ci_failed,sig/sig-Yocto",
          "draft": false,
          "mergeable": true
        },
        ...
      ]
    }
    ```

- Pull Request指派者列表
  - 功能介绍

    获取所有Pull Request指派者的列表
  - URI

    GET /pulls/assignees
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配Pull Request的指派者
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/pulls/assignees?keyword=alex
    ```

    输出示例
    ```
    {
      "total": 2,
      "page": 1,
      "per_page": 20,
      "data": [
        "AlexZ11",
        "alexanderbill"
      ]
    }
    ```

- Pull Request提交人列表
  - 功能介绍

    获取所有Pull Request提交人的列表
  - URI

    GET /pulls/authors
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配Pull Request的提交人
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/pulls/authors?keyword=al
    ```

    输出示例
    ```
    {
      "total": 67,
      "page": 1,
      "per_page": 20,
      "data": [
        "a_night_of_baldness",
        "alapha",
        "albert-lee-7",
        "alec-z",
        "alexanderbill",
        "alexicy",
        "AlexZ11",
        "AlfredEinstein",
        "algorithmofdish",
        "Ali-JR",
        "aliceye666",
        "alichinese",
        "alignment",
        "aliyutaozi",
        "alize029",
        "Allen-Maker",
        "allen-shi",
        "almighty1982",
        "almzmx",
        "alonelur"
      ]
    }
    ```

- Pull Request分支列表
  - 功能介绍

    获取所有Pull Request的分支列表
  - URI

    GET /pulls/refs
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配Pull Request的指定分支
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/pulls/refs?keyword=openeuler-22.09
    ```

    输出示例
    ```
    {
      "total": 5,
      "page": 1,
      "per_page": 10,
      "data": [
        "Multi-Version_obs-server-2.10.11_openEuler-22.09",
        "openEuler-22.09",
        "openEuler-22.09-HCK",
        "openEuler-22.09-next",
        "sync-pr214-openEuler-22.03-LTS-to-openEuler-22.09"
      ]
    }
    ```

- Pull Request标签列表
  - 功能介绍

    获取所有Pull Request的标签列表
  - URI

    GET /pulls/labels
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配Pull Request的标签
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/pulls/labels?keyword=kind
    ```

    输出示例
    ```
    {
      "total": 8,
      "page": 1,
      "per_page": 20,
      "data": [
        "kind/bug",
        "kind/cleanup",
        "kind/design",
        "kind/docs",
        "kind/documentation",
        "kind/enhancement",
        "kind/feature",
        "kind/wait_for_update"
      ]
    }
    ```

- Pull Request仓库列表
  - 功能介绍

    获取所有Pull Request的仓库列表
  - URI

    GET /pulls/repos
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | sig | 否 | string | issue所属SIG
    | keyword | 否 | string | 模糊匹配Pull Request的所属仓库
    | page | 否 | int | 页数，默认1
    | per_page | 否 | int | 每页数量，默认10，最大100
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/pulls/repos?sig=tc
    ```

    输出示例
    ```
    {
      "total": 2,
      "page": 1,
      "per_page": 10,
      "data": [
        "openeuler/community",
        "openeuler/TC"
      ]
    }
    ```

- SIG列表
  - 功能介绍

    获取所有SIG的列表
  - URI

    GET /pulls/sigs
  - 请求参数
    | 参数 | 是否必选 | 参数类型 | 描述
    | :---: | :---: | :---: | :---
    | keyword | 否 | string | 模糊匹配Pull Request的所属SIG
  - 示例

    输入示例
    ```
    GET https://ipb.osinfra.cn/pulls/sigs?keyword=com
    ```

    输出示例
    ```
    {
      "code": 200,
      "msg": "请求成功",
      "data": [
        "Compiler",
        "Computing",
        "security-committee",
        "sig-compat-winapp",
        "sig-Compatibility-Infra",
        "sig-compliance",
        "sig-confidential-computing",
        "user-committee"
      ]
    }
    ```
