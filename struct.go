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

func StructToMap(src interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	val := make(map[string]interface{})
	err = json.Unmarshal(bytes, &val)
	if err != nil {
		return nil, err
	}
	return val, nil
}
