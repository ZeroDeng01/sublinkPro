import i18n from 'i18n';

const getUnknownNodeName = () => i18n.t('nodes.common.unknownNode', { defaultValue: 'Unknown node' });

export const getNodeDisplayName = (node) => {
  if (!node) return getUnknownNodeName();

  if (node.PreviewName) {
    return node.PreviewName;
  }

  if (node.EffectiveName) {
    return node.EffectiveName;
  }

  if (node.NameMode === 'remark' && node.Name) {
    return node.Name;
  }

  if (node.NameMode === 'link' && node.LinkName) {
    return node.LinkName;
  }

  return node.Name || node.LinkName || node.OriginalName || getUnknownNodeName();
};
