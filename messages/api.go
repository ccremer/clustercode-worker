package messages

import json2 "encoding/json"

type PreTask struct {
    JobId     string
    File      string
    Priority  int
    SliceSize int    `json:"slice_size"`
    FileHash  string `json:"file_hash"`
    Args      []string
}

func (value *PreTask) FromJson(json string) error {
    arr := []byte(json)
    err := json2.Unmarshal(arr, value)
    return err
}

func (value *PreTask) ToJson() (string, error) {
    json, err := json2.Marshal(&value)
    if err == nil {
        return string(json[:]), nil
    } else {
        return "", err
    }
}
