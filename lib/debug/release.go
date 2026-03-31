//go:build !debug

package debug

const Debug = false

func PrintLog(format string, a ...interface{}) {}