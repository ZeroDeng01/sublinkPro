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
import RouterIcon from '@mui/icons-material/Router';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import BlockIcon from '@mui/icons-material/Block';

// 协议颜色映射
const protocolColors = {
  vmess: '#1976d2',
  vless: '#7b1fa2',
  trojan: '#d32f2f',
  ss: '#2e7d32',
  ssr: '#e64a19',
  hysteria: '#f9a825',
  hysteria2: '#ef6c00',
  tuic: '#0277bd',
  wireguard: '#88171a',
  naiveproxy: '#5d4037',
  anytls: '#20a84c',
  socks5: '#116ea4',
  socks: '#dd4984'
};

// 协议显示名称映射
const protocolLabels = {
  vmess: 'VMess',
  vless: 'VLESS',
  trojan: 'Trojan',
  ss: 'SS',
  ssr: 'SSR',
  hysteria: 'Hysteria',
  hysteria2: 'Hysteria2',
  tuic: 'TUIC',
  wireguard: 'WireGuard',
  naiveproxy: 'NaiveProxy',
  anytls: 'AnyTLS',
  socks5: 'SOCKS5',
  socks: 'SOCKS'
};

/**
 * 节点协议过滤器
 * 用于订阅设置中按协议类型过滤节点（白名单/黑名单模式）
 */
export default function NodeProtocolFilter({ protocolOptions, whitelistValue, blacklistValue, onWhitelistChange, onBlacklistChange }) {
  const theme = useTheme();
  const [expanded, setExpanded] = useState(false);

  // 解析逗号分隔的协议字符串为数组
  const parseProtocolString = (str) => {
    if (!str) return [];
    return str
      .split(',')
      .map((p) => p.trim().toLowerCase())
      .filter((p) => p);
  };

  // 将数组转换为逗号分隔的字符串
  const toProtocolString = (arr) => {
    return arr.join(',');
  };

  const whitelistProtocols = parseProtocolString(whitelistValue);
  const blacklistProtocols = parseProtocolString(blacklistValue);
  const hasAnyRules = whitelistProtocols.length > 0 || blacklistProtocols.length > 0;

  // 获取协议选项（确保有颜色和显示名称）
  const getProtocolOption = (protocol) => {
    const p = protocol.toLowerCase();
    return {
      value: p,
      label: protocolLabels[p] || p.toUpperCase(),
      color: protocolColors[p] || '#757575'
    };
  };

  // 构建选项列表
  const options = (protocolOptions || []).map((p) => getProtocolOption(p));

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
          <RouterIcon color="primary" fontSize="small" />
          <Typography variant="subtitle2" fontWeight={600}>
            协议类型过滤
          </Typography>
          {hasAnyRules && (
            <Typography variant="caption" color="textSecondary">
              (白名单 {whitelistProtocols.length} / 黑名单 {blacklistProtocols.length} 个协议)
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
              按节点协议类型过滤。<strong>黑名单优先级高于白名单</strong>：黑名单协议的节点会被排除，剩余节点必须匹配白名单协议才会保留。
            </Typography>
          </Alert>

          {/* 白名单协议选择 */}
          <Box sx={{ mb: 2 }}>
            <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
              <CheckCircleOutlineIcon color="success" fontSize="small" />
              <Typography variant="subtitle2" fontWeight={600} color="success.main">
                白名单协议
              </Typography>
            </Stack>
            <Autocomplete
              multiple
              options={options}
              getOptionLabel={(option) => option.label || option}
              value={whitelistProtocols.map(getProtocolOption)}
              onChange={(e, newValue) => onWhitelistChange(toProtocolString(newValue.map((v) => v.value || v)))}
              isOptionEqualToValue={(option, value) => (option.value || option) === (value.value || value)}
              filterSelectedOptions
              renderTags={(value, getTagProps) =>
                value.map((option, index) => {
                  const { key, ...tagProps } = getTagProps({ index });
                  return (
                    <Chip
                      key={key}
                      label={option.label || option}
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
                <Box component="li" {...props} key={option.value}>
                  <Box
                    sx={{
                      width: 12,
                      height: 12,
                      borderRadius: '50%',
                      backgroundColor: option.color || '#1976d2',
                      mr: 1
                    }}
                  />
                  {option.label}
                </Box>
              )}
              renderInput={(params) => <TextField {...params} placeholder="选择白名单协议（只保留这些协议的节点）" size="small" />}
            />
            <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
              仅保留使用白名单协议的节点
            </Typography>
          </Box>

          <Divider sx={{ my: 2 }} />

          {/* 黑名单协议选择 */}
          <Box>
            <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1 }}>
              <BlockIcon color="error" fontSize="small" />
              <Typography variant="subtitle2" fontWeight={600} color="error.main">
                黑名单协议
              </Typography>
            </Stack>
            <Autocomplete
              multiple
              options={options}
              getOptionLabel={(option) => option.label || option}
              value={blacklistProtocols.map(getProtocolOption)}
              onChange={(e, newValue) => onBlacklistChange(toProtocolString(newValue.map((v) => v.value || v)))}
              isOptionEqualToValue={(option, value) => (option.value || option) === (value.value || value)}
              filterSelectedOptions
              renderTags={(value, getTagProps) =>
                value.map((option, index) => {
                  const { key, ...tagProps } = getTagProps({ index });
                  return (
                    <Chip
                      key={key}
                      label={option.label || option}
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
                <Box component="li" {...props} key={option.value}>
                  <Box
                    sx={{
                      width: 12,
                      height: 12,
                      borderRadius: '50%',
                      backgroundColor: option.color || '#1976d2',
                      mr: 1
                    }}
                  />
                  {option.label}
                </Box>
              )}
              renderInput={(params) => <TextField {...params} placeholder="选择黑名单协议（排除这些协议的节点）" size="small" />}
            />
            <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
              使用黑名单协议的节点将被排除
            </Typography>
          </Box>
        </Box>
      </Collapse>
    </Paper>
  );
}

NodeProtocolFilter.propTypes = {
  protocolOptions: PropTypes.array,
  whitelistValue: PropTypes.string,
  blacklistValue: PropTypes.string,
  onWhitelistChange: PropTypes.func.isRequired,
  onBlacklistChange: PropTypes.func.isRequired
};
