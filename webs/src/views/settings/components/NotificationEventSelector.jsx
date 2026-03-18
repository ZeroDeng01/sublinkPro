import PropTypes from 'prop-types';
import { useMemo } from 'react';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

const severityColorMap = {
  success: 'success',
  error: 'error',
  warning: 'warning',
  info: 'info'
};

const severityLabelMap = {
  success: '成功',
  error: '错误',
  warning: '警告',
  info: '信息'
};

export default function NotificationEventSelector({ value, eventOptions, onChange, disabled = false, description }) {
  const selectedKeys = value || [];

  const groupedOptions = useMemo(() => {
    const groups = new Map();
    (eventOptions || []).forEach((event) => {
      const categoryKey = event.category || 'other';
      if (!groups.has(categoryKey)) {
        groups.set(categoryKey, {
          key: categoryKey,
          name: event.categoryName || '其他事件',
          events: []
        });
      }
      groups.get(categoryKey).events.push(event);
    });
    return Array.from(groups.values());
  }, [eventOptions]);

  const importantKeys = useMemo(
    () => (eventOptions || []).filter((event) => event.severity === 'error' || event.category === 'security').map((event) => event.key),
    [eventOptions]
  );

  const selectedCount = selectedKeys.length;
  const totalCount = eventOptions?.length || 0;

  const updateKeys = (nextKeys) => {
    if (disabled) return;
    onChange(nextKeys);
  };

  const toggleEvent = (eventKey) => {
    const currentSet = new Set(selectedKeys);
    if (currentSet.has(eventKey)) {
      currentSet.delete(eventKey);
    } else {
      currentSet.add(eventKey);
    }
    updateKeys((eventOptions || []).map((event) => event.key).filter((key) => currentSet.has(key)));
  };

  const toggleGroup = (group, checked) => {
    const currentSet = new Set(selectedKeys);
    group.events.forEach((event) => {
      if (checked) {
        currentSet.add(event.key);
      } else {
        currentSet.delete(event.key);
      }
    });
    updateKeys((eventOptions || []).map((event) => event.key).filter((key) => currentSet.has(key)));
  };

  return (
    <Stack spacing={2}>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} alignItems={{ sm: 'center' }} justifyContent="space-between">
        <Box>
          <Typography variant="subtitle1">触发事件</Typography>
          <Typography variant="body2" color="text.secondary">
            {description || '为当前通知渠道勾选要自动发送的事件，测试发送不受这里的选择影响。'}
          </Typography>
        </Box>
        <Chip
          color={selectedCount === 0 ? 'default' : 'primary'}
          variant={selectedCount === totalCount && totalCount > 0 ? 'filled' : 'outlined'}
          label={`${selectedCount}/${totalCount} 已选`}
        />
      </Stack>

      <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
        <Button
          size="small"
          variant="outlined"
          disabled={disabled || totalCount === 0}
          onClick={() => updateKeys((eventOptions || []).map((event) => event.key))}
        >
          全选
        </Button>
        <Button size="small" variant="outlined" disabled={disabled || importantKeys.length === 0} onClick={() => updateKeys(importantKeys)}>
          仅错误与安全
        </Button>
        <Button size="small" color="inherit" disabled={disabled || selectedCount === 0} onClick={() => updateKeys([])}>
          清空
        </Button>
      </Stack>

      {selectedCount === 0 && (
        <Alert severity="warning" variant="outlined">
          当前没有选中任何自动触发事件。保存后渠道仍可启用，但不会自动推送业务通知。
        </Alert>
      )}

      {groupedOptions.map((group) => {
        const groupSelectedCount = group.events.filter((event) => selectedKeys.includes(event.key)).length;
        const groupChecked = group.events.length > 0 && groupSelectedCount === group.events.length;
        const groupIndeterminate = groupSelectedCount > 0 && groupSelectedCount < group.events.length;

        return (
          <Paper key={group.key} variant="outlined" sx={{ overflow: 'hidden' }}>
            <Stack
              direction={{ xs: 'column', md: 'row' }}
              spacing={1}
              alignItems={{ md: 'center' }}
              justifyContent="space-between"
              sx={{ px: 2, py: 1.5, backgroundColor: 'rgba(0,0,0,0.02)' }}
            >
              <Box>
                <Typography variant="subtitle2">{group.name}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {groupSelectedCount}/{group.events.length} 个事件已选
                </Typography>
              </Box>
              <FormControlLabel
                sx={{ mr: 0 }}
                label="全选分组"
                control={
                  <Checkbox
                    checked={groupChecked}
                    indeterminate={groupIndeterminate}
                    disabled={disabled}
                    onChange={(event) => toggleGroup(group, event.target.checked)}
                  />
                }
              />
            </Stack>

            <Divider />

            <Stack spacing={1.25} sx={{ p: 2 }}>
              {group.events.map((event) => {
                const checked = selectedKeys.includes(event.key);

                return (
                  <Box
                    key={event.key}
                    role="button"
                    tabIndex={disabled ? -1 : 0}
                    onClick={() => toggleEvent(event.key)}
                    onKeyDown={(keyboardEvent) => {
                      if (disabled) return;
                      if (keyboardEvent.key === 'Enter' || keyboardEvent.key === ' ') {
                        keyboardEvent.preventDefault();
                        toggleEvent(event.key);
                      }
                    }}
                    sx={{
                      border: '1px solid',
                      borderColor: checked ? 'primary.main' : 'divider',
                      borderRadius: 2,
                      p: 1.5,
                      cursor: disabled ? 'default' : 'pointer',
                      backgroundColor: checked ? 'rgba(25, 118, 210, 0.06)' : 'transparent',
                      transition: 'all 0.2s ease'
                    }}
                  >
                    <Stack direction="row" spacing={1.5} alignItems="flex-start">
                      <Checkbox
                        checked={checked}
                        disabled={disabled}
                        onClick={(event) => event.stopPropagation()}
                        onChange={() => toggleEvent(event.key)}
                        sx={{ mt: -0.75, ml: -0.75 }}
                      />
                      <Box sx={{ flexGrow: 1 }}>
                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ sm: 'center' }} useFlexGap>
                          <Typography variant="body2" sx={{ fontWeight: 600 }}>
                            {event.name}
                          </Typography>
                          <Chip
                            size="small"
                            color={severityColorMap[event.severity] || 'default'}
                            variant="outlined"
                            label={severityLabelMap[event.severity] || '事件'}
                          />
                        </Stack>
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                          {event.description}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 0.75, fontFamily: 'monospace' }}>
                          {event.key}
                        </Typography>
                      </Box>
                    </Stack>
                  </Box>
                );
              })}
            </Stack>
          </Paper>
        );
      })}
    </Stack>
  );
}

NotificationEventSelector.propTypes = {
  description: PropTypes.string,
  disabled: PropTypes.bool,
  eventOptions: PropTypes.arrayOf(
    PropTypes.shape({
      category: PropTypes.string,
      categoryName: PropTypes.string,
      description: PropTypes.string,
      key: PropTypes.string,
      name: PropTypes.string,
      severity: PropTypes.string
    })
  ),
  onChange: PropTypes.func.isRequired,
  value: PropTypes.arrayOf(PropTypes.string)
};
