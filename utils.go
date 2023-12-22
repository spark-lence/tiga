package tiga

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"text/template"
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

func URLTemplateRender(urlTemplate string, params map[string]interface{}) (string,error) {
	// 创建一个新的模板
	tmpl, err := template.New("url").Parse(urlTemplate)
	if err != nil {
		return "", fmt.Errorf("parse url template error: %w", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		return "", fmt.Errorf("execute url template error: %w", err)
	}

	// 打印或使用结果
	renderedURL := buf.String()
	return renderedURL, nil
}
