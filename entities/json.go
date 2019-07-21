package entities

import (
	json2 "encoding/json"
	xml2 "encoding/xml"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

func (m *Media) UnmarshalJSON(j []byte) error {
	var rawStrings map[string]string

	err := json2.Unmarshal(j, &rawStrings)
	if err != nil {
		return err
	}

	for k, v := range rawStrings {
		if strings.ToLower(k) == "path" {
			u, err := url.Parse(v)
			if err != nil {
				return err
			}
			m.Path = u
		}
	}

	return nil
}

//var Validator *schema.Validator

func FromXml(xml string, value interface{}) error {
	if true {
		arr := []byte(xml)
		err := xml2.Unmarshal(arr, &value)
		return err
	}
	return nil
}

func ToXml(value interface{}) (string, error) {
	xml, err := xml2.Marshal(&value)
	if err == nil {
		return string(xml[:]), nil
	} else {
		return "", err
	}
}

func FromJson(json string, value interface{}) error {
	arr := []byte(json)
	return json2.Unmarshal(arr, &value)
}

func ToJson(value interface{}) (string, error) {
	json, err := json2.Marshal(&value)
	if err == nil {
		return string(json[:]), nil
	} else {
		return "", err
	}
}

func failOnDeserialize(err error, payload []byte) {
	if err != nil {
		log.WithFields(log.Fields{
			"payload": string(payload),
			"error":   err,
			"help":    "Try purging the invalid messages (they have not been ack'ed) and try again.",
		}).Fatal("Could not deserialize message.")
	}
}
