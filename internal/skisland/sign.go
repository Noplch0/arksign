package skisland

import (
	"arkSign/internal/settings"
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// 常量定义
const (
	urlGetTokenByPwd = "https://as.hypergryph.com/user/auth/v1/token_by_phone_password"
	urlOauth         = "https://as.hypergryph.com/user/oauth2/v2/grant"
	urlCerd          = "https://zonai.skland.com/api/v1/user/auth/generate_cred_by_code"
	urlPlayerInfo    = "https://zonai.skland.com/api/v1/game/player/binding"
	urlVerify        = "https://as.hypergryph.com/user/info/v1/basic"
)

// 游戏签到地址映射
var signUrlMap = map[string]string{
	"arknights": "https://zonai.skland.com/api/v1/game/attendance",
	"endfield":  "https://zonai.skland.com/api/v1/game/endfield/attendance",
}

type CharacterInfo struct {
	AppCode string 
	UID     string
	GameId  string
	Server  string
	Name    string
}

// ... 省略 loginInfo, OauthInfo, CerdInfo, header 等结构体定义 (与前文一致) ...

type loginInfo struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type OauthInfo struct {
	Token    string `json:"token"`
	AppCode  string `json:"appCode"`
	TypeCode int    `json:"type"`
}

type CerdInfo struct {
	Kind int    `json:"kind"`
	Code string `json:"code"`
}

type header struct {
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
	Did       string `json:"dId"`
	Vname     string `json:"vName"`
}

type nHeader struct {
	Sign string `json:"sign"`
	header
}

type headerAgent struct {
	Cred       string `header:"cred"`
	Agent      string `header:"User-Agent"`
	Encoding   string `header:"Accept-Encoding"`
	Connection string `header:"Connection"`
	nHeader
}

// --- 辅助工具函数 ---

func agent(cred string, header2 nHeader) headerAgent {
	return headerAgent{
		cred,
		"Skland/1.17.0 (com.hypergryph.skland; build:101700050; Android 34; ) Okhttp/4.11.0",
		"gzip",
		"close",
		header2,
	}
}

func setHeader() header {
	return header{
		Platform:  "",
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		Did:       "",
		Vname:     "",
	}
}

func getStrRespBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

func getRespBody(resp string, index string) interface{} {
	reJson := gojsonq.New().FromString(resp)
	return reJson.Find(index)
}

func string2Header(text string) (http.Header, error) {
	var headers map[string]interface{}
	err := json.NewDecoder(strings.NewReader(text)).Decode(&headers)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	header := make(http.Header)
	for k, v := range headers {
		if strValue, ok := v.(string); ok {
			header.Set(k, strValue)
		}
	}
	return header, nil
}

func EncodeSignCode(code string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(code))
	sha := hex.EncodeToString(h.Sum(nil))

	hash := md5.New()
	hash.Write([]byte(sha))
	return hex.EncodeToString(hash.Sum(nil))
}

// --- 核心业务函数 ---

func GetToken(phone string, passwd string) (string, error) {
	var accountInfo = loginInfo{Phone: phone, Password: passwd}
	accountJson, _ := json.Marshal(accountInfo)
	resp, err := http.Post(urlGetTokenByPwd, "application/json", bytes.NewBuffer(accountJson))
	if err != nil {
		return "", errors.New("login Failed")
	}
	respString := getStrRespBody(resp)
	if resp.StatusCode != 200 {
		return "", errors.New("login Failed(status code not 200)")
	}
	s := getRespBody(respString, "data.token")
	return s.(string), nil
}

func VerifyToken(token string) bool {
	params := url.Values{}
	params.Set("token", token)
	resp, err := http.Get(urlVerify + "?" + params.Encode())
	if err != nil {
		return false
	}
	respString := getStrRespBody(resp)
	result := getRespBody(respString, "msg").(string)
	return result == "OK"
}

func GetOauth(token string) string {
	info := OauthInfo{token, "4ca99fa6b56cc2ba", 0}
	js, _ := json.Marshal(info)
	resp, _ := http.Post(urlOauth, "application/json", bytes.NewBuffer(js))
	return getRespBody(getStrRespBody(resp), "data.code").(string)
}

func GetCerd(code string) (string, string) {
	info := CerdInfo{1, code}
	js, _ := json.Marshal(info)
	resp, _ := http.Post(urlCerd, "application/json", bytes.NewBuffer(js))
	respString := getStrRespBody(resp)
	cred := getRespBody(respString, "data.cred")
	fixToken := getRespBody(respString, "data.token")
	return cred.(string), fixToken.(string)
}

