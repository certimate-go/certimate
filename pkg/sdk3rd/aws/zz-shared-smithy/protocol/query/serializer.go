package query

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type serializer struct {
	omitEmpty bool
}

func NewSerializer() *serializer {
	return &serializer{}
}

func (s *serializer) UseOmitEmptyValue() {
	s.omitEmpty = true
}

func (s *serializer) Serialize(v any) ([]byte, error) {
	m, err := s.SerializeToMap(v)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	for k, v := range m {
		values.Set(k, v)
	}

	return []byte(values.Encode()), nil
}

func (s *serializer) SerializeToMap(v any) (map[string]string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	raw := make(map[string]any)
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	s.tidy(raw)

	m := make(map[string]string)
	s.flatten("", raw, m)
	return m, nil
}

func (s *serializer) flatten(prefix string, val any, out map[string]string) {
	switch v := val.(type) {
	case map[string]any:
		for k, sub := range v {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			s.flatten(key, sub, out)
		}

	case []any:
		for i, elem := range v {
			index := strconv.Itoa(i + 1)
			key := prefix + ".member." + index
			s.flatten(key, elem, out)
		}

	default:
		out[prefix] = fmt.Sprintf("%v", v)
	}
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
