package gurps

import (
	"reflect"
	"strings"

	"github.com/richardwilkes/toolbox/txt"
)

type scriptNameMapper struct{}

func (n scriptNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string { //nolint:gocritic // API requires it
	return uncapitalizeScriptName(f.Name)
}

func (n scriptNameMapper) MethodName(_ reflect.Type, m reflect.Method) string { //nolint:gocritic // API requires it
	return uncapitalizeScriptName(m.Name)
}

func uncapitalizeScriptName(s string) string {
	if strings.EqualFold(s, "id") {
		return "id"
	}
	if strings.EqualFold(s, "parentid") {
		return "parentID"
	}
	return txt.FirstToLower(s)
}
