package env

import (
	"fmt"
	"os"
	"strconv"
)

// 以字符串形式读取指定环境变量的值。
//
// 入参：
//   - envVar：环境变量。
//
// 出参：
//   - 环境变量值。
func GetString(envVar string) string {
	return GetOrDefaultString(envVar, "")
}

// 以字符串形式读取指定环境变量的值。
//
// 入参：
//   - envVar：环境变量。
//   - defaultValue: 默认值。
//
// 出参：
//   - 环境变量值。如果指定环境变量不存在、或者值为零值，则返回默认值。
func GetOrDefaultString(envVar, defaultValue string) string {
	return getOrDefault(envVar, defaultValue, parseString)
}

// 以整数形式读取指定环境变量的值。
//
// 入参：
//   - envVar：环境变量。
//
// 出参：
//   - 环境变量值。
func GetInt(envVar string) int {
	return GetOrDefaultInt(envVar, 0)
}

// 以整数形式读取指定环境变量的值。
//
// 入参：
//   - envVar：环境变量。
//   - defaultValue: 默认值。
//
// 出参：
//   - 环境变量值。如果指定环境变量不存在、或者值的类型不是整数，则返回默认值。
func GetOrDefaultInt(envVar string, defaultValue int) int {
	return getOrDefault(envVar, defaultValue, strconv.Atoi)
}

// 以布尔形式读取指定环境变量的值。
//
// 入参：
//   - envVar：环境变量。
//
// 出参：
//   - 环境变量值。
func GetBool(envVar string) bool {
	return GetOrDefaultBool(envVar, false)
}

// 以布尔形式读取指定环境变量的值。
//
// 入参：
//   - envVar：环境变量。
//   - defaultValue: 默认值。
//
// 出参：
//   - 环境变量值。如果指定环境变量不存在、或者值的类型不是布尔，则返回默认值。
func GetOrDefaultBool(envVar string, defaultValue bool) bool {
	return getOrDefault(envVar, defaultValue, strconv.ParseBool)
}

func getOrDefault[T any](envVar string, defaultValue T, parser func(string) (T, error)) T {
	v, err := parser(os.Getenv(envVar))
	if err != nil {
		return defaultValue
	}

	return v
}

func parseString(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("empty string")
	}

	return s, nil
}
