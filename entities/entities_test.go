package entities

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/url"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

var serializationTests = []struct {
	name     string
	expected interface{}
	result   interface{}
	testFile string
}{
	{
		"SliceAddedEvent_WithArgs",
		&SliceAddedEvent{
			Args:    []string{"arg1", "arg with space"},
			JobID:   "620b8251-52a1-4ecd-8adc-4fb280214bba",
			SliceNr: 34,
		},
		&SliceAddedEvent{},
		"slice_added_event_1.xml",
	},
	{
		"SliceAddedEvent_WithoutArgs",
		&SliceAddedEvent{
			JobID:   "620b8251-52a1-4ecd-8adc-4fb280214bba",
			SliceNr: 34,
		},
		&SliceAddedEvent{},
		"slice_added_event_2.xml",
	},
	{
		"SliceCompletedEvent_WithStreams_OneLine",
		&SliceCompletedEvent{
			JobID: "620b8251-52a1-4ecd-8adc-4fb280214bba",
			StdStreams: []StdStream{
				{FD: StdOutFileDescriptor, Line: "This is from stdout"},
			},
		},
		&SliceCompletedEvent{},
		"slice_completed_event_1.xml",
	},
	{
		"SliceCompletedEvent_WithStreams_MultipleLines",
		&SliceCompletedEvent{
			JobID: "620b8251-52a1-4ecd-8adc-4fb280214bba",
			StdStreams: []StdStream{
				{FD: StdErrFileDescriptor, Line: "This is from stderr"},
				{FD: StdOutFileDescriptor, Line: "This is from stdout"},
			},
		},
		&SliceCompletedEvent{},
		"slice_completed_event_2.xml",
	},
	{
		"SliceCompletedEvent_WithoutStreams",
		&SliceCompletedEvent{
			JobID: "620b8251-52a1-4ecd-8adc-4fb280214bba",
		},
		&SliceCompletedEvent{},
		"slice_completed_event_3.xml",
	},
}

func TestMedia_Priority_ShouldReturnPort(t *testing.T) {
	uri, err := url.Parse("clustercode://base_dir:12/path")
	assert.NoError(t, err)
	subject := Media{Path: uri,}

	result := subject.Priority()
	assert.Equal(t, 12, result)
}

func TestMedia_GetSubstitutedPath_ShouldReplaceBasePath_Relative(t *testing.T) {
	uri, err := url.Parse("clustercode://base_dir:12/path")
	assert.NoError(t, err)
	subject := Media{Path: uri,}

	result := subject.GetSubstitutedPath("replacement")
	assert.Equal(t, "replacement/12/path", result)
}

func TestMedia_GetSubstitutedPath_ShouldReplaceBasePath_Absolute(t *testing.T) {
	uri, err := url.Parse("clustercode://base_dir:12/path/another")
	assert.NoError(t, err)
	subject := Media{Path: uri,}

	result := subject.GetSubstitutedPath("/replacement")
	assert.Equal(t, "/replacement/12/path/another", result)
}

func updateGoldenFileIfNecessary(t *testing.T, content string, path string) {
	if *update {
		t.Log("update golden file")
		if err := ioutil.WriteFile(path, []byte(content+"\n"), 0644); err != nil {
			t.Fatalf("failed to update golden file: %s", err)
		}
	}
}
