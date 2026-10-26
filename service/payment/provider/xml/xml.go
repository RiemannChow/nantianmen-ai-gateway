// Package xml provides minimal flat-map XML marshalling/unmarshalling for the
// WeChat Pay APIv2 text protocol (no namespaces, attributes, or nesting).
package xml

import (
	"encoding/xml"
	"strings"
)

type flat struct {
	XMLName xml.Name
	Values  []kv `xml:",any"`
}

type kv struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// MarshalMap renders a flat key/value map as <root><k>v</k>...</root>.
func MarshalMap(root string, m map[string]string) []byte {
	var b strings.Builder
	b.WriteString("<" + root + ">")
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for _, k := range keys {
		b.WriteString("<" + k + ">")
		b.WriteString(escape(m[k]))
		b.WriteString("</" + k + ">")
	}
	b.WriteString("</" + root + ">")
	return []byte(b.String())
}

// ParseMap decodes a flat <root><k>v</k></root> document into a map.
func ParseMap(data []byte) (map[string]string, error) {
	var f flat
	if err := xml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(f.Values))
	for _, item := range f.Values {
		out[item.XMLName.Local] = item.Value
	}
	return out, nil
}

func escape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;")
	return r.Replace(s)
}
