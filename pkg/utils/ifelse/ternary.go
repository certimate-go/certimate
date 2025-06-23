package ifelse

// 三目条件函数。
//
// 入参:
//   - condition: 条件。
//   - consequent: 条件为真时返回的值。
//   - alternative: 条件为假时返回的值。
//
// 出参:
//   - 若 condition 的为真，将返回 consequent；否则，将返回 alternative。
func Ternary[T any](condition bool, consequent, alternative T) T {
	if condition {
		return consequent
	} else {
		return alternative
	}
}

// 与 [Ternary] 类似，但返回值支持延迟计算函数。
//
// 入参:
//   - condition: 条件。
//   - consequentFunc: 条件为真时返回的计算函数。
//   - alternativeFunc: 条件为假时返回的计算函数。
//
// 出参:
//   - 若 condition 的为真，将返回 consequentFunc 的计算结果；否则，将返回 alternativeFunc 的计算结果。
func TernaryFunc[T any](condition bool, consequentFunc, alternativeFunc func() T) T {
	if condition {
		return consequentFunc()
	} else {
		return alternativeFunc()
	}
}
