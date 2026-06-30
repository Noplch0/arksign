package settings

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ReadPassword 读取密码，回显 '*' 掩码
// 跨平台支持 Linux / macOS / Windows
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// 获取终端文件描述符
	fd := int(os.Stdin.Fd())

	// 保存终端原始状态
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		// 非终端环境（如管道重定向），回退到简单行读取
		return readLineFallback()
	}

	// 恢复终端状态（同时恢复回显）
	defer func() {
		term.Restore(fd, oldState)
		// 额外的安全措施：确保回显已恢复
		fmt.Print("\r")
	}()

	// 自定义回显 '*' 并逐字符读取
	os.Stdout.Write([]byte(prompt)) // 重新打印提示（MakeRaw 可能清除了）
	// 由于 MakeRaw 已经设置了 raw 模式，直接用 syscall.Read 逐字节读取

	var password []byte
	buf := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			break
		}

		switch buf[0] {
		case '\r', '\n':
			// 回车键：在 raw 模式下 Enter 发送 \r
			os.Stdout.Write([]byte{'\r', '\n'})
			return string(password), nil
		case 127, 8: // Backspace (127) / Delete (8)
			if len(password) > 0 {
				password = password[:len(password)-1]
				os.Stdout.Write([]byte{'\b', ' ', '\b'})
			}
		case 3: // Ctrl+C
			os.Stdout.Write([]byte{'\r', '\n'})
			os.Exit(1)
		default:
			if buf[0] >= 32 { // 可打印字符
				password = append(password, buf[0])
				os.Stdout.Write([]byte{'*'})
			}
		}
	}
	return string(password), nil
}

// readLineFallback 非终端环境的回退方案
func readLineFallback() (string, error) {
	var buf [256]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return "", err
	}
	result := string(buf[:n])
	// 去除尾部 \r\n
	for len(result) > 0 && (result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
		result = result[:len(result)-1]
	}
	return result, nil
}
