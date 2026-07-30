package restjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"
)

type namePolicy int

const (
	namePolicyNone namePolicy = iota
	namePolicyCamelCase
)

type serializer struct {
	namePolicy namePolicy
	omitEmpty  bool
}

func NewSerializer() *serializer {
	return &serializer{}
}

func (s *serializer) UseDefaultNamePolicy() {
	s.namePolicy = namePolicyNone
}

func (s *serializer) UseCamelCaseNamePolicy() {
	s.namePolicy = namePolicyCamelCase
}

func (s *serializer) UseOmitEmptyValue() {
	s.omitEmpty = true
}

func (s *serializer) Serialize(v any) ([]byte, error) {
	m, err := s.SerializeToMap(v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func (s *serializer) SerializeToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	raw := make(map[string]any)
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	s.tidy(raw)

	switch s.namePolicy {
	case namePolicyNone:
		return raw, nil

	case namePolicyCamelCase:
		raw = s.deepCamelize(raw).(map[string]any)
		return raw, nil

	default:
		return nil, fmt.Errorf("unsupported name policy: %v", s.namePolicy)
	}
}

func (s *serializer) deepCamelize(v any) any {
	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Map:
		{
			if val.Type().Key().Kind() != reflect.String {
				return v
			}

			newMap := make(map[string]any)
			for _, key := range val.MapKeys() {
				nkey := s.toCamelCase(key.String())
				elem := val.MapIndex(key).Interface()
				newMap[nkey] = s.deepCamelize(elem)
			}
			return newMap
		}

	case reflect.Slice, reflect.Array:
		{
			newSlice := make([]any, val.Len())
			for i := 0; i < val.Len(); i++ {
				newSlice[i] = s.deepCamelize(val.Index(i).Interface())
			}
			return newSlice
		}

	default:
		return v
	}
}

func (s *serializer) toCamelCase(str string) string {
	if str == "" {
		return str
	}

	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func (s *serializer) tidy(m map[string]any) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
			continue
		}

		switch val := v.(type) {
		case string:
			if s.omitEmpty && len(val) == 0 {
				delete(m, k)
			}

		case map[string]any:
			if s.omitEmpty && len(val) == 0 {
				delete(m, k)
			} else {
				s.tidy(val)
			}

		case []any:
			if s.omitEmpty && len(val) == 0 {
				delete(m, k)
			} else {
				for _, item := range val {
					if item == nil {
						continue
					}
					if subm, ok := item.(map[string]any); ok {
						s.tidy(subm)
					}
				}
			}
		}
	}
}
