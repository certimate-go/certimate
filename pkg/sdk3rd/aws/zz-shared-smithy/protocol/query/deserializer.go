package query

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	smithytime "github.com/aws/smithy-go/time"

	xreflect "github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/reflect"
)

type deserializer struct {
	action string
}

func NewDeserializer(action string) *deserializer {
	return &deserializer{action: action}
}

func (d *deserializer) Deserialize(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return xml.UnmarshalError("xml: Unmarshal(nil)")
	}

	elem := xreflect.Indirect(rv)
	if elem.Kind() != reflect.Struct {
		return xml.Unmarshal(data, v)
	}

	dec := xml.NewDecoder(bytes.NewReader(data))

	if d.action != "" {
		_, err := d.advanceStartElement(dec, d.action+"Response")
		if err != nil {
			return fmt.Errorf("cannot find root element: <%sResponse>", d.action)
		}

		start, err := d.advanceStartElement(dec, d.action+"Result")
		if err != nil {
			return fmt.Errorf("cannot find start element: <%sResult>", d.action)
		}

		if err := d.decodeStruct(dec, start, elem); err != nil {
			return err
		}

		return nil
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if se, ok := tok.(xml.StartElement); ok {
			if err := d.decodeStruct(dec, se, elem); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *deserializer) advanceStartElement(dec *xml.Decoder, name string) (xml.StartElement, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return xml.StartElement{}, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == name {
				return t, nil
			}
		}
	}
}

func (d *deserializer) decodeStruct(dec *xml.Decoder, start xml.StartElement, rv reflect.Value) error {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return d.decodeStruct(dec, start, rv.Elem())
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	fields := map[string]reflect.Value{}
	for i := 0; i < rv.NumField(); i++ {
		ft := rv.Type().Field(i)
		name := ft.Name
		fields[name] = rv.Field(i)
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			field, ok := fields[t.Name.Local]
			if !ok {
				if err := dec.Skip(); err != nil {
					return err
				}
				continue
			}

			if err := d.decodeValue(dec, t, field); err != nil {
				return err
			}

		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
			if d.action != "" && (t.Name.Local == d.action+"Result" || t.Name.Local == d.action+"Response") {
				return nil
			}
		}
	}
}

func (d *deserializer) decodeValue(dec *xml.Decoder, start xml.StartElement, rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Ptr:
		{
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			return d.decodeValue(dec, start, rv.Elem())
		}

	case reflect.Struct:
		{
			if rv.Type() == reflect.TypeOf(time.Time{}) {
				s, err := d.decodeCharData(dec)
				if err != nil {
					return err
				}

				t, err := smithytime.ParseDateTime(s)
				if err != nil {
					return fmt.Errorf("cannot convert '%s' to time.Time: %w", s, err)
				}

				rv.Set(reflect.ValueOf(t))
				return nil
			}

			return d.decodeStruct(dec, start, rv)
		}

	case reflect.Slice:
		{
			return d.decodeSlice(dec, start, rv)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s, err := d.decodeCharData(dec)
		if err != nil {
			return err
		}

		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert '%s' to int64: %w", s, err)
		}

		rv.SetInt(i)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s, err := d.decodeCharData(dec)
		if err != nil {
			return err
		}

		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert '%s' to uint64: %w", s, err)
		}

		rv.SetUint(u)
		return nil

	case reflect.Float32, reflect.Float64:
		s, err := d.decodeCharData(dec)
		if err != nil {
			return err
		}

		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("cannot convert '%s' to float: %w", s, err)
		}

		rv.SetFloat(f)
		return nil

	case reflect.String:
		{
			s, err := d.decodeCharData(dec)
			if err != nil {
				return err
			}

			rv.SetString(s)
			return nil
		}

	case reflect.Bool:
		{
			s, err := d.decodeCharData(dec)
			if err != nil {
				return err
			}

			b, err := strconv.ParseBool(s)
			if err != nil {
				return fmt.Errorf("cannot convert '%s' to bool: %w", s, err)
			}

			rv.SetBool(b)
			return nil
		}

	default:
		return fmt.Errorf("unsupported kind: %s", rv.Kind())
	}
}

func (d *deserializer) decodeSlice(dec *xml.Decoder, start xml.StartElement, rv reflect.Value) error {
	elemType := rv.Type().Elem()

	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}

		case xml.StartElement:
			if t.Name.Local != "member" {
				return fmt.Errorf("expected <member>, got <%s>", t.Name.Local)
			}

			newElem := reflect.New(elemType).Elem()
			if err := d.decodeValue(dec, t, newElem); err != nil {
				return err
			}

			rv.Set(reflect.Append(rv, newElem))
		}
	}
}

func (d *deserializer) decodeCharData(dec *xml.Decoder) (string, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", err
	}

	cd, ok := tok.(xml.CharData)
	if !ok {
		return "", fmt.Errorf("expected char data")
	}

	return string(cd), nil
}
