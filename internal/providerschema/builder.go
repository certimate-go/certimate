package providerschema

import "fmt"

type FieldOption func(*Field)

func LabelKey(key string) FieldOption {
	return func(f *Field) { f.LabelKey = key }
}

func PlaceholderKey(key string) FieldOption {
	return func(f *Field) { f.PlaceholderKey = key }
}

func TooltipKey(key string) FieldOption {
	return func(f *Field) { f.TooltipKey = key }
}

func ExtraKey(key string) FieldOption {
	return func(f *Field) { f.ExtraKey = key }
}

func TooltipHtml() FieldOption {
	return func(f *Field) { f.TooltipHtml = true }
}

func ExtraHtml() FieldOption {
	return func(f *Field) { f.ExtraHtml = true }
}

func Default(value any) FieldOption {
	return func(f *Field) {
		f.Default = value
		f.hasDefault = true
	}
}

func Secret() FieldOption {
	return func(f *Field) { f.Secret = true }
}

func Required() FieldOption {
	return func(f *Field) { f.Required = true }
}

func Span(columns int) FieldOption {
	return func(f *Field) { f.Span = columns }
}

func Min(value float64) FieldOption {
	return func(f *Field) { v := value; f.Min = &v }
}

func Max(value float64) FieldOption {
	return func(f *Field) { v := value; f.Max = &v }
}

func WithFilterMode(mode FilterMode) FieldOption {
	return func(f *Field) { f.FilterMode = mode }
}

func Options(values ...string) FieldOption {
	return func(f *Field) {
		opts := make([]Option, len(values))
		for i, v := range values {
			opts[i] = Option{Value: v}
		}
		f.Options = opts
	}
}

func OptionsWith(opts ...Option) FieldOption {
	return func(f *Field) { f.Options = opts }
}

type condKind int

const (
	condVisible condKind = iota
	condRequired
)

type visibilityBuilder struct {
	field string
	kind  condKind
}

func (b *visibilityBuilder) apply(c Condition) FieldOption {
	return func(f *Field) {
		switch b.kind {
		case condVisible:
			f.VisibleWhen = append(f.VisibleWhen, c)
		case condRequired:
			f.RequiredWhen = append(f.RequiredWhen, c)
		}
	}
}

func VisibleWhen(field string) *visibilityBuilder {
	return &visibilityBuilder{field: field, kind: condVisible}
}

func RequiredWhen(field string) *visibilityBuilder {
	return &visibilityBuilder{field: field, kind: condRequired}
}

func (b *visibilityBuilder) Equals(value string) FieldOption {
	return b.apply(Condition{Field: b.field, Op: OpEquals, Values: []string{value}})
}

func (b *visibilityBuilder) Not(value string) FieldOption {
	return b.apply(Condition{Field: b.field, Op: OpNotEquals, Values: []string{value}})
}

func (b *visibilityBuilder) In(values ...string) FieldOption {
	return b.apply(Condition{Field: b.field, Op: OpIn, Values: values})
}

func (b *visibilityBuilder) NotIn(values ...string) FieldOption {
	return b.apply(Condition{Field: b.field, Op: OpNotIn, Values: values})
}

type validateBuilder struct {
	cond Condition
	set  bool
}

func (b *validateBuilder) on(op Op, values []string) *validateBuilder {
	b.cond = Condition{Field: b.cond.Field, Op: op, Values: values}
	b.set = true
	return b
}

func (b *validateBuilder) Equals(value string) *validateBuilder {
	return b.on(OpEquals, []string{value})
}

func (b *validateBuilder) Not(value string) *validateBuilder {
	return b.on(OpNotEquals, []string{value})
}

func (b *validateBuilder) In(values ...string) *validateBuilder { return b.on(OpIn, values) }

func (b *validateBuilder) NotIn(values ...string) *validateBuilder { return b.on(OpNotIn, values) }

func (b *validateBuilder) Validator(name Validator, params ...ValidatorParam) FieldOption {
	rule := ValidationRule{
		Condition: b.cond,
		Name:      name,
		Params:    validatorParams(params),
	}
	return func(f *Field) { f.ValidateWhen = append(f.ValidateWhen, rule) }
}

type ValidatorParam struct {
	Key   string
	Value any
}

func WithParam(key string, value any) ValidatorParam {
	return ValidatorParam{Key: key, Value: value}
}

func validatorParams(params []ValidatorParam) map[string]any {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]any, len(params))
	for _, p := range params {
		out[p.Key] = p.Value
	}
	return out
}

func ValidateWhen(field string) *validateBuilder {
	return &validateBuilder{cond: Condition{Field: field}}
}

