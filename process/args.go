package process

import (
	"github.com/thoas/go-funk"
	"strings"
)

func (p *Process) replaceFields(args []string) []string {

	fields := map[string]string{
		"${input_dir}":  p.config.Input.Dir,
		"${output_dir}": p.config.Output.Dir,
		"${tmp_dir}":    p.config.Output.TmpDir,
	}

	return funk.Map(args, func(arg string) string {
		tmp := arg
		funk.ForEach(fields, func(k string, v string) {
			tmp = strings.Replace(tmp, k, v, -1)
		})
		return tmp
	}).([]string)
}
