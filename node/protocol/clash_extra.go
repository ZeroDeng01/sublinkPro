package protocol

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"sublink/internal/yamlutil"

	"gopkg.in/yaml.v3"
)

var (
	proxyYAMLFieldsOnce sync.Once
	proxyYAMLFields     map[string]struct{}
)

type proxyYAML Proxy

// proxyYAMLCodec lets a registered protocol own additional Clash fields.
type proxyYAMLCodec interface {
	decodeProxyYAML(*yaml.Node, *Proxy) error
	encodeProxyYAML(Proxy) any
	proxyYAMLFieldNames() map[string]struct{}
}

// UnmarshalYAML decodes modeled proxy fields and retains every unknown
// top-level Clash attribute for a subsequent Clash or Mihomo export.
func (p *Proxy) UnmarshalYAML(value *yaml.Node) error {
	var common proxyYAML
	if err := yamlutil.DecodeWithStringBools(value, &common); err != nil {
		return err
	}
	decoded := Proxy(common)
	var protocolFields map[string]struct{}
	if codec, ok := getProtocolByName(decoded.Type).(proxyYAMLCodec); ok {
		if err := codec.decodeProxyYAML(value, &decoded); err != nil {
			return err
		}
		protocolFields = codec.proxyYAMLFieldNames()
	}

	extra := cloneYAMLNode(value)
	extra.Content = extra.Content[:0]
	if value.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(value.Content); index += 2 {
			key := value.Content[index]
			if isKnownProxyYAMLField(key.Value, protocolFields) {
				continue
			}
			clonedKey := cloneYAMLNode(key)
			clonedValue := cloneYAMLNode(value.Content[index+1])
			extra.Content = append(extra.Content, &clonedKey, &clonedValue)
		}
	}
	decoded.ClashExtra = extra
	*p = decoded
	return nil
}

// MarshalYAML serializes modeled fields first, then reapplies retained
// attributes that are still unknown to the current Proxy model.
func (p Proxy) MarshalYAML() (any, error) {
	var fields any = proxyYAML(p)
	var protocolFields map[string]struct{}
	if codec, ok := getProtocolByName(p.Type).(proxyYAMLCodec); ok {
		fields = codec.encodeProxyYAML(p)
		protocolFields = codec.proxyYAMLFieldNames()
	}
	encoded, err := yaml.Marshal(fields)
	if err != nil {
		return nil, err
	}

	var document yaml.Node
	if err := yaml.Unmarshal(encoded, &document); err != nil {
		return nil, err
	}
	mapping := yamlMappingNode(&document)
	if mapping == nil {
		return nil, fmt.Errorf("proxy YAML did not encode as a mapping")
	}
	merged := cloneYAMLNode(mapping)
	if extra := yamlMappingNode(&p.ClashExtra); extra != nil {
		for index := 0; index+1 < len(extra.Content); index += 2 {
			key := extra.Content[index]
			if isKnownProxyYAMLField(key.Value, protocolFields) {
				continue
			}
			clonedKey := cloneYAMLNode(key)
			clonedValue := cloneYAMLNode(extra.Content[index+1])
			merged.Content = append(merged.Content, &clonedKey, &clonedValue)
		}
	}
	return merged, nil
}

// EncodeClashExtra converts a proxy's retained attributes into the text stored
// with a node. Empty retained attributes require no persistent payload.
func EncodeClashExtra(proxy Proxy) (string, error) {
	mapping := yamlMappingNode(&proxy.ClashExtra)
	if mapping == nil || len(mapping.Content) == 0 {
		return "", nil
	}
	encoded, err := yaml.Marshal(mapping)
	if err != nil {
		return "", fmt.Errorf("marshal retained Clash attributes: %w", err)
	}
	return string(encoded), nil
}

// ApplyClashExtra restores persisted unmodeled attributes onto a proxy for a
// Clash or Mihomo output. Invalid persisted data is returned to the caller.
func ApplyClashExtra(proxy *Proxy, encoded string) error {
	if proxy == nil || strings.TrimSpace(encoded) == "" {
		return nil
	}
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(encoded), &document); err != nil {
		return fmt.Errorf("parse retained Clash attributes: %w", err)
	}
	mapping := yamlMappingNode(&document)
	if mapping == nil {
		return fmt.Errorf("retained Clash attributes must be a mapping")
	}
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		k := mapping.Content[index].Value
		v := mapping.Content[index+1].Value
		if k == "hop-interval" || k == "hop_interval" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				proxy.HopInterval = n
			}
		} else if k == "ports" && v != "" {
			proxy.Ports = v
		}
	}
	proxy.ClashExtra = cloneYAMLNode(mapping)
	return nil
}

// Modeled fields take precedence over retained attributes.
func isKnownProxyYAMLField(key string, protocolFields map[string]struct{}) bool {
	proxyYAMLFieldsOnce.Do(func() {
		proxyYAMLFields = yamlFieldNames(reflect.TypeFor[Proxy]())
	})
	if _, known := proxyYAMLFields[key]; known {
		return true
	}
	_, known := protocolFields[key]
	return known
}

func yamlFieldNames(structType reflect.Type) map[string]struct{} {
	fields := make(map[string]struct{}, structType.NumField())
	for index := 0; index < structType.NumField(); index++ {
		tag, _, _ := strings.Cut(structType.Field(index).Tag.Get("yaml"), ",")
		if tag != "" && tag != "-" {
			fields[tag] = struct{}{}
		}
	}
	return fields
}

// yamlMappingNode returns the mapping payload from a YAML document or mapping node.
func yamlMappingNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}
	return node
}

// cloneYAMLNode creates an independent YAML node tree before it is retained
// outside the source document or merged into a later output document.
func cloneYAMLNode(node *yaml.Node) yaml.Node {
	if node == nil {
		return yaml.Node{}
	}
	clone := *node
	if len(node.Content) > 0 {
		clone.Content = make([]*yaml.Node, len(node.Content))
		for index, child := range node.Content {
			clonedChild := cloneYAMLNode(child)
			clone.Content[index] = &clonedChild
		}
	}
	return clone
}
