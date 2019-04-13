package schema

import (
	"errors"
	"fmt"
	"github.com/jbussdieker/golibxml"
	"github.com/krolaw/xsd"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"unsafe"
)

type (
	Validator struct {
		schema *xsd.Schema
	}
)

func NewXmlValidator(pathToXsd string) *Validator {
	v := &Validator{}
	v.LoadXmlSchema(pathToXsd)
	return v
}

func (v *Validator) LoadXmlSchema(pathToXsd string) {
	log.WithField("path", pathToXsd).Debug("Loading schema")
	xsdfile, err := ioutil.ReadFile(pathToXsd)
	if err != nil {
		log.Fatal(err)
	}
	xsdSchema, err := xsd.ParseSchema(xsdfile)
	if err != nil {
		log.Fatal(err)
	}
	v.schema = xsdSchema
}

func (v *Validator) ValidateXml(xml *string) (bool, error) {
	if v.schema == nil {
		log.Fatal("schema is not loaded")
	}
	doc := golibxml.ParseDoc(*xml)
	if doc == nil {
		return false, errors.New("provided XML string does not seem to be valid XML")
	}
	defer doc.Free()

	// golibxml._Ctype_xmlDocPtr can't be cast to xsd.DocPtr, even though they are both
	// essentially _Ctype_xmlDocPtr.  Using unsafe gets around this.
	if err := v.schema.Validate(xsd.DocPtr(unsafe.Pointer(doc.Ptr))); err != nil {
		return false, errors.New(fmt.Sprintln(err))
	} else {
		return true, nil
	}
}
