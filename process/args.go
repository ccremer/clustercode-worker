package process

import (
    "github.com/micro/go-config"
    "github.com/thoas/go-funk"
    "strings"
)

func replaceFields(args []string) ([]string) {

    fields := map[string]string {
        "${input_dir}": config.Get("input", "dir").String("/input"),
        "${output_dir}": config.Get("output", "dir").String("/output"),
        "${tmp_dir}": config.Get("output", "tmpdir").String("/var/tmp/clustercode"),
    }

    return funk.Map(args, func(arg string) string {
        tmp := arg
        funk.ForEach(fields, func(k string, v string) {
            tmp = strings.Replace(tmp, k , v, -1)
        })
        return tmp
    }).([]string)
}
