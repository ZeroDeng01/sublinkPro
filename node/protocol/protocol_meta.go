package protocol

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// FieldMeta 描述一个协议字段在通用表单和元数据输出中的展示方式。
type FieldMeta struct {
	// Name 是字段标识，通常使用协议结构体中的导出字段名或点路径。
	Name string `json:"name"`
	// Label 是给 UI 或外部调用方展示的字段名称。
	Label string `json:"label"`
	// Type 表示字段值类型，当前约定为 string、int 或 bool。
	Type string `json:"type"`
	// Group 用于把字段归类到同一个展示分组；为空时表示不强制分组。
	Group string `json:"group,omitempty"`
	// Description 是字段的补充说明，供外部表单或文档直接展示。
	Description string `json:"description,omitempty"`
	// Placeholder 是字段为空时的占位提示文本。
	Placeholder string `json:"placeholder,omitempty"`
	// Options 是可选值列表；为空时表示调用方可自由输入。
	Options []string `json:"options,omitempty"`
	// Advanced 标记该字段是否应被视为高级选项。
	Advanced bool `json:"advanced,omitempty"`
	// Secret 标记该字段是否包含敏感信息，调用方通常应避免明文展示。
	Secret bool `json:"secret,omitempty"`
	// Multiline 标记该字段是否更适合以多行文本方式输入。
	Multiline bool `json:"multiline,omitempty"`
}

// ProtocolMeta 是单个已注册协议的对外元数据摘要。
type ProtocolMeta struct {
	// Name 是协议的规范化内部名称。
	Name string `json:"name"`
	// Label 是面向用户展示的协议名称。
	Label string `json:"label"`
	// Color 是协议对应的展示色值。
	Color string `json:"color"`
	// Icon 是协议对应的简短图标文本。
	Icon string `json:"icon"`
	// Fields 是该协议可编辑字段列表；若协议未显式声明，可能来自原型结构体反射推断。
	Fields []FieldMeta `json:"fields"`
}

// LinkIdentity 表示从节点链接中提取出的规范化识别信息。
type LinkIdentity struct {
	// Protocol 是已注册的协议名称；通过 ExtractLinkIdentity 提取时，若协议实现未填写，该字段会被自动补齐。
	Protocol string
	// Name 是节点当前名称；当协议本身缺少名称时可能为空。
	Name string
	// Address 是 Host 与 Port 拼接后的地址表示，通常为 host:port。
	Address string
	// Host 是从链接中提取出的主机名或服务器地址。
	Host string
	// Port 是从链接中提取出的端口字符串；不可用时可能为空。
	Port string
}

// Protocol 定义一个可注册协议需要提供的基础编解码与元数据能力。
// 同一实现返回的 Prototype、DecodeLink、EncodeLink 和 ExtractIdentity 应使用一致的具体类型语义。
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

// ProxyCapable 表示协议支持与通用 Proxy 结构互相转换。
type ProxyCapable interface {
	ToProxy(Urls, OutputConfig) (Proxy, error)
	CanHandleProxy(Proxy) bool
	FromProxy(Proxy) (string, error)
}

// SurgeCapable 表示协议支持导出为 Surge 节点行。
// 返回值依次为生成的节点行、节点名称以及错误信息。
type SurgeCapable interface {
	ToSurgeLine(string, OutputConfig) (string, string, error)
}

// ProtocolSpec 是 Protocol 的通用实现，适用于通过函数组合注册协议元信息的场景。
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

// Prototype 返回注册时保存的协议原型值，用于外部推断字段结构或生成默认元数据。
func (p *ProtocolSpec) Prototype() interface{} {
	return p.prototype
}

// Fields 返回协议声明字段列表的外层切片副本。
// 调用方应将返回结果视为只读数据，以避免误用共享的嵌套切片内容。
func (p *ProtocolSpec) Fields() []FieldMeta {
	return append([]FieldMeta(nil), p.fields...)
}

func (p *ProtocolSpec) NameFieldPath() string {
	return p.nameFieldPath
}

