// numbers.go 提供 library 包内简单数值边界工具。
package library

// maxInt 返回两个整数中的较大值。
func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

// minInt 返回两个整数中的较小值。
func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
