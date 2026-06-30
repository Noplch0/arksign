package main

import (
	"arkSign/internal/settings"
	"arkSign/internal/skisland"
	"flag"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// 构建时注入的版本信息
var (
	Version   = "dev"
	BuildTime = "unknown"
)

var a = flag.Bool("a", false, "添加用户模式")
var t = flag.String("t", "4:30", `每日运行签到任务时间,例如"8:30"`)
var o = flag.Bool("o", false, "只运行一次")
var v = flag.Bool("v", false, "显示版本信息")

func main() {
	flag.Parse()

	if *v {
		fmt.Printf("arkSign %s (built %s)\n", Version, BuildTime)
		return
	}

	if *a {
		fmt.Print("请输入添加账号：")
		var phone string
		fmt.Scanln(&phone)
		passwd, err := settings.ReadPassword("请输入密码：")
		if err != nil {
			fmt.Println("读取密码失败:", err)
			return
		}
		settings.AddAcountData(phone, passwd)
		data, num := settings.GetAccountData("configs/accounts.json")
		if num != 0 {
			skisland.DoAll(data,false)
		}
		return
	}
	if *o {
		data, num := settings.GetAccountData("configs/accounts.json")
		if num != 0 {
			skisland.DoAll(data,false)
		}

	} else {
		data, num := settings.GetAccountData("configs/accounts.json")
		if num != 0 {
			
			hour, minute, err := settings.ParseTime(*t)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("将在每日%02d:%02d签到\n", hour, minute)
			cronstring := fmt.Sprintf("%d %d * * *", minute, hour)
			c := cron.New()
			c.AddFunc(cronstring, func() {
				skisland.DoAll(data,false)
			})
			c.Start()
			for {
				time.Sleep(time.Second)
			}
		}
	}
}
