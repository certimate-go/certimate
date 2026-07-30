package amzjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	smithytime "github.com/aws/smithy-go/time"

	xreflect "github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/reflect"
)

type deserializer struct {
	useEpochTime bool
}

func NewDeserializer() *deserializer {
	return &deserializer{}
}

func (d *deserializer) UseEpochTime() {
	d.useEpochTime = true
}

func (d *deserializer) Deserialize(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}

	elem := xreflect.Indirect(rv)
	if elem.Kind() != reflect.Struct {
		return json.Unmarshal(data, v)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var raw map[string]any
	if err := dec.Decode(&raw); err != nil {
		return err
	}
	return d.decodeStruct(raw, elem)
}

func (d *deserializer) decodeStruct(dataM map[string]any, rv reflect.Value) error {
	t := rv.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		key := field.Name
		val, ok := dataM[key]
		if !ok {
			continue
		}

		fv := rv.Field(i)
		if err := d.decodeValue(val, fv); err != nil {
			return fmt.Errorf("failed to set field %s: %w", key, err)
		}
	}

	return nil
}

func (d *deserializer) decodeValue(dataV any, rv reflect.Value) error {
	if dataV == nil {
		return nil
	}

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return d.decodeValue(dataV, rv.Elem())
	}

	switch rv.Kind() {
	case reflect.Struct:
		{
			if rv.Type() == reflect.TypeOf(time.Time{}) {
				t, err := d.toTime(dataV)
				if err != nil {
					return fmt.Errorf("expected time.Time, got %T: %w", dataV, err)
				}

				rv.Set(reflect.ValueOf(t))
				return nil
			}

			m, ok := dataV.(map[string]any)
			if !ok {
				return fmt.Errorf("expected object for struct, got %T", dataV)
			}

			return d.decodeStruct(m, rv)
		}

	case reflect.Slice:
		{
			arr, ok := dataV.([]any)
			if !ok {
				return fmt.Errorf("expected array for slice, got %T", dataV)
			}

			sl := reflect.MakeSlice(rv.Type(), len(arr), len(arr))
			for i := range arr {
				if err := d.decodeValue(arr[i], sl.Index(i)); err != nil {
					return err
				}
			}

			rv.Set(sl)
			return nil
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			n, err := d.toInt64(dataV)
			if err != nil {
				return fmt.Errorf("expected integer, got %T: %w", dataV, err)
			}

			rv.SetInt(n)
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			n, err := d.toUint64(dataV)
			if err != nil {
				return fmt.Errorf("expected unsigned integer, got %T: %w", dataV, err)
			}

			rv.SetUint(n)
			return nil
		}

	case reflect.Float32, reflect.Float64:
		{
			n, err := d.toFloat64(dataV)
			if err != nil {
				return fmt.Errorf("expected float, got %T: %w", dataV, err)
			}

			rv.SetFloat(n)
			return nil
		}

	case reflect.String:
		{
			s, ok := dataV.(string)
			if !ok {
				return fmt.Errorf("expected string, got %T", dataV)
			}

			rv.SetString(s)
			return nil
		}

	case reflect.Bool:
		{
			b, ok := dataV.(bool)
			if !ok {
				return fmt.Errorf("expected bool, got %T", dataV)
			}

			rv.SetBool(b)
			return nil
		}

	default:
		return fmt.Errorf("unsupported kind: %s", rv.Kind())
	}
}

func (d *deserializer) toInt64(val any) (int64, error) {
	switch t := val.(type) {
	case json.Number:
		return t.Int64()
	case float64:
		return int64(t), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", val)
	}
}

func (d *deserializer) toUint64(val any) (uint64, error) {
	switch t := val.(type) {
	case json.Number:
		i, err := t.Int64()
		return uint64(i), err
	case float64:
		return uint64(t), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", val)
	}
}

func (d *deserializer) toFloat64(val any) (float64, error) {
	switch t := val.(type) {
	case json.Number:
		return t.Float64()
	case float64:
		return t, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}

func (d *deserializer) toTime(val any) (time.Time, error) {
	switch t := val.(type) {
	case string:
		return smithytime.ParseDateTime(t)
	default:
		if d.useEpochTime {
			if sec, err := d.toFloat64(val); err == nil {
				return smithytime.ParseEpochSeconds(sec), nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", val)
	}
}
