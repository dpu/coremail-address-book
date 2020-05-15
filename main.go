package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	protocol		  string
	host              string
	cookieCoreMail    string
	cookieCoreMailSid string
)

func init() {
	flag.StringVar(&protocol, "protocol", "http", "Coremail login url protocol")
	flag.StringVar(&host, "host", "mail.dlpu.edu.cn", "Coremail host")
	flag.StringVar(&cookieCoreMail, "coremail_cookie", "", "Coremail value in Request Cookie")
	flag.StringVar(&cookieCoreMailSid, "coremail_sid", "", "Coremail.sid value in Request Cookie")
}

func main() {
	flag.Parse()
	url := buildURL(protocol, host, cookieCoreMailSid)

	respCount := getResp(url, 0, 1, fmt.Sprintf("Coremail=%s", cookieCoreMail))

	var responseCount Response
	json.Unmarshal([]byte(respCount), &responseCount)

	f, err := os.Create(host + ".email_list.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	w.Write([]string{"name", "email"})

	for i := 0; i < responseCount.Total; i += 100 {
		respList := getResp(url, i, 100, fmt.Sprintf("Coremail=%s", cookieCoreMail))
		var responseList Response;
		json.Unmarshal([]byte(respList), &responseList)
		for _, item := range responseList.Items {
			w.Write([]string{item.Name, item.Email})
		}
	}
	w.Flush()
}

func buildURL(protocol string, host string, sid string) string {
	return protocol + "://" + host + "/coremail/s/json?func=oab%3AlistEx&sid=" + sid;
}

func buildPostData(start int, limit int) string {
	return fmt.Sprintf("{\"dn\":\"a\",\"returnAttrs\":[\"true_name\",\"email\"],\"start\":%d,\"limit\":%d,\"defaultReturnMeetingRoom\":false}", start, limit)
}

func getResp(url string, start int, limit int, cookie string) string {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(buildPostData(start, limit)))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(body)
}

type Response struct {
	Code  string `json:"code"`
	Total int    `json:"total"`
	Items []Item `json:"var"`
}

type Item struct {
	Name  string `json:"true_name"`
	Email string `json:"email"`
}