package skisland

import (
	"arkSign/internal/settings"
	"fmt"
)

func VerifyPassword(phone, password string) (string, error) {
	token, err := GetToken(phone, password)
	if err != nil {
		return token, err
	} else {
		return token, nil
	}
}

func VerifyAccount(data settings.AccountData) bool {
	fmt.Printf("Verifying account:%#v\n", data.Phone)
	return VerifyToken(data.Token)
}

func RefreshToken(data *settings.AccountData) bool {
	if !VerifyAccount(*data) {
		var err error
		(*data).Token, err = GetToken((*data).Phone, (*data).Passwd)
		if err != nil {
			fmt.Println("账号或者密码错误！")
			fmt.Println(err)
			return false
		}
		fmt.Printf("已替换为:%s\n", data.Token)
		return true
	} else {
		fmt.Printf("Token(%#v) is valid\n", (*data).Token)
		return true
	}
}
