package models

import (
	"reflect"
)

var funcs = map[string]interface{}{
	"isMale": func(value reflect.Value) reflect.Value {
		return reflect.ValueOf(value.Uint() == 1)
	},
}

func FromProtoConvertor(proto interface{}, to interface{}) interface{} {
	protoRef := reflect.ValueOf(proto)
	if protoRef.Kind() == reflect.Ptr && protoRef.Elem().Kind() == reflect.Struct {
		protoRef = protoRef.Elem()
	}
	toRef := reflect.ValueOf(to)
	if toRef.Kind() == reflect.Ptr && toRef.Elem().Kind() == reflect.Struct {
		toRef = toRef.Elem()
	}

	for i := 0; i < toRef.NumField(); i++ {
		toTypeField := toRef.Type().Field(i)

		tagValue, ok := toTypeField.Tag.Lookup("proto")
		if !ok {
			tagValue = toTypeField.Name
		}

		protoField := protoRef.FieldByName(tagValue)
		if protoField.IsValid() {
			toRefField := toRef.Field(i)
			if protoFunc, ok := funcs[toTypeField.Tag.Get("proto_func")]; ok {
				toRefField.Set(protoFunc.(func(value reflect.Value) reflect.Value)(protoField))
			} else {
				toRefField.Set(protoField)
			}
		}
	}

	return to
}

func ToProtoConvertor(from interface{}, proto interface{}) interface{} {
	fromRef := reflect.ValueOf(from)
	if fromRef.Kind() == reflect.Ptr && fromRef.Elem().Kind() == reflect.Struct {
		fromRef = fromRef.Elem()
	}
	protoRef := reflect.ValueOf(proto)
	if protoRef.Kind() == reflect.Ptr && protoRef.Elem().Kind() == reflect.Struct {
		protoRef = protoRef.Elem()
	}

	for i := 0; i < fromRef.NumField(); i++ {
		fromField := fromRef.Field(i)
		fromTypeField := fromRef.Type().Field(i)

		tagValue, ok := fromTypeField.Tag.Lookup("proto")
		if !ok {
			tagValue = fromTypeField.Name
		}

		protoRefField := protoRef.FieldByName(tagValue)
		if protoRefField.IsValid() {
			if protoFunc, ok := funcs[fromTypeField.Tag.Get("proto_func")]; ok {
				protoRefField.Set(protoFunc.(func(value reflect.Value) reflect.Value)(fromField))
			} else {
				protoRefField.Set(fromField)
			}
		}
	}

	return proto
}
