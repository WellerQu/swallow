package swallow

import "reflect"

type desc interface {
	TableName() string
}

func getTableNameByModel(model interface{}) string {
	if dv, ok := model.(desc); ok {
		return dv.TableName()
	}

	typeInfo := reflect.TypeOf(model)
	if typeInfo.Kind() == reflect.Ptr {
		typeInfo = typeInfo.Elem()
	}

	return typeInfo.Name()
}
