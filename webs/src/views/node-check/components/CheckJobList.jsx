import { Fragment } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Alert,
  Box,
  Button,
  Chip,
  FormControlLabel,
  IconButton,
  Paper,
  Stack,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Tooltip,
  Typography
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import { formatDateTime } from 'i18n/locales';

const tableWidth = '@container detection-center (min-width: 960px)';
const wrapText = { overflowWrap: 'anywhere' };
const columns = [
  { key: 'jobName', width: '16%' },
  { key: 'plan', width: '10%' },
  { key: 'scope', width: '22%' },
  { key: 'trigger', width: '14%' },
  { key: 'runTimes' },
  { key: 'autoEnabled', width: 80, align: 'center' },
  { key: 'actions', width: 122, align: 'right' }
];

function JobName({ job, legacyLabel, runningLabel }) {
  return (
    <Stack gap={0.5} sx={{ minWidth: 0 }}>
      <Typography variant="body2" sx={{ ...wrapText, fontWeight: 600 }}>
        {job.name}
      </Typography>
      {(job.legacy || job.running) && (
        <Stack direction="row" flexWrap="wrap" gap={0.5}>
          {job.legacy && <Chip label={legacyLabel} size="small" variant="outlined" />}
          {job.running && <Chip label={runningLabel} size="small" color="info" />}
        </Stack>
      )}
    </Stack>
  );
}

function JobValue({ children, detail }) {
  return (
    <>
      <Typography variant="body2" sx={{ ...wrapText, lineHeight: 1.4 }}>
        {children}
      </Typography>
      {detail && (
        <Typography variant="caption" color="text.secondary" sx={{ ...wrapText, display: 'block', lineHeight: 1.4 }}>
          {detail}
        </Typography>
      )}
    </>
  );
}

function CardField({ label, children, sx }) {
  return (
    <Box sx={{ minWidth: 0, ...sx }}>
      <Typography component="dt" variant="caption" color="text.secondary" sx={{ lineHeight: 1.35 }}>
        {label}
      </Typography>
      <Box component="dd" sx={{ m: 0 }}>
        {children}
      </Box>
    </Box>
  );
}

function JobActions({ actions, jobName, compact }) {
  return (
    <Stack direction="row" gap={0.5} flexWrap={compact ? 'nowrap' : 'wrap'} justifyContent={compact ? 'flex-end' : 'flex-start'}>
      {actions.map(({ label, Icon, ...props }) => {
        const name = `${label} · ${jobName}`;
        return compact ? (
          <Tooltip key={label} title={label}>
            <span>
              <IconButton aria-label={name} size="small" {...props}>
                <Icon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
        ) : (
          <Button key={label} aria-label={name} size="small" startIcon={<Icon fontSize="small" />} sx={{ minHeight: 44 }} {...props}>
            {label}
          </Button>
        );
      })}
    </Stack>
  );
}

