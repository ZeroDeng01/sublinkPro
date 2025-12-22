import { useState, useEffect, useCallback, useRef } from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';

import ChainFlowBuilder from './ChainFlowBuilder';
import MobileChainBuilder from './MobileChainBuilder';

/**
 * 单条链式代理规则编辑器
 * 使用画板式交互配置代理链和目标节点
 * 移动端使用简化的卡片式配置
 */
export default function ChainRuleEditor({
  value,
  onChange,
  nodes = [],
  fields = [],
  operators = [],
  groupTypes = [],
  templateGroups = [],
  isMobile = false
}) {
  const [name, setName] = useState(value?.name || '');
  const [enabled, setEnabled] = useState(value?.enabled ?? true);
  const [chainConfig, setChainConfig] = useState([]);
  const [targetConfig, setTargetConfig] = useState({ type: 'specified_node', conditions: null, nodeId: undefined });

  // 使用 ref 跟踪更新来源
  const skipNextUpdate = useRef(false);

  // 解析外部数据
  useEffect(() => {
    if (value && !skipNextUpdate.current) {
      setName(value.name || '');
      setEnabled(value.enabled ?? true);

      // 解析 chainConfig
      if (value.chainConfig) {
        try {
          const parsed = typeof value.chainConfig === 'string' ? JSON.parse(value.chainConfig) : value.chainConfig;
          setChainConfig(Array.isArray(parsed) ? parsed : []);
        } catch {
          setChainConfig([]);
        }
      } else {
        setChainConfig([]);
      }

      // 解析 targetConfig
      if (value.targetConfig) {
        try {
          const parsed = typeof value.targetConfig === 'string' ? JSON.parse(value.targetConfig) : value.targetConfig;
          setTargetConfig({
            type: parsed.type || 'specified_node',
            conditions: parsed.conditions || null,
            nodeId: parsed.nodeId,
            endPosition: parsed.endPosition // 恢复结束节点位置
          });
        } catch {
          setTargetConfig({ type: 'specified_node', conditions: null, nodeId: undefined });
        }
      } else {
        setTargetConfig({ type: 'specified_node', conditions: null, nodeId: undefined });
      }
    }
    skipNextUpdate.current = false;
  }, [value]);

  // 通知父组件数据变化
  const notifyChange = useCallback(
    (newName, newEnabled, newChainConfig, newTargetConfig) => {
      skipNextUpdate.current = true;
      const output = {
        name: newName,
        enabled: newEnabled,
        chainConfig: JSON.stringify(newChainConfig),
        targetConfig: JSON.stringify({
          type: newTargetConfig.type,
          conditions: newTargetConfig.type === 'conditions' ? newTargetConfig.conditions : undefined,
          nodeId: newTargetConfig.type === 'specified_node' ? newTargetConfig.nodeId : undefined,
          endPosition: newTargetConfig.endPosition // 保存结束节点位置
        })
      };
      onChange?.(output);
    },
    [onChange]
  );

  // 名称变化
  const handleNameChange = useCallback(
    (e) => {
      const newName = e.target.value;
      setName(newName);
      notifyChange(newName, enabled, chainConfig, targetConfig);
    },
    [enabled, chainConfig, targetConfig, notifyChange]
  );

  // 启用状态变化
  const handleEnabledChange = useCallback(
    (e) => {
      const newEnabled = e.target.checked;
      setEnabled(newEnabled);
      notifyChange(name, newEnabled, chainConfig, targetConfig);
    },
    [name, chainConfig, targetConfig, notifyChange]
  );

  // 代理链配置变化
  const handleChainConfigChange = useCallback(
    (newConfig) => {
      setChainConfig(newConfig);
      notifyChange(name, enabled, newConfig, targetConfig);
    },
    [name, enabled, targetConfig, notifyChange]
  );

  // 目标配置变化
  const handleTargetConfigChange = useCallback(
    (newConfig) => {
      setTargetConfig(newConfig);
      notifyChange(name, enabled, chainConfig, newConfig);
    },
    [name, enabled, chainConfig, notifyChange]
  );

  return (
    <Box sx={{ height: '100%' }}>
      <Stack spacing={2}>
        {/* 基本信息 - 移动端堆叠显示 */}
        <Stack direction={isMobile ? 'column' : 'row'} spacing={2} alignItems={isMobile ? 'stretch' : 'center'} sx={{ pt: 1 }}>
          <TextField
            size="small"
            label="规则名称"
            value={name}
            onChange={handleNameChange}
            sx={{ flex: 1 }}
            placeholder="例如：香港中转美国"
            fullWidth={isMobile}
          />
          <FormControlLabel
            control={<Switch checked={enabled} onChange={handleEnabledChange} />}
            label="启用"
            sx={isMobile ? { alignSelf: 'flex-start' } : {}}
          />
        </Stack>

        {/* 画板式代理链配置 - 移动端使用简化版 */}
        <Typography variant="body2" color="text.secondary">
          {isMobile
            ? '配置多级代理链路：入口 → 中间节点 → 目标节点'
            : '点击「添加代理节点」构建多级代理链，点击节点进行编辑。建议链路不超过 3 级。'}
        </Typography>

        {isMobile ? (
          // 移动端使用简化的卡片式配置
          <MobileChainBuilder
            chainConfig={chainConfig}
            targetConfig={targetConfig}
            onChainConfigChange={handleChainConfigChange}
            onTargetConfigChange={handleTargetConfigChange}
            nodes={nodes}
            fields={fields}
            operators={operators}
            groupTypes={groupTypes}
            templateGroups={templateGroups}
          />
        ) : (
          // 桌面端使用流程图画板
          <ChainFlowBuilder
            chainConfig={chainConfig}
            targetConfig={targetConfig}
            onChainConfigChange={handleChainConfigChange}
            onTargetConfigChange={handleTargetConfigChange}
            nodes={nodes}
            fields={fields}
            operators={operators}
            groupTypes={groupTypes}
            templateGroups={templateGroups}
          />
        )}
      </Stack>
    </Box>
  );
}
