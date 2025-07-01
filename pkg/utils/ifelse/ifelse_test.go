package ifelse_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/utils/ifelse"
)

func TestIfTrue(t *testing.T) {
	result := ifelse.If[string](true).
		Then("true branch").
		Else("false branch")

	if result != "true branch" {
		t.Errorf("Expected 'true branch', got '%s'", result)
	}
}

func TestIfFalse(t *testing.T) {
	result := ifelse.If[string](false).
		Then("true branch").
		Else("false branch")

	if result != "false branch" {
		t.Errorf("Expected 'false branch', got '%s'", result)
	}
}

func TestElseIfFirstMatch(t *testing.T) {
	result := ifelse.If[string](false).
		Then("should not run").
		ElseIf(true).
		Then("elseif branch").
		Else("should not run")

	if result != "elseif branch" {
		t.Errorf("Expected 'elseif branch', got '%s'", result)
	}
}

func TestElseIfSecondMatch(t *testing.T) {
	result := ifelse.If[string](false).
		Then("should not run").
		ElseIf(false).
		Then("should not run").
		ElseIf(true).
		Then("second elseif").
		Else("should not run")

	if result != "second elseif" {
		t.Errorf("Expected 'second elseif', got '%s'", result)
	}
}

func TestMultipleConditions(t *testing.T) {
	result := ifelse.If[string](1 > 2).
		Then("impossible").
		ElseIf(2+2 == 5).
		Then("false math").
		ElseIf(3*3 == 9).
		Then("correct math").
		Else("fallback")

	if result != "correct math" {
		t.Errorf("Expected 'correct math', got '%s'", result)
	}
}

func TestAllConditionsFalse(t *testing.T) {
	result := ifelse.If[int](false).
		Then(1).
		ElseIf(false).
		Then(2).
		ElseIf(false).
		Then(3).
		Else(99)

	if result != 99 {
		t.Errorf("Expected 99, got %d", result)
	}
}

func TestLazyEvaluationThen(t *testing.T) {
	called := []string{}

	result := ifelse.If[string](true).
		ThenFunc(func() string {
			called = append(called, "then")
			return "then"
		}).
		ElseIf(true).
		ThenFunc(func() string {
			called = append(called, "elseif")
			return "elseif"
		}).
		ElseFunc(func() string {
			called = append(called, "else")
			return "else"
		})

	// 验证结果和调用情况
	if result != "then" {
		t.Errorf("Expected 'then', got '%s'", result)
	}

	if len(called) != 1 || called[0] != "then" {
		t.Errorf("Expected only 'then' called, got %v", called)
	}
}

func TestLazyEvaluationElseIf(t *testing.T) {
	called := []string{}

	result := ifelse.If[string](false).
		ThenFunc(func() string {
			called = append(called, "then")
			return "then"
		}).
		ElseIf(true).
		ThenFunc(func() string {
			called = append(called, "elseif")
			return "elseif"
		}).
		ElseFunc(func() string {
			called = append(called, "else")
			return "else"
		})

	// 验证结果和调用情况
	if result != "elseif" {
		t.Errorf("Expected 'elseif', got '%s'", result)
	}

	if len(called) != 1 || called[0] != "elseif" {
		t.Errorf("Expected only 'elseif' called, got %v", called)
	}
}

func TestLazyEvaluationElse(t *testing.T) {
	called := []string{}

	result := ifelse.If[string](false).
		ThenFunc(func() string {
			called = append(called, "then")
			return "then"
		}).
		ElseIf(false).
		ThenFunc(func() string {
			called = append(called, "elseif")
			return "elseif"
		}).
		ElseFunc(func() string {
			called = append(called, "else")
			return "else"
		})

	// 验证结果和调用情况
	if result != "else" {
		t.Errorf("Expected 'else', got '%s'", result)
	}

	if len(called) != 1 || called[0] != "else" {
		t.Errorf("Expected only 'else' called, got %v", called)
	}
}

func TestMixedValueAndFunc(t *testing.T) {
	result := ifelse.If[int](false).
		Then(0).
		ElseIf(false).
		ThenFunc(func() int {
			return 1
		}).
		ElseIf(true).
		Then(2).
		Else(3)

	if result != 2 {
		t.Errorf("Expected 2, got %d", result)
	}
}

func TestComplexNumericLogic(t *testing.T) {
	x := 15
	result := ifelse.If[string](x < 10).
		Then("single digit").
		ElseIf(x < 20).
		Then("teens").
		ElseIf(x < 30).
		Then("twenties").
		Else("older")

	if result != "teens" {
		t.Errorf("Expected 'teens', got '%s'", result)
	}
}
