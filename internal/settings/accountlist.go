package settings

import (
	"arkSign/internal/crypto"
	"encoding/json"
	"fmt"
	"os"

	"github.com/thedevsaddam/gojsonq"
	"golang.org/x/term"
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
	_ = EnsureFileExists(filename, "")

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
				return err == nil
			} else {
				return false
			}
		}
	}
	accounts.List = append(accounts.List, AccountData{phone, passwd, ""})
	err := SaveAccountData("configs/accounts.json", accounts)
	return err == nil
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
		_, _ = os.Stdin.Read(b)
		return data, 0
	} else {
		return data, len(data.List)
	}
}

const accountsPerPage = 3

// DeleteAccountData 交互式删除账号，支持翻页和单键选择
func DeleteAccountData() {
	accounts, _ := ReadAccountData("configs/accounts.json")
	if len(accounts.List) == 0 {
		fmt.Println("未检测到已添加的账号！")
		return
	}

	// 进入 raw 模式以捕获单键输入
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Println("无法进入终端raw模式:", err)
		return
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	page := 0
	totalPages := (len(accounts.List) + accountsPerPage - 1) / accountsPerPage

	for {
		// 校正页码（删除后可能越界）
		if totalPages == 0 {
			fmt.Print("\r\n没有账号了，退出。\r\n")
			return
		}
		if page >= totalPages {
			page = totalPages - 1
		}
		if page < 0 {
			page = 0
		}

		// 清屏并绘制界面
		drawPage(accounts.List, page, totalPages)

		// 读取单键
		buf := make([]byte, 1)
		_, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		key := buf[0]
		switch key {
		case '0':
			// 退出
			fmt.Print("\r\n已退出。\r\n")
			return
		case '-':
			if page > 0 {
				page--
			}
		case '=':
			if page < totalPages-1 {
				page++
			}
		case '1', '2', '4':
			// 计算实际索引：键 '1'→偏移0, '2'→偏移1, '4'→偏移2
			offset := keyOffset(key)
			idx := page*accountsPerPage + offset
			if idx < len(accounts.List) {
				phone := maskPhoneLocal(accounts.List[idx].Phone)
				// 读取确认输入（需要恢复终端以便读取行）
				_ = term.Restore(fd, oldState)
				result, _ := PromptForConfirmation("确定要删除账号 " + phone + " 吗")
				if result {
					accounts.List = append(accounts.List[:idx], accounts.List[idx+1:]...)
					_ = SaveAccountData("configs/accounts.json", accounts)
					totalPages = (len(accounts.List) + accountsPerPage - 1) / accountsPerPage
				}
				// 重新进入 raw 模式
				oldState, err = term.MakeRaw(fd)
				if err != nil {
					fmt.Println("无法恢复终端模式:", err)
					return
				}
			}
		}
	}
}

// keyOffset 将键字符映射到页内偏移
func keyOffset(key byte) int {
	switch key {
	case '1':
		return 0
	case '2':
		return 1
	case '4':
		return 2
	default:
		return -1
	}
}

// drawPage 绘制当前页面的账号列表
func drawPage(list []AccountData, page, totalPages int) {
	// 清屏
	fmt.Print("\r\033[2J\033[H")

	fmt.Printf("=== 账号管理 (共 %d 个账号) [第 %d/%d 页] ===\r\n\r\n", len(list), page+1, totalPages)

	start := page * accountsPerPage
	end := start + accountsPerPage
	if end > len(list) {
		end = len(list)
	}

	// 键映射：位置 0→'1', 1→'2', 2→'4'
	keys := []byte{'1', '2', '4'}
	for i := start; i < end; i++ {
		offset := i - start
		fmt.Printf("  %c. %s\r\n", keys[offset], maskPhoneLocal(list[i].Phone))
	}

	fmt.Print("\r\n========================================\r\n")
	fmt.Print("  [1/2/4] 删除对应账号  [0] 退出  [-=] 翻页\r\n")
	fmt.Print("========================================\r\n")
}

// maskPhoneLocal 隐藏手机号中间四位
func maskPhoneLocal(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
