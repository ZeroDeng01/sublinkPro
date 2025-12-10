import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
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
import { isoToFlag } from '../utils';

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
  countryFilter,
  setCountryFilter,
  groupOptions,
  sourceOptions,
  countryOptions,
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
  countryFilter: PropTypes.array.isRequired,
  setCountryFilter: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  sourceOptions: PropTypes.array.isRequired,
  countryOptions: PropTypes.array.isRequired,
  onReset: PropTypes.func.isRequired
};
