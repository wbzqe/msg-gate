package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/toolkits/web/param"
	"github.com/wbzqe/msg-gate/config"
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type Content struct {
	Content string `json:"content"`
}

type MsgPost struct {
	ToUser  string  `json:"touser"`
	MsgType string  `json:"msgtype"`
	AgentID string  `json:"agentid"`
	Text    Content `json:"text"`
}

func HdQywx(w http.ResponseWriter, r *http.Request) {
	tos := param.MustString(r, "tos")
	content := param.MustString(r, "content")

	if userList := strings.Split(tos, ","); len(userList) > 1 {
		tos = strings.Join(userList, "|")
	}

	err := SendQywx(tos, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Error(w, "success", http.StatusOK)
	}
}

func SendQywx(tos string, content string) error {
	//content := "[P0][OK][192.168.11.26_ofmon][][【critical】与主mysql同步延迟超过10s！ all(#3) seconds_behind_master port=3306 0>10][O1 2017-04-17 08:55:00]"
	content = strings.Replace(content, "][", "\n", -1)
	if content[0] == '[' {
		content = content[1:]
	}

	if content[len(content)-1] == ']' {
		content = content[:len(content)-1]
	}

	text := Content{}
	text.Content = content

	cfg := config.Config()
	msg := MsgPost{
		ToUser:  tos,
		MsgType: "text",
		AgentID: cfg.Qywx.AgentId,
		Text:    text,
	}

	token, err := GetAccessToken()
	if err != nil {
		log.Printf("get token failed!")
		return err
	}

	url := "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=" + token.AccessToken

	result, err := WxPost(url, msg)
	if err != nil {
		log.Printf("request qywx failed: %v", err)
	}
	return fmt.Errorf(string(result))
}

func GetAccessToken() (AccessToken, error) {

	value, found := TokenCache.Get("token")
	if found {
		accessToken, ok := value.(AccessToken)
		if !ok {
			return accessToken, fmt.Errorf("token parse failed! ")
		}
		return accessToken, nil
	}

	cfg := config.Config()

	WxAccessTokenUrl := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + cfg.Qywx.CorpID + "&corpsecret=" + cfg.Qywx.Secret

	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	newAccess := AccessToken{}

	result, err := client.Get(WxAccessTokenUrl)
	if err != nil {
		return newAccess, err
	}

	res, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return newAccess, err
	}

	err = json.Unmarshal(res, &newAccess)
	if err != nil {
		return newAccess, err
	}

	if newAccess.ExpiresIn == 0 || newAccess.AccessToken == "" {
		return newAccess, fmt.Errorf("ErrCode: %v, ErrMsg: %v", newAccess.ErrCode, newAccess.ErrMsg)
	}

	TokenCache.Set("token", newAccess, time.Duration(newAccess.ExpiresIn)*time.Second)

	return newAccess, nil

}

//微信请求数据
func WxPost(url string, data MsgPost) (string, error) {
	jsonBody, err := encodeJson(data)
	if err != nil {
		return "", err
	}

	r, err := http.Post(url, "application/json;charset=utf-8", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}

//json序列化(禁止 html 符号转义)
func encodeJson(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
