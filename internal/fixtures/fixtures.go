// Package fixtures defines test/fixtures-manifest.json, which
// tools/extract_fixtures writes and the e2e harness (internal/e2e) reads.
package fixtures

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vega/ts-json-schema-generator-go/internal/config"
)

// Entry is one fixture invocation in the manifest.
type Entry struct {
	Name       string         `json:"name"`
	Types      []string       `json:"types,omitempty"`
	Config     map[string]any `json:"config,omitempty"`
	MainTsOnly bool           `json:"mainTsOnly,omitempty"`
	Skip       string         `json:"skip,omitempty"`
}

// harnessKeys are config.Config options that the harness derives from the
// entry itself, so a fixture config cannot set them.
var harnessKeys = map[string]bool{"path": true, "type": true}

// configFields maps each fixture config key (the JSON name of a config.Config
// field) to the field's type.
var configFields = func() map[string]reflect.Type {
	fields := map[string]reflect.Type{}
	configType := reflect.TypeFor[config.Config]()
	for i := range configType.NumField() {
		field := configType.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" || harnessKeys[name] {
			continue
		}
		fields[name] = field.Type
	}
	return fields
}()

// ValidateConfig reports an error if cfg has a key that is not a
// config.Config option, or a value of the wrong kind for its option.
func ValidateConfig(cfg map[string]any) error {
	for key, val := range cfg {
		fieldType, known := configFields[key]
		if !known {
			return fmt.Errorf("unsupported config key %q", key)
		}
		switch {
		case fieldType.Kind() == reflect.String:
			if _, ok := val.(string); !ok {
				return fmt.Errorf("config key %q is not a string", key)
			}
		case fieldType.Kind() == reflect.Bool:
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("config key %q is not a boolean", key)
			}
		case fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.String:
			list, ok := val.([]any)
			if !ok {
				return fmt.Errorf("config key %q is not an array", key)
			}
			for _, item := range list {
				if _, ok := item.(string); !ok {
					return fmt.Errorf("config key %q contains non-string %v", key, item)
				}
			}
		default:
			return fmt.Errorf("config key %q has unsupported type %s", key, fieldType)
		}
	}
	return nil
}
