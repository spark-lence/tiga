package tiga

import (
	"encoding/json"
)

func StructToJsonStr(src interface{}) (string, error) {
	bytes, err := json.Marshal(src)
	if err != nil {
		return "", err
	}
	return string(bytes), nil

}
