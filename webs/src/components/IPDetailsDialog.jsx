import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';

// material-ui
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import Link from '@mui/material/Link';
import Tooltip from '@mui/material/Tooltip';

// icons
import CloseIcon from '@mui/icons-material/Close';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import PublicIcon from '@mui/icons-material/Public';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import BusinessIcon from '@mui/icons-material/Business';
import DnsIcon from '@mui/icons-material/Dns';
import CachedIcon from '@mui/icons-material/Cached';

// localStorage 缓存键名
const IP_CACHE_KEY = 'sublink_ip_info_cache';

// 缓存有效期：7天
const CACHE_TTL_MS = 7 * 24 * 60 * 60 * 1000;

/**
 * 从 localStorage 读取 IP 缓存
 */
const getIPCache = () => {
  try {
    const cached = localStorage.getItem(IP_CACHE_KEY);
    return cached ? JSON.parse(cached) : {};
  } catch {
    return {};
  }
};

/**
 * 将 IP 缓存写入 localStorage
 */
const setIPCache = (cache) => {
  try {
    localStorage.setItem(IP_CACHE_KEY, JSON.stringify(cache));
  } catch {
    // localStorage 满了或不可用时静默失败
  }
};

/**
 * 获取指定 IP 的缓存数据
 */
const getCachedIPInfo = (ip) => {
  const cache = getIPCache();
  const entry = cache[ip];
  if (!entry) return null;
  // 检查是否过期
  if (Date.now() - entry.timestamp > CACHE_TTL_MS) {
    // 清理过期缓存
    delete cache[ip];
    setIPCache(cache);
    return null;
  }
  return entry.data;
};

/**
 * 缓存 IP 信息
 */
const cacheIPInfo = (ip, data) => {
  const cache = getIPCache();
  cache[ip] = { data, timestamp: Date.now() };
  // 限制缓存数量，防止 localStorage 膨胀（最多保留100条）
  const keys = Object.keys(cache);
  if (keys.length > 100) {
    // 删除最早的缓存
    const sortedKeys = keys.sort((a, b) => cache[a].timestamp - cache[b].timestamp);
    sortedKeys.slice(0, keys.length - 100).forEach((key) => delete cache[key]);
  }
  setIPCache(cache);
};

/**
 * IP详情弹窗组件
 * 通过第三方API查询IP详细信息，支持缓存避免重复请求
 */
