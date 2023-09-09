package wsdl

import (
	"os"
	"path/filepath"
	"testing"
)

func glob(pat string) []string {
	s, err := filepath.Glob(pat)
	if err != nil {
		panic(err)
	}
	return s
}

func TestParse(t *testing.T) {
	for _, filename := range glob("testdata/*.wsdl") {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Error(err)
			continue
		}
		def, err := Parse(data)
		if err != nil {
			t.Errorf("parse %s: %s", filename, err)
		}
		t.Logf("\n%s", def)
	}
}
