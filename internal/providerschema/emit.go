package providerschema

const SchemaVersion = "form/v1"

type Envelope struct {
	SchemaVersion string         `json:"schemaVersion"`
	Provider      string         `json:"provider"`
	Category      Category       `json:"category"`
	Schema        EnvelopeSchema `json:"schema"`
}

type EnvelopeSchema struct {
	Columns []Column `json:"columns"`
}

type Column struct {
	Name           string           `json:"name"`
	ValueType      ValueType        `json:"valueType"`
	LabelKey       string           `json:"labelKey,omitempty"`
	PlaceholderKey string           `json:"placeholderKey,omitempty"`
	TooltipKey     string           `json:"tooltipKey,omitempty"`
	ExtraKey       string           `json:"extraKey,omitempty"`
	TooltipHtml    bool             `json:"tooltipHtml,omitempty"`
	ExtraHtml      bool             `json:"extraHtml,omitempty"`
	Default        any              `json:"default,omitempty"`
	Secret         bool             `json:"secret,omitempty"`
	Required       bool             `json:"required,omitempty"`
	Options        []Option         `json:"options,omitempty"`
	Min            *float64         `json:"min,omitempty"`
	Max            *float64         `json:"max,omitempty"`
	FilterMode     FilterMode       `json:"filterMode,omitempty"`
	Span           int              `json:"span,omitempty"`
	Dependencies   []string         `json:"dependencies,omitempty"`
	VisibleWhen    []Condition      `json:"visibleWhen,omitempty"`
	RequiredWhen   []Condition      `json:"requiredWhen,omitempty"`
	ValidateWhen   []ValidationRule `json:"validateWhen,omitempty"`
}

func Emit(s *Schema) *Envelope {
	columns := make([]Column, len(s.Fields))
	for i := range s.Fields {
		columns[i] = emitColumn(&s.Fields[i])
	}
	return &Envelope{
		SchemaVersion: SchemaVersion,
		Provider:      s.Provider,
		Category:      s.Category,
		Schema:        EnvelopeSchema{Columns: columns},
	}
}

func emitColumn(f *Field) Column {
	return Column{
		Name:           f.Name,
		ValueType:      f.ValueType,
		LabelKey:       f.LabelKey,
		PlaceholderKey: f.PlaceholderKey,
		TooltipKey:     f.TooltipKey,
		ExtraKey:       f.ExtraKey,
		TooltipHtml:    f.TooltipHtml,
		ExtraHtml:      f.ExtraHtml,
		Default:        f.Default,
		Secret:         f.Secret,
		Required:       f.Required,
		Options:        emitOptions(f),
		Min:            f.Min,
		Max:            f.Max,
		FilterMode:     f.FilterMode,
		Span:           f.Span,
		Dependencies:   dependencies(f),
		VisibleWhen:    f.VisibleWhen,
		RequiredWhen:   f.RequiredWhen,
		ValidateWhen:   f.ValidateWhen,
	}
}

func emitOptions(f *Field) []Option {
	switch f.ValueType {
	case ValueTypeSelect, ValueTypeRadio, ValueTypeAutocomplete:
		return f.Options
	}
	return nil
}

func dependencies(f *Field) []string {
	seen := make(map[string]struct{})
	var deps []string
	add := func(field string) {
		if field == "" {
			return
		}
		if _, ok := seen[field]; ok {
			return
		}
		seen[field] = struct{}{}
		deps = append(deps, field)
	}
	for _, c := range f.VisibleWhen {
		add(c.Field)
	}
	for _, c := range f.RequiredWhen {
		add(c.Field)
	}
	for _, r := range f.ValidateWhen {
		add(r.Field)
	}
	return deps
}
