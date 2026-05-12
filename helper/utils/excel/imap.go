package excel

import (
	"pulse/helper/utils/toolkit/timex"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type (
	ImportMap[T any] interface {
		Maps(rows [][]string) ([]*T, error)
		Map(row []string) (*T, error)
	}

	importMap[T any] struct {
	}

	importTag struct {
		Index  int    `json:"index"`
		Format string `json:"format"`
	}
)

func NewImportMap[T any]() ImportMap[T] {
	return &importMap[T]{}
}

func (s *importMap[T]) Maps(rows [][]string) ([]*T, error) {
	var results []*T
	for _, row := range rows {
		result, err := s.Map(row)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *importMap[T]) Map(row []string) (*T, error) {
	if len(row) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	var result T
	m := len(row)
	v := reflect.ValueOf(&result).Elem()
	t := v.Type()

	numField := v.NumField()
	for i := 0; i < numField; i++ {
		field := v.Field(i)
		tag, err := s.getTags(t.Field(i).Tag.Get("map"), t.Field(i).Name, numField)
		if err != nil {
			return nil, err
		}
		if tag.Index >= m {
			continue
		}
		value := strings.TrimSpace(row[tag.Index])

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int:
			intVal, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("error converting field %s to int: %w", t.Field(i).Name, err)
			}
			field.SetInt(int64(intVal))
		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				date, err := s.date(value, tag.Format)
				if err != nil {
					return nil, fmt.Errorf("error parsing value %q to *time.Time: %w", value, err)
				}
				field.Set(reflect.ValueOf(date))
			}
		case reflect.Ptr:
			if field.Type().Elem() == reflect.TypeOf(time.Time{}) {
				if len(value) == 0 {
					continue
				}
				date, err := s.date(value, tag.Format)
				if err != nil {
					return nil, fmt.Errorf("error parsing value %q to *time.Time: %w", value, err)
				}
				field.Set(reflect.ValueOf(&date))
			}
		default:
			return nil, fmt.Errorf("unsupported field type: %s", field.Kind())
		}
	}
	return &result, nil
}

func (s *importMap[T]) getTags(tagName, fieldName string, total int) (*importTag, error) {
	if tagName == "" {
		return nil, fmt.Errorf("tagName is empty")
	}
	var tag importTag
	for _, part := range strings.Split(tagName, ";") {
		parts := strings.Split(part, ":")
		if len(parts) != 2 {
			continue
		}
		switch strings.ToLower(parts[0]) {
		case "index":
			index, err := strconv.Atoi(parts[1])
			if err != nil || index >= total {
				logx.Errorf("invalid or missing index for field %s", fieldName)
				continue
			}
			tag.Index = index
		case "format":
			tag.Format = strings.TrimSpace(parts[1])
		}
	}
	return &tag, nil
}

func (s *importMap[T]) date(value, format string) (time.Time, error) {
	if format == "" {
		format = time.DateOnly
	}
	date, err := timex.ParseInLocal(format, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing value %q to *time.Time: %w", value, err)
	}
	return date, nil
}