// DecodeLink 使用协议自身的解码函数解析链接；当协议未提供解码能力时返回错误。
func (p *ProtocolSpec) DecodeLink(link string) (interface{}, error) {
	if p.decode == nil {
		return nil, fmt.Errorf("protocol %s does not support decoding", p.name)
	}
	return p.decode(link)
}

// EncodeLink 使用协议自身的编码函数生成链接；当值类型不匹配或未提供编码能力时返回错误。
func (p *ProtocolSpec) EncodeLink(value interface{}) (string, error) {
	if p.encode == nil {
		return "", fmt.Errorf("protocol %s does not support encoding", p.name)
	}
	return p.encode(value)
}

// ExtractIdentity 从协议对象中提取 LinkIdentity；当协议未提供该能力或值类型不匹配时返回错误。
func (p *ProtocolSpec) ExtractIdentity(value interface{}) (LinkIdentity, error) {
	if p.identity == nil {
		return LinkIdentity{}, fmt.Errorf("protocol %s does not provide identity extraction", p.name)
	}
	return p.identity(value)
}

// ProxyProtocolSpec 在 ProtocolSpec 之上补充 ProxyCapable 能力。
type ProxyProtocolSpec struct {
	*ProtocolSpec
	toProxy        func(Urls, OutputConfig) (Proxy, error)
	canHandleProxy func(Proxy) bool
	fromProxy      func(Proxy) (string, error)
}

// ToProxy 将协议链接转换为通用 Proxy；当协议未配置导出能力时返回错误。
func (p *ProxyProtocolSpec) ToProxy(link Urls, config OutputConfig) (Proxy, error) {
	if p.toProxy == nil {
		return Proxy{}, fmt.Errorf("protocol %s does not support proxy export", p.name)
	}
	return p.toProxy(link, config)
}

// CanHandleProxy 判断当前协议是否能处理给定的 Proxy 输入。
func (p *ProxyProtocolSpec) CanHandleProxy(proxy Proxy) bool {
	if p.canHandleProxy == nil {
		return false
	}
	return p.canHandleProxy(proxy)
}

// FromProxy 将通用 Proxy 重新编码为协议链接；当协议未配置导入能力时返回错误。
func (p *ProxyProtocolSpec) FromProxy(proxy Proxy) (string, error) {
	if p.fromProxy == nil {
		return "", fmt.Errorf("protocol %s does not support proxy import", p.name)
	}
	return p.fromProxy(proxy)
}

// ProxySurgeProtocolSpec 在 ProxyProtocolSpec 之上补充 SurgeCapable 能力。
type ProxySurgeProtocolSpec struct {
	*ProxyProtocolSpec
	toSurgeLine func(string, OutputConfig) (string, string, error)
}

// ToSurgeLine 将协议链接导出为 Surge 节点行，并返回生成的节点名称；未配置该能力时返回错误。
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
	registryMu        sync.RWMutex
	protocolsByName   = make(map[string]Protocol)
	protocolsByAlias  = make(map[string]Protocol)
	aliasMatchers     []aliasMatcher
	protocolList      []Protocol
	proxyCapables     []ProxyCapable
	protocolMetaCache []ProtocolMeta
	protocolMetaDirty = true
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

// MustRegisterProtocol 注册协议元信息及其别名，并同步刷新导出能力索引。
// 这里会对名称和别名做去重校验，重复注册会直接 panic，以避免运行时分发出现歧义。
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

// detectProtocol 根据链接前缀匹配协议实现。
// 别名前缀会按长度优先排序，避免短前缀提前命中覆盖更具体的协议别名。
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

// rebuildProtocolMetaCacheLocked 在注册表变更后重建协议元数据缓存。
// 若协议未显式声明字段元信息，则会从原型结构体反射推断一份基础字段描述。
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

// newProtocolSpec 将具体协议的编解码函数包装为统一的元数据描述，供注册表与 UI 共用。
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

// newProxyProtocolSpec 为支持 Clash Proxy 导入导出的协议补齐统一适配层。
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

