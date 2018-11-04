package messaging

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestDeserializePreTask(t *testing.T) {

    json := string(`
        {
            "file": "0/path/to/file.ext",
            "args": [
                "arg1",
                "arg with space"
            ],
            "job_id": "620b8251-52a1-4ecd-8adc-4fb280214bba",
            "file_hash": "b8934ef001960cafc224be9f1e1ca82c",
            "priority": 1,
            "slice_size": 120
        }
        `)
    expected := PreTask{
        Args:      []string{"arg1", "arg with space"},
        File:      "0/path/to/file.ext",
        JobID:     "620b8251-52a1-4ecd-8adc-4fb280214bba",
        Priority:  1,
        SliceSize: 120,
        FileHash:  "b8934ef001960cafc224be9f1e1ca82c",
    }

    result := PreTask{}
    FromJson(json, &result)

    assert.Equal(t, expected, result)
}
