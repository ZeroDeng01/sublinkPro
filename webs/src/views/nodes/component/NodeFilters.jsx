import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';

// utils
import { isoToFlag, STATUS_OPTIONS } from '../utils';

/**
 * 节点过滤器工具栏
 */
export default function NodeFilters({
  searchQuery,
  setSearchQuery,
  groupFilter,
  setGroupFilter,
  sourceFilter,
  setSourceFilter,
  maxDelay,
  setMaxDelay,
  minSpeed,
  setMinSpeed,
  speedStatusFilter,
  setSpeedStatusFilter,
  delayStatusFilter,
  setDelayStatusFilter,
  countryFilter,
  setCountryFilter,
  tagFilter,
  setTagFilter,
  protocolFilter,
  setProtocolFilter,
  groupOptions,
  sourceOptions,
  countryOptions,
  tagOptions,
  protocolOptions,
  onReset
}) {
  return (
    <Stack direction="row" spacing={2} sx={{ mb: 2 }} flexWrap="wrap" useFlexGap>
      <FormControl size="small" sx={{ minWidth: 120 }}>
        <InputLabel>分组</InputLabel>
        <Select value={groupFilter} label="分组" onChange={(e) => setGroupFilter(e.target.value)} variant={'outlined'}>
          <MenuItem value="">全部</MenuItem>
          <MenuItem value="未分组">未分组</MenuItem>
          {groupOptions.map((group) => (
            <MenuItem key={group} value={group}>
              {group}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <TextField
        size="small"
        placeholder="搜索节点备注或链接"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        sx={{ minWidth: 200 }}
      />
      <FormControl size="small" sx={{ minWidth: 120 }}>
        <InputLabel>来源</InputLabel>
        <Select value={sourceFilter} label="来源" onChange={(e) => setSourceFilter(e.target.value)} variant={'outlined'}>
          <MenuItem value="">全部</MenuItem>
          {sourceOptions.map((source) => (
            <MenuItem key={source} value={source}>
              {source === 'manual' ? '手动添加' : source}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 120 }}>
        <InputLabel>协议</InputLabel>
        <Select value={protocolFilter} label="协议" onChange={(e) => setProtocolFilter(e.target.value)} variant={'outlined'}>
          <MenuItem value="">全部</MenuItem>
          {protocolOptions.map((protocol) => (
            <MenuItem key={protocol} value={protocol}>
              {protocol.toUpperCase()}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 100 }}>
        <InputLabel>延迟状态</InputLabel>
        <Select value={delayStatusFilter} label="延迟状态" onChange={(e) => setDelayStatusFilter(e.target.value)}>
          {STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 100 }}>
        <InputLabel>速度状态</InputLabel>
        <Select value={speedStatusFilter} label="速度状态" onChange={(e) => setSpeedStatusFilter(e.target.value)}>
          {STATUS_OPTIONS.map((opt) => (
            <MenuItem key={opt.value} value={opt.value}>
              {opt.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <TextField
        size="small"
        placeholder="最大延迟"
        type="number"
        value={maxDelay}
        onChange={(e) => setMaxDelay(e.target.value)}
        sx={{ width: 150 }}
        InputProps={{ endAdornment: <InputAdornment position="end">ms</InputAdornment> }}
      />
      <TextField
        size="small"
        placeholder="最低速度"
        type="number"
        value={minSpeed}
        onChange={(e) => setMinSpeed(e.target.value)}
        sx={{ width: 150 }}
        InputProps={{ endAdornment: <InputAdornment position="end">MB/s</InputAdornment> }}
      />
      {countryOptions.length > 0 && (
        <Autocomplete
          multiple
          size="small"
          options={countryOptions}
          value={countryFilter}
          onChange={(e, newValue) => setCountryFilter(newValue)}
          sx={{ minWidth: 150 }}
          getOptionLabel={(option) => `${isoToFlag(option)} ${option}`}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                {isoToFlag(option)} {option}
              </li>
            );
          }}
          renderTags={(value, getTagProps) =>
            value.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return <Chip key={key} label={`${isoToFlag(option)} ${option}`} size="small" {...tagProps} />;
            })
          }
          renderInput={(params) => <TextField {...params} label="国家代码" placeholder="选择国家" />}
        />
      )}
      {tagOptions && tagOptions.length > 0 && (
        <Autocomplete
          multiple
          size="small"
          options={tagOptions}
          value={tagFilter}
          onChange={(e, newValue) => setTagFilter(newValue)}
          sx={{ minWidth: 150 }}
          getOptionLabel={(option) => option.name || option}
          isOptionEqualToValue={(option, value) => option.name === (value.name || value)}
          renderOption={(props, option) => {
            const { key, ...otherProps } = props;
            return (
              <li key={key} {...otherProps}>
                <Box
                  sx={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    backgroundColor: option.color || '#1976d2',
                    mr: 1,
                    flexShrink: 0
                  }}
                />
                {option.name}
              </li>
            );
          }}
          renderTags={(value, getTagProps) =>
            value.map((option, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return (
                <Chip
                  key={key}
                  label={option.name || option}
                  size="small"
                  sx={{
                    backgroundColor: option.color || '#1976d2',
                    color: '#fff',
                    '& .MuiChip-deleteIcon': { color: 'rgba(255,255,255,0.7)' }
                  }}
                  {...tagProps}
                />
              );
            })
          }
          renderInput={(params) => <TextField {...params} label="标签" placeholder="选择标签" />}
        />
      )}
      <Button onClick={onReset}>重置</Button>
    </Stack>
  );
}

NodeFilters.propTypes = {
  searchQuery: PropTypes.string.isRequired,
  setSearchQuery: PropTypes.func.isRequired,
  groupFilter: PropTypes.string.isRequired,
  setGroupFilter: PropTypes.func.isRequired,
  sourceFilter: PropTypes.string.isRequired,
  setSourceFilter: PropTypes.func.isRequired,
  maxDelay: PropTypes.string.isRequired,
  setMaxDelay: PropTypes.func.isRequired,
  minSpeed: PropTypes.string.isRequired,
  setMinSpeed: PropTypes.func.isRequired,
  speedStatusFilter: PropTypes.string.isRequired,
  setSpeedStatusFilter: PropTypes.func.isRequired,
  delayStatusFilter: PropTypes.string.isRequired,
  setDelayStatusFilter: PropTypes.func.isRequired,
  countryFilter: PropTypes.array.isRequired,
  setCountryFilter: PropTypes.func.isRequired,
  tagFilter: PropTypes.array.isRequired,
  setTagFilter: PropTypes.func.isRequired,
  protocolFilter: PropTypes.string.isRequired,
  setProtocolFilter: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  sourceOptions: PropTypes.array.isRequired,
  countryOptions: PropTypes.array.isRequired,
  tagOptions: PropTypes.array.isRequired,
  protocolOptions: PropTypes.array.isRequired,
  onReset: PropTypes.func.isRequired
};
