package protocol

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	proxyYAMLFieldsOnce sync.Once
	proxyYAMLFields     map[string]struct{}
)

// UnmarshalYAML decodes modeled proxy fields and retains every unknown
// top-level Clash attribute for a subsequent Clash or Mihomo export.
func (p *Proxy) UnmarshalYAML(value *yaml.Node) error {
	type proxyAlias Proxy
	var decoded proxyAlias
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	extra := cloneYAMLNode(value)
	extra.Content = extra.Content[:0]
	if value.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(value.Content); index += 2 {
			key := value.Content[index]
			if _, known := knownProxyYAMLFields()[key.Value]; known {
				continue
			}
			clonedKey := cloneYAMLNode(key)
			clonedValue := cloneYAMLNode(value.Content[index+1])
			extra.Content = append(extra.Content, &clonedKey, &clonedValue)
		}
	}
	decoded.ClashExtra = extra
	*p = Proxy(decoded)
	return nil
}

// MarshalYAML serializes modeled fields first, then reapplies retained
// attributes that are still unknown to the current Proxy model.
func (p Proxy) MarshalYAML() (any, error) {
	type proxyAlias Proxy
	encoded, err := yaml.Marshal(proxyAlias(p))
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
			if _, known := knownProxyYAMLFields()[key.Value]; known {
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
	proxy.ClashExtra = cloneYAMLNode(mapping)
	return nil
}

// knownProxyYAMLFields builds the current model's YAML key set once, so newly
// modeled fields automatically stop being treated as retained extras.
func knownProxyYAMLFields() map[string]struct{} {
	proxyYAMLFieldsOnce.Do(func() {
		proxyYAMLFields = make(map[string]struct{})
		proxyType := reflect.TypeOf(Proxy{})
		for index := 0; index < proxyType.NumField(); index++ {
			tag := strings.Split(proxyType.Field(index).Tag.Get("yaml"), ",")[0]
			if tag != "" && tag != "-" {
				proxyYAMLFields[tag] = struct{}{}
			}
		}
	})
	return proxyYAMLFields
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
