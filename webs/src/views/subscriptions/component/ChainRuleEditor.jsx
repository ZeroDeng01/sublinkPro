import { useState, useEffect, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';

import ChainFlowBuilder from './ChainFlowBuilder';
import MobileChainBuilder from './MobileChainBuilder';

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
  const { t } = useTranslation();
  const [name, setName] = useState(value?.name || '');
  const [enabled, setEnabled] = useState(value?.enabled ?? true);
  const [chainConfig, setChainConfig] = useState([]);
  const [targetConfig, setTargetConfig] = useState({ type: 'specified_node', conditions: null, nodeId: undefined });

  const skipNextUpdate = useRef(false);

  useEffect(() => {
    if (value && !skipNextUpdate.current) {
      setName(value.name || '');
      setEnabled(value.enabled ?? true);

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

      if (value.targetConfig) {
        try {
          const parsed = typeof value.targetConfig === 'string' ? JSON.parse(value.targetConfig) : value.targetConfig;
          setTargetConfig({
            type: parsed.type || 'specified_node',
            conditions: parsed.conditions || null,
            nodeId: parsed.nodeId,
            endPosition: parsed.endPosition
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
          endPosition: newTargetConfig.endPosition
        })
      };
      onChange?.(output);
    },
    [onChange]
  );

  const handleNameChange = useCallback(
    (e) => {
      const newName = e.target.value;
      setName(newName);
      notifyChange(newName, enabled, chainConfig, targetConfig);
    },
    [enabled, chainConfig, targetConfig, notifyChange]
  );

  const handleEnabledChange = useCallback(
    (e) => {
      const newEnabled = e.target.checked;
      setEnabled(newEnabled);
      notifyChange(name, newEnabled, chainConfig, targetConfig);
    },
    [name, chainConfig, targetConfig, notifyChange]
  );

  const handleChainConfigChange = useCallback(
    (newConfig) => {
      setChainConfig(newConfig);
      notifyChange(name, enabled, newConfig, targetConfig);
    },
    [name, enabled, targetConfig, notifyChange]
  );

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
        <Stack direction={isMobile ? 'column' : 'row'} spacing={2} alignItems={isMobile ? 'stretch' : 'center'} sx={{ pt: 1 }}>
          <TextField
            size="small"
            label={t('subscriptions.chain.ruleName')}
            value={name}
            onChange={handleNameChange}
            sx={{ flex: 1 }}
            placeholder={t('subscriptions.chain.ruleNamePlaceholder')}
            fullWidth={isMobile}
          />
          <FormControlLabel
            control={<Switch checked={enabled} onChange={handleEnabledChange} />}
            label={t('common.enabled')}
            sx={isMobile ? { alignSelf: 'flex-start' } : {}}
          />
        </Stack>

        <Typography variant="body2" color="text.secondary">
          {isMobile ? t('subscriptions.chain.mobileDescription') : t('subscriptions.chain.desktopDescription')}
        </Typography>

        {isMobile ? (
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