type builder struct {
	schema Schema
	err    error
}

func New(provider string, category Category) *builder {
	return &builder{schema: Schema{Provider: provider, Category: category}}
}

func (b *builder) Field(name string, vt ValueType, opts ...FieldOption) *builder {
	if b.err != nil {
		return b
	}
	f := Field{Name: name, ValueType: vt}
	for _, opt := range opts {
		opt(&f)
	}
	b.schema.Fields = append(b.schema.Fields, f)
	return b
}

func (b *builder) Build() (*Schema, error) {
	if b.err != nil {
		return nil, b.err
	}
	if err := validate(&b.schema); err != nil {
		return nil, err
	}
	schema := b.schema
	return &schema, nil
}

func validate(s *Schema) error {
	if s.Provider == "" {
		return fmt.Errorf("providerschema: provider must not be empty")
	}

	declared := make(map[string]struct{}, len(s.Fields))
	for _, f := range s.Fields {
		if f.Name == "" {
			return fmt.Errorf("providerschema: field name must not be empty")
		}
		if _, dup := declared[f.Name]; dup {
			return fmt.Errorf("providerschema: duplicate field name %q", f.Name)
		}
		declared[f.Name] = struct{}{}
	}

	for i := range s.Fields {
		if err := validateField(&s.Fields[i], declared); err != nil {
			return err
		}
	}
	return nil
}

func validateField(f *Field, declared map[string]struct{}) error {
	if _, ok := validValueTypes[f.ValueType]; !ok {
		return fmt.Errorf("providerschema: field %q has unknown valueType %q", f.Name, f.ValueType)
	}

	switch f.ValueType {
	case ValueTypeSelect, ValueTypeRadio:
		if len(f.Options) == 0 {
			return fmt.Errorf("providerschema: field %q of valueType %q requires Options", f.Name, f.ValueType)
		}
	case ValueTypeNumber:
		if f.Min != nil && f.Max != nil && *f.Min > *f.Max {
			return fmt.Errorf("providerschema: field %q has min (%v) greater than max (%v)", f.Name, *f.Min, *f.Max)
		}
	}

	if f.hasDefault {
		if err := validateDefault(f); err != nil {
			return err
		}
	}

	for _, c := range f.VisibleWhen {
		if err := validateCondition(c, f.Name, "visibleWhen", declared); err != nil {
			return err
		}
	}
	for _, c := range f.RequiredWhen {
		if err := validateCondition(c, f.Name, "requiredWhen", declared); err != nil {
			return err
		}
	}
	for _, r := range f.ValidateWhen {
		if err := validateCondition(r.Condition, f.Name, "validateWhen", declared); err != nil {
			return err
		}
		if _, ok := validValidators[r.Name]; !ok {
			return fmt.Errorf("providerschema: field %q validateWhen uses unknown validator %q", f.Name, r.Name)
		}
	}
	return nil
}

func validateCondition(c Condition, owner, clause string, declared map[string]struct{}) error {
	if _, ok := declared[c.Field]; !ok {
		return fmt.Errorf("providerschema: field %q %s references undeclared discriminator %q", owner, clause, c.Field)
	}
	if len(c.Values) == 0 {
		return fmt.Errorf("providerschema: field %q %s on %q has no values", owner, clause, c.Field)
	}
	switch c.Op {
	case OpEquals, OpNotEquals, OpIn, OpNotIn:
	default:
		return fmt.Errorf("providerschema: field %q %s on %q has unknown op %q", owner, clause, c.Field, c.Op)
	}
	return nil
}

func validateDefault(f *Field) error {
	v := f.Default
	switch f.ValueType {
	case ValueTypeText, ValueTypeTextarea, ValueTypeSecret, ValueTypeCode, ValueTypeAutocomplete:
		if _, ok := v.(string); !ok {
			return fmt.Errorf("providerschema: field %q default must be a string, got %T", f.Name, v)
		}
	case ValueTypeSelect, ValueTypeRadio:
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("providerschema: field %q default must be a string, got %T", f.Name, v)
		}
		if !optionContains(f.Options, s) {
			return fmt.Errorf("providerschema: field %q default %q is not among its options", f.Name, s)
		}
	case ValueTypeSwitch:
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("providerschema: field %q default must be a bool, got %T", f.Name, v)
		}
	case ValueTypeNumber:
		switch n := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			_ = n
		default:
			return fmt.Errorf("providerschema: field %q default must be a number, got %T", f.Name, v)
		}
	}
	return nil
}

func optionContains(opts []Option, value string) bool {
	for _, o := range opts {
		if o.Value == value {
			return true
		}
	}
	return false
}
