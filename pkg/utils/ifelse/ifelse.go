package ifelse

type ifExpr[T any] struct {
	condition bool
}

type thenExpr[T any] struct {
	condition  bool
	consequent T
}

// 示例：
//
//	result := ifelse.If[T](condition).Then(consequent).Else(alternative)
func If[T any](condition bool) *ifExpr[T] {
	return &ifExpr[T]{
		condition: condition,
	}
}

func (e *ifExpr[T]) Then(consequent T) *thenExpr[T] {
	return &thenExpr[T]{
		condition:  e.condition,
		consequent: consequent,
	}
}

func (e *thenExpr[T]) Else(alternative T) T {
	if e.condition {
		return e.consequent
	}

	return alternative
}
