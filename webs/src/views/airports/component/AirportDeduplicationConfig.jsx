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
  Chip,
  CircularProgress,
  Alert,
  Paper,
  Stack,
  Collapse
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import { useTheme } from '@mui/material/styles';
import { getProtocolMeta } from 'api/subscriptions';

/**
 * 机场去重规则配置组件
 * 与订阅去重不同，机场拉取时不支持通用字段模式（因为通用字段在拉取时尚未补充）
 * @param {Object} props
 * @param {string} props.value - 当前去重规则配置(JSON字符串)
 * @param {Function} props.onChange - 配置变化回调
 */
function AirportDeduplicationConfig({ value, onChange }) {
  const theme = useTheme();
  // 元数据状态
  const [protocolMeta, setProtocolMeta] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [expanded, setExpanded] = useState(false);

  // 配置状态
  const [config, setConfig] = useState({
    mode: 'none',
    protocolRules: {}
  });

  // 加载协议元数据
  useEffect(() => {
    const fetchMeta = async () => {
      try {
        setLoading(true);
        const protoRes = await getProtocolMeta();
        setProtocolMeta(protoRes.data || []);
        setError(null);
      } catch (err) {
        setError('加载协议元数据失败');
        console.error('加载去重元数据失败:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchMeta();
  }, []);

  // 解析初始值
  useEffect(() => {
    if (value) {
      try {
        const parsed = JSON.parse(value);
        setConfig({
          mode: parsed.mode || 'none',
          protocolRules: parsed.protocolRules || {}
        });
        // 如果已配置去重规则，自动展开
        if (parsed.mode && parsed.mode !== 'none') {
          setExpanded(true);
        }
      } catch (err) {
        console.error('解析去重配置失败:', err);
      }
    }
  }, [value]);

  // 配置变化时通知父组件
  const updateConfig = (newConfig) => {
    setConfig(newConfig);
    // 如果是none模式，传空字符串
    if (newConfig.mode === 'none') {
      onChange('');
    } else {
      onChange(JSON.stringify(newConfig));
    }
  };

  // 模式切换
  const handleModeChange = (event) => {
    const newMode = event.target.value;
    updateConfig({
      ...config,
      mode: newMode
    });
  };

  // 协议字段勾选
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

  // 获取协议已选字段数
  const getProtocolSelectedCount = (protoName) => {
    return (config.protocolRules[protoName] || []).length;
  };

  // 获取总选择数
  const getTotalSelectedCount = () => {
    return Object.values(config.protocolRules).reduce((sum, fields) => sum + (fields?.length || 0), 0);
  };

  // 获取配置状态描述
  const getConfigStatus = () => {
    if (config.mode === 'none') {
      return '内容哈希全库去重';
    }
    const count = getTotalSelectedCount();
    if (count === 0) {
      return '按协议去重（未配置字段）';
    }
    return `按协议去重（${count} 个字段）`;
  };

  return (
    <Paper
      elevation={0}
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
          p: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          background: `linear-gradient(145deg, ${theme.palette.mode === 'dark' ? '#1a2027' : '#f5f5f5'} 0%, ${theme.palette.mode === 'dark' ? '#121417' : '#fafafa'} 100%)`,
          cursor: 'pointer',
          '&:hover': {
            bgcolor: 'action.hover'
          }
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <FilterAltIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600}>
            节点入库去重
          </Typography>
          <Chip
            size="small"
            label={getConfigStatus()}
            color={config.mode === 'none' ? 'default' : 'primary'}
            variant={config.mode === 'none' ? 'outlined' : 'filled'}
          />
        </Stack>
        <Stack direction="row" alignItems="center" spacing={0.5}>
          {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
        </Stack>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ p: 2, pt: 1 }}>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
              <CircularProgress size={24} />
              <Typography sx={{ ml: 1 }}>加载配置...</Typography>
            </Box>
          ) : error ? (
            <Alert severity="error">{error}</Alert>
          ) : (
            <>
              {/* 模式选择 */}
              <FormControl component="fieldset" sx={{ mb: 2 }}>
                <FormLabel component="legend">去重模式</FormLabel>
                <RadioGroup row value={config.mode} onChange={handleModeChange}>
                  <FormControlLabel value="none" control={<Radio size="small" />} label="默认（内容哈希全库去重）" />
                  <FormControlLabel value="protocol" control={<Radio size="small" />} label="按协议字段去重（仅本次拉取）" />
                </RadioGroup>
              </FormControl>

              {/* 协议特定字段选择 */}
              {config.mode === 'protocol' && (
                <Box>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    为每个协议配置去重字段（当多个节点的选定字段值完全相同时，仅保留第一个）：
                  </Typography>
                  {protocolMeta.map((proto) => (
                    <Accordion key={proto.name} sx={{ mb: 1 }} defaultExpanded={getProtocolSelectedCount(proto.name) > 0}>
                      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                        <Typography sx={{ fontWeight: 500 }}>{proto.label}</Typography>
                        {getProtocolSelectedCount(proto.name) > 0 && (
                          <Chip size="small" label={`已选 ${getProtocolSelectedCount(proto.name)} 个`} color="primary" sx={{ ml: 1 }} />
                        )}
                      </AccordionSummary>
                      <AccordionDetails>
                        <FormGroup row>
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
                </Box>
              )}

              {/* 提示信息 */}
              <Alert variant="standard" severity={config.mode === 'none' ? 'info' : 'warning'} sx={{ mt: 2 }}>
                <Typography variant="body2">
                  {config.mode === 'none' ? (
                    <>
                      使用<strong>节点内容哈希进行全库去重</strong>，确保数据库中不会存储内容完全相同的节点。
                      此方式不受链接参数顺序影响，不同机场间的重复节点也会被过滤。
                    </>
                  ) : (
                    <>
                      按协议字段去重<strong>仅在本次拉取的节点内部生效</strong>，
                      不会与数据库中已有的其他节点进行比较。
                      未配置字段的协议将继续使用默认的内容哈希全库去重。
                    </>
                  )}
                </Typography>
              </Alert>
            </>
          )}
        </Box>
      </Collapse>
    </Paper>
  );
}

AirportDeduplicationConfig.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired
};

export default AirportDeduplicationConfig;
