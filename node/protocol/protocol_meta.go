package protocol

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type FieldMeta struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Group       string   `json:"group,omitempty"`
	Description string   `json:"description,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Options     []string `json:"options,omitempty"`
	Advanced    bool     `json:"advanced,omitempty"`
	Secret      bool     `json:"secret,omitempty"`
	Multiline   bool     `json:"multiline,omitempty"`
}

type ProtocolMeta struct {
	Name   string      `json:"name"`
	Label  string      `json:"label"`
	Color  string      `json:"color"`
	Icon   string      `json:"icon"`
	Fields []FieldMeta `json:"fields"`
}

type LinkIdentity struct {
	Protocol string
	Name     string
	Address  string
	Host     string
	Port     string
}

type Protocol interface {
	Name() string
	Aliases() []string
	Label() string
	Color() string
	Icon() string
	Prototype() interface{}
	Fields() []FieldMeta
	NameFieldPath() string
	DecodeLink(string) (interface{}, error)
	EncodeLink(interface{}) (string, error)
	ExtractIdentity(interface{}) (LinkIdentity, error)
}

type ProxyCapable interface {
	ToProxy(Urls, OutputConfig) (Proxy, error)
	CanHandleProxy(Proxy) bool
	FromProxy(Proxy) (string, error)
}

type SurgeCapable interface {
	ToSurgeLine(string, OutputConfig) (string, string, error)
}

type ProtocolSpec struct {
	name          string
	aliases       []string
	label         string
	color         string
	icon          string
	prototype     interface{}
	fields        []FieldMeta
	nameFieldPath string
	decode        func(string) (interface{}, error)
	encode        func(interface{}) (string, error)
	identity      func(interface{}) (LinkIdentity, error)
}

func (p *ProtocolSpec) Name() string {
	return p.name
}

func (p *ProtocolSpec) Aliases() []string {
	return append([]string(nil), p.aliases...)
}

func (p *ProtocolSpec) Label() string {
	return p.label
}

func (p *ProtocolSpec) Color() string {
	return p.color
}

func (p *ProtocolSpec) Icon() string {
	return p.icon
}

func (p *ProtocolSpec) Prototype() interface{} {
	return p.prototype
}

func (p *ProtocolSpec) Fields() []FieldMeta {
	return append([]FieldMeta(nil), p.fields...)
}

func (p *ProtocolSpec) NameFieldPath() string {
	return p.nameFieldPath
}

func (p *ProtocolSpec) DecodeLink(link string) (interface{}, error) {
	if p.decode == nil {
		return nil, fmt.Errorf("protocol %s does not support decoding", p.name)
	}
	return p.decode(link)
}

func (p *ProtocolSpec) EncodeLink(value interface{}) (string, error) {
	if p.encode == nil {
		return "", fmt.Errorf("protocol %s does not support encoding", p.name)
	}
	return p.encode(value)
}

func (p *ProtocolSpec) ExtractIdentity(value interface{}) (LinkIdentity, error) {
	if p.identity == nil {
		return LinkIdentity{}, fmt.Errorf("protocol %s does not provide identity extraction", p.name)
	}
	return p.identity(value)
}

type ProxyProtocolSpec struct {
	*ProtocolSpec
	toProxy       func(Urls, OutputConfig) (Proxy, error)
	canHandleProxy func(Proxy) bool
	fromProxy     func(Proxy) (string, error)
}

func (p *ProxyProtocolSpec) ToProxy(link Urls, config OutputConfig) (Proxy, error) {
	if p.toProxy == nil {
		return Proxy{}, fmt.Errorf("protocol %s does not support proxy export", p.name)
	}
	return p.toProxy(link, config)
}

func (p *ProxyProtocolSpec) CanHandleProxy(proxy Proxy) bool {
	if p.canHandleProxy == nil {
		return false
	}
	return p.canHandleProxy(proxy)
}

func (p *ProxyProtocolSpec) FromProxy(proxy Proxy) (string, error) {
	if p.fromProxy == nil {
		return "", fmt.Errorf("protocol %s does not support proxy import", p.name)
	}
	return p.fromProxy(proxy)
}

type ProxySurgeProtocolSpec struct {
	*ProxyProtocolSpec
	toSurgeLine func(string, OutputConfig) (string, string, error)
}

func (p *ProxySurgeProtocolSpec) ToSurgeLine(link string, config OutputConfig) (string, string, error) {
	if p.toSurgeLine == nil {
		return "", "", fmt.Errorf("protocol %s does not support Surge export", p.name)
	}
	return p.toSurgeLine(link, config)
}

type aliasMatcher struct {
	prefix   string
	protocol Protocol
}

var (
	registryMu           sync.RWMutex
	protocolsByName      = make(map[string]Protocol)
	protocolsByAlias     = make(map[string]Protocol)
	aliasMatchers        []aliasMatcher
	protocolList         []Protocol
	proxyCapables        []ProxyCapable
	protocolMetaCache    []ProtocolMeta
	protocolMetaDirty    = true
)

func normalizeProtocolName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func proxyTypeMatches(proxy Proxy, names ...string) bool {
	proxyType := normalizeProtocolName(proxy.Type)
	for _, name := range names {
		if proxyType == normalizeProtocolName(name) {
			return true
		}
	}
	return false
}

func normalizeAlias(alias string) string {
	alias = strings.ToLower(strings.TrimSpace(alias))
	if alias == "" {
		return ""
	}
	if !strings.Contains(alias, "://") {
		alias += "://"
	}
	return alias
}

func sortAliasMatchersLocked() {
	sort.SliceStable(aliasMatchers, func(i, j int) bool {
		return len(aliasMatchers[i].prefix) > len(aliasMatchers[j].prefix)
	})
}

func MustRegisterProtocol(protocol Protocol) {
	if protocol == nil {
		panic("cannot register nil protocol")
	}

	name := normalizeProtocolName(protocol.Name())
	if name == "" {
		panic("cannot register protocol with empty name")
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := protocolsByName[name]; exists {
		panic(fmt.Sprintf("protocol %s already registered", name))
	}

	protocolsByName[name] = protocol
	protocolList = append(protocolList, protocol)

	localAliases := map[string]struct{}{}
	aliases := append(protocol.Aliases(), name)
	for _, alias := range aliases {
		normalized := normalizeAlias(alias)
		if normalized == "" {
			continue
		}
		if _, seen := localAliases[normalized]; seen {
			continue
		}
		localAliases[normalized] = struct{}{}
		if existing, exists := protocolsByAlias[normalized]; exists {
			panic(fmt.Sprintf("alias %s already registered by protocol %s", normalized, existing.Name()))
		}
		protocolsByAlias[normalized] = protocol
		aliasMatchers = append(aliasMatchers, aliasMatcher{prefix: normalized, protocol: protocol})
	}
	sortAliasMatchersLocked()

	if proxyProtocol, ok := protocol.(ProxyCapable); ok {
		proxyCapables = append(proxyCapables, proxyProtocol)
	}

	protocolMetaDirty = true
}

func getProtocolByName(name string) Protocol {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return protocolsByName[normalizeProtocolName(name)]
}

func detectProtocol(link string) Protocol {
	registryMu.RLock()
	defer registryMu.RUnlock()

	linkLower := strings.ToLower(strings.TrimSpace(link))
	for _, matcher := range aliasMatchers {
		if strings.HasPrefix(linkLower, matcher.prefix) {
			return matcher.protocol
		}
	}
	return nil
}

func getProxyProtocol(proxy Proxy) ProxyCapable {
	registryMu.RLock()
	defer registryMu.RUnlock()
	for _, proxyCapable := range proxyCapables {
		if proxyCapable.CanHandleProxy(proxy) {
			return proxyCapable
		}
	}
	return nil
}

func rebuildProtocolMetaCacheLocked() {
	if !protocolMetaDirty {
		return
	}

	metas := make([]ProtocolMeta, 0, len(protocolList))
	for _, protocol := range protocolList {
		fields := protocol.Fields()
		if len(fields) == 0 {
			if prototype := protocol.Prototype(); prototype != nil {
				fields = extractFields(prototype)
			}
		}
		metas = append(metas, ProtocolMeta{
			Name:   protocol.Name(),
			Label:  protocol.Label(),
			Color:  protocol.Color(),
			Icon:   protocol.Icon(),
			Fields: fields,
		})
	}

	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Name < metas[j].Name
	})

	protocolMetaCache = metas
	protocolMetaDirty = false
}

func buildIdentity(protocolName, name, host, port string) LinkIdentity {
	return LinkIdentity{
		Protocol: protocolName,
		Name:     name,
		Host:     host,
		Port:     port,
		Address:  fmt.Sprintf("%s:%s", host, port),
	}
}

func newProtocolSpec[T any](
	name string,
	aliases []string,
	label string,
	color string,
	icon string,
	prototype T,
	nameFieldPath string,
	decode func(string) (T, error),
	encode func(T) string,
	identity func(T) LinkIdentity,
	fieldMetas ...FieldMeta,
) *ProtocolSpec {
	return &ProtocolSpec{
		name:          name,
		aliases:       aliases,
		label:         label,
		color:         color,
		icon:          icon,
		prototype:     prototype,
		fields:        append([]FieldMeta(nil), fieldMetas...),
		nameFieldPath: nameFieldPath,
		decode: func(link string) (interface{}, error) {
			return decode(link)
		},
		encode: func(value interface{}) (string, error) {
			typed, ok := value.(T)
			if !ok {
				return "", fmt.Errorf("invalid protocol value type %T for %s", value, name)
			}
			return encode(typed), nil
		},
		identity: func(value interface{}) (LinkIdentity, error) {
			typed, ok := value.(T)
			if !ok {
				return LinkIdentity{}, fmt.Errorf("invalid protocol identity type %T for %s", value, name)
			}
			return identity(typed), nil
		},
	}
}

func newProxyProtocolSpec[T any](
	base *ProtocolSpec,
	toProxy func(Urls, OutputConfig) (Proxy, error),
	canHandle func(Proxy) bool,
	convert func(Proxy) T,
	encode func(T) string,
) *ProxyProtocolSpec {
	return &ProxyProtocolSpec{
		ProtocolSpec:   base,
		toProxy:        toProxy,
		canHandleProxy: canHandle,
		fromProxy: func(proxy Proxy) (string, error) {
			return encode(convert(proxy)), nil
		},
	}
}

func newProxySurgeProtocolSpec[T any](
	base *ProtocolSpec,
	toProxy func(Urls, OutputConfig) (Proxy, error),
	canHandle func(Proxy) bool,
	convert func(Proxy) T,
	encode func(T) string,
	toSurgeLine func(string, OutputConfig) (string, string, error),
) *ProxySurgeProtocolSpec {
	return &ProxySurgeProtocolSpec{
		ProxyProtocolSpec: newProxyProtocolSpec(base, toProxy, canHandle, convert, encode),
		toSurgeLine:       toSurgeLine,
	}
}

func InitProtocolMeta() {
	registryMu.Lock()
	defer registryMu.Unlock()
	rebuildProtocolMetaCacheLocked()
}

func GetAllProtocolMeta() []ProtocolMeta {
	InitProtocolMeta()

	registryMu.RLock()
	defer registryMu.RUnlock()

	metas := make([]ProtocolMeta, len(protocolMetaCache))
	copy(metas, protocolMetaCache)
	return metas
}

func ExtractNodeNameFromFields(protocolName string, fields map[string]interface{}) string {
	protocol := getProtocolByName(protocolName)
	if protocol == nil || fields == nil {
		return ""
	}

	fieldPath := protocol.NameFieldPath()
	if fieldPath == "" {
		return ""
	}

	if value, ok := fields[fieldPath].(string); ok {
		return value
	}
	return ""
}

func ExtractLinkIdentity(link string) (LinkIdentity, error) {
	protocol := detectProtocol(link)
	if protocol == nil {
		return LinkIdentity{}, fmt.Errorf("不支持的协议类型")
	}

	decoded, err := protocol.DecodeLink(link)
	if err != nil {
		return LinkIdentity{}, err
	}

	identity, err := protocol.ExtractIdentity(decoded)
	if err != nil {
		return LinkIdentity{}, err
	}
	if identity.Protocol == "" {
		identity.Protocol = protocol.Name()
	}
	return identity, nil
}

func DecodeProtocolObject(link string) (interface{}, string, error) {
	protocol := detectProtocol(link)
	if protocol == nil {
		return nil, "", fmt.Errorf("不支持的协议类型")
	}

	decoded, err := protocol.DecodeLink(link)
	if err != nil {
		return nil, protocol.Name(), err
	}
	return decoded, protocol.Name(), nil
}

func EncodeProxyLink(proxy Proxy) (string, error) {
	proxyProtocol := getProxyProtocol(proxy)
	if proxyProtocol == nil {
		return "", fmt.Errorf("unsupported proxy type: %s", proxy.Type)
	}
	return proxyProtocol.FromProxy(proxy)
}

func extractFields(v interface{}) []FieldMeta {
	var fields []FieldMeta
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fields
	}

	extractFieldsRecursive(t, "", &fields)
	return fields
}

func extractFieldsRecursive(t reflect.Type, prefix string, fields *[]FieldMeta) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		jsonTag := field.Tag.Get("json")
		label := strings.Split(jsonTag, ",")[0]
		if label == "" || label == "-" {
			label = field.Name
		}

		switch field.Type.Kind() {
		case reflect.String:
			*fields = append(*fields, FieldMeta{Name: fieldName, Label: label, Type: "string"})
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			*fields = append(*fields, FieldMeta{Name: fieldName, Label: label, Type: "int"})
		case reflect.Bool:
			*fields = append(*fields, FieldMeta{Name: fieldName, Label: label, Type: "bool"})
		case reflect.Struct:
			extractFieldsRecursive(field.Type, fieldName, fields)
		}
	}
}

func GetProtocolFieldValue(protoObj interface{}, fieldPath string) string {
	if protoObj == nil {
		return ""
	}

	v := reflect.ValueOf(protoObj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	parts := strings.Split(fieldPath, ".")
	for _, part := range parts {
		if v.Kind() != reflect.Struct {
			return ""
		}
		v = v.FieldByName(part)
		if !v.IsValid() {
			return ""
		}
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Interface:
		if v.IsNil() {
			return ""
		}
		return fmt.Sprintf("%v", v.Interface())
	default:
		return ""
	}
}

func GetProtocolFromLink(link string) string {
	if link == "" {
		return "unknown"
	}

	protocol := detectProtocol(link)
	if protocol == nil {
		return "other"
	}
	return protocol.Name()
}

func GetProtocolLabel(name string) string {
	protocol := getProtocolByName(name)
	if protocol == nil {
		return name
	}
	return protocol.Label()
}

func GetProtocolLabelFromLink(link string) string {
	protocolName := GetProtocolFromLink(link)
	if protocolName == "unknown" {
		return "未知"
	}
	if protocolName == "other" {
		return "其他"
	}
	return GetProtocolLabel(protocolName)
}

func GetAllProtocolNames() []string {
	InitProtocolMeta()

	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(protocolMetaCache))
	for _, meta := range protocolMetaCache {
		names = append(names, meta.Name)
	}
	return names
}

func GetProtocolMeta(name string) *ProtocolMeta {
	InitProtocolMeta()

	registryMu.RLock()
	defer registryMu.RUnlock()

	normalized := normalizeProtocolName(name)
	for i := range protocolMetaCache {
		if protocolMetaCache[i].Name == normalized {
			meta := protocolMetaCache[i]
			return &meta
		}
	}
	return nil
}

func RenameNodeLink(link string, newName string) string {
	if strings.TrimSpace(link) == "" || strings.TrimSpace(newName) == "" {
		return link
	}

	protocol := detectProtocol(link)
	if protocol == nil {
		return link
	}

	decoded, err := protocol.DecodeLink(link)
	if err != nil {
		return link
	}

	v := reflect.ValueOf(decoded)
	if !v.IsValid() {
		return link
	}
	if v.Kind() != reflect.Ptr {
		clone := reflect.New(v.Type())
		clone.Elem().Set(v)
		v = clone
	}
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return renameFragmentOnly(link, newName)
	}

	if fieldPath := protocol.NameFieldPath(); fieldPath != "" {
		if err := setFieldValue(v.Elem(), fieldPath, newName); err == nil {
			if encoded, encodeErr := protocol.EncodeLink(v.Elem().Interface()); encodeErr == nil {
				return encoded
			}
		}
	}

	return renameFragmentOnly(link, newName)
}

func renameFragmentOnly(link string, newName string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	u.Fragment = newName
	return u.String()
}
