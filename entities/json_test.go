package entities

import (
	"github.com/ccremer/clustercode-worker/schema"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestSerializeXml(t *testing.T) {
	for _, tt := range serializationTests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.testFile)

			// serialize
			xmlString, err := ToXml(tt.expected)
			assert.NoError(t, err)

			updateGoldenFileIfNecessary(t, xmlString, path)

			// verify from existing file
			xmlBytes, err := ioutil.ReadFile(path)
			assert.NoError(t, err)
			assert.Equal(t, string(xmlBytes), xmlString+"\n")
		})
	}
}

func TestDeserializeXml(t *testing.T) {
	Validator = schema.NewXmlValidator("../schema/clustercode_v1.xsd")
	for _, tt := range serializationTests {
		t.Run(tt.name, func(t *testing.T) {

			// get XML
			path := filepath.Join("testdata", tt.testFile)
			rawXmlBytes, ioErr := ioutil.ReadFile(path)
			assert.NoError(t, ioErr)
			xml := string(rawXmlBytes)

			// deserialize
			xmlErr := FromXml(xml, tt.result)
			assert.NoError(t, xmlErr)

			// verify
			assert.Equal(t, tt.expected, tt.result)
		})
	}
}