// newProxySurgeProtocolSpec 在 Proxy 导入导出基础上，再补充 Surge 节点行导出能力。
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

// InitProtocolMeta 在缓存脏标记存在时重建协议元数据缓存。
// 调用方通常无需手动预热，读取类接口会在内部按需调用它。
func InitProtocolMeta() {
	registryMu.Lock()
	defer registryMu.Unlock()
	rebuildProtocolMetaCacheLocked()
}

// GetAllProtocolMeta 返回当前已注册协议的元数据列表。
// 返回切片是新的外层切片，内容按协议名称排序；调用方应将其中字段切片视为只读数据。
func GetAllProtocolMeta() []ProtocolMeta {
	InitProtocolMeta()

	registryMu.RLock()
	defer registryMu.RUnlock()

	metas := make([]ProtocolMeta, len(protocolMetaCache))
	copy(metas, protocolMetaCache)
	return metas
}

// ExtractNodeNameFromFields 根据协议定义的 NameFieldPath 从字段映射中提取节点名称。
// 当协议不存在、字段映射为空、名称字段未声明或值不是字符串时，返回空字符串。
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

// ExtractLinkIdentity 解析任意已注册协议链接，并返回用于去重或展示的 LinkIdentity。
// 若协议实现未填写 Protocol 字段，该函数会自动补上当前注册协议名；不支持的协议或解码失败会返回错误。
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

// DecodeProtocolObject 根据链接自动识别协议并返回对应的协议对象与规范化协议名。
// 若链接无法识别协议，返回 nil、空协议名和错误；若识别成功但解码失败，仍会返回已识别的协议名。
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

// EncodeProxyLink 根据 Proxy.Type 选择已注册的 ProxyCapable 实现，并将 Proxy 编码为协议链接。
// 当没有任何协议声明可处理该 Proxy 时，返回错误。
func EncodeProxyLink(proxy Proxy) (string, error) {
	proxyProtocol := getProxyProtocol(proxy)
	if proxyProtocol == nil {
		return "", fmt.Errorf("unsupported proxy type: %s", proxy.Type)
	}
	return proxyProtocol.FromProxy(proxy)
}

// extractFields 从协议原型结构体递归提取基础字段信息，用于没有显式 FieldMeta 的协议兜底展示。
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

// extractFieldsRecursive 递归展开导出字段，并用点路径表示嵌套结构字段。
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

// GetProtocolFieldValue 按点路径读取协议对象中的导出字段值，并返回字符串形式结果。
// fieldPath 使用 Go 导出字段名而不是 JSON 标签；当对象为空、路径不存在或最终值类型不受支持时返回空字符串。
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

// GetProtocolFromLink 根据链接内容返回识别到的协议名称。
// 空字符串输入返回 unknown，无法识别但非空的输入返回 other。
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

// GetProtocolLabel 返回协议名称对应的展示标签；若协议未注册，则回退为原始名称。
func GetProtocolLabel(name string) string {
	protocol := getProtocolByName(name)
	if protocol == nil {
		return name
	}
	return protocol.Label()
}

// GetProtocolLabelFromLink 根据链接返回协议展示标签。
// 空输入映射为“未知”，无法识别的非空链接映射为“其他”。
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

// GetAllProtocolNames 返回当前已注册协议的名称列表。
// 返回顺序与协议元数据缓存一致，即按协议名称排序后的顺序。
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

// GetProtocolMeta 根据协议名称返回对应的元数据值副本。
// 返回结果中的嵌套切片仍应视为只读；名称会先做规范化处理，协议不存在时返回 nil。
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

// RenameNodeLink 优先通过协议对象的 NameFieldPath 重编码节点名称，失败时再退回到仅修改 URL fragment。
// 当输入为空、协议不受支持、解码失败或重编码失败时，会尽量返回原始链接而不是中断调用流程。
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

// renameFragmentOnly 作为兜底策略，仅修改链接 fragment，不尝试理解协议内部结构。
func renameFragmentOnly(link string, newName string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	u.Fragment = newName
	return u.String()
}
