import { useState, useEffect } from 'react';
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

// icons
import CloseIcon from '@mui/icons-material/Close';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import PublicIcon from '@mui/icons-material/Public';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import BusinessIcon from '@mui/icons-material/Business';
import DnsIcon from '@mui/icons-material/Dns';

/**
 * IP详情弹窗组件
 * 通过第三方API查询IP详细信息
 */
export default function IPDetailsDialog({ open, onClose, ip, onCopy }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [ipInfo, setIpInfo] = useState(null);

  useEffect(() => {
    if (open && ip) {
      fetchIPDetails();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, ip]);

  const fetchIPDetails = async () => {
    if (!ip) return;
    setLoading(true);
    setError(null);
    try {
      // 使用 ip-api.com 的 JSON API (免费，支持CORS)
      const response = await fetch(
        `http://ip-api.com/json/${ip}?lang=zh-CN&fields=status,message,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query`
      );
      const data = await response.json();
      if (data.status === 'success') {
        setIpInfo(data);
      } else {
        setError(data.message || '查询IP信息失败');
      }
    } catch (err) {
      setError('网络请求失败: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

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
