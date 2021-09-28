package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	protocol          string
	host              string
	cookieCoreMail    string
	cookieCoreMailSid string
)

type Response struct {
	Code  string `json:"code"`
	Total int    `json:"total"`
	Items []Item `json:"var"`
}

type Item struct {
	Name  string `json:"true_name"`
	Email string `json:"email"`
}

func init() {
	flag.StringVar(&protocol, "protocol", "http", "Coremail login url protocol")
	flag.StringVar(&host, "host", "mail.dlpu.edu.cn", "Coremail host")
	flag.StringVar(&cookieCoreMail, "coremail_cookie", "", "Coremail value in Request Cookie")
	flag.StringVar(&cookieCoreMailSid, "coremail_sid", "", "Coremail.sid value in Request Cookie")
}

func getAllGroup() [][]string {
	url := buildURL("oab:getDirectories")
	postData := fmt.Sprintf(`{dn: "a", attrIds: ["email"]}`)
	resp := getResp(url, fmt.Sprintf("Coremail=%s", cookieCoreMail), postData)

	//使用正则匹配出所有的组群
	reg := regexp.MustCompile(`(?s)id":"(.{32})".*?"name":"(.*?)"`)
	result1 := reg.FindAllStringSubmatch(resp, -1)
	// fmt.Println("==================")
	// fmt.Println(result1)
	// fmt.Println("==================")
	// for i := 0; i < len(result1); i++ {
	// 	fmt.Println(result1[i][1], result1[i][2])
	// }
	// os.Exit(0)
	return result1
}

func main() {
	flag.Parse()

	// 先获取组群
	allGroupID := getAllGroup()

	f, err := os.Create(host + ".email_list.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	w.Write([]string{"name", "email"})

	// 根据组群ID获取组群下的所有通讯录
	for i := 0; i < len(allGroupID); i++ {
		url := buildURL("oab:listEx")
		// a表示all，"/"后面就是组群ID
		postData := fmt.Sprintf(`{"dn":"a/%s","returnAttrs":["true_name","email"],"start":%d,"limit":%d,"defaultReturnMeetingRoom":false}`, allGroupID[i][1], 0, 1)
		fmt.Println("======", postData)
		respCount := getResp(url, fmt.Sprintf("Coremail=%s", cookieCoreMail), postData)

		var responseCount Response
		json.Unmarshal([]byte(respCount), &responseCount)

		for j := 0; j < responseCount.Total; j += 100 {
			postData := fmt.Sprintf(`{"dn":"a/%s","returnAttrs":["true_name","email"],"start":%d,"limit":%d,"defaultReturnMeetingRoom":false}`, allGroupID[i][1], j, 100)
			respList := getResp(url, fmt.Sprintf("Coremail=%s", cookieCoreMail), postData)
			var responseList Response
			json.Unmarshal([]byte(respList), &responseList)
			for _, item := range responseList.Items {
				// 联系人前面添加组群名，针对同名的人更便于区分
				w.Write([]string{allGroupID[i][2] + "-" + item.Name, item.Email})
			}
		}
	}

	w.Flush()
}

func buildURL(fc string) string {
	return protocol + "://" + host + "/coremail/s/json?func=" + fc + "&sid=" + cookieCoreMailSid
}

func getResp(url string, cookie, postData string) string {
	client := &http.Client{}

	fmt.Printf("[%s]\n", url)
	req, err := http.NewRequest("POST", url, strings.NewReader(postData))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()

	fmt.Printf("[%s]\n", string(body))
	return string(body)
}
