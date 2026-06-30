package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// 密钥种子经过 SHA-256 派生为 32 字节 AES-256 密钥
// 不使用明文密钥防止 strings 命令直接提取
var appKey = func() []byte {
	// 混淆种子：分段存储，运行时拼接
	seed := []byte{
		0x61, 0x72, 0x6b, 0x53, 0x69, 0x67, 0x6e, 0x40, // arkSign@
		0x32, 0x30, 0x32, 0x34, 0x21, 0x48, 0x79, 0x70, // 2024!Hyp
		0x65, 0x72, 0x67, 0x72, 0x79, 0x70, 0x68, 0x2e, // ergyph.
		0x53, 0x6b, 0x6c, 0x61, 0x6e, 0x64, 0x23, 0x24, // Skland#$
	}
	hash := sha256.Sum256(seed)
	return hash[:]
}()

// encryptedWrapper 加密数据的 JSON 包装结构
type encryptedWrapper struct {
	N string `json:"n"` // base64 编码的 GCM nonce (12 bytes)
	D string `json:"d"` // base64 编码的密文
}

// Encrypt 使用 AES-256-GCM 加密明文数据，返回加密包装的 JSON 字节
func Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(appKey)
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建 GCM 失败: %w", err)
	}

	// 生成随机 nonce (12 bytes for GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("生成 nonce 失败: %w", err)
	}

	// 加密：Seal 会自动附加认证标签
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	wrapper := encryptedWrapper{
		N: base64.StdEncoding.EncodeToString(nonce),
		D: base64.StdEncoding.EncodeToString(ciphertext),
	}

	return json.Marshal(wrapper)
}

// Decrypt 解密加密包装的 JSON 数据，返回原始明文
// 如果数据格式不正确或认证失败，返回错误
func Decrypt(encryptedJSON []byte) ([]byte, error) {
	var wrapper encryptedWrapper
	if err := json.Unmarshal(encryptedJSON, &wrapper); err != nil {
		return nil, fmt.Errorf("解析加密包装失败: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(wrapper.N)
	if err != nil {
		return nil, fmt.Errorf("解码 nonce 失败: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(wrapper.D)
	if err != nil {
		return nil, fmt.Errorf("解码密文失败: %w", err)
	}

	block, err := aes.NewCipher(appKey)
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建 GCM 失败: %w", err)
	}

	// 解密：Open 会验证认证标签
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("解密失败(密钥不匹配或数据已损坏): %w", err)
	}

	return plaintext, nil
}
