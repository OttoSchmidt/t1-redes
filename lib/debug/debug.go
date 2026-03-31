//go:build debug

package debug

import "fmt"

const Debug = true

func PrintLog(format string, a ...interface{}) {
    fmt.Printf("[DEBUG] " + format, a...)
}