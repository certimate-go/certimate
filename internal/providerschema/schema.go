package providerschema

type Category string

const (
	CategoryDeploy Category = "deploy"
	CategoryAccess Category = "access"
	CategoryApply  Category = "apply"
	CategoryNotify Category = "notify"
)

type ValueType string

const (
	ValueTypeText         ValueType = "text"
	ValueTypeTextarea     ValueType = "textarea"
	ValueTypeSelect       ValueType = "select"
	ValueTypeRadio        ValueType = "radio"
	ValueTypeSwitch       ValueType = "switch"
	ValueTypeNumber       ValueType = "number"
	ValueTypeSecret       ValueType = "secret"
	ValueTypeCode         ValueType = "code"
	ValueTypeAutocomplete ValueType = "autocomplete"
)

var validValueTypes = map[ValueType]struct{}{
	ValueTypeText:         {},
	ValueTypeTextarea:     {},
	ValueTypeSelect:       {},
	ValueTypeRadio:        {},
	ValueTypeSwitch:       {},
	ValueTypeNumber:       {},
	ValueTypeSecret:       {},
	ValueTypeCode:         {},
	ValueTypeAutocomplete: {},
}

type FilterMode string

const (
	FilterModeFuzzy  FilterMode = "fuzzy"
	FilterModePrefix FilterMode = "prefix"
	FilterModeNone   FilterMode = "none"
)

type Validator string

const (
	ValidatorDomain     Validator = "domain"
	ValidatorHostname   Validator = "hostname"
	ValidatorIPv4       Validator = "ipv4"
	ValidatorIPv6       Validator = "ipv6"
	ValidatorPort       Validator = "port"
	ValidatorURL        Validator = "url"
	ValidatorURLHttp    Validator = "url_http"
	ValidatorURLHttps   Validator = "url_https"
	ValidatorJsonObject Validator = "json_object"
	ValidatorRegex      Validator = "regex"
	ValidatorCron       Validator = "cron"
)

var validValidators = map[Validator]struct{}{
	ValidatorDomain:     {},
	ValidatorHostname:   {},
	ValidatorIPv4:       {},
	ValidatorIPv6:       {},
	ValidatorPort:       {},
	ValidatorURL:        {},
	ValidatorURLHttp:    {},
	ValidatorURLHttps:   {},
	ValidatorJsonObject: {},
	ValidatorRegex:      {},
	ValidatorCron:       {},
}

type Op string

const (
	OpEquals    Op = "eq"
	OpNotEquals Op = "ne"
	OpIn        Op = "in"
	OpNotIn     Op = "notIn"
)

type Condition struct {
	Field  string   `json:"field"`
	Op     Op       `json:"op"`
	Values []string `json:"values,omitempty"`
}

type ValidationRule struct {
	Condition
	Name   Validator      `json:"validator"`
	Params map[string]any `json:"params,omitempty"`
}

type Option struct {
	Value    string `json:"value"`
	LabelKey string `json:"labelKey,omitempty"`
}

type Field struct {
	Name string `json:"name"`

	ValueType ValueType `json:"valueType"`

	LabelKey       string `json:"labelKey,omitempty"`
	PlaceholderKey string `json:"placeholderKey,omitempty"`
	TooltipKey     string `json:"tooltipKey,omitempty"`
	ExtraKey       string `json:"extraKey,omitempty"`

	TooltipHtml bool `json:"tooltipHtml,omitempty"`
	ExtraHtml   bool `json:"extraHtml,omitempty"`

	Default    any `json:"default,omitempty"`
	hasDefault bool

	Secret   bool `json:"secret,omitempty"`
	Required bool `json:"required,omitempty"`

	Options []Option `json:"options,omitempty"`

	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`

	FilterMode FilterMode `json:"filterMode,omitempty"`

	Span int `json:"span,omitempty"`

	VisibleWhen  []Condition      `json:"visibleWhen,omitempty"`
	RequiredWhen []Condition      `json:"requiredWhen,omitempty"`
	ValidateWhen []ValidationRule `json:"validateWhen,omitempty"`
}

type Schema struct {
	Provider string   `json:"provider"`
	Category Category `json:"category"`
	Fields   []Field  `json:"fields"`
}
