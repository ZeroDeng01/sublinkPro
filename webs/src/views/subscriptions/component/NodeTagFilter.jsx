import { useState } from 'react';
import PropTypes from 'prop-types';

// material-ui
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import { useTheme } from '@mui/material/styles';

// icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import BlockIcon from '@mui/icons-material/Block';

/**
 * 节点标签过滤器
 * 用于订阅设置中按标签过滤节点（白名单/黑名单模式）
 */
export default function NodeTagFilter({ tagOptions, whitelistValue, blacklistValue, onWhitelistChange, onBlacklistChange }) {
  const theme = useTheme();
  const [expanded, setExpanded] = useState(false);

  // 解析逗号分隔的标签字符串为数组
  const parseTagString = (str) => {
    if (!str) return [];
    return str
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t);
  };

  // 将数组转换为逗号分隔的字符串
  const toTagString = (arr) => {
    return arr.map((t) => t.name || t).join(',');
  };

  const whitelistTags = parseTagString(whitelistValue);
  const blacklistTags = parseTagString(blacklistValue);
  const hasAnyRules = whitelistTags.length > 0 || blacklistTags.length > 0;

  // 获取标签选项的完整信息（包含颜色）
  const getTagOption = (tagName) => {
    return (tagOptions || []).find((t) => t.name === tagName) || { name: tagName, color: '#1976d2' };
  };

  return (
    <Paper
      elevation={0}
      sx={{
        mb: 2,
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
          <LocalOfferIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600}>
            节点标签过滤
          </Typography>
          {hasAnyRules && (
            <Typography variant="caption" color="textSecondary">
              (白名单 {whitelistTags.length} / 黑名单 {blacklistTags.length} 个标签)
            </Typography>
          )}
        </Stack>
        <Stack direction="row" alignItems="center" spacing={0.5}>
          {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
        </Stack>
      </Box>

      <Collapse in={expanded} timeout="auto">
        <Box sx={{ p: 2, pt: 1 }}>
          {/* 说明 */}
          <Alert variant={'standard'} severity="info" sx={{ mb: 2 }}>
            <Typography variant="body2">
              按节点标签过滤。<strong>黑名单优先级高于白名单</strong>：匹配黑名单标签的节点会被排除，然后必须匹配白名单标签的节点才会保留。
            </Typography>
          </Alert>

          {/* 白名单标签选择 */}
          <Box sx={{ mb: 2 }}>
            <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
              <CheckCircleOutlineIcon color="success" fontSize="small" />
              <Typography variant="subtitle2" fontWeight={600} color="success.main">
                白名单标签
              </Typography>
            </Stack>
            <Autocomplete
              multiple
              options={tagOptions || []}
              getOptionLabel={(option) => option.name || option}
              value={whitelistTags.map(getTagOption)}
              onChange={(e, newValue) => onWhitelistChange(toTagString(newValue))}
              isOptionEqualToValue={(option, value) => (option.name || option) === (value.name || value)}
              filterSelectedOptions
              renderTags={(value, getTagProps) =>
                value.map((option, index) => {
                  const { key, ...tagProps } = getTagProps({ index });
                  return (
                    <Chip
                      key={key}
                      label={option.name || option}
                      size="small"
                      sx={{
                        backgroundColor: option.color || '#4caf50',
                        color: '#fff',
                        '& .MuiChip-deleteIcon': { color: 'rgba(255,255,255,0.7)' }
                      }}
                      {...tagProps}
                    />
                  );
                })
              }
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.name}>
                  <Box
                    sx={{
                      width: 12,
                      height: 12,
                      borderRadius: '50%',
                      backgroundColor: option.color || '#1976d2',
                      mr: 1
                    }}
                  />
                  {option.name}
                </Box>
              )}
              renderInput={(params) => <TextField {...params} placeholder="选择白名单标签（节点必须包含其中一个标签）" size="small" />}
            />
            <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
              节点必须包含白名单中的至少一个标签才会被保留
            </Typography>
          </Box>

          <Divider sx={{ my: 2 }} />

          {/* 黑名单标签选择 */}
          <Box>
            <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
              <BlockIcon color="error" fontSize="small" />
              <Typography variant="subtitle2" fontWeight={600} color="error.main">
                黑名单标签
              </Typography>
            </Stack>
            <Autocomplete
              multiple
              options={tagOptions || []}
              getOptionLabel={(option) => option.name || option}
              value={blacklistTags.map(getTagOption)}
              onChange={(e, newValue) => onBlacklistChange(toTagString(newValue))}
              isOptionEqualToValue={(option, value) => (option.name || option) === (value.name || value)}
              filterSelectedOptions
              renderTags={(value, getTagProps) =>
                value.map((option, index) => {
                  const { key, ...tagProps } = getTagProps({ index });
                  return (
                    <Chip
                      key={key}
                      label={option.name || option}
                      size="small"
                      sx={{
                        backgroundColor: option.color || '#f44336',
                        color: '#fff',
                        '& .MuiChip-deleteIcon': { color: 'rgba(255,255,255,0.7)' }
                      }}
                      {...tagProps}
                    />
                  );
                })
              }
              renderOption={(props, option) => (
                <Box component="li" {...props} key={option.name}>
                  <Box
                    sx={{
                      width: 12,
                      height: 12,
                      borderRadius: '50%',
                      backgroundColor: option.color || '#1976d2',
                      mr: 1
                    }}
                  />
                  {option.name}
                </Box>
              )}
              renderInput={(params) => <TextField {...params} placeholder="选择黑名单标签（包含这些标签的节点将被排除）" size="small" />}
            />
            <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
              包含黑名单标签的节点会被排除
            </Typography>
          </Box>
        </Box>
      </Collapse>
    </Paper>
  );
}

NodeTagFilter.propTypes = {
  tagOptions: PropTypes.array,
  whitelistValue: PropTypes.string,
  blacklistValue: PropTypes.string,
  onWhitelistChange: PropTypes.func.isRequired,
  onBlacklistChange: PropTypes.func.isRequired
};
