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

const urlGetTokenByPwd = "https://as.hypergryph.com/user/auth/v1/token_by_phone_password"
const urlOauth = "https://as.hypergryph.com/user/oauth2/v2/grant"
const urlCerd = "https://zonai.skland.com/api/v1/user/auth/generate_cred_by_code"
const urlPlayerInfo = "https://zonai.skland.com/api/v1/game/player/binding"
const urlSign = "https://zonai.skland.com/api/v1/game/attendance"
const urlVerify = "https://as.hypergryph.com/user/info/v1/basic"

type loginInfo struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type OauthInfo struct {
	Token    string `json:"token"`
	Appcode  string `json:"appCode"`
	TypeCode int    `json:"type"`
}

func getOauthInfo(token string) OauthInfo {
	return OauthInfo{
		token,
		"4ca99fa6b56cc2ba",
		0,
	}
}

type CerdInfo struct {
	Kind int    `json:"kind"`
	Code string `json:"code"`
}

func getCerdInfo(code string) CerdInfo {
	return CerdInfo{
		1,
		code,
	}
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
	var result = reJson.Find(index)
	return result
}

func GetToken(phone string, passwd string) (string, error) {
	var accountInfo = loginInfo{Phone: phone, Password: passwd}
	accountJson, _ := json.Marshal(accountInfo)
	resp, err := http.Post(urlGetTokenByPwd, "application/json", bytes.NewBuffer(accountJson))
	if err != nil {
		return "", errors.New("login Failed(incorrect phone or password)")
	}
	respString := getStrRespBody(resp)
	if resp.StatusCode != 200 {
		return "", errors.New("login Failed(incorrect phone? or password)")
	}
	s := getRespBody(respString, "data.token")
	result := s.(string)
	return result, nil
}

func VerifyToken(token string) bool {
	params := url.Values{}
	parseUrl, err := url.Parse(urlVerify)
	if err != nil {
		log.Println(err)
		return false
	}
	params.Set("token", token)
	parseUrl.RawQuery = params.Encode()
	urlVerifyWithParams := parseUrl.String()
	resp, err := http.Get(urlVerifyWithParams)
	if err != nil {
		log.Println(err)
		return false
	}
	respString := getStrRespBody(resp)
	s := getRespBody(respString, "msg")
	result := s.(string)
	if result != "OK" {
		fmt.Printf("token:%#v已失效\n", token)
		return false
	}
	return true
}

func GetOauth(token string) string {
	info := getOauthInfo(token)
	js, _ := json.Marshal(info)
	resp, err := http.Post(urlOauth, "application/json", bytes.NewBuffer(js))
	if err != nil {
		log.Println(err)
	}
	respString := getStrRespBody(resp)
	s := getRespBody(respString, "data.code")
	result := s.(string)
	return result
}

func GetCerd(code string) (string, string) {
	info := getCerdInfo(code)
	js, _ := json.Marshal(info)
	resp, err := http.Post(urlCerd, "application/json", bytes.NewBuffer(js))
	if err != nil {
		log.Println(err)
	}
	respString := getStrRespBody(resp)
	cred := getRespBody(respString, "data.cred")
	fixToken := getRespBody(respString, "data.token")

	return cred.(string), fixToken.(string)
}

func EncodeSignCode(code string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(code))
	sha := hex.EncodeToString(h.Sum(nil))

	hash := md5.New()
	hash.Write([]byte(sha))
	result := hex.EncodeToString(hash.Sum(nil))
	return result
}

func string2Header(text string) (http.Header, error) {
	var headers map[string]interface{}
	err := json.NewDecoder(strings.NewReader(text)).Decode(&headers)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	header := make(http.Header)
	for k, v := range headers {
		strValue, ok := v.(string)
		if ok {
			header.Set(k, strValue)
		} else {
			return nil, fmt.Errorf("error parsing JSON: unknown type: %T", v)
		}
	}
	return header, nil
}

