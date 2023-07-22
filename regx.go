package tiga

import (
	"errors"
	"regexp"
	"strconv"
)

var numberRegx *regexp.Regexp = regexp.MustCompile(`-?\d+(\.\d+)?`)

func GetNumberByRegx(content string) (int64, error) {
	value := numberRegx.FindStringSubmatch(content)
	if len(value) > 0 {
		v := value[0]
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return num, err

		}
		return num, nil
	}
	return 0, errors.New("not match value")
}

func GetFloatRegx(content string) (float64, error) {
	value := numberRegx.FindStringSubmatch(content)
	if len(value) > 0 {
		v := value[0]
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return num, err

		}
		return num, nil
	}
	return 0.0, errors.New("not match value")
}
