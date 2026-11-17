package workflowtemplate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/certimate-go/certimate/internal/providerschema"
)

const (
	ParamErrCodeRequired               = "required"
	ParamErrCodeInvalidDomain          = "invalidDomain"
	ParamErrCodeAccessRequired         = "accessRequired"
	ParamErrCodeAccessNotFound         = "accessNotFound"
	ParamErrCodeAccessProviderMismatch = "accessProviderMismatch"
	ParamErrCodeProviderUnavailable    = "providerUnavailable"
	ParamErrCodeConfigInvalid          = "configInvalid"
	ParamErrCodeBrokenTemplate         = "brokenTemplate"
)

func findColumn(envelope *providerschema.Envelope, name string) *providerschema.Column {
	for i := range envelope.Schema.Columns {
		if envelope.Schema.Columns[i].Name == name {
			return &envelope.Schema.Columns[i]
		}
	}
	return nil
}

func matchCondition(cond providerschema.Condition, getValue func(string) any) bool {
	value := fmt.Sprintf("%v", getValue(cond.Field))
	for _, v := range cond.Values {
		if v == value {
			switch cond.Op {
			case "eq", "in":
				return true
			case "ne", "notIn":
				return false
			}
		}
	}

	switch cond.Op {
	case "eq", "in":
		return false
	case "ne", "notIn":
		return true
	default:
		return true
	}
}

func isEmptyValue(v any) bool {
	switch tv := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(tv) == ""
	default:
		return false
	}
}

func validateProviderConfig(envelope *providerschema.Envelope, config map[string]any) error {
	defaults := make(map[string]any)
	for i := range envelope.Schema.Columns {
		col := &envelope.Schema.Columns[i]
		if col.DefaultOnMissing {
			defaults[col.Name] = col.Default
		}
	}

	getValue := func(field string) any {
		if v, ok := config[field]; ok && v != nil {
			return v
		}
		return defaults[field]
	}

	for i := range envelope.Schema.Columns {
		col := &envelope.Schema.Columns[i]

		visible := true
		for _, cond := range col.VisibleWhen {
			if !matchCondition(cond, getValue) {
				visible = false
				break
			}
		}
		if !visible {
			continue
		}

		required := col.Required
		if !required {
			for _, cond := range col.RequiredWhen {
				if matchCondition(cond, getValue) {
					required = true
					break
				}
			}
		}

		value := getValue(col.Name)
		if required && isEmptyValue(value) {
			return fmt.Errorf("the field %q is required", col.Name)
		}
		if isEmptyValue(value) {
			continue
		}

		for _, rule := range col.ValidateWhen {
			if !matchCondition(rule.Condition, getValue) {
				continue
			}
			if !runValidator(rule.Name, rule.Params, value) {
				return fmt.Errorf("the field %q is invalid", col.Name)
			}
		}

		if err := checkValueType(col, value); err != nil {
			return err
		}
	}

	return nil
}

func runValidator(name providerschema.Validator, params map[string]any, value any) bool {
	s := fmt.Sprintf("%v", value)
	switch name {
	case "positiveInteger":
		n, err := strconv.ParseFloat(s, 64)
		return err == nil && n > 0 && n == float64(int64(n))
	case "domain":
		allowWildcard := false
		if v, ok := params["allowWildcard"].(bool); ok {
			allowWildcard = v
		}
		return isValidDomainName(s, allowWildcard)
	case "hostname_or_url":
		return isValidDomainName(s, false) || strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
	case "hostname":
		return isValidDomainName(s, false)
	case "ipv4":
		return isValidIPv4(s)
	case "ipv6":
		return strings.Contains(s, ":") && strings.Count(s, ":") >= 2
	case "port":
		n, err := strconv.Atoi(s)
		return err == nil && n >= 1 && n <= 65535
	case "url":
		return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
	case "url_http":
		return strings.HasPrefix(s, "http://")
	case "url_https":
		return strings.HasPrefix(s, "https://")
	case "json_object":
		trimmed := strings.TrimSpace(s)
		return strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
	case "cron":
		return len(strings.Fields(s)) >= 5
	case "regex":
		return true
	default:
		return true
	}
}

func checkValueType(col *providerschema.Column, value any) error {
	switch col.ValueType {
	case providerschema.ValueTypeSwitch:
		if _, ok := value.(bool); !ok {
			if s, ok := value.(string); !ok || (s != "true" && s != "false") {
				return fmt.Errorf("the field %q must be a boolean", col.Name)
			}
		}

	case providerschema.ValueTypeNumber:
		n, err := toFloat64(value)
		if err != nil {
			return fmt.Errorf("the field %q must be a number", col.Name)
		}
		if col.NumericValue == "positiveInteger" && (n <= 0 || n != float64(int64(n))) {
			return fmt.Errorf("the field %q must be a positive integer", col.Name)
		}
		if col.Min != nil && n < *col.Min {
			return fmt.Errorf("the field %q must not be less than %v", col.Name, *col.Min)
		}
		if col.Max != nil && n > *col.Max {
			return fmt.Errorf("the field %q must not be greater than %v", col.Name, *col.Max)
		}

	case providerschema.ValueTypeSelect, providerschema.ValueTypeRadio:
		s := fmt.Sprintf("%v", value)
		if len(col.Options) > 0 && !col.AllowCustomValue {
			found := false
			for _, opt := range col.Options {
				if opt.Value == s {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("the value of field %q is not allowed", col.Name)
			}
		}
	}

	return nil
}

func toFloat64(value any) (float64, error) {
	switch tv := value.(type) {
	case float64:
		return tv, nil
	case float32:
		return float64(tv), nil
	case int:
		return float64(tv), nil
	case int32:
		return float64(tv), nil
	case int64:
		return float64(tv), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(tv), 64)
	default:
		return 0, fmt.Errorf("not a number")
	}
}

func isValidDomainName(s string, allowWildcard bool) bool {
	if allowWildcard && strings.HasPrefix(s, "*.") {
		s = s[2:]
	}

	if s == "" || len(s) > 253 {
		return false
	}

	labels := strings.Split(s, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, r := range label {
			if !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' || r == '-') {
				return false
			}
		}
	}

	return true
}

func isValidIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if part == "" || len(part) > 3 {
			return false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || n > 255 {
			return false
		}
		if len(part) > 1 && part[0] == '0' {
			return false
		}
	}
	return true
}
