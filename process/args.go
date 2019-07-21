package process

import (
	"github.com/thoas/go-funk"
	"strings"
)

func (p *Process) replaceFields(args []string, replacements map[string]string) []string {

	return funk.Map(args, func(arg string) string {
		tmp := arg
		funk.ForEach(replacements, func(k string, v string) {
			tmp = strings.ReplaceAll(tmp, k, v)
		})
		return tmp
	}).([]string)
}
