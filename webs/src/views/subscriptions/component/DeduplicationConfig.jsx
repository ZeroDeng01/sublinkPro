import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import {
  Box,
  FormControl,
  FormLabel,
  RadioGroup,
  FormControlLabel,
  Radio,
  Checkbox,
  FormGroup,
  Typography,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Collapse,
  Chip,
  CircularProgress,
  Alert,
  Paper,
  Stack
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import { useTheme } from '@mui/material/styles';
import { getProtocolMeta, getNodeFieldsMeta } from 'api/subscriptions';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

function DeduplicationConfig({ value, onChange }) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);
  const insetHighlight = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';

  const [protocolMeta, setProtocolMeta] = useState([]);
  const [nodeFieldsMeta, setNodeFieldsMeta] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [expanded, setExpanded] = useState(false);
  const [config, setConfig] = useState({
    mode: 'none',
    commonFields: [],
    protocolRules: {}
  });

  useEffect(() => {
    const fetchMeta = async () => {
      try {
        setLoading(true);
        const [protoRes, nodeRes] = await Promise.all([getProtocolMeta(), getNodeFieldsMeta()]);
        setProtocolMeta(protoRes.data || []);
        setNodeFieldsMeta(nodeRes.data?.fields || []);
        setError(null);
      } catch (err) {
        setError('加载元数据失败');
        console.error('加载去重元数据失败:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchMeta();
  }, []);

  useEffect(() => {
    if (value) {
      try {
        const parsed = JSON.parse(value);
        setConfig({
          mode: parsed.mode || 'none',
          commonFields: parsed.commonFields || [],
          protocolRules: parsed.protocolRules || {}
        });
        if (parsed.mode && parsed.mode !== 'none') {
          setExpanded(true);
        }
      } catch (err) {
        console.error('解析去重配置失败:', err);
      }
    }
  }, [value]);

  const updateConfig = (newConfig) => {
    setConfig(newConfig);
    if (newConfig.mode === 'none') {
      onChange('');
    } else {
      onChange(JSON.stringify(newConfig));
    }
  };

  const handleModeChange = (event) => {
    const newMode = event.target.value;
    updateConfig({
      ...config,
      mode: newMode
    });
  };

  const handleCommonFieldChange = (fieldName) => {
    const newFields = config.commonFields.includes(fieldName)
      ? config.commonFields.filter((f) => f !== fieldName)
      : [...config.commonFields, fieldName];

    updateConfig({
      ...config,
      commonFields: newFields
    });
  };

  const handleProtocolFieldChange = (protoName, fieldName) => {
    const currentFields = config.protocolRules[protoName] || [];
    const newFields = currentFields.includes(fieldName) ? currentFields.filter((f) => f !== fieldName) : [...currentFields, fieldName];

    updateConfig({
      ...config,
      protocolRules: {
        ...config.protocolRules,
        [protoName]: newFields
      }
    });
  };

  const getProtocolSelectedCount = (protoName) => (config.protocolRules[protoName] || []).length;

  const getStatusLabel = () => {
    if (config.mode === 'none') return '未启用';
    if (config.mode === 'common') return `通用 ${config.commonFields.length} 项`;
    const total = Object.values(config.protocolRules).reduce((sum, fields) => sum + (fields?.length || 0), 0);
    return `协议 ${total} 项`;
  };

  const headerHoverSurface = isDark ? withAlpha(palette.background.paper, 0.2) : withAlpha(palette.primary.main, 0.04);
  const contentSurface = isDark
    ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.08)} 0%, ${dialogSurface} 100%)`
    : 'none';

  const sectionCardSx = {
    p: 1.75,
    borderRadius: 2,
    bgcolor: nestedPanelSurface,
    border: '1px solid',
    borderColor: panelBorder,
    boxShadow: insetHighlight
  };

  const optionGroupSx = {
    gap: 1,
    '& .MuiFormControlLabel-root': {
      m: 0,
      minHeight: 40,
      px: 1.25,
      py: 0.375,
      border: '1px solid',
      borderColor: panelBorder,
      borderRadius: 1.5,
      bgcolor: dialogSurface,
      transition: 'border-color 0.2s ease, background-color 0.2s ease',
      '&:hover': {
        borderColor: withAlpha(palette.primary.main, isDark ? 0.3 : 0.18)
      },
      '& .MuiFormControlLabel-label': {
        color: secondaryText,
        fontSize: theme.typography.body2.fontSize
      }
    }
  };

  const nestedAccordionSx = {
    mb: 0,
    bgcolor: dialogSurface,
    border: '1px solid',
    borderColor: panelBorder,
    borderRadius: 2,
    overflow: 'hidden',
    boxShadow: insetHighlight,
    '&:before': { display: 'none' },
    '& .MuiAccordionSummary-root': {
      minHeight: 48,
      bgcolor: mutedPanelSurface,
      '&:hover': {
        bgcolor: headerHoverSurface
      },
      '&.Mui-expanded': {
        minHeight: 48,
        bgcolor: nestedPanelSurface,
        borderBottom: '1px solid',
        borderColor: panelBorder
      }
    },
    '& .MuiAccordionSummary-content': {
      alignItems: 'center',
      gap: 1.25,
      '&.Mui-expanded': {
        margin: '12px 0'
      }
    },
    '& .MuiAccordionSummary-expandIconWrapper': {
      color: tertiaryText
    },
    '& .MuiAccordionDetails-root': {
      px: 2,
      py: 1.75,
      bgcolor: dialogSurface,
      backgroundImage: isDark ? `linear-gradient(180deg, ${withAlpha(palette.background.paper, 0.06)} 0%, ${dialogSurface} 100%)` : 'none'
    }
  };

  return (
    <Paper
      elevation={0}
      sx={{
        mb: 0,
        border: '1px solid',
        borderColor: panelBorder,
        borderRadius: 2,
        overflow: 'hidden',
        bgcolor: dialogSurface,
        backgroundImage: dialogSurfaceGradient,
        boxShadow: insetHighlight
      }}
    >
      <Box
        sx={{
          px: 1.75,
          py: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          bgcolor: expanded ? nestedPanelSurface : mutedPanelSurface,
          borderBottom: expanded ? '1px solid' : 'none',
          borderColor: panelBorder,
          cursor: 'pointer',
          transition: 'background-color 0.2s ease, border-color 0.2s ease',
          '&:hover': {
            bgcolor: expanded ? nestedPanelSurface : headerHoverSurface
          }
        }}
        onClick={() => setExpanded((prev) => !prev)}
      >
        <Stack direction="row" alignItems="center" spacing={1.25} sx={{ minWidth: 0, flex: 1 }}>
          <FilterAltIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600} sx={{ color: primaryText }}>
            节点去重规则
          </Typography>
          <Chip
            size="small"
            variant="outlined"
            label={getStatusLabel()}
            sx={{
              height: 22,
              color: config.mode === 'none' ? tertiaryText : palette.primary.main,
              bgcolor:
                config.mode === 'none'
                  ? withAlpha(palette.text.secondary, isDark ? 0.12 : 0.05)
                  : withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
              borderColor:
                config.mode === 'none'
                  ? withAlpha(palette.divider, isDark ? 0.8 : 0.95)
                  : withAlpha(palette.primary.main, isDark ? 0.28 : 0.16)
            }}
          />
        </Stack>
        <Box sx={{ display: 'flex', alignItems: 'center', color: tertiaryText, ml: 1 }}>
          <ExpandMoreIcon
            fontSize="small"
            sx={{ transform: expanded ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s ease' }}
          />
        </Box>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ px: 2.25, py: 2.25, bgcolor: dialogSurface, backgroundImage: contentSurface }}>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 1, p: 1.5, color: secondaryText }}>
              <CircularProgress size={24} />
              <Typography variant="body2">加载配置...</Typography>
            </Box>
          ) : error ? (
            <Alert variant="outlined" severity="error">
              {error}
            </Alert>
          ) : (
            <Stack spacing={2.25}>
              <Typography variant="body2" sx={{ color: secondaryText }}>
                在预览和导出前统一处理重复节点，可按通用字段组合去重，也可以为协议分别设置去重字段。
              </Typography>

              <Box sx={sectionCardSx}>
                <FormControl component="fieldset">
                  <FormLabel component="legend" sx={{ color: tertiaryText, '&.Mui-focused': { color: tertiaryText } }}>
                    去重模式
                  </FormLabel>
                  <RadioGroup row value={config.mode} onChange={handleModeChange} sx={{ mt: 0.75 }}>
                    <FormControlLabel value="none" control={<Radio size="small" />} label="不启用" />
                    <FormControlLabel value="common" control={<Radio size="small" />} label="通用字段去重" />
                    <FormControlLabel value="protocol" control={<Radio size="small" />} label="按协议去重" />
                  </RadioGroup>
                </FormControl>
              </Box>

              {config.mode === 'common' && (
                <Box sx={sectionCardSx}>
                  <Stack spacing={1.25}>
                    <Typography variant="body2" sx={{ color: secondaryText }}>
                      选择用于判断节点是否重复的字段（多选组合）：
                    </Typography>
                    <FormGroup row sx={optionGroupSx}>
                      {nodeFieldsMeta.map((field) => (
                        <FormControlLabel
                          key={field.name}
                          control={
                            <Checkbox
                              size="small"
                              checked={config.commonFields.includes(field.name)}
                              onChange={() => handleCommonFieldChange(field.name)}
                            />
                          }
                          label={field.label}
                        />
                      ))}
                    </FormGroup>
                    {config.commonFields.length > 0 && (
                      <Typography variant="caption" sx={{ color: palette.primary.main }}>
                        已选择：{config.commonFields.map((f) => nodeFieldsMeta.find((m) => m.name === f)?.label || f).join(' + ')}
                      </Typography>
                    )}
                  </Stack>
                </Box>
              )}

              {config.mode === 'protocol' && (
                <Box sx={sectionCardSx}>
                  <Stack spacing={1.25} sx={{ mb: 1.5 }}>
                    <Typography variant="body2" sx={{ color: secondaryText }}>
                      为每个协议配置去重字段（未配置的协议不进行去重）：
                    </Typography>
                  </Stack>
                  <Stack spacing={1.25}>
                    {protocolMeta.map((proto) => (
                      <Accordion key={proto.name} sx={nestedAccordionSx} defaultExpanded={getProtocolSelectedCount(proto.name) > 0}>
                        <AccordionSummary expandIcon={<ExpandMoreIcon fontSize="small" />}>
                          <Typography sx={{ fontWeight: 600, color: primaryText }}>{proto.label}</Typography>
                          {getProtocolSelectedCount(proto.name) > 0 && (
                            <Chip
                              size="small"
                              label={`已选 ${getProtocolSelectedCount(proto.name)} 个`}
                              variant="outlined"
                              sx={{
                                ml: 0.5,
                                color: palette.primary.main,
                                bgcolor: withAlpha(palette.primary.main, isDark ? 0.14 : 0.06),
                                borderColor: withAlpha(palette.primary.main, isDark ? 0.28 : 0.16)
                              }}
                            />
                          )}
                        </AccordionSummary>
                        <AccordionDetails>
                          <FormGroup row sx={optionGroupSx}>
                            {(proto.fields || []).map((field) => (
                              <FormControlLabel
                                key={field.name}
                                control={
                                  <Checkbox
                                    size="small"
                                    checked={(config.protocolRules[proto.name] || []).includes(field.name)}
                                    onChange={() => handleProtocolFieldChange(proto.name, field.name)}
                                  />
                                }
                                label={field.label}
                              />
                            ))}
                          </FormGroup>
                        </AccordionDetails>
                      </Accordion>
                    ))}
                  </Stack>
                </Box>
              )}

              {config.mode !== 'none' && (
                <Alert
                  variant="outlined"
                  severity="info"
                  sx={{
                    borderColor: withAlpha(palette.info.main, isDark ? 0.28 : 0.18),
                    bgcolor: withAlpha(palette.info.main, isDark ? 0.12 : 0.05),
                    boxShadow: insetHighlight
                  }}
                >
                  <Typography variant="body2" sx={{ color: primaryText }}>
                    去重规则会在节点预览和订阅输出时应用，当多个节点的选定字段值完全相同时，仅保留第一个节点。
                  </Typography>
                </Alert>
              )}
            </Stack>
          )}
        </Box>
      </Collapse>
    </Paper>
  );
}

DeduplicationConfig.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired
};

export default DeduplicationConfig;
