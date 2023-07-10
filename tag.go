package tiga

import (
	"fmt"
	"reflect"
)

func TagsTransformer(object interface{}, srcTagName string, targetTagName string) {
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
}
