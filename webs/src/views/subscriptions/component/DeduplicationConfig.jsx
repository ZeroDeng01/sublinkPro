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
  Alert
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { getProtocolMeta, getNodeFieldsMeta } from 'api/subscriptions';

/**
 * 去重规则配置组件
 * @param {Object} props
 * @param {string} props.value - 当前去重规则配置(JSON字符串)
 * @param {Function} props.onChange - 配置变化回调
 */
function DeduplicationConfig({ value, onChange }) {
  // 元数据状态
  const [protocolMeta, setProtocolMeta] = useState([]);
  const [nodeFieldsMeta, setNodeFieldsMeta] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // 配置状态
  const [config, setConfig] = useState({
    mode: 'none',
    commonFields: [],
    protocolRules: {}
  });

  // 加载元数据
  useEffect(() => {
    const fetchMeta = async () => {
      try {
        setLoading(true);
        const [protoRes, nodeRes] = await Promise.all([getProtocolMeta(), getNodeFieldsMeta()]);
        setProtocolMeta(protoRes.data || []);
        setNodeFieldsMeta(nodeRes.data || []);
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

  // 解析初始值
  useEffect(() => {
    if (value) {
      try {
        const parsed = JSON.parse(value);
        setConfig({
          mode: parsed.mode || 'none',
          commonFields: parsed.commonFields || [],
          protocolRules: parsed.protocolRules || {}
        });
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

  // 通用字段勾选
  const handleCommonFieldChange = (fieldName) => {
    const newFields = config.commonFields.includes(fieldName)
      ? config.commonFields.filter((f) => f !== fieldName)
      : [...config.commonFields, fieldName];
    updateConfig({
      ...config,
      commonFields: newFields
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

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
        <CircularProgress size={24} />
        <Typography sx={{ ml: 1 }}>加载配置...</Typography>
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  return (
    <Box>
      {/* 模式选择 */}
      <FormControl component="fieldset" sx={{ mb: 2 }}>
        <FormLabel component="legend">去重模式</FormLabel>
        <RadioGroup row value={config.mode} onChange={handleModeChange}>
          <FormControlLabel value="none" control={<Radio size="small" />} label="不启用" />
          <FormControlLabel value="common" control={<Radio size="small" />} label="通用字段去重" />
          <FormControlLabel value="protocol" control={<Radio size="small" />} label="按协议去重" />
        </RadioGroup>
      </FormControl>

      {/* 通用字段选择 */}
      {config.mode === 'common' && (
        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
            选择用于判断节点是否重复的字段（多选组合）：
          </Typography>
          <FormGroup row>
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
            <Typography variant="caption" color="primary" sx={{ mt: 1, display: 'block' }}>
              已选择：{config.commonFields.map((f) => nodeFieldsMeta.find((m) => m.name === f)?.label || f).join(' + ')}
            </Typography>
          )}
        </Box>
      )}

      {/* 协议特定字段选择 */}
      {config.mode === 'protocol' && (
        <Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
            为每个协议配置去重字段（未配置的协议不进行去重）：
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
      {config.mode !== 'none' && (
        <Alert variant={'standard'} severity="info" sx={{ mt: 2 }}>
          去重规则会在节点预览和订阅输出时应用，当多个节点的选定字段值完全相同时，仅保留第一个节点。
        </Alert>
      )}
    </Box>
  );
}

DeduplicationConfig.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired
};

export default DeduplicationConfig;
