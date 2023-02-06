package config

import (
	"strings"

	"github.com/luisdavim/synthetic-checker/pkg/template"
)

// TemplatedString is a custom string type that allows setting its value as a go text template.
// The template is automatically rendered when unmarsheling fields of this type.
// see the template package for more details.
type TemplatedString string

func (t TemplatedString) String() string {
	return string(t)
}

// UnmarshalJSON automatically renders the template set on the TemplatedString value
func (t *TemplatedString) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	buf := new(strings.Builder)
	err := template.Render(buf, string(data[1:len(data)-1]), nil)
	*t = TemplatedString(buf.String())

	return err
}
