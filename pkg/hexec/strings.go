package hexec

import "strings"

func SplitEnvVar(v string) (string, string) {
	name, value, _ := strings.Cut(v, "=")
	return name, value
}
