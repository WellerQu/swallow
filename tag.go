package swallow

import (
	"reflect"
	"regexp"
)

// TagPrimaryKey 主键标识标签
const TagPrimaryKey = "PRIMARY_KEY"

// TagColumnName 列名标签
const TagColumnName = "column"

// TagIgnoreMember 将会忽略的成员
const TagIgnoreMember = "-"

func parseTag(f reflect.StructField) (tags map[string]interface{}) {
	tags = make(map[string]interface{})

	if tagValue, ok := f.Tag.Lookup("swallow"); ok {
		regColumn := regexp.MustCompile(`column:(.*?);`)
		if regColumn.MatchString(tagValue) {
			tags[TagColumnName] = regColumn.FindStringSubmatch(tagValue)[1]
		}

		regIsPrimary := regexp.MustCompile(`PRIMARY_KEY;`)
		if regIsPrimary.MatchString(tagValue) {
			tags[TagPrimaryKey] = true
		}

		regIgnore := regexp.MustCompile(`-;`)
		if regIgnore.MatchString(tagValue) {
			tags[TagIgnoreMember] = true
		}
	}

	return
}