// GetCharacterList 改动点：增加了对终末地角色的过滤
func GetCharacterList(cred string, key string) []CharacterInfo {
	h1 := setHeader()
	u, _ := url.Parse(urlPlayerInfo)
	jsoncode, _ := json.Marshal(h1)
	originCode := u.Path + u.Query().Encode() + h1.Timestamp + string(jsoncode)
	sign := EncodeSignCode(originCode, key)

	nh := nHeader{Sign: sign, header: h1}
	h2 := agent(cred, nh)
	headerjson, _ := json.Marshal(h2)
	headers, _ := string2Header(string(headerjson))

	req, _ := http.NewRequest("GET", urlPlayerInfo, nil)
	req.Header = headers
	resp, _ := http.DefaultClient.Do(req)
	
	body := getStrRespBody(resp)
	data := getRespBody(body, "data.list")
	
	var charList []CharacterInfo
	if data == nil {
		return charList
	}

	for _, appItem := range data.([]interface{}) {
		app := appItem.(map[string]any)
		appCode := app["appCode"].(string)
		bindingList := app["bindingList"].([]interface{})

		for _, b := range bindingList {
			binding := b.(map[string]any)
			
			// --- 新增：针对终末地未创建角色情况的过滤逻辑 ---
			if appCode == "endfield" {
				outerName, _ := binding["nickName"].(string)
				defaultRole, hasDefault := binding["defaultRole"].(map[string]any)
				
				// 如果外层昵称为空，且没有默认角色（或默认角色昵称也为空），则认为没玩这个游戏
				if outerName == "" && (!hasDefault || defaultRole == nil || defaultRole["nickname"] == "") {
					fmt.Println("未找到终末地角色数据，已自动跳过。")
					continue 
				}
			}
			// --------------------------------------------

			info := CharacterInfo{
				AppCode: appCode,
				UID:     binding["uid"].(string),
				GameId:  fmt.Sprintf("%v", binding["gameId"]),
				Server:  binding["channelName"].(string),
			}

			// 名字提取
			if n, ok := binding["nickName"].(string); ok && n != "" {
				info.Name = n
			} else if dr, ok := binding["defaultRole"].(map[string]any); ok {
				info.Name = dr["nickname"].(string)
			} else {
				info.Name = "玩家"
			}
			charList = append(charList, info)
		}
	}
	return charList
}

func DoSign(cred string, key string, char CharacterInfo) (map[string]string, error) {
	targetUrl, ok := signUrlMap[char.AppCode]
	if !ok {
		return nil, fmt.Errorf("不支持的游戏类型: %s", char.AppCode)
	}

	u, _ := url.Parse(targetUrl)
	bodyData := map[string]string{"uid": char.UID, "gameId": char.GameId}
	jsonBody, _ := json.Marshal(bodyData)
	
	h1 := setHeader()
	h1Json, _ := json.Marshal(h1)
	originCode := u.Path + string(jsonBody) + h1.Timestamp + string(h1Json)
	sign := EncodeSignCode(originCode, key)

	h2 := agent(cred, nHeader{Sign: sign, header: h1})
	h2Json, _ := json.Marshal(h2)
	headers, _ := string2Header(string(h2Json))

	req, _ := http.NewRequest("POST", targetUrl, bytes.NewBuffer(jsonBody))
	req.Header = headers
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	respString := getStrRespBody(resp)
	code := int(getRespBody(respString, "code").(float64))

	if code == 10001 {
		return nil, fmt.Errorf("今日已签到")
	} else if code != 0 {
		msg := getRespBody(respString, "message").(string)
		return nil, fmt.Errorf("失败: %s", msg)
	}

	awardList := make(map[string]string)
	if char.AppCode == "endfield" {
		awardIds := getRespBody(respString, "data.awardIds").([]interface{})
		resMap := getRespBody(respString, "data.resourceInfoMap").(map[string]interface{})
		for _, item := range awardIds {
			id := item.(map[string]interface{})["id"].(string)
			if res, exists := resMap[id].(map[string]interface{}); exists {
				name := res["name"].(string)
				count := strconv.Itoa(int(res["count"].(float64)))
				awardList[name] = count
			}
		}
	} else {
		awards := getRespBody(respString, "data.awards").([]interface{})
		for _, item := range awards {
			obj := item.(map[string]interface{})
			name := obj["resource"].(map[string]interface{})["name"].(string)
			count := strconv.Itoa(int(obj["count"].(float64)))
			awardList[name] = count
		}
	}
	return awardList, nil
}

func GetAwardlist(awardlist map[string]string) string {
	var result string
	for k, v := range awardlist {
		result += fmt.Sprintf("\n - %s \t %s", k, v)
	}
	return result + "\n"
}

func DoAll(data settings.AccountList, isshowtimes bool) {
	success := 0
	failed := 0

	for i := range data.List {
		if !RefreshToken(&data.List[i]) {
			fmt.Printf("账号 %s 认证失败\n", data.List[i].Phone)
			continue
		}

		oauth := GetOauth(data.List[i].Token)
		cred, fixToken := GetCerd(oauth)
		chars := GetCharacterList(cred, fixToken)
		
		for _, char := range chars {
			result, err := DoSign(cred, fixToken, char)
			gameName := "明日方舟"
			if char.AppCode == "endfield" {
				gameName = "终末地"
			}

			if err != nil {
				failed++
				fmt.Printf("[%s] %s %s: %v\n", gameName, char.Server, char.Name, err)
			} else {
				success++
				fmt.Printf("[%s] %s %s 签到成功！奖励：%s", gameName, char.Server, char.Name, GetAwardlist(result))
			}
		}
	}

	settings.SaveAccountData("configs/accounts.json", data)
	if isshowtimes {
		fmt.Printf("\n任务结束：成功 %d, 失败 %d\n", success, failed)
	}
}

