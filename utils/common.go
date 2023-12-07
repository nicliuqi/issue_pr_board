package utils

import (
	"encoding/json"
	"fmt"
	"github.com/TestsLing/aj-captcha-go/util"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/chenhg5/collection"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func JsonToSlice(str string) []map[string]interface{} {
	var temSlice []map[string]interface{}
	err := json.Unmarshal([]byte(str), &temSlice)
	if err != nil {
		logs.Error(err)
		logs.Error("Parse string to slice error, the string is:", str)
		return nil
	}
	return collection.Collect(str).ToMapArray()
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
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/openeuler/community/git/trees/master?access_token=%s"+
		"&recursive=1", os.Getenv("AccessToken"))
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("Fail to get sigs mapping, err: %v", err)
		return nil, nil
	}
	body, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		logs.Error("Fail to close response body of getting sigs mapping, err:", err)
		return nil, nil
	}
	treeMap := JsonToMap(string(body))
	if treeMap == nil {
		return nil, nil
	}
	sigs := map[string][]string{}
	repos := map[string]string{}
	for _, value := range treeMap["tree"].([]interface{}) {
		path := value.(map[string]interface{})["path"]
		pathSlices := strings.Split(path.(string), "/")
		if len(pathSlices) == 5 && strings.HasPrefix(path.(string), "sig") &&
			strings.HasSuffix(path.(string), ".yaml") {
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

func CheckAuth(authorization string) bool {
	ts := time.Now().Unix()
	authJson := util.AesDecrypt(authorization, os.Getenv("AUTH_SECRET"))
	if authJson == "" {
		return false
	}
	authMap := collection.Collect(authJson).ToMap()
	secret, ok := authMap["secret"].(string)
	if !ok {
		return false
	}
	timestamp, ok2 := authMap["timestamp"].(float64)
	if !ok2 {
		return false
	}
	interval, _ := beego.AppConfig.Int64("verifyinterval")
	if ts-int64(timestamp) < 0 {
		return false
	}
	if ts-int64(timestamp) > interval {
		return false
	}
	if secret != os.Getenv("AUTH_SECRET") {
		return false
	} else {
		return true
	}
}
