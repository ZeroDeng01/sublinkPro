import { useState, useMemo } from 'react';
import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import TextField from '@mui/material/TextField';
import CircularProgress from '@mui/material/CircularProgress';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

/**
 * 可搜索的节点选择组件
 * 初始只加载前N个节点，其他需通过搜索查找
 */
export default function SearchableNodeSelect({
  nodes = [],
  loading = false,
  value = null,
  onChange,
  displayField = 'Name',
  valueField = 'Link',
  label = '选择节点',
  placeholder = '搜索节点...',
  helperText = '',
  freeSolo = false,
  limit = 50,
  disabled = false,
  ...props
}) {
  const [inputValue, setInputValue] = useState('');

  // 获取初始显示的节点（前N个）
  const limitedNodes = useMemo(() => {
    return nodes.slice(0, limit);
  }, [nodes, limit]);

  // 根据搜索过滤节点
  const filteredOptions = useMemo(() => {
    if (!inputValue) {
      return limitedNodes;
    }

    const searchLower = inputValue.toLowerCase();
    const filtered = nodes.filter((node) => {
      const name = (node[displayField] || '').toLowerCase();
      const link = (node.Link || '').toLowerCase();
      const group = (node.Group || '').toLowerCase();
      return name.includes(searchLower) || link.includes(searchLower) || group.includes(searchLower);
    });

    // 返回搜索结果，限制数量
    return filtered.slice(0, limit);
  }, [nodes, inputValue, displayField, limit, limitedNodes]);

  // 确保当前选中的值在选项中
  const optionsWithSelected = useMemo(() => {
    if (!value) return filteredOptions;

    // 检查当前值是否已在选项中
    const isInOptions = filteredOptions.some((opt) => opt[valueField] === (typeof value === 'string' ? value : value[valueField]));

    if (isInOptions) return filteredOptions;

    // 如果当前值不在选项中，将其添加到开头
    if (typeof value === 'object' && value !== null) {
      return [value, ...filteredOptions];
    }

    // 如果是字符串值，创建一个临时对象
    if (typeof value === 'string') {
      const nodeFromFull = nodes.find((n) => n[valueField] === value);
      if (nodeFromFull) {
        return [nodeFromFull, ...filteredOptions];
      }
    }

    return filteredOptions;
  }, [filteredOptions, value, valueField, nodes]);

  // 是否有更多节点未显示
  const hasMoreNodes = nodes.length > limit;
  const hiddenCount = nodes.length - limit;

  return (
    <Autocomplete
      freeSolo={freeSolo}
      options={optionsWithSelected}
      loading={loading}
      disabled={disabled}
      getOptionLabel={(option) => {
        if (typeof option === 'string') return option;
        return option[displayField] || option[valueField] || '';
      }}
      value={value}
      inputValue={inputValue}
      onInputChange={(event, newInputValue) => {
        setInputValue(newInputValue);
      }}
      onChange={(event, newValue) => {
        onChange?.(newValue);
      }}
      onBlur={() => {
        // freeSolo 模式下，失焦时如果 inputValue 与当前 value 不同，则同步给父组件
        if (freeSolo) {
          const currentValueStr = typeof value === 'string' ? value : value?.[displayField] || '';
          if (inputValue !== currentValueStr) {
            // 如果 inputValue 为空，传递空字符串；否则传递输入的内容
            onChange?.(inputValue || '');
          }
        }
      }}
      isOptionEqualToValue={(option, value) => {
        if (!option || !value) return false;
        if (typeof option === 'string' || typeof value === 'string') {
          return option === value;
        }
        return option[valueField] === value[valueField];
      }}
      noOptionsText={inputValue ? '未找到匹配的节点' : '输入关键词搜索节点'}
      ListboxProps={{
        sx: {
          maxHeight: 300,
          '& .MuiAutocomplete-option:last-child':
            hasMoreNodes && !inputValue
              ? {
                  borderTop: '1px dashed',
                  borderColor: 'divider'
                }
              : {}
        }
      }}
      renderOption={(props, option, { index }) => {
        const isLastItem = index === optionsWithSelected.length - 1;
        return (
          <>
            <Box component="li" {...props} key={option.ID || option[valueField]}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                <Typography variant="body2" noWrap sx={{ maxWidth: '60%' }}>
                  {option[displayField] || '未知'}
                </Typography>
                <Typography variant="caption" color="textSecondary" sx={{ ml: 2 }}>
                  {option.Group || '未分组'}
                </Typography>
              </Box>
            </Box>
            {/* 在列表末尾显示更多节点提示 */}
            {isLastItem && hasMoreNodes && !inputValue && (
              <Box
                sx={{
                  px: 2,
                  py: 1.5,
                  bgcolor: 'action.hover',
                  borderTop: '1px solid',
                  borderColor: 'divider',
                  textAlign: 'center'
                }}
              >
                <Typography variant="caption" color="primary" sx={{ fontWeight: 500 }}>
                  💡 还有 {hiddenCount} 个节点未显示，请输入关键词搜索
                </Typography>
              </Box>
            )}
          </>
        );
      }}
      renderInput={(params) => (
        <TextField
          {...params}
          label={label}
          placeholder={placeholder}
          helperText={
            helperText ||
            (hasMoreNodes ? (
              <Typography component="span" variant="caption" color="primary" sx={{ fontWeight: 500 }}>
                ⚠️ 仅显示前 {limit} 个节点（共 {nodes.length} 个），输入关键词搜索更多
              </Typography>
            ) : (
              ''
            ))
          }
          InputProps={{
            ...params.InputProps,
            endAdornment: (
              <>
                {loading ? <CircularProgress color="inherit" size={20} /> : null}
                {params.InputProps.endAdornment}
              </>
            )
          }}
        />
      )}
      {...props}
    />
  );
}

SearchableNodeSelect.propTypes = {
  nodes: PropTypes.array,
  loading: PropTypes.bool,
  value: PropTypes.oneOfType([PropTypes.object, PropTypes.string]),
  onChange: PropTypes.func,
  displayField: PropTypes.string,
  valueField: PropTypes.string,
  label: PropTypes.string,
  placeholder: PropTypes.string,
  helperText: PropTypes.string,
  freeSolo: PropTypes.bool,
  limit: PropTypes.number,
  disabled: PropTypes.bool
};
