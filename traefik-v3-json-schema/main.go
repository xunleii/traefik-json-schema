package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"reflect"

	huma "github.com/danielgtaylor/huma/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/traefik/traefik/v3/pkg/config/static"
)

func main() {
	schema := JsonSchema{
		ID:    "https://json.schemastore.org/traefik-v3.json",
		Title: "Traefik v3 Static Configuration",
	}

	schema.Definitions = huma.NewMapRegistry("#/$defs/", func(t reflect.Type, hint string) string {
		name := huma.DefaultSchemaNamer(t, hint)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if t.PkgPath() == "" {
			return name
		}
		name = fmt.Sprintf("%s%s", path.Base(t.PkgPath()), name)
		return name
	})
	schema.Schema = huma.SchemaFromType(schema.Definitions, reflect.TypeOf(static.Configuration{}))
	json, err := json.Marshal(&schema)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(json))
}

type JsonSchema struct {
	ID          string        `yaml:"$id"`
	Title       string        `yaml:"title"`
	Schema      *huma.Schema  `yaml:",squash"`
	Definitions huma.Registry `yaml:"$defs"`
}

func (j *JsonSchema) MarshalJSON() ([]byte, error) {
	schema := map[string]interface{}{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "yaml", Result: &schema})
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(j)
	if err != nil {
		return nil, err
	}

	// NOTE: in order to have a JSON Schema that is human-readable, we need a custom ordering of root fields
	orderedFields := []string{"$id", "$schema", "title", "properties", "type", "$defs"}
	buf := bytes.NewBufferString("{")
	buf.WriteString("\"$schema\": \"http://json-schema.org/draft-07/schema#\"")
	for _, field := range orderedFields {
		if _, ok := schema[field]; ok {
			buf.WriteString(",")

			raw, err := json.Marshal(schema[field])
			if err != nil {
				return nil, err
			}
			buf.WriteString(fmt.Sprintf("\"%s\": %s", field, string(raw)))
		}
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}
