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
	args := []string{"${input_dir}/0/movie.mp4", "-hide_banner"}

	result := p.replaceFields(args)

	assert.Equal(t, expected, result)
}

func TestReplaceFields_ShouldReplaceOutputDir(t *testing.T) {
	p := Process{config: &cfg,}
	expected := []string{"/output/0/movie.mp4", "-hide_banner"}
	args := []string{"${output_dir}/0/movie.mp4", "-hide_banner"}

	result := p.replaceFields(args)

	assert.Equal(t, expected, result)
}

func TestReplaceFields_ShouldReplaceAllVariables(t *testing.T) {
	p := Process{config: &cfg,}
	expected := []string{"/input/output/movie.mp4", "/output"}
	args := []string{"${input_dir}${output_dir}/movie.mp4", "${output_dir}"}

	result := p.replaceFields(args)

	assert.Equal(t, expected, result)
}