func GetCharacterList(cerd string, key string) map[int]map[string]string {
	header1 := setHeader()
	parses, _ := url.Parse(urlPlayerInfo)
	query := parses.Query().Encode()
	path := parses.Path
	jsoncode, _ := json.Marshal(header1)
	originCode := path + query + header1.Timestamp + string(jsoncode)
	var newheader map[string]interface{}
	if err := json.Unmarshal(jsoncode, &newheader); err != nil {
		fmt.Println(err)
	}
	newheader["sign"] = EncodeSignCode(originCode, key)

	nh := nHeader{
		Sign:   newheader["sign"].(string),
		header: header1,
	}

	header2 := agent(cerd, nh)
	headerjson, _ := json.Marshal(header2)
	headertext := string(headerjson)

	req, err := http.NewRequest("GET", urlPlayerInfo, nil)
	if err != nil {
		fmt.Println(err)
	}
	headers, _ := string2Header(headertext)
	req.Header = headers
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	body := getStrRespBody(resp)
	data := getRespBody(body, "data.list")
	charlist := make(map[int]map[string]string)
	var a = 0
	for _, k := range data.([]interface{}) {
		for _, c := range k.(map[string]any)["bindingList"].([]interface{}) {
			submap := make(map[string]string)
			cinfo := c.(map[string]any)
			submap["uid"] = cinfo["uid"].(string)
			submap["gameId"] = cinfo["channelMasterId"].(string)
			submap["server"] = cinfo["channelName"].(string)
			submap["name"] = cinfo["nickName"].(string)
			charlist[a] = submap
		}
	}
	return charlist
}

func DoSign(cred string, key string, charinfo map[string]string) (map[string]string, error) {
	parses, _ := url.Parse(urlSign)
	path := parses.Path
	body := make(map[string]string)
	body["uid"] = charinfo["uid"]
	body["gameId"] = charinfo["gameId"]
	jsonbody, _ := json.Marshal(body)
	intent := string(jsonbody)
	header1 := setHeader()
	jsoncode, _ := json.Marshal(header1)
	originCode := path + intent + header1.Timestamp + string(jsoncode)
	sign := EncodeSignCode(originCode, key)
	h2 := agent(cred, nHeader{
		Sign:   sign,
		header: header1,
	})
	headerjson, _ := json.Marshal(h2)
	headertext := string(headerjson)
	headers, _ := string2Header(headertext)
	req, err := http.NewRequest("POST", urlSign, bytes.NewBuffer(jsonbody))
	if err != nil {
		return nil, err
	}
	req.Header = headers
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	respString := getStrRespBody(resp)
	message := getRespBody(respString, "code")
	if int(message.(float64)) == 10001 {
		return nil, fmt.Errorf("%s", "今日已签到！")
	} else {
		awarddata := getRespBody(respString, "data.awards")
		awardlist := make(map[string]string)
		for _, item := range awarddata.([]interface{}) {
			name := item.(map[string]interface{})["resource"].(map[string]interface{})["name"].(string)
			count := strconv.Itoa(int(item.(map[string]interface{})["count"].(float64)))
			awardlist[name] = count
		}
		return awardlist, nil
	}
}

func GetAwardlist(awardlist map[string]string) string {
	var result string
	for k, v := range awardlist {
		result += "\n - " + k + "\t" + v + "\n"
	}
	return result
}

func DoAll(data settings.AccountList) {

	success := 0
	failed := 0
	for i := range data.List {
		if !RefreshToken(&data.List[i]) {
			fmt.Println("刷新token失败!")
			return
		}
		oauth := GetOauth(data.List[i].Token)
		cred, fixToken := GetCerd(oauth)
		charlist := GetCharacterList(cred, fixToken)
		for _, char := range charlist {
			result, err := DoSign(cred, fixToken, char)
			if err != nil {
				failed++
				fmt.Printf("%s %s ", char["server"], char["name"])
				fmt.Println(err)
			} else {
				success++
				fmt.Printf("%s %s 签到成功\n", char["server"], char["name"])
				fmt.Printf("本次签到获取奖励：%s", GetAwardlist(result))
			}
		}
	}
	settings.SaveAccountData("configs/accounts.json", data)
	fmt.Printf("本次签到成功 %d 次 失败 %d 次\n", success, failed)

	return
}
