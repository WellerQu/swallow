package swallow

import (
	"reflect"
)

// Field 模型字段描述
type Field struct {
	name         string
	tags         map[string]interface{}
	reflectField reflect.StructField
	reflectValue reflect.Value
}

func getFieldsByModel(model interface{}) (fields []Field) {
	var typeInfo reflect.Type
	var valueInfo reflect.Value

	if m, ok := model.(*reflect.Value); ok {
		// 反射法则二: 根据反射对象还原出接口对象
		typeInfo = reflect.TypeOf(m.Interface())
		valueInfo = *m
	} else {
		typeInfo = reflect.TypeOf(model)
		valueInfo = reflect.ValueOf(model)
	}

	if typeInfo.Kind() == reflect.Ptr {
		typeInfo = typeInfo.Elem()
		valueInfo = valueInfo.Elem()
	}

	numField := typeInfo.NumField()

	for i := 0; i < numField; i++ {
		// 反射法则一: 根据接口对象反射出反射对象
		// 取结构成员的反射对象
		field := typeInfo.Field(i)

		// 取结构成员值得反射对象
		var value reflect.Value
		if field.Anonymous {
			value = valueInfo.FieldByName(field.Type.Name())
		} else {
			value = valueInfo.FieldByName(field.Name)
		}

		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() == reflect.Struct {
			// 反射法则三: 想要通过发射修改原始值, 需要提供指向原始值的指针对象
			fields = append(fields, getFieldsByModel(&value)...)
			continue
		}

		tags := parseTag(field)

		fields = append(fields, Field{name: field.Name, tags: tags, reflectField: field, reflectValue: value})
	}

	return
}

func getPrimaryKeyByModel(model interface{}) (field Field, exist bool) {
	fields := getFieldsByModel(model)
	for _, field := range fields {
		if field.isPrimaryKey() {
			return field, true
		}
	}

	exist = false

	return
}

func (f *Field) isPrimaryKey() bool {
	_, ok := f.tags[TagPrimaryKey]

	return ok
}

func (f *Field) isBlank() bool {
	switch f.reflectValue.Kind() {
	case reflect.String:
		return f.reflectValue.String() == ""
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return f.reflectValue.Int() == 0
	case reflect.Float32, reflect.Float64:
		return f.reflectValue.Float() == 0
	case reflect.Bool:
		return false
	}

	return true
}

func (f *Field) columnName() string {
	if v, ok := f.tags[TagColumnName]; ok {
		return v.(string)
	}

	return f.name
}

func (f *Field) isIgnore() bool {
	_, ok := f.tags[TagIgnoreMember]
	return ok
}
