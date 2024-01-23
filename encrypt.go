package tiga

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
)

func encryptAESCTR(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// DecryptAES 解密函数
func DecryptAES(key []byte, cipherHex string, ivKey string) (string, error) {
	cipherText, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建和使用相同的IV
	iv := []byte(ivKey)
	if len(ivKey) == 0 {
		return "", errors.New("IV cannot be empty")
	}
	if len(iv) != aes.BlockSize {
		return "", errors.New("IV length must be equal to block size")
	}

	// 解密器
	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(cipherText))
	stream.XORKeyStream(plaintext, cipherText)

	return string(plaintext), nil
}
func EncryptAESWithAutoIV(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 随机初始化向量
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 加密器
	stream := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plaintext))
	stream.XORKeyStream(cipherText, []byte(plaintext))
	// 返回带有IV的加密文本
	return hex.EncodeToString(iv) + hex.EncodeToString(cipherText), nil
}
func EncryptAES(key []byte, plaintext string, ivKey string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := []byte(ivKey)
	if ivKey == "" {
		iv = make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return "", err
		}
	}
	// 加密器
	stream := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plaintext))
	stream.XORKeyStream(cipherText, []byte(plaintext))

	// 返回带有IV的加密文本
	return hex.EncodeToString(cipherText), nil
}

func EncryptStructAES(key []byte, data interface{}, ivKey string) error {
	// 获取结构体的反射值对象
	val := reflect.ValueOf(data).Elem()

	// 遍历结构体的所有字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)
		val := field.String()
		// 检查字段是否有特定的标签并且是字符串类型
		if processTag, ok := typeField.Tag.Lookup("aes"); ok && processTag == "true" && field.Kind() == reflect.String {
			// 获取字段的值
			// val := field.String()
			// fmt.Printf("encrypt %s,value %s", typeField.Name, val)
			val, err := EncryptAES(key, val, ivKey)
			if err != nil {
				return fmt.Errorf("encrypt %s error,%w", typeField.Name, err)
			}
			field.SetString(val)
		}
	}
	return nil
}
func DecryptStructAES(key []byte, data interface{}, ivKey string) error {
	// 获取结构体的反射值对象
	val := reflect.ValueOf(data).Elem()

	// 遍历结构体的所有字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)
		val := field.String()
		// 检查字段是否有特定的标签并且是字符串类型
		if processTag, ok := typeField.Tag.Lookup("aes"); ok && processTag == "true" && field.Kind() == reflect.String {
			// 获取字段的值
			if val == "" {
				continue
			}
			val, err := DecryptAES(key, val, ivKey)
			if err != nil {
				return fmt.Errorf("encrypt %s error,%w", typeField.Name, err)
			}
			field.SetString(val)
		}
	}
	return nil
}
func GenerateRandomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result += string(letters[num.Int64()])
	}
	return result, nil
}
func ComputeHmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
