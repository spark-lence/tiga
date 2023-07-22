package tiga

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GetNowDatetime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
func GetNowDate() string {
	return time.Now().Format("2006-01-02")
}
func GetMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
func Map2String(data map[string]interface{}) (string, error) {
	str, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("marshal with error: %+v\n", err)
		return "", err
	}

	return string(str), nil
}
func ThousandsSeparatorToInt(src string) int {
	numStr := strings.Replace(src, ",", "", -1)

	// 将结果转换为 int
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}
	return num

}
