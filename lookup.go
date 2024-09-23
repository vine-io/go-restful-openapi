package restspec

import (
	"reflect"
	"strings"

	"github.com/emicklei/go-restful/v3"
)

func asParamType(kind int) string {
	switch {
	case kind == restful.PathParameterKind:
		return "path"
	case kind == restful.QueryParameterKind:
		return "query"
	case kind == restful.BodyParameterKind:
		return "body"
	case kind == restful.HeaderParameterKind:
		return "header"
	case kind == restful.FormParameterKind:
		return "formData"
	}
	return ""
}

func ReadSample(sample any) func(b *restful.RouteBuilder) {
	return func(b *restful.RouteBuilder) {
		rt := reflect.TypeOf(sample)
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			tag := field.Tag.Get("in")
			if tag == "" {
				continue
			}

			in := inMap(strings.Trim(tag, `"`))
			desc := in["description"]
			if _, ok := in["body"]; ok {
				ft := field.Type
				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				fv := reflect.New(ft).Elem().Interface()
				b.Reads(fv)
			}
			if text, ok := in["path"]; ok {
				param := restful.PathParameter(text, desc)
				paramKeyFrom(param, field)
				b.Param(param)
			}
			if text, ok := in["query"]; ok {
				parts := strings.Split(text, ",")
				for _, part := range parts {
					param := restful.QueryParameter(part, desc)
					paramKeyFrom(param, field)
					b.Param(param)
				}
			}
			if text, ok := in["header"]; ok {
				parts := strings.Split(text, ",")
				for _, part := range parts {
					param := restful.HeaderParameter(part, desc)
					paramKeyFrom(param, field)
					b.Param(param)
				}
			}
		}
	}
}

func paramKeyFrom(param *restful.Parameter, field reflect.StructField) {
	var dataType, format string
	st := field.Type

	t := st.Kind()
	if t == reflect.Slice || t == reflect.Array {
		dataType = "array"
		if st.Elem().Kind() == reflect.Uint8 {

		} else {
			format = st.Elem().Kind().String()
		}
	} else {
		dataType = st.String()
		if len(st.Name()) == 0 { // unnamed type
			// If it is an array, remove the leading []
			dataType = strings.TrimPrefix(dataType, "[]")
			// Swagger UI has special meaning for [
			dataType = strings.Replace(dataType, "[]", "||", -1)
		}
	}

	param.DataType(jsonSchemaType(dataType))
	if len(format) != 0 {
		param.DataFormat(format)
	}
}

func inMap(text string) map[string]string {
	maps := make(map[string]string)
	parts := strings.Split(strings.TrimSpace(text), ";")
	for _, part := range parts {
		cuts := strings.Split(strings.TrimSpace(part), "=")
		if len(cuts) != 2 {
			continue
		}
		maps[cuts[0]] = cuts[1]
	}

	return maps
}