export default function CheckJobList({ jobs, profiles, timezone, busy, onEdit, onRun, onRemove, onToggle }) {
  const { t, i18n } = useTranslation();
  const language = i18n.resolvedLanguage || i18n.language;
  const date = (value) =>
    !value || String(value).startsWith('0001-') ? '—' : formatDateTime(value, language, { dateStyle: 'short', timeStyle: 'medium' }) || '—';

  if (!jobs.length) {
    return (
      <Typography color="text.secondary" sx={{ p: 3, textAlign: 'center' }}>
        {t('detectionCenter.noJobs')}
      </Typography>
    );
  }

  const profileNames = new Map(profiles.map(({ id, name }) => [id, name]));
  const rows = jobs.map((job) => {
    const name = <JobName job={job} legacyLabel={t('detectionCenter.legacy')} runningLabel={t('tasks.status.running')} />;
    const plan = <JobValue>{profileNames.get(job.profileId) || t('detectionCenter.missingPlan')}</JobValue>;
    const scope = (
      <JobValue detail={[...(job.groups || []), ...(job.tags || [])].join(' · ')}>{t(`detectionCenter.scopes.${job.scope}`)}</JobValue>
    );
    const schedule = [job.trigger === 'scheduled' ? job.cronExpr || '—' : '', timezone].filter(Boolean).join(' · ');
    const trigger = <JobValue detail={schedule}>{t(`detectionCenter.triggers.${job.trigger}`)}</JobValue>;
    const runTimes = [
      { label: t('detectionCenter.lastRun'), value: date(job.lastRunTime) },
      { label: t('detectionCenter.nextRun'), value: date(job.nextRunTime) }
    ];
    const toggle = (
      <Switch
        checked={job.enabled}
        disabled={busy}
        onChange={(_, enabled) => onToggle(job, enabled)}
        slotProps={{ input: { 'aria-label': `${t('detectionCenter.autoEnabled')} · ${job.name}` } }}
      />
    );
    const actions = [
      { label: t('common.edit'), Icon: EditIcon, onClick: () => onEdit(job), disabled: busy },
      { label: t('detectionCenter.run'), Icon: PlayArrowIcon, onClick: () => onRun(job), disabled: busy || job.running },
      { label: t('common.delete'), Icon: DeleteIcon, onClick: () => onRemove(job), disabled: busy || job.running, color: 'error' }
    ];
    const warning = job.lastError && (
      <Alert severity="warning" sx={{ ...wrapText, py: 0, '& .MuiAlert-message': { minWidth: 0 } }}>
        {job.lastError}
      </Alert>
    );
    return { job, name, plan, scope, trigger, runTimes, toggle, actions, warning };
  });

  return (
    <>
      <Paper
        variant="outlined"
        sx={{ display: 'none', overflow: 'hidden', bgcolor: 'background.paper', [tableWidth]: { display: 'block' } }}
      >
        <Table
          size="small"
          aria-label={t('detectionCenter.tabs.jobs')}
          sx={{ tableLayout: 'fixed', '& .MuiTableCell-root': { px: 1.5, py: 0.875, borderColor: 'divider' } }}
        >
          <colgroup>
            {columns.map(({ key, width }) => (
              <col key={key} style={{ width }} />
            ))}
          </colgroup>
          <TableHead sx={{ bgcolor: 'background.default' }}>
            <TableRow>
              {columns.map(({ key, align }) => (
                <TableCell key={key} scope="col" align={align} sx={{ '&.MuiTableCell-head': { color: 'text.secondary', fontWeight: 600 } }}>
                  {t(`detectionCenter.${key}`)}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody sx={{ '& > :last-child > .MuiTableCell-root': { borderBottom: 0 } }}>
            {rows.map(({ job, name, plan, scope, trigger, runTimes, toggle, actions, warning }) => (
              <Fragment key={job.id}>
                <TableRow hover sx={warning ? { '& > .MuiTableCell-root': { borderBottom: 0 } } : undefined}>
                  <TableCell component="th" scope="row">
                    {name}
                  </TableCell>
                  <TableCell>{plan}</TableCell>
                  <TableCell>{scope}</TableCell>
                  <TableCell>{trigger}</TableCell>
                  <TableCell sx={{ fontVariantNumeric: 'tabular-nums' }}>
                    {runTimes.map(({ label, value }) => (
                      <Typography key={label} variant="caption" display="block" sx={{ ...wrapText, lineHeight: 1.4 }}>
                        <Box component="span" sx={{ color: 'text.secondary', mr: 0.75 }}>
                          {label}
                        </Box>
                        {value}
                      </Typography>
                    ))}
                  </TableCell>
                  <TableCell align="center">{toggle}</TableCell>
                  <TableCell align="right">
                    <JobActions actions={actions} jobName={job.name} compact />
                  </TableCell>
                </TableRow>
                {warning && (
                  <TableRow>
                    <TableCell colSpan={columns.length} sx={{ '&&': { pt: 0 } }}>
                      {warning}
                    </TableCell>
                  </TableRow>
                )}
              </Fragment>
            ))}
          </TableBody>
        </Table>
      </Paper>
      <Box
        sx={{
          display: 'grid',
          gap: 1,
          gridTemplateColumns: 'minmax(0, 1fr)',
          alignItems: 'start',
          '@container detection-center (min-width: 680px)': { gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' },
          [tableWidth]: { display: 'none' }
        }}
      >
        {rows.map(({ job, name, plan, scope, trigger, runTimes, toggle, actions, warning }) => (
          <Paper component="article" aria-label={job.name} key={job.id} variant="outlined" sx={{ p: 1.5, bgcolor: 'background.default' }}>
            <Stack gap={1}>
              <Stack direction="row" justifyContent="space-between" alignItems="center" gap={1}>
                {name}
                <FormControlLabel
                  control={toggle}
                  label={t('detectionCenter.autoEnabled')}
                  slotProps={{ typography: { variant: 'caption' } }}
                  sx={{ m: 0, minHeight: 44, flexShrink: 0, '& .MuiSwitch-root': { ml: -0.75 } }}
                />
              </Stack>
              <Box component="dl" sx={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: 1, m: 0 }}>
                <CardField label={t('detectionCenter.plan')}>{plan}</CardField>
                <CardField label={t('detectionCenter.trigger')}>{trigger}</CardField>
                <CardField label={t('detectionCenter.scope')} sx={{ gridColumn: '1 / -1' }}>
                  {scope}
                </CardField>
                {runTimes.map(({ label, value }) => (
                  <CardField key={label} label={label} sx={{ fontVariantNumeric: 'tabular-nums' }}>
                    <JobValue>{value}</JobValue>
                  </CardField>
                ))}
              </Box>
              {warning}
              <Box sx={{ borderTop: 1, borderColor: 'divider' }}>
                <JobActions actions={actions} jobName={job.name} />
              </Box>
            </Stack>
          </Paper>
        ))}
      </Box>
    </>
  );
}
