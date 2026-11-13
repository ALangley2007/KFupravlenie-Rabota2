package config

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const requiredOccurrences = 1

type xmlValue struct {
	XMLName    xml.Name
	Text       string     `xml:",chardata"`
	Attributes []xml.Attr `xml:",any,attr"`
	Children   []xmlValue `xml:",any"`
}

type xmlConfig struct {
	XMLName    xml.Name   `xml:"config"`
	Package    []xmlValue `xml:"package"`
	Repository []xmlValue `xml:"repository"`
	TestMode   []xmlValue `xml:"test_mode"`
	Filter     []xmlValue `xml:"filter"`
	Unknown    []xmlValue `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
	Text       string     `xml:",chardata"`
}

func decode(reader io.Reader) (map[string]string, error) {
	decoder := xml.NewDecoder(reader)
	root, err := nextElement(decoder)
	if err != nil {
		return nil, err
	}
	var document xmlConfig
	if err := decoder.DecodeElement(&document, &root); err != nil {
		return nil, err
	}
	if _, err := nextElement(decoder); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("допускается только один корневой элемент")
	}
	return document.values()
}

func nextElement(decoder *xml.Decoder) (xml.StartElement, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return xml.StartElement{}, err
		}
		switch value := token.(type) {
		case xml.StartElement:
			return value, nil
		case xml.Comment, xml.ProcInst:
			continue
		case xml.CharData:
			if strings.TrimSpace(string(value)) == "" {
				continue
			}
		}
		return xml.StartElement{}, fmt.Errorf("лишние данные вне корневого элемента")
	}
}

func (document xmlConfig) values() (map[string]string, error) {
	if document.XMLName.Space != "" || document.Attributes != nil {
		return nil, fmt.Errorf("config: атрибуты и пространства имён запрещены")
	}
	if document.Unknown != nil || strings.TrimSpace(document.Text) != "" {
		return nil, fmt.Errorf("config: неизвестный параметр или лишний текст")
	}
	fields := []struct {
		name     string
		elements []xmlValue
	}{
		{"package", document.Package},
		{"repository", document.Repository},
		{"test_mode", document.TestMode},
		{"filter", document.Filter},
	}
	values := make(map[string]string)
	for _, field := range fields {
		value, err := fieldValue(field.name, field.elements)
		if err != nil {
			return nil, err
		}
		values[field.name] = value
	}
	return values, nil
}

func fieldValue(name string, elements []xmlValue) (string, error) {
	if len(elements) != requiredOccurrences {
		return "", fmt.Errorf("%s: требуется ровно одно вхождение", name)
	}
	value := elements[0]
	if value.XMLName.Space != "" || value.Attributes != nil || value.Children != nil {
		return "", fmt.Errorf("%s: нужен текст без атрибутов и вложенных тегов", name)
	}
	return value.Text, nil
}
