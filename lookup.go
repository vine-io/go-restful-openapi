package restspec

import (
	"reflect"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/ggicci/httpin"
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
	case kind == restful.MultiPartFormParameterKind:
		return "multipartFormData"
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
			if _, ok := in["body"]; ok {
				ft := field.Type
				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				fv := reflect.New(ft).Elem().Interface()
				b.Reads(fv)
			}
			if text, ok := in["path"]; ok {
				param := restful.PathParameter(text, "")
				setParamFrom(param, field, in)
				b.Param(param)
			}
			if text, ok := in["query"]; ok {
				parts := strings.Split(text, ",")
				for _, part := range parts {
					param := restful.QueryParameter(part, "")
					setParamFrom(param, field, in)
					b.Param(param)
				}
			}
			if text, ok := in["form"]; ok {
				parts := strings.Split(text, ",")
				for _, part := range parts {
					var param *restful.Parameter
					if isFileField(field.Type) {
						param = restful.MultiPartFormParameter(part, "").DataFormat("binary")
						if desc, ok := in["description"]; ok {
							param.Description(desc)
						}
						if _, ok := in["required"]; ok {
							param.Required(true)
						}
					} else {
						param = restful.FormParameter(part, "")
						setParamFrom(param, field, in)
					}
					b.Param(param)
				}
			}
			if text, ok := in["header"]; ok {
				parts := strings.Split(text, ",")
				for _, part := range parts {
					param := restful.HeaderParameter(part, "")
					setParamFrom(param, field, in)
					b.Param(param)
				}
			}
		}
	}
}

func setParamFrom(param *restful.Parameter, field reflect.StructField, in map[string]string) {
	var dataType, format string
	st := field.Type

	t := st.Kind()
	if t == reflect.Slice || t == reflect.Array {
		dataType = arrayType
		if st.Elem().Kind() == reflect.Uint8 {
			dataType = "byte"
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
		param.DataFormat(jsonSchemaFormat(format))
	}

	if desc, ok := in["description"]; ok {
		param.Description(desc)
	}
	if defaultValue, ok := in["default"]; ok {
		param.DefaultValue(defaultValue)
	}
	if _, ok := in["required"]; ok {
		param.Required(true)
	}
}

func inMap(text string) map[string]string {
	maps := make(map[string]string)
	parts := strings.Split(strings.TrimSpace(text), ";")
	for _, part := range parts {
		cuts := strings.Split(strings.TrimSpace(part), "=")
		var value string
		if len(cuts) > 1 {
			value = cuts[1]
		}
		maps[cuts[0]] = value
	}

	return maps
}

func isFileField(rt reflect.Type) bool {
	if rt.Kind() == reflect.Ptr {
		return isFileField(rt.Elem())
	}
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		return isFileField(rt.Elem())
	}
	return rt.Kind() == reflect.TypeOf(httpin.File{}).Kind()
}
