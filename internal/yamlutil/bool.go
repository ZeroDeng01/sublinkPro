package yamlutil

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var boolFields sync.Map // map[reflect.Type]map[string]bool; entries are immutable

// DecodeWithStringBools also accepts the strings "true" and "false" for bool
// fields in a flat struct. Other values retain yaml.v3's normal validation.
// Normalization copies only changed nodes, leaving the input and untyped data intact.
func DecodeWithStringBools(value *yaml.Node, target any) error {
	typ := reflect.TypeOf(target)
	if typ == nil || typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return value.Decode(target)
	}
	typ = typ.Elem()
	fields, ok := boolFields.Load(typ)
	if !ok {
		keys := make(map[string]bool)
		for i := range typ.NumField() {
			field := typ.Field(i)
			if field.PkgPath != "" || field.Type.Kind() != reflect.Bool {
				continue
			}
			key, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
			if key == "" {
				key = strings.ToLower(field.Name)
			}
			if key != "-" {
				keys[key] = true
			}
		}
		fields, _ = boolFields.LoadOrStore(typ, keys)
	}
	keys, ok := fields.(map[string]bool)
	if !ok {
		return fmt.Errorf("invalid cached YAML boolean fields for %v", typ)
	}
	return normalizeBoolFields(value, keys, make(map[*yaml.Node]*yaml.Node)).Decode(target)
}

// Only visit struct mappings and their YAML merges, never arbitrary nested maps.
func normalizeBoolFields(value *yaml.Node, fields map[string]bool, memo map[*yaml.Node]*yaml.Node) *yaml.Node {
	if value == nil {
		return nil
	}
	if cached, ok := memo[value]; ok {
		return cached
	}
	// Preserve cyclic references for the decoder to reject, and visit shared
	// merge sources only once rather than expanding an alias graph repeatedly.
	memo[value] = value
	if value.Kind == yaml.AliasNode {
		alias := normalizeBoolFields(value.Alias, fields, memo)
		if alias != value.Alias {
			copy := *value
			copy.Alias = alias
			memo[value] = &copy
			return &copy
		}
		return value
	}
	var normalized *yaml.Node
	replace := func(index int, replacement *yaml.Node) {
		if replacement == value.Content[index] {
			return
		}
		if normalized == nil {
			copy := *value
			copy.Content = append([]*yaml.Node(nil), value.Content...)
			normalized = &copy
		}
		normalized.Content[index] = replacement
	}
	switch value.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(value.Content); i += 2 {
			key, child := value.Content[i], value.Content[i+1]
			if key.Tag == "!!merge" {
				replace(i+1, normalizeBoolFields(child, fields, memo))
			} else if fields[key.Value] {
				scalar := child
				if scalar.Kind == yaml.AliasNode {
					scalar = scalar.Alias
				}
				if scalar != nil && scalar.Kind == yaml.ScalarNode && scalar.Tag == "!!str" && (scalar.Value == "true" || scalar.Value == "false") {
					copy := *scalar
					copy.Tag = "!!bool"
					copy.Style = 0
					replace(i+1, &copy)
				}
			}
		}
	case yaml.SequenceNode: // A merge may contain an ordered list of mappings.
		for i, child := range value.Content {
			replace(i, normalizeBoolFields(child, fields, memo))
		}
	}
	if normalized != nil {
		memo[value] = normalized
		return normalized
	}
	return value
}
