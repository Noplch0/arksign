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
	fmt.Printf("正在验证账号 %s ...\n", maskPhone(data.Phone))
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
		fmt.Println("Token 已刷新")
		return true
	} else {
		fmt.Println("Token 有效")
		return true
	}
}

// maskPhone 隐藏手机号中间四位
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
