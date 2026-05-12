package filter

import (
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"maps"
	"reflect"
	"strings"
)

const (
	methodResNum = 2
)

const (
	OptIgnore    = "-"
	OptOmitempty = "omitempty"
	OptDive      = "dive"
	OptWildcard  = "wildcard"
)

const (
	flagIgnore = 1 << iota
	flagOmiEmpty
	flagDive
	flagWildcard
)

type (
	StructToMap interface {
		Convert(m any) (map[string]any, error)
		ConvertWithAppend(m any, mx map[string]any) map[string]any
	}

	structToMap struct {
		*StructToMapOption
	}

	StructToMapOption struct {
		tag        string
		methodName string
	}

	SMOption func(s *StructToMapOption)
)

// WithTag sets a custom struct tag.
func WithTag(tag string) SMOption {
	return func(m *StructToMapOption) {
		m.tag = tag
	}
}

// WithMethodName sets a method name for conversion.
func WithMethodName(methodName string) SMOption {
	return func(m *StructToMapOption) {
		m.methodName = methodName
	}
}

// NewStructToMap initializes the StructToMap implementation.
func NewStructToMap(opts ...SMOption) StructToMap {
	r := &structToMap{
		StructToMapOption: &StructToMapOption{
			tag: "form",
		},
	}
	for _, opt := range opts {
		opt(r.StructToMapOption)
	}
	return r
}

// Convert transforms a struct into a map[string]any.
func (s *structToMap) Convert(m any) (map[string]any, error) {
	v := reflect.ValueOf(m)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, fmt.Errorf("%s is a nil pointer", v.Kind().String())
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// Ensure the input is a struct.
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct but %s", v.Kind().String())
	}

	t := v.Type()
	res := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)

		// Ignore unexported fields
		if fieldType.PkgPath != "" {
			continue
		}
		// Read tag
		tagVal, flag := readTag(fieldType, s.tag)

		if flag&flagIgnore != 0 {
			continue
		}

		fieldValue := v.Field(i)
		if flag&flagOmiEmpty != 0 && fieldValue.IsZero() {
			continue
		}

		if fieldValue.Type() == reflect.TypeOf(&timestamppb.Timestamp{}) {
			ts, ok := fieldValue.Interface().(*timestamppb.Timestamp)
			if ok && ts != nil && !ts.AsTime().IsZero() {
				res[tagVal] = ts.AsTime()
			}
			continue
		}

		if fieldValue.Type() == reflect.TypeOf(timestamppb.Timestamp{}) {
			ts, ok := fieldValue.Interface().(timestamppb.Timestamp)
			if ok && !ts.AsTime().IsZero() {
				res[tagVal] = ts.AsTime()
			}
			continue
		}

		// Handle nil pointers in fields
		if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
			continue
		}
		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		// Process field based on its type
		switch fieldValue.Kind() {
		case reflect.Slice, reflect.Array:
			if s.methodName != "" {
				if key, value, err := callFunc(fieldValue, s.methodName); err == nil {
					res[key] = value
					continue
				}
			}
			res[tagVal] = fieldValue.Interface()
		case reflect.Struct:
			if s.methodName != "" {
				if key, value, err := callFunc(fieldValue, s.methodName); err == nil {
					res[key] = value
					continue
				}
			}

			// Recursively convert struct fields
			deepRes, deepErr := s.Convert(fieldValue.Interface())
			if deepErr != nil {
				return nil, deepErr
			}
			if flag&flagDive != 0 {
				for k, v := range deepRes {
					res[k] = v
				}
			} else {
				res[tagVal] = deepRes
			}
		case reflect.Map, reflect.Chan:
			res[tagVal] = fieldValue.Interface()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			res[tagVal] = fieldValue.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			res[tagVal] = fieldValue.Uint()
		case reflect.Float32, reflect.Float64:
			res[tagVal] = fieldValue.Float()
		case reflect.String:
			if flag&flagWildcard != 0 {
				res[tagVal] = "%" + fieldValue.String() + "%"
			} else {
				res[tagVal] = fieldValue.String()
			}
		case reflect.Bool:
			res[tagVal] = fieldValue.Bool()
		case reflect.Complex64, reflect.Complex128:
			res[tagVal] = fieldValue.Complex()
		case reflect.Interface:
			res[tagVal] = fieldValue.Interface()
		default:
			// Ignore unsupported types
		}
	}
	return res, nil
}

// ConvertWithAppend appends additional map data to the converted struct map.
func (s *structToMap) ConvertWithAppend(m any, mx map[string]any) map[string]any {
	n, err := s.Convert(m)
	if err != nil {
		fmt.Println("Error converting struct to map:", err)
		return nil
	}
	if mx != nil {
		maps.Copy(n, mx)
	}
	return n
}

// readTag parses the struct field tag.
func readTag(f reflect.StructField, tag string) (string, int) {
	val, ok := f.Tag.Lookup(tag)
	fieldTag := f.Name // Default to struct field name if no tag
	flag := 0

	if !ok || val == "" {
		flag |= flagIgnore
		return "", flag
	}

	opts := strings.Split(val, ",")
	if opts[0] != "" {
		fieldTag = opts[0]
	}

	for _, opt := range opts {
		switch opt {
		case OptIgnore:
			flag |= flagIgnore
		case OptOmitempty:
			flag |= flagOmiEmpty
		case OptDive:
			flag |= flagDive
		case OptWildcard:
			flag |= flagWildcard
		}
	}
	return fieldTag, flag
}

// callFunc calls a method on a struct field and returns (string, any).
func callFunc(fv reflect.Value, methodName string) (string, any, error) {
	methodRes := fv.MethodByName(methodName).Call([]reflect.Value{})
	if len(methodRes) != methodResNum {
		return "", nil, fmt.Errorf("wrong method %s, should return (string, any)", methodName)
	}
	if methodRes[0].Kind() != reflect.String {
		return "", nil, fmt.Errorf("wrong method %s, first return value should be string", methodName)
	}
	return methodRes[0].String(), methodRes[1].Interface(), nil
}
