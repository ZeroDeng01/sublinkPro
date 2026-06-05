import { useState, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Collapse from '@mui/material/Collapse';
import IconButton from '@mui/material/IconButton';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

// icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import FingerprintIcon from '@mui/icons-material/Fingerprint';

/**
 * 节点名称唯一化配置组件
 * 通过添加机场标识前缀防止多机场间节点名称重复
 */
export default function NodeNameUniquifyConfig({ enabled, prefix, intraUniquify, airportId, onChange }) {
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState(false);
  const previewNodeName = t('components.nodeNameUniquify.preview.samplePrimary');
  const secondPreviewNodeName = t('components.nodeNameUniquify.preview.sampleSecondary');
  const displayPrefix = prefix.trim() || (airportId ? `[A${airportId}]` : t('components.nodeNameUniquify.defaultPrefixPending'));

  const getSettingRowSx = (active) => ({
    px: 1.5,
    py: 1,
    borderRadius: 1.5,
    border: '1px solid',
    borderColor: 'divider',
    bgcolor: active ? 'background.default' : 'action.disabledBackground',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 1.5
  });

  // 根据当前配置计算预览结果
  const previewResult = useMemo(() => {
    if (!enabled) {
      return previewNodeName;
    }
    return displayPrefix + previewNodeName;
  }, [enabled, displayPrefix, previewNodeName]);

  const duplicatePreviewGroups = useMemo(() => {
    if (!intraUniquify) {
      return [];
    }

    const previewPrefix = enabled ? displayPrefix : '';

    return [
      {
        label: t('components.nodeNameUniquify.preview.duplicateGroup', { name: previewNodeName }),
        items: [`${previewPrefix}${previewNodeName}-1`, `${previewPrefix}${previewNodeName}-2`]
      },
      {
        label: t('components.nodeNameUniquify.preview.duplicateGroup', { name: secondPreviewNodeName }),
        items: [`${previewPrefix}${secondPreviewNodeName}-1`, `${previewPrefix}${secondPreviewNodeName}-2`]
      }
    ];
  }, [enabled, intraUniquify, displayPrefix, previewNodeName, secondPreviewNodeName, t]);

  // 自动展开面板（当开启功能时）
  useEffect(() => {
    if (enabled || intraUniquify) {
      setExpanded(true);
    }
  }, [enabled, intraUniquify]);

  const handleEnabledChange = (event) => {
    onChange({ enabled: event.target.checked, prefix, intraUniquify });
  };

  const handlePrefixChange = (event) => {
    onChange({ enabled, prefix: event.target.value, intraUniquify });
  };

  const handleIntraUniquifyChange = (event) => {
    onChange({ enabled, prefix, intraUniquify: event.target.checked });
  };

  return (
    <Box
      sx={{
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 2,
        overflow: 'hidden'
      }}
    >
      {/* 标题栏 */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 2,
          py: 1,
          bgcolor: 'background.default',
          cursor: 'pointer',
          '&:hover': { bgcolor: 'action.hover' }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 0.75, minWidth: 0, flex: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0, width: '100%' }}>
            <FingerprintIcon color="action" fontSize="small" />
            <Box sx={{ minWidth: 0 }}>
              <Typography variant="body2" fontWeight={500}>
                {t('components.nodeNameUniquify.title')}
              </Typography>
              <Typography
                variant="caption"
                sx={{
                  display: 'block',
                  color: 'text.secondary'
                }}
              >
                {t('components.nodeNameUniquify.subtitle')}
              </Typography>
            </Box>
          </Box>
          {(enabled || intraUniquify) && (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.75 }}>
              {enabled && (
                <Typography
                  variant="caption"
                  sx={{
                    px: 1,
                    py: 0.25,
                    borderRadius: 1,
                    bgcolor: 'primary.main',
                    color: 'primary.contrastText'
                  }}
                >
                  {t('components.nodeNameUniquify.crossAirport.title')}
                </Typography>
              )}
              {intraUniquify && (
                <Typography
                  variant="caption"
                  sx={{
                    px: 1,
                    py: 0.25,
                    borderRadius: 1,
                    bgcolor: 'secondary.main',
                    color: 'secondary.contrastText'
                  }}
                >
                  {t('components.nodeNameUniquify.intraAirport.badge')}
                </Typography>
              )}
            </Box>
          )}
        </Box>
        <IconButton size="small">{expanded ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />}</IconButton>
      </Box>

      {/* 展开内容 */}
      <Collapse in={expanded}>
        <Box sx={{ p: 2, borderTop: '1px solid', borderColor: 'divider' }}>
          <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 2 }}>
            {t('components.nodeNameUniquify.description')}
          </Typography>

          <Box sx={{ display: 'grid', gap: 1.25, mb: 2 }}>
            <Box sx={{ ...getSettingRowSx(enabled), display: 'grid', gap: 1.25 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1.5 }}>
                <Box sx={{ minWidth: 0 }}>
                  <Typography variant="body2" fontWeight={500}>
                    {t('components.nodeNameUniquify.crossAirport.title')}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {t('components.nodeNameUniquify.crossAirport.description')}
                  </Typography>
                </Box>
                <Switch size="small" checked={enabled} onChange={handleEnabledChange} />
              </Box>
              {enabled && (
                <TextField
                  fullWidth
                  size="small"
                  label={t('components.nodeNameUniquify.crossAirport.prefixLabel')}
                  placeholder={
                    airportId
                      ? t('components.nodeNameUniquify.crossAirport.placeholderWithId', { id: airportId })
                      : t('components.nodeNameUniquify.crossAirport.placeholderPending')
                  }
                  value={prefix}
                  onChange={handlePrefixChange}
                  helperText={t('components.nodeNameUniquify.crossAirport.helper')}
                />
              )}
            </Box>

            <Box sx={getSettingRowSx(intraUniquify)}>
              <Box sx={{ minWidth: 0 }}>
                <Typography variant="body2" fontWeight={500}>
                  {t('components.nodeNameUniquify.intraAirport.title')}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {t('components.nodeNameUniquify.intraAirport.description')}
                </Typography>
              </Box>
              <Switch size="small" checked={intraUniquify} onChange={handleIntraUniquifyChange} />
            </Box>
          </Box>

          {/* 预览效果 */}
          <Box
            sx={{
              p: 1.5,
              borderRadius: 1,
              bgcolor: enabled || intraUniquify ? 'action.selected' : 'action.disabledBackground'
            }}
          >
            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 0.5 }}>
              {t('components.nodeNameUniquify.preview.crossAirportTitle')}
            </Typography>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
              <Typography variant="body2" sx={{ fontFamily: 'monospace', color: 'text.secondary', textDecoration: 'line-through' }}>
                {previewNodeName}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                →
              </Typography>
              <Typography
                variant="body2"
                sx={{
                  fontFamily: 'monospace',
                  fontWeight: 500,
                  color: enabled ? 'primary.main' : 'text.disabled'
                }}
              >
                {previewResult}
              </Typography>
            </Box>

            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mt: 0.75 }}>
              {t(
                enabled
                  ? 'components.nodeNameUniquify.preview.enabledDescription'
                  : 'components.nodeNameUniquify.preview.disabledDescription'
              )}
            </Typography>

            {intraUniquify && (
              <Box sx={{ mt: 1.5 }}>
                <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 0.5 }}>
                  {t('components.nodeNameUniquify.preview.intraAirportTitle')}
                </Typography>
                <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 1 }}>
                  {t('components.nodeNameUniquify.preview.intraAirportDescription')}
                </Typography>

                {duplicatePreviewGroups.map((group) => (
                  <Box key={group.label} sx={{ '&:not(:last-of-type)': { mb: 1.25 } }}>
                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                      {group.label}
                    </Typography>
                    {group.items.map((item) => (
                      <Typography key={item} variant="body2" sx={{ fontFamily: 'monospace', color: 'primary.main', fontWeight: 500 }}>
                        {item}
                      </Typography>
                    ))}
                  </Box>
                ))}
              </Box>
            )}
          </Box>
        </Box>
      </Collapse>
    </Box>
  );
}

NodeNameUniquifyConfig.propTypes = {
  enabled: PropTypes.bool.isRequired,
  prefix: PropTypes.string,
  intraUniquify: PropTypes.bool,
  airportId: PropTypes.number,
  onChange: PropTypes.func.isRequired
};

NodeNameUniquifyConfig.defaultProps = {
  prefix: '',
  intraUniquify: false,
  airportId: 0
};
