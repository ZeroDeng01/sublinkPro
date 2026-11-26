import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { visuallyHidden } from '@mui/utils';
import ContentCopy from '@mui/icons-material/ContentCopy';
import LockOutlined from '@mui/icons-material/LockOutlined';
import LockOpenOutlined from '@mui/icons-material/LockOpenOutlined';

import { copyText } from 'utils/clipboard';
import { requiresSharedCredentials } from '../gatewaySettings';

const wildcardHosts = new Set(['0.0.0.0', '::', '0:0:0:0:0:0:0:0']);

function endpointFor(listener) {
  const bound = listener.boundAddress || '';
  const separator = bound.lastIndexOf(':');
  const host = (separator > 0 ? bound.slice(0, separator) : listener.listenAddress).replace(/^\[|\]$/g, '');
  const port = separator > 0 ? bound.slice(separator + 1) : listener.port;
  return {
    port,
    wildcard: wildcardHosts.has(host),
    loopback: /^(127\.|::ffff:127\.)/i.test(host) || ['localhost', '::1', '0:0:0:0:0:0:0:1'].includes(host.toLowerCase()),
    address: host.includes(':') ? `[${host}]:${port}` : `${host}:${port}`
  };
}

export default function GatewayOverview({ status, settings, goToTab, onStart, onStop, busy, showMessage }) {
  const { t } = useTranslation();
  const label = (key) => t(`settings.socks5.page.${key}`);
  if (!status) return null;

  const running = status.listeners.filter((listener) => listener.running);
  const hasEnabledListener = status.listeners.some((listener) => listener.enabled);
  const needsAuth = requiresSharedCredentials(settings, status.listeners) && (!settings.username || !settings.hasPassword);

  const copyAddress = async (address) => {
    const copied = await copyText(address);
    showMessage(label(copied ? 'copied' : 'copyFailed'), copied ? 'success' : 'error');
  };

  const canStart = status.enabled && hasEnabledListener && !needsAuth;
  const next = !hasEnabledListener ? 'listeners' : needsAuth ? 'auth' : 'listeners';
  const heading = canStart ? 'resumeTitle' : !hasEnabledListener ? 'setupListener' : needsAuth ? 'setupAuth' : 'setupEnable';
  const shown = status.running ? running : status.listeners.slice(0, 3);
  const listenerCard = (
    <Card variant="outlined" sx={{ minWidth: 0 }}>
      <CardContent sx={{ p: { xs: 2, sm: 3 }, '&:last-child': { pb: { xs: 2, sm: 3 } } }}>
        <Stack spacing={2}>
          <Stack
            direction={{ xs: 'column', sm: 'row' }}
            spacing={1}
            alignItems={{ xs: 'flex-start', sm: 'center' }}
            justifyContent="space-between"
          >
            <Box>
              <Typography component="h2" variant="h4">
                {label(status.running ? 'connectTitle' : 'savedListenersTitle')}
              </Typography>
              <Typography color="text.secondary" variant="body2" sx={{ mt: 0.5 }}>
                {label(status.running ? 'connectHelp' : 'savedListenersHelp')}
              </Typography>
            </Box>
            {status.running && (
              <Button color="error" disabled={busy} onClick={onStop}>
                {label('stop')}
              </Button>
            )}
          </Stack>
          {shown.length === 0 ? (
            <Typography variant="body2" color="text.secondary">
              {t('settings.socks5.listeners.empty')}
            </Typography>
          ) : (
            <Box
              sx={{
                display: 'grid',
                gap: 1.5,
                gridTemplateColumns: {
                  xs: 'minmax(0, 1fr)',
                  lg: status.running && shown.length > 1 ? 'repeat(2, minmax(0, 1fr))' : 'minmax(0, 1fr)'
                }
              }}
            >
              {shown.map((listener) => {
                const { address, port, wildcard, loopback } = endpointFor(listener);
                const protocol = listener.protocol || 'socks5';
                const auth = listener.effectiveRequireAuth ?? listener.requireAuth ?? settings.requireAuth;
                const hints = [
                  wildcard ? label('wildcardHelp') : loopback ? label('loopbackHelp') : '',
                  protocol === 'mixed' ? label('mixedHelp') : ''
                ]
                  .filter(Boolean)
                  .join(' ');
                const titleId = `gateway-overview-listener-${listener.id}`;
                return (
                  <Stack
                    key={listener.id}
                    component="article"
                    aria-labelledby={titleId}
                    spacing={1.25}
                    sx={{
                      minWidth: 0,
                      p: { xs: 1.5, sm: 2 },
                      border: 1,
                      borderColor: 'divider',
                      borderRadius: 2,
                      bgcolor: 'background.default'
                    }}
                  >
                    <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" alignItems="center">
                      <Typography id={titleId} component="h3" variant="h5" sx={{ overflowWrap: 'anywhere', minWidth: 0 }}>
                        {listener.id}
                      </Typography>
                      {!listener.enabled && <Chip size="small" variant="outlined" color="default" label={t('common.disabled')} />}
                    </Stack>
                    <Stack component="dl" direction="row" spacing={0.75} useFlexGap flexWrap="wrap" sx={{ m: 0 }}>
                      <Box component="div">
                        <Box component="dt" sx={visuallyHidden}>
                          {label('protocolLabel')}
                        </Box>
                        <Box component="dd" sx={{ m: 0 }}>
                          <Chip
                            size="small"
                            color="primary"
                            variant="light"
                            label={protocol === 'mixed' ? label('mixedChoice') : t(`settings.socks5.listeners.protocols.${protocol}`)}
                          />
                        </Box>
                      </Box>
                      <Box component="div" sx={{ maxWidth: '100%' }}>
                        <Box component="dt" sx={visuallyHidden}>
                          {label('authLabel')}
                        </Box>
                        <Box component="dd" sx={{ m: 0 }}>
                          <Chip
                            size="small"
                            variant="outlined"
                            color="default"
                            icon={auth ? <LockOutlined /> : <LockOpenOutlined />}
                            label={label(auth ? 'requiresAuth' : 'noAuth')}
                            sx={{
                              minHeight: 24,
                              height: 'auto',
                              '& .MuiChip-icon': { color: 'inherit' },
                              '& .MuiChip-label': { whiteSpace: 'normal', overflowWrap: 'anywhere' }
                            }}
                          />
                        </Box>
                      </Box>
                    </Stack>
                    <Stack
                      direction="row"
                      spacing={1}
                      useFlexGap
                      flexWrap="wrap"
                      alignItems="center"
                      sx={{
                        px: 1.25,
                        py: 1,
                        minWidth: 0,
                        border: 1,
                        borderColor: 'divider',
                        borderRadius: 1.5,
                        bgcolor: 'background.paper'
                      }}
                    >
                      <Box sx={{ minWidth: 0 }}>
                        <Typography variant="caption" color="text.secondary">
                          {label('proxyAddress')}
                        </Typography>
                        <Typography
                          variant="body1"
                          sx={{
                            fontFamily: 'monospace',
                            fontSize: { xs: '1rem', sm: '1.125rem' },
                            fontWeight: 600,
                            letterSpacing: 0,
                            overflowWrap: 'anywhere'
                          }}
                        >
                          {status.running && wildcard ? t('settings.socks5.page.externalAddress', { port }) : address}
                        </Typography>
                      </Box>
                      {status.running && !wildcard && (
                        <Button
                          size="small"
                          variant="outlined"
                          aria-label={label('copyAddress')}
                          startIcon={<ContentCopy fontSize="small" />}
                          onClick={() => copyAddress(address)}
                        >
                          {label('copyShort')}
                        </Button>
                      )}
                    </Stack>
                    {status.running && hints && (
                      <Typography variant="caption" color="text.secondary">
                        {hints}
                      </Typography>
                    )}
                  </Stack>
                );
              })}
            </Box>
          )}
          {!status.running && status.listeners.length > shown.length && (
            <Button size="small" sx={{ alignSelf: 'flex-start' }} onClick={() => goToTab('listeners')}>
              {t('settings.socks5.page.showAllListeners', { count: status.listeners.length })}
            </Button>
          )}
        </Stack>
      </CardContent>
    </Card>
  );

  if (status.running) return listenerCard;

  return (
    <Box
      sx={{
        display: 'grid',
        gap: 2,
        gridTemplateColumns: { xs: 'minmax(0, 1fr)', lg: 'minmax(0, 2fr) minmax(0, 3fr)' }
      }}
    >
      <Card variant="outlined" sx={{ bgcolor: 'background.default' }}>
        <CardContent sx={{ p: { xs: 2, sm: 3 }, '&:last-child': { pb: { xs: 2, sm: 3 } } }}>
          <Stack spacing={1.5} alignItems="flex-start">
            <Typography component="h2" variant="h4">
              {label(heading)}
            </Typography>
            <Typography color="text.secondary" variant="body2">
              {label(`${heading}Help`)}
            </Typography>
            <Button variant="contained" disabled={busy} onClick={canStart ? onStart : () => goToTab(next)}>
              {label(canStart ? 'start' : next === 'auth' ? 'configureAuth' : 'configureListeners')}
            </Button>
          </Stack>
        </CardContent>
      </Card>
      {listenerCard}
    </Box>
  );
}
