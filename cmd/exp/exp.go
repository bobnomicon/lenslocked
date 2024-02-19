package main

import (
	"fmt"
	"strings"
)

func Join(vals ...string) string {
	var sb strings.Builder
	for i, s := range vals {
		sb.WriteString(s)
		if i < len(vals) - 1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}

func main() {
	strings := []string{"the", "quick", "brown", "fox"}
	fmt.Println(Join(strings...))
}