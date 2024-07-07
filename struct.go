package tiga

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func StructToJsonStr(src interface{}) (string, error) {
	bytes, err := json.Marshal(src)
	if err != nil {
		return "", err
	}
	return string(bytes), nil

}
func DeepCopy(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	srcVal := reflect.ValueOf(src)
	return deepCopyReflect(srcVal).Interface()
}

func deepCopyReflect(srcVal reflect.Value) reflect.Value {
	switch srcVal.Kind() {
	case reflect.Ptr:
		if srcVal.IsNil() {
			return reflect.Zero(srcVal.Type())
		}
		dstVal := reflect.New(srcVal.Elem().Type())
		dstVal.Elem().Set(deepCopyReflect(srcVal.Elem()))
		return dstVal

	case reflect.Interface:
		if srcVal.IsNil() {
			return reflect.Zero(srcVal.Type())
		}
		dstVal := deepCopyReflect(srcVal.Elem())
		return dstVal.Convert(srcVal.Type())

	case reflect.Struct:
		dstVal := reflect.New(srcVal.Type()).Elem()
		for i := 0; i < srcVal.NumField(); i++ {
			fieldVal := srcVal.Field(i)
			if fieldVal.CanSet() {
				dstVal.Field(i).Set(deepCopyReflect(fieldVal))
			}
		}
		return dstVal

	case reflect.Slice:
		if srcVal.IsNil() {
			return reflect.Zero(srcVal.Type())
		}
		dstVal := reflect.MakeSlice(srcVal.Type(), srcVal.Len(), srcVal.Cap())
		for i := 0; i < srcVal.Len(); i++ {
			dstVal.Index(i).Set(deepCopyReflect(srcVal.Index(i)))
		}
		return dstVal

	case reflect.Map:
		if srcVal.IsNil() {
			return reflect.Zero(srcVal.Type())
		}
		dstVal := reflect.MakeMapWithSize(srcVal.Type(), srcVal.Len())
		for _, key := range srcVal.MapKeys() {
			dstVal.SetMapIndex(deepCopyReflect(key), deepCopyReflect(srcVal.MapIndex(key)))
		}
		return dstVal

	default:
		return srcVal
	}
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
func ValueToString(src interface{}) (string, error) {
	switch v := src.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, int32, int64, float64, float32, bool:
		return fmt.Sprintf("%v", v), nil
	default:
		value, err := json.Marshal(src)
		if err != nil {
			return "", err
		}
		return string(value), nil
	}
}
func StringToValue(src string, dst interface{}) error {
	return json.Unmarshal([]byte(src), dst)
}

func IsTagExists(v interface{}, tagName string, tagValue string) bool {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	exists := false
	for i := 0; i < val.NumField(); i++ {
		if typ.Field(i).Tag.Get(tagName) == tagValue {
			exists = true
		}

	}
	return exists
}
func SearchTag(v interface{}, tagName string, tagValue string) []interface{} {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	vals := make([]interface{}, 0)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.Struct {
			if !field.CanInterface() {
				continue
			}
			// Recursively search in embedded struct
			vals = append(vals, SearchTag(field.Interface(), tagName, tagValue)...)
		} else {
			if tagVal := typ.Field(i).Tag.Get(tagName); tagVal == tagValue {
				vals = append(vals, field.Interface())
			}
		}
	}
	return vals
}

func IsArray(v interface{}) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice
}
func GetElementCount(models interface{}) (int, error) {
	val := reflect.ValueOf(models)
	if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
		return val.Len(), nil
	} else {
		return 0, fmt.Errorf("provided interface is not an array or slice")
	}
}
func GetFirstElement(models interface{}) (interface{}, error) {
	val := reflect.ValueOf(models)
	if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
		return val.Index(0).Interface(), nil
	} else {
		return nil, fmt.Errorf("provided interface is not an array or slice")
	}
}

func GetArrayOrSlice(arr interface{}) []interface{} {
	// 获取反射值对象
	val := reflect.ValueOf(arr)
	dst := make([]interface{}, 0)
	// 确认val是数组或切片
	if val.Kind() != reflect.Array && val.Kind() != reflect.Slice {
		return nil
	}

	// 遍历数组或切片的元素
	for i := 0; i < val.Len(); i++ {
		// 获取元素的反射值
		element := val.Index(i)

		// 使用Interface()来获取元素的实际值
		dst = append(dst, element.Interface())
	}
	return dst
}
