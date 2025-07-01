package ifelse

type branch[T any] struct {
	cond bool
	fn   func() T
}

type ifExpr[T any] struct {
	cond bool
}

type thenExpr[T any] struct {
	branches []branch[T]
}

type elseIfExpr[T any] struct {
	branches []branch[T]
	cond     bool
}

// 用法示例：
//
//	result := ifelse.If[string](age < 18).Then("child").
//	  ElseIf(age < 60).Then("adult").
//	  ElseIf(age < 120).Then("senior").
//	  Else("invalid")
func If[T any](condition bool) *ifExpr[T] {
	return &ifExpr[T]{cond: condition}
}

func (e *ifExpr[T]) Then(consequent T) *thenExpr[T] {
	return &thenExpr[T]{
		branches: []branch[T]{
			{cond: e.cond, fn: func() T { return consequent }},
		},
	}
}

func (e *ifExpr[T]) ThenFunc(consequent func() T) *thenExpr[T] {
	return &thenExpr[T]{
		branches: []branch[T]{
			{cond: e.cond, fn: consequent},
		},
	}
}

func (e *thenExpr[T]) ElseIf(condition bool) *elseIfExpr[T] {
	return &elseIfExpr[T]{
		branches: e.branches,
		cond:     condition,
	}
}

func (e *elseIfExpr[T]) Then(alternative T) *thenExpr[T] {
	branch := branch[T]{cond: e.cond, fn: func() T { return alternative }}
	return &thenExpr[T]{
		branches: append(e.branches, branch),
	}
}

func (e *elseIfExpr[T]) ThenFunc(alternativeFunc func() T) *thenExpr[T] {
	branch := branch[T]{cond: e.cond, fn: alternativeFunc}
	return &thenExpr[T]{
		branches: append(e.branches, branch),
	}
}

func (e *thenExpr[T]) Else(alternative T) T {
	for _, b := range e.branches {
		if b.cond {
			return b.fn()
		}
	}
	return alternative
}

func (e *thenExpr[T]) ElseFunc(alternativeFunc func() T) T {
	for _, b := range e.branches {
		if b.cond {
			return b.fn()
		}
	}
	return alternativeFunc()
}
