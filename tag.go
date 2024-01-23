package tiga

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DatetimeRangeType interface {
	String() string
}
type QueryTags struct {
}

func (q QueryTags) parseQueryTag(tag string) map[string]string {
	parts := strings.Split(tag, ";")
	conditions := make(map[string]string)

	for _, part := range parts {
		conds := strings.Split(part, ":")
		conditions[conds[0]] = conds[1]

	}

	return conditions
}
func (q QueryTags) getPaginationValues(v interface{}) (int32, int64, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return 0, 0, fmt.Errorf("expected a struct")
	}

	pageValue := val.FieldByName("Page")
	pageSizeValue := val.FieldByName("PageSize")

	if !pageValue.IsValid() || !pageSizeValue.IsValid() {
		return 0, 0, fmt.Errorf("fields not found")
	}

	if pageValue.Kind() != reflect.Int32 || pageSizeValue.Kind() != reflect.Int64 {
		return 0, 0, fmt.Errorf("fields have incorrect types")
	}

	return int32(pageValue.Int()), pageSizeValue.Int(), nil
}
func (q QueryTags) computeDatetimeRange(period DatetimeRangeType) (time.Time, time.Time) {
	var start time.Time = time.Time{}
	var end time.Time = time.Time{}
	switch period.String() {
	case "MIN":
		start = time.Now().Add(-1 * time.Minute)
	case "HOUR":
		start = time.Now().Add(-1 * time.Hour)
	case "DAY":
		start = time.Now().Add(-1 * time.Hour * 24)
	case "MONTH":
		start = time.Now().Add(-1 * time.Hour * 30 * 24)
	}
	if !start.IsZero() {
		end = time.Now().UTC()
		start = start.UTC()
	}
	return start, end
}
func (q QueryTags)isEmptyValue(v reflect.Value) bool {
    switch v.Kind() {
    case reflect.Ptr, reflect.Slice, reflect.Map:
        return v.IsNil() // 检查指针、切片、映射是否为空
    case reflect.String:
        return v.Len() == 0 // 检查字符串是否为空
    case reflect.Invalid:
        return true // reflect.Invalid 表示零值
	case reflect.Int32:
		return v.Int() == 0
    }
    return false
}

func (q QueryTags) BuildConditions(base *gorm.DB, conditions interface{}) *gorm.DB {
	if conditions == nil {
		return nil
	}
	val := reflect.ValueOf(conditions)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return base
	}
	ok := false
	page, pageSize, _ := q.getPaginationValues(conditions)
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		// 判断字段是否可导出
		if typeField.PkgPath == "" && valueField.CanInterface() {
			queryTag := typeField.Tag.Get("query")
			if queryTag != "" {
				conditions := q.parseQueryTag(queryTag)
				// skip为特定的值时，跳过该字段
				if skip, ok := conditions["skip"]; ok && skip == val.String() {
					continue
				}
				// 如果字段类型为时间范围类型，则根据时间范围类型计算时间范围
				if queryField, has := conditions["period"]; has {
					start, end := q.computeDatetimeRange(valueField.Interface().(DatetimeRangeType))
					base = base.Where(fmt.Sprintf("%s BETWEEN ? AND ?", queryField), start, end)
					ok = true
					continue
				}

				condition := conditions["condition"]
				if !q.isEmptyValue(valueField) {
					base = base.Where(condition, valueField.Interface())
					ok = true
				}
			}
		}
	}
	if ((int64(page) - 1) *pageSize) > 0 {
		base = base.Offset((int(page) - 1) * int(pageSize)).Limit(int(pageSize))
		ok = true
	}
	if !ok {
		return nil
	}
	return base
}
func TagsTransformer(object interface{}, srcTagName string, targetTagName string) interface{} {
	// Get the reflect value of the struct
	v := reflect.ValueOf(object).Elem()

	// Get the type of the struct
	t := v.Type()

	// Iterate over the struct fields
	for i := 0; i < v.NumField(); i++ {
		// Get the field type
		fieldType := t.Field(i)

		// Get the current tags of the field
		tags := fieldType.Tag

		// Check if the field doesn't have a bson tag
		if tags.Get(targetTagName) == "" {
			// Add bson tag to the field
			newTags := reflect.StructTag(fmt.Sprintf(`%s:"%s" %s`, targetTagName, fieldType.Name, tags.Get(srcTagName)))
			fieldType.Tag = newTags
		}
	}
	return object
}
