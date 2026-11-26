import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import ExpandMore from '@mui/icons-material/ExpandMore';

import CandidateNodePool from './CandidateNodePool';
import SpecificNodeSelector from './SpecificNodeSelector';
import { gatewayInputId, numericFields } from '../gatewaySettings';

export function SettingsGroup({ title, openOnError, children }) {
  const details = useRef(null);
  useEffect(() => {
    if (openOnError) details.current.open = true;
  }, [openOnError]);

  return (
    <Box
      component="details"
      ref={details}
      sx={{
        border: 1,
        borderColor: 'divider',
        borderRadius: 2,
        bgcolor: 'background.default',
        '& > summary': {
          listStyle: 'none',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          p: 1.75
        },
        '& > summary::-webkit-details-marker': { display: 'none' },
        '&[open] > summary svg': { transform: 'rotate(180deg)' }
      }}
    >
      <Typography component="summary" variant="subtitle2">
        {title}
        <ExpandMore fontSize="small" aria-hidden />
      </Typography>
      <Box sx={{ px: 1.75, pb: 1.75 }}>{children}</Box>
    </Box>
  );
}

export default function GatewaySettingsFields({ section, form, update, error, busy }) {
  const { t } = useTranslation();
  useEffect(() => {
    if (error?.tab !== section) return undefined;
    const frame = window.requestAnimationFrame(() => document.getElementById(gatewayInputId(error.field))?.focus());
    return () => window.cancelAnimationFrame(frame);
  }, [error, section]);

  const label = (key, options) => t(`settings.socks5.${key}`, options);
  const invalid = (field) => error?.field === field;
  const helper = (field, fallback) => (invalid(field) ? label(error.message, error.params) : fallback);
  const text = (field, props) => (
    <TextField
      id={gatewayInputId(field)}
      fullWidth
      size="small"
      label={label(`form.${field}`)}
      value={form[field]}
      onChange={(event) => update({ [field]: event.target.value })}
      error={invalid(field)}
      {...props}
      disabled={busy || props?.disabled}
      helperText={helper(field, props?.helperText)}
    />
  );
  const number = (field) => {
    const { min, max } = numericFields[field];
    return text(field, {
      type: 'number',
      helperText: label(`form.${field}Helper`),
      slotProps: { htmlInput: { min, max } }
    });
  };
  const toggle = (field, options = {}) => (
    <FormControlLabel
      control={
        <Switch
          id={gatewayInputId(field)}
          disabled={busy}
          checked={Boolean(form[field])}
          onChange={(_, checked) => (options.onChange ? options.onChange(checked) : update({ [field]: checked }))}
        />
      }
      label={label(`form.${field}`)}
    />
  );
  const group = (name, children) => (
    <SettingsGroup title={label(`page.groups.${name}`)} openOnError={error?.section === name ? error : null}>
      {children}
    </SettingsGroup>
  );
  const grid = (fields) => (
    <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: 'minmax(0, 1fr)', sm: 'repeat(2, minmax(0, 1fr))' } }}>{fields}</Box>
  );

  let title = label(`page.groups.${section}`);
  let contents;
  if (section === 'listeners') {
    contents = (
      <>
        <Box>
          {toggle('enabled')}
          <FormHelperText>{label('page.enabledHelper')}</FormHelperText>
        </Box>
        {group(
          'limits',
          grid(
            ['maxConnections', 'maxConnectionsPerClient', 'idleTimeoutSeconds', 'maxConnectionDurationSeconds'].map((field) => (
              <Box key={field}>{number(field)}</Box>
            ))
          )
        )}
      </>
    );
  } else if (section === 'routing') {
    const specific = form.selection === 'specific';
    contents = (
      <>
        {text('selection', {
          select: true,
          helperText: label('page.selectionHelp'),
          children: Object.entries({ best: 'Best', random: 'Random', round_robin: 'RoundRobin', smart: 'Smart', specific: 'Specific' }).map(
            ([value, name]) => (
              <MenuItem key={value} value={value}>
                {label(`form.selection${name}`)}
              </MenuItem>
            )
          )
        })}
        {specific && (
          <>
            <SpecificNodeSelector
              inputId={gatewayInputId('nodeId')}
              value={Number(form.nodeId) || 0}
              onChange={(nodeId) => update({ nodeId })}
              disabled={busy}
              error={invalid('nodeId')}
            />
            <Box>
              {toggle('specificFallback', {
                onChange: (checked) =>
                  update({
                    specificFallback: checked,
                    maxAttempts: checked ? Math.max(Number(form.maxAttempts) || 1, 2) : form.maxAttempts
                  })
              })}
              <FormHelperText>{label('form.specificFallbackHelper')}</FormHelperText>
            </Box>
          </>
        )}
        {(!specific || form.specificFallback) && <CandidateNodePool value={form} onChange={update} disabled={busy} />}
        {group(
          'advancedRouting',
          <Stack spacing={2}>
            {grid(
              <>
                {number('maxAttempts')}
                {number('dialTimeoutSeconds')}
                {number('failureCooldownSeconds')}
              </>
            )}
            <Box>
              {toggle('stickySessionEnabled')}
              <FormHelperText>{label(specific ? 'form.stickySessionSpecificHelper' : 'form.stickySessionEnabledHelper')}</FormHelperText>
            </Box>
            {(form.stickySessionEnabled || invalid('stickySessionTtlSeconds')) &&
              grid(
                <>
                  {text('stickySessionMode', {
                    select: true,
                    children: [
                      <MenuItem key="client_ip" value="client_ip">
                        {label('form.stickySessionModeClientIp')}
                      </MenuItem>,
                      <MenuItem key="username" value="username">
                        {label('form.stickySessionModeUsername')}
                      </MenuItem>
                    ]
                  })}
                  {number('stickySessionTtlSeconds')}
                </>
              )}
          </Stack>
        )}
      </>
    );
  } else if (section === 'auth') {
    const username = form.username.trim() || 'proxy';
    contents = (
      <>
        <Alert severity="warning">{label('alerts.security')}</Alert>
        <Box>
          {toggle('requireAuth')}
          <FormHelperText error={invalid('requireAuth')}>{helper('requireAuth', label('page.sharedAuthHelp'))}</FormHelperText>
        </Box>
        {grid(
          <>
            {text('username', {
              autoComplete: 'off',
              helperText: form.username.includes(':') ? label('form.httpUsernameWarning') : undefined
            })}
            {text('password', {
              type: 'password',
              autoComplete: 'new-password',
              disabled: form.clearPassword,
              helperText: label(form.hasPassword ? 'form.passwordSaved' : 'form.passwordHelper')
            })}
          </>
        )}
        {form.hasPassword && (
          <Box>
            {toggle('clearPassword', { onChange: (checked) => update({ clearPassword: checked, password: '' }) })}
            {invalid('clearPassword') && <FormHelperText error>{helper('clearPassword')}</FormHelperText>}
          </Box>
        )}
        {group(
          'usernameRouting',
          <Stack spacing={1.5}>
            <Typography color="text.secondary" variant="body2">
              {label('page.usernameRouting.help')}
            </Typography>
            <Box
              component="dl"
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: 'minmax(0, 1fr)', sm: 'fit-content(38%) minmax(0, 1fr)' },
                columnGap: 2,
                rowGap: { xs: 0.5, sm: 1 },
                maxWidth: { sm: 820 },
                m: 0
              }}
            >
              {Object.entries({ listener: username, profile: `${username}@japan`, client: `${username}@japan.user01` }).map(
                ([kind, example]) => (
                  <Box key={kind} sx={{ display: 'contents' }}>
                    <Typography component="dt" variant="body2" sx={{ fontFamily: 'monospace', overflowWrap: 'anywhere' }}>
                      {example}
                    </Typography>
                    <Typography component="dd" color="text.secondary" variant="body2" sx={{ m: 0, mb: { xs: 0.75, sm: 0 } }}>
                      {label(`page.usernameRouting.${kind}`)}
                    </Typography>
                  </Box>
                )
              )}
            </Box>
            <Typography color="text.secondary" variant="body2">
              {label('page.usernameRouting.independentAccounts')}
            </Typography>
          </Stack>
        )}
      </>
    );
  } else {
    title = null;
    contents = group(
      'probeSettings',
      <Stack spacing={2}>
        <Box>
          {toggle('healthCheckEnabled')}
          <FormHelperText>{label('form.healthCheckEnabledHelper')}</FormHelperText>
        </Box>
        {(form.healthCheckEnabled || error?.section === 'probeSettings') &&
          grid(
            <>
              {number('healthCheckIntervalSeconds')}
              {number('healthCheckTimeoutSeconds')}
            </>
          )}
      </Stack>
    );
  }

  return (
    <Card variant="outlined">
      <CardContent sx={{ p: { xs: 1.75, sm: 2.5 }, '&:last-child': { pb: { xs: 1.75, sm: 2.5 } } }}>
        <Stack spacing={2} component="fieldset" disabled={busy} sx={{ border: 0, p: 0, m: 0, minWidth: 0 }}>
          {title && (
            <Typography variant="h4" component="h2">
              {title}
            </Typography>
          )}
          {contents}
        </Stack>
      </CardContent>
    </Card>
  );
}
