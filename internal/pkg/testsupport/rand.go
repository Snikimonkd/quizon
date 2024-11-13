// nolint:gosec
package testsupport

import "math/rand"

// RandInt64 - генерирует случайное число для тестов
func RandInt64() int64 {
	return rand.Int63()
}
