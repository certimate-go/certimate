package logging

import (
	"log/slog"

	types "ganigeorgiev.com/pocketbase/pocketbase/tools/types"
)

type Record struct {
	slog.Record
}

func (r Record) Data() types.JSONMap[any] {
	data := make(map[string]any, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		if err := r.resolveAttr(data, a); err != nil {
			return false
		}
		return true
	})

	return types.JSONMap[any](data)
}

func (r Record) resolveAttr(data map[string]any, attr slog.Attr) error {
	attr.Value = attr.Value.Resolve()

	if attr.Equal(slog.Attr{}) {
		return nil
	}

	switch attr.Value.Kind() {
	case slog.KindGroup:
		{
			attrs := attr.Value.Group()
			if len(attrs) == 0 {
				return nil
			}

			groupData := make(map[string]any, len(attrs))

			for _, subAttr := range attrs {
				r.resolveAttr(groupData, subAttr)
			}

			if len(groupData) > 0 {
				data[attr.Key] = groupData
			}
		}

	default:
		{
			switch v := attr.Value.Any().(type) {
			case error:
				data[attr.Key] = v.Error()
			default:
				data[attr.Key] = v
			}
		}
	}

	return nil
}
