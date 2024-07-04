package settings

import (
	"encoding/json"
	"github.com/thedevsaddam/gojsonq"
	"os"
)

type AccountData struct {
	Phone  string `json:"phone"`
	Passwd string `json:"passwd"`
	Token  string `json:"token"`
}

type AccountList struct {
	List []AccountData `json:"accounts"`
}

func ReadAccountData(filename string) (AccountList, error) {
	EnsureFileExists(filename, `{"accounts": []}`)
	js := gojsonq.New().File(filename)
	accounts := js.Find("accounts")
	var accountlist AccountList
	accountlist.List = []AccountData{}
	for _, r := range accounts.([]interface{}) {
		inf := r.(map[string]interface{})
		var temp AccountData
		temp.Phone = inf["phone"].(string)
		temp.Passwd = inf["passwd"].(string)
		temp.Token = inf["token"].(string)
		accountlist.List = append(accountlist.List, temp)
	}
	return accountlist, nil
}

func SaveAccountData(filename string, data AccountList) error {
	jdata, _ := json.MarshalIndent(data, "", "  ")
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(jdata)
	return nil
}

func AddAcountData(phone string, passwd string) bool {
	accounts, _ := ReadAccountData("configs/accounts.json")
	for i := range accounts.List {
		if accounts.List[i].Phone == phone {
			result, _ := PromptForConfirmation("该账号已存在！要更新密码吗")
			if result {
				accounts.List[i].Passwd = passwd
				accounts.List[i].Token = ""
				err := SaveAccountData("configs/accounts.json", accounts)
				if err != nil {
					return false
				}
				return true
			} else {
				return false
			}
		}
	}
	accounts.List = append(accounts.List, AccountData{phone, passwd, ""})
	err := SaveAccountData("configs/accounts.json", accounts)
	if err != nil {
		return false
	}
	return true
}
