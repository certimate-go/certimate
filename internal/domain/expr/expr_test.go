package expr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/certimate-go/certimate/internal/domain/expr"
)

func TestUnmarshalExpr(t *testing.T) {
	type testInput struct {
		data []byte
	}
	testCases := []struct {
		name        string
		input       testInput
		expected    *expr.Expr
		expectedErr bool
	}{
		{
			name: "case1",
			input: testInput{
				data: []byte(`{"left":{"left":{"selector":{"id":"ODnYSOXB6HQP2_vz6JcZE","name":"certificate.validity","type":"boolean"},"type":"var"},"operator":"eq","right":{"type":"const","value":"true","valueType":"boolean"},"type":"comparison"},"operator":"and","right":{"left":{"selector":{"id":"ODnYSOXB6HQP2_vz6JcZE","name":"certificate.daysLeft","type":"number"},"type":"var"},"operator":"eq","right":{"type":"const","value":"2","valueType":"number"},"type":"comparison"},"type":"logical"}`),
			},
			expectedErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := expr.UnmarshalExpr(tc.input.data)
			if tc.expectedErr {
				assert.Error(t, err, "Case: %-20s", tc.name)
			} else {
				assert.NoError(t, err, "Case: %-20s", tc.name)
			}
		})
	}
}

func TestEvalExpr(t *testing.T) {
	type testInput struct {
		data      []byte
		variables map[string]map[string]any
	}
	testCases := []struct {
		name        string
		input       testInput
		expected    *expr.EvalResult
		expectedErr bool
	}{
		{
			name: "case1",
			input: testInput{
				data: []byte(`{"left":{"left":{"selector":{"id":"ODnYSOXB6HQP2_vz6JcZE","name":"certificate.validity","type":"boolean"},"type":"var"},"operator":"eq","right":{"type":"const","value":"true","valueType":"boolean"},"type":"comparison"},"operator":"and","right":{"left":{"selector":{"id":"ODnYSOXB6HQP2_vz6JcZE","name":"certificate.daysLeft","type":"number"},"type":"var"},"operator":"eq","right":{"type":"const","value":"2","valueType":"number"},"type":"comparison"},"type":"logical"}`),
				variables: map[string]map[string]any{
					"ODnYSOXB6HQP2_vz6JcZE": {
						"certificate.validity": "true",
						"certificate.daysLeft": "2",
					},
				},
			},
			expectedErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			temp, err := expr.UnmarshalExpr(tc.input.data)
			assert.NoError(t, err, "failed to unmarshal expression")

			result, err := temp.Eval(tc.input.variables)
			if tc.expectedErr {
				assert.Error(t, err, "Case: %-20s", tc.name)
			} else {
				assert.NoError(t, err, "Case: %-20s", tc.name)
				assert.True(t, result.Value.(bool))
			}
		})
	}
}

func TestEvalLogicalExpr(t *testing.T) {
	t.Run("And", func(t *testing.T) {
		logicalExpr := expr.LogicalExpr{
			Left: expr.ConstantExpr{
				Type:      "const",
				Value:     "true",
				ValueType: "boolean",
			},
			Operator: expr.And,
			Right: expr.ConstantExpr{
				Type:      "const",
				Value:     "true",
				ValueType: "boolean",
			},
		}
		result, err := logicalExpr.Eval(nil)
		assert.NoError(t, err, "failed to evaluate logical expression")
		assert.True(t, result.Value.(bool))
	})

	t.Run("Or", func(t *testing.T) {
		orExpr := expr.LogicalExpr{
			Left: expr.ConstantExpr{
				Type:      "const",
				Value:     "true",
				ValueType: "boolean",
			},
			Operator: expr.Or,
			Right: expr.ConstantExpr{
				Type:      "const",
				Value:     "true",
				ValueType: "boolean",
			},
		}
		result, err := orExpr.Eval(nil)
		assert.NoError(t, err, "failed to evaluate logical expression")
		assert.True(t, result.Value.(bool))
	})
}
