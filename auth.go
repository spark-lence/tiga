package tiga

import (
	"encoding/json"
	"fmt"
)

func GenerateJWT(payload interface{}, secret string) (string, error) {
	headers := make(map[string]string)
	headers["alg"] = "HS256"
	headers["typ"] = "JWT"
	// 生成header
	header, err := json.Marshal(headers)
	if err != nil {
		return "", err
	}
	// 生成payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	// 生成signature
	sig:=fmt.Sprintf("%s.%s", Bytes2BaseURL64(header), Bytes2BaseURL64(payloadBytes))
	// 生成token
	return fmt.Sprintf("%s.%s",sig,ComputeHmacSha256(sig, secret)) , nil
}
