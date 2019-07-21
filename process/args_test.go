package process

import (
	"github.com/ccremer/clustercode-worker/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

var cfg = config.CreateDefaultConfig()

func TestReplaceFields_ShouldReplaceInputDir(t *testing.T) {
	p := Process{config: &cfg,}
	expected := []string{"/input/0/movie.mp4", "-hide_banner"}
	args := []string{"${INPUT}/0/movie.mp4", "-hide_banner"}

	result := p.replaceFields(args, map[string]string{
		"${INPUT}":  p.config.Input.Dir,
		"${OUTPUT}": p.config.Output.Dir,
		"${TMP}":    p.config.Output.TmpDir,
	})

	assert.Equal(t, expected, result)
}

func TestReplaceFields_ShouldReplaceOutputDir(t *testing.T) {
	p := Process{config: &cfg,}
	expected := []string{"/output/0/movie.mp4", "-hide_banner"}
	args := []string{"${OUTPUT}/0/movie.mp4", "-hide_banner"}

	result := p.replaceFields(args, map[string]string{
		"${INPUT}":  p.config.Input.Dir,
		"${OUTPUT}": p.config.Output.Dir,
		"${TMP}":    p.config.Output.TmpDir,
	})

	assert.Equal(t, expected, result)
}

func TestReplaceFields_ShouldReplaceAllVariables(t *testing.T) {
	p := Process{config: &cfg,}
	expected := []string{"/input/output/movie.mp4", "/output"}
	args := []string{"${INPUT}${OUTPUT}/movie.mp4", "${OUTPUT}"}

	result := p.replaceFields(args, map[string]string{
		"${INPUT}":  p.config.Input.Dir,
		"${OUTPUT}": p.config.Output.Dir,
		"${TMP}":    p.config.Output.TmpDir,
	})

	assert.Equal(t, expected, result)
}