export default function IPDetailsDialog({ open, onClose, ip, onCopy }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [ipInfo, setIpInfo] = useState(null);
  const [fromCache, setFromCache] = useState(false);

  const fetchIPDetails = useCallback(async () => {
    if (!ip) return;

    // 检查 localStorage 缓存
    const cachedData = getCachedIPInfo(ip);
    if (cachedData) {
      setIpInfo(cachedData);
      setFromCache(true);
      setError(null);
      setLoading(false);
      return;
    }

    // 缓存未命中或已过期，发起后端请求
    setLoading(true);
    setError(null);
    setFromCache(false);

    try {
      // 调用后端API（后端会处理多级缓存和第三方API调用）
      const { getIPDetails } = await import('api/nodes');
      const response = await getIPDetails(ip);

      // 成功（code === 200 时返回，否则被拦截器 reject）
      const data = response.data;
      // 转换为前端需要的格式（保持字段名一致性）
      const ipData = {
        status: 'success',
        query: data.ip,
        country: data.country,
        countryCode: data.countryCode,
        region: data.region,
        regionName: data.regionName,
        city: data.city,
        zip: data.zip,
        lat: data.lat,
        lon: data.lon,
        timezone: data.timezone,
        isp: data.isp,
        org: data.org,
        as: data.as,
        provider: data.provider
      };
      // 缓存成功结果到 localStorage
      cacheIPInfo(ip, ipData);
      setIpInfo(ipData);
    } catch (err) {
      // 业务错误或网络错误都通过 err.message 获取
      setError(err.message || '查询IP信息失败');
    } finally {
      setLoading(false);
    }
  }, [ip]);

  useEffect(() => {
    if (open && ip) {
      fetchIPDetails();
    }
  }, [open, ip, fetchIPDetails]);

  const handleCopy = () => {
    if (ip && onCopy) {
      onCopy(ip);
    }
  };

  // 格式化IP显示（IPv6截断）
  // const formatIP = (ipAddr) => {
  //   if (!ipAddr) return '-';
  //   // IPv6地址通常很长，截断显示
  //   if (ipAddr.includes(':') && ipAddr.length > 25) {
  //     return ipAddr.substring(0, 22) + '...';
  //   }
  //   return ipAddr;
  // };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ m: 0, p: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Stack direction="row" spacing={1} alignItems="center">
          <PublicIcon color="primary" />
          <Typography variant="h6">IP 详情</Typography>
          {/* 缓存状态指示 */}
          {fromCache && ipInfo && (
            <Tooltip title="数据来自缓存，点击刷新获取最新信息">
              <Chip
                icon={<CachedIcon sx={{ fontSize: '14px !important' }} />}
                label="已缓存"
                size="small"
                variant="outlined"
                color="success"
                onClick={() => {
                  // 清除当前IP的缓存并重新获取
                  const cache = getIPCache();
                  delete cache[ip];
                  setIPCache(cache);
                  fetchIPDetails();
                }}
                sx={{
                  height: 22,
                  fontSize: 11,
                  cursor: 'pointer',
                  '&:hover': { bgcolor: 'success.lighter' }
                }}
              />
            </Tooltip>
          )}
        </Stack>
        <IconButton aria-label="close" onClick={onClose} size="small">
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <Divider />
      <DialogContent>
        {/* 数据来源提示 */}
        <Box sx={{ mb: 2, p: 1.5, bgcolor: 'info.lighter', borderRadius: 1, border: '1px solid', borderColor: 'info.light' }}>
          <Typography variant="caption" color="info.main" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <span>ℹ️</span>
            IP信息由{' '}
            <Link href="http://ip-api.com" target="_blank" rel="noopener noreferrer" sx={{ fontWeight: 'bold' }}>
              ip-api.com
            </Link>{' '}
            提供（免费服务，数据仅供参考）
          </Typography>
        </Box>

        {/* IP地址显示 */}
        <Box sx={{ mb: 2, p: 2, bgcolor: 'action.hover', borderRadius: 2 }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Stack direction="row" spacing={1} alignItems="center" sx={{ minWidth: 0, flex: 1 }}>
              <DnsIcon color="primary" fontSize="small" />
              <Typography
                variant="body1"
                fontWeight="bold"
                sx={{
                  fontFamily: 'monospace',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}
                title={ip}
              >
                {ip || '-'}
              </Typography>
            </Stack>
            <IconButton size="small" onClick={handleCopy} title="复制IP">
              <ContentCopyIcon fontSize="small" />
            </IconButton>
          </Stack>
        </Box>

        {/* 加载状态 */}
        {loading && (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        )}

        {/* 错误状态 */}
        {error && !loading && (
          <Box sx={{ textAlign: 'center', py: 3 }}>
            <Typography color="error" gutterBottom>
              {error}
            </Typography>
            <Link href={`https://ipinfo.io/${ip}`} target="_blank" rel="noopener noreferrer" sx={{ fontSize: '0.875rem' }}>
              在 ipinfo.io 查看 →
            </Link>
          </Box>
        )}

        {/* IP信息详情 */}
        {ipInfo && !loading && (
          <Stack spacing={2}>
            {/* 位置信息 */}
            <Box>
              <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
                <LocationOnIcon color="secondary" fontSize="small" />
                <Typography variant="subtitle2" color="textSecondary">
                  位置信息
                </Typography>
              </Stack>
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                <Chip label={ipInfo.country || '未知'} size="small" color="primary" variant="outlined" />
                <Chip label={ipInfo.regionName || ipInfo.region || '未知'} size="small" variant="outlined" />
                <Chip label={ipInfo.city || '未知'} size="small" variant="outlined" />
                {ipInfo.zip && <Chip label={`邮编: ${ipInfo.zip}`} size="small" variant="outlined" />}
              </Stack>
            </Box>

            {/* 运营商信息 */}
            <Box>
              <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
                <BusinessIcon color="info" fontSize="small" />
                <Typography variant="subtitle2" color="textSecondary">
                  运营商信息
                </Typography>
              </Stack>
              <Stack spacing={0.5}>
                {ipInfo.isp && (
                  <Typography variant="body2">
                    <strong>ISP:</strong> {ipInfo.isp}
                  </Typography>
                )}
                {ipInfo.org && (
                  <Typography variant="body2">
                    <strong>组织:</strong> {ipInfo.org}
                  </Typography>
                )}
                {ipInfo.as && (
                  <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                    <strong>AS:</strong> {ipInfo.as}
                  </Typography>
                )}
              </Stack>
            </Box>

            {/* 其他信息 */}
            {(ipInfo.timezone || (ipInfo.lat && ipInfo.lon)) && (
              <Box>
                <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1 }}>
                  其他信息
                </Typography>
                <Stack spacing={0.5}>
                  {ipInfo.timezone && (
                    <Typography variant="body2">
                      <strong>时区:</strong> {ipInfo.timezone}
                    </Typography>
                  )}
                  {ipInfo.lat && ipInfo.lon && (
                    <Typography variant="body2">
                      <strong>坐标:</strong> {ipInfo.lat}, {ipInfo.lon}
                    </Typography>
                  )}
                </Stack>
              </Box>
            )}

            {/* 外部链接 */}
            <Box sx={{ pt: 1, borderTop: '1px solid', borderColor: 'divider' }}>
              <Stack direction="row" spacing={2}>
                <Link href={`https://ipinfo.io/${ip}`} target="_blank" rel="noopener noreferrer" sx={{ fontSize: '0.75rem' }}>
                  ipinfo.io 详情 →
                </Link>
                <Link href={`https://bgp.he.net/ip/${ip}`} target="_blank" rel="noopener noreferrer" sx={{ fontSize: '0.75rem' }}>
                  BGP 查询 →
                </Link>
                <Link href={`https://ippure.com/?ip=${ip}`} target="_blank" rel="noopener noreferrer" sx={{ fontSize: '0.75rem' }}>
                  IP Pure 详情 →
                </Link>
              </Stack>
            </Box>
          </Stack>
        )}
      </DialogContent>
    </Dialog>
  );
}

IPDetailsDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  ip: PropTypes.string,
  onCopy: PropTypes.func
};
