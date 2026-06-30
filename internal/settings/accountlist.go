package settings

import (
	"arkSign/internal/crypto"
	"encoding/json"
	"fmt"
	"os"

	"github.com/thedevsaddam/gojsonq"
)

type AccountData struct {
	Phone  string `json:"phone"`
	Passwd string `json:"passwd"`
	Token  string `json:"token"`
}

type AccountList struct {
	List []AccountData `json:"accounts"`
}

// readAccountDataPlaintext 回退函数：读取旧版明文 JSON 文件
func readAccountDataPlaintext(rawBytes []byte) (AccountList, error) {
	js := gojsonq.New().FromString(string(rawBytes))
	accounts := js.Find("accounts")
	var accountlist AccountList
	accountlist.List = []AccountData{}
	if accounts == nil {
		return accountlist, nil
	}
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

func ReadAccountData(filename string) (AccountList, error) {
	EnsureFileExists(filename, "")

	rawBytes, err := os.ReadFile(filename)
	if err != nil {
		return AccountList{}, err
	}

	// 空文件：返回空列表
	if len(rawBytes) == 0 {
		return AccountList{List: []AccountData{}}, nil
	}

	// 尝试解密（新格式）
	plaintext, decErr := crypto.Decrypt(rawBytes)
	if decErr == nil {
		var accountlist AccountList
		if err := json.Unmarshal(plaintext, &accountlist); err != nil {
			return AccountList{}, fmt.Errorf("解析解密后的账户数据失败: %w", err)
		}
		if accountlist.List == nil {
			accountlist.List = []AccountData{}
		}
		return accountlist, nil
	}

	// 解密失败，回退到明文 JSON（兼容旧文件）
	return readAccountDataPlaintext(rawBytes)
}

func SaveAccountData(filename string, data AccountList) error {
	if data.List == nil {
		data.List = []AccountData{}
	}
	jdata, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化账户数据失败: %w", err)
	}

	encrypted, err := crypto.Encrypt(jdata)
	if err != nil {
		return fmt.Errorf("加密账户数据失败: %w", err)
	}

	return os.WriteFile(filename, encrypted, 0600)
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

func GetAccountData(filepath string) (AccountList, int) {
	data, err := ReadAccountData(filepath)
	if err != nil {
		fmt.Println(err)
	}
	if len(data.List) == 0 {
		fmt.Println("未检测到已添加的账号！")
		fmt.Printf("Press any key to exit...")
		b := make([]byte, 1)
		os.Stdin.Read(b)
		return data, 0
	} else {
		return data, len(data.List)
	}
}
