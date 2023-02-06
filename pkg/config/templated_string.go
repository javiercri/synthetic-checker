package config

import (
	"strings"

	"github.com/luisdavim/synthetic-checker/pkg/template"
)

type TemplatedString string

func (t TemplatedString) String() string {
	return string(t)
}

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
