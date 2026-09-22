package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

func TestMVPAllSchemasAndExamples(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(testRepoPath("config/schemas"), "*.schema.json"))
	if err != nil || len(paths) < 13 {
		t.Fatal("missing schemas")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			compiler := jsonschema.NewCompiler()
			compiler.AssertFormat()
			if err := compiler.AddResource("https://empresa-security.local/schema.json", doc); err != nil {
				t.Fatal(err)
			}
			if _, err := compiler.Compile("https://empresa-security.local/schema.json"); err != nil {
				t.Fatal(err)
			}
		})
	}
	examples, err := filepath.Glob(filepath.Join(testRepoPath("config/examples"), "*.json"))
	if err != nil || len(examples) == 0 {
		t.Fatal("missing examples")
	}
	for _, path := range examples {
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		for _, suffix := range []string{"-p1", "-p2", "-check"} {
			name = strings.TrimSuffix(name, suffix)
		}
		if err := validateJSONSchemaFile(name+".schema.json", path); err != nil {
			t.Fatal(err)
		}
	}
}
