import { useMemo, useState } from 'react';
import { useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';
import useMediaQuery from '@mui/material/useMediaQuery';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import LocationOnIcon from '@mui/icons-material/LocationOn';
import TouchAppIcon from '@mui/icons-material/TouchApp';
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';

const SORT_FIELDS = {
  ip: 'ip',
  region: 'region',
  count: 'count',
  date: 'date'
};

const SORT_ORDERS = {
  asc: 'asc',
  desc: 'desc'
};

const parseIPv4 = (value) => {
  const parts = String(value || '')
    .trim()
    .split('.');

  if (parts.length !== 4) return null;

  const octets = parts.map((part) => {
    if (!/^\d+$/.test(part)) return null;
    const octet = Number(part);
    return octet >= 0 && octet <= 255 ? octet : null;
  });

  return octets.every((octet) => octet !== null) ? octets : null;
};

const compareIPv4Octets = (left, right) => {
  for (let index = 0; index < left.length; index += 1) {
    if (left[index] !== right[index]) return left[index] - right[index];
  }

  return 0;
};

const compareIpValues = (leftValue, rightValue) => {
  const leftIp = String(leftValue || '').trim();
  const rightIp = String(rightValue || '').trim();
  const leftIPv4 = parseIPv4(leftIp);
  const rightIPv4 = parseIPv4(rightIp);

  if (leftIPv4 && rightIPv4) return compareIPv4Octets(leftIPv4, rightIPv4);
  if (leftIPv4) return -1;
  if (rightIPv4) return 1;

  return leftIp.localeCompare(rightIp, undefined, { numeric: true, sensitivity: 'base' });
};

const normalizeCount = (value) => Number(value) || 0;

const normalizeDate = (value) => {
  const timestamp = Date.parse(String(value || ''));
  return Number.isNaN(timestamp) ? 0 : timestamp;
};

export default function AccessLogsDialog({ open, logs, onClose, loading = false, title }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [sortField, setSortField] = useState(SORT_FIELDS.date);
  const [sortOrder, setSortOrder] = useState(SORT_ORDERS.desc);
  const [searchKeyword, setSearchKeyword] = useState('');
  const { isDark } = useResolvedColorScheme();
  const { palette, dialogSurface, dialogSurfaceGradient, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(
    theme,
    isDark
  );
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  const rowHoverSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.05);
  const rowBorder = isDark ? withAlpha(palette.divider, 0.58) : withAlpha(palette.divider, 0.78);
  const ipSurface = isDark ? withAlpha(palette.primary.main, 0.14) : withAlpha(palette.primary.main, 0.06);
  const ipBorder = withAlpha(palette.primary.main, isDark ? 0.26 : 0.16);
  const countChipSurface = withAlpha(palette.primary.main, isDark ? 0.18 : 0.08);
  const countChipBorder = withAlpha(palette.primary.main, isDark ? 0.34 : 0.18);
  const sortControlSurface = isDark ? withAlpha(palette.background.paper, 0.72) : withAlpha(palette.background.default, 0.92);
  const sortFocusBorder = withAlpha(palette.primary.main, isDark ? 0.44 : 0.28);

  const unknownSourceLabel = t('subscriptions.accessLogs.unknownSource');

  const sortOptions = useMemo(
    () => [
      { value: SORT_FIELDS.ip, label: t('subscriptions.accessLogs.sort.fields.ip') },
      { value: SORT_FIELDS.region, label: t('subscriptions.accessLogs.sort.fields.region') },
      { value: SORT_FIELDS.count, label: t('subscriptions.accessLogs.sort.fields.count') },
      { value: SORT_FIELDS.date, label: t('subscriptions.accessLogs.sort.fields.date') }
    ],
    [t]
  );

  const filteredLogs = useMemo(() => {
    const keyword = searchKeyword.trim().toLowerCase();
    if (!keyword) return logs;

    return logs.filter((log) => {
      const ip = String(log.IP || '').toLowerCase();
      const region = String(log.Addr || unknownSourceLabel).toLowerCase();

      return ip.includes(keyword) || region.includes(keyword);
    });
  }, [logs, searchKeyword, unknownSourceLabel]);

  const sortedLogs = useMemo(() => {
    const compareLogs = (leftLog, rightLog) => {
      let comparison = 0;

      if (sortField === SORT_FIELDS.ip) {
        comparison = compareIpValues(leftLog.IP, rightLog.IP);
      } else if (sortField === SORT_FIELDS.region) {
        const leftRegion = String(leftLog.Addr || unknownSourceLabel);
        const rightRegion = String(rightLog.Addr || unknownSourceLabel);
        comparison = leftRegion.localeCompare(rightRegion, undefined, { numeric: true, sensitivity: 'base' });
      } else if (sortField === SORT_FIELDS.count) {
        comparison = normalizeCount(leftLog.Count) - normalizeCount(rightLog.Count);
      } else if (sortField === SORT_FIELDS.date) {
        comparison = normalizeDate(leftLog.Date) - normalizeDate(rightLog.Date);
      }

      return sortOrder === SORT_ORDERS.desc ? comparison * -1 : comparison;
    };

    return filteredLogs
      .map((log, index) => ({ log, index }))
      .sort((left, right) => {
        const comparison = compareLogs(left.log, right.log);
        return comparison === 0 ? left.index - right.index : comparison;
      })
      .map(({ log }) => log);
  }, [filteredLogs, sortField, sortOrder, unknownSourceLabel]);

  const hasSearchKeyword = searchKeyword.trim().length > 0;

  const handleSort = (field) => {
    if (field === sortField) {
      setSortOrder((currentOrder) => (currentOrder === SORT_ORDERS.asc ? SORT_ORDERS.desc : SORT_ORDERS.asc));
      return;
    }

    setSortField(field);
    setSortOrder(field === SORT_FIELDS.count || field === SORT_FIELDS.date ? SORT_ORDERS.desc : SORT_ORDERS.asc);
  };

  const handleSortFieldChange = (field) => {
    setSortField(field);
    setSortOrder(field === SORT_FIELDS.count || field === SORT_FIELDS.date ? SORT_ORDERS.desc : SORT_ORDERS.asc);
  };

  const dialogPaperSx = {
    borderRadius: isMobile ? 0 : 3,
    overflow: 'hidden',
    bgcolor: dialogSurface,
    backgroundImage: dialogSurfaceGradient,
    border: '1px solid',
    borderColor: panelBorder
  };

  const titleSx = {
    px: 2.5,
    py: 2,
    bgcolor: mutedPanelSurface,
    borderBottom: '1px solid',
    borderColor: panelBorder,
    boxShadow: `inset 0 -1px 0 ${withAlpha(palette.divider, 0.42)}`
  };

  const actionsSx = {
    px: 2.5,
    py: 1.5,
    bgcolor: mutedPanelSurface,
    borderTop: '1px solid',
    borderColor: panelBorder
  };

  const countChipSx = {
    ml: 1,
    bgcolor: countChipSurface,
    color: palette.primary.main,
    border: '1px solid',
    borderColor: countChipBorder,
    fontWeight: 600
  };

  const mobileSortControlSx = {
    minWidth: 0,
    flex: 1,
    '& .MuiInputLabel-root': {
      color: secondaryText
    },
    '& .MuiOutlinedInput-root': {
      bgcolor: sortControlSurface,
      color: primaryText,
      borderRadius: 2,
      '& fieldset': { borderColor: rowBorder },
      '&:hover fieldset': { borderColor: sortFocusBorder },
      '&.Mui-focused fieldset': { borderColor: sortFocusBorder }
    },
    '& .MuiSelect-icon': {
      color: secondaryText
    }
  };

  const directionToggleSx = {
    flexShrink: 0,
    bgcolor: sortControlSurface,
    borderRadius: 2,
    '& .MuiToggleButton-root': {
      minWidth: 48,
      px: 1.25,
      color: secondaryText,
      borderColor: rowBorder,
      '&.Mui-selected': {
        color: palette.primary.main,
        bgcolor: withAlpha(palette.primary.main, isDark ? 0.18 : 0.1)
      },
      '&.Mui-selected:hover': {
        bgcolor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.14)
      }
    }
  };

  const searchFieldSx = {
    mb: 1.5,
    '& .MuiOutlinedInput-root': {
      bgcolor: sortControlSurface,
      color: primaryText,
      borderRadius: 2,
      '& fieldset': { borderColor: rowBorder },
      '&:hover fieldset': { borderColor: sortFocusBorder },
      '&.Mui-focused fieldset': { borderColor: sortFocusBorder }
    },
    '& .MuiInputBase-input::placeholder': {
      color: tertiaryText,
      opacity: 1
    }
  };

  const renderSortableHeader = (field, label, sx, align) => (
    <TableCell sx={{ whiteSpace: 'nowrap', ...sx }} align={align} sortDirection={sortField === field ? sortOrder : false}>
      <TableSortLabel
        active={sortField === field}
        direction={sortField === field ? sortOrder : SORT_ORDERS.asc}
        onClick={() => handleSort(field)}
        sx={{ whiteSpace: 'nowrap', '& .MuiTableSortLabel-icon': { flexShrink: 0 } }}
      >
        {label}
      </TableSortLabel>
    </TableCell>
  );

  const renderSearchField = () => (
    <TextField
      fullWidth
      size="small"
      value={searchKeyword}
      onChange={(event) => setSearchKeyword(event.target.value)}
      placeholder={t('subscriptions.accessLogs.search.placeholder')}
      sx={searchFieldSx}
      slotProps={{
        htmlInput: {
          'aria-label': t('subscriptions.accessLogs.search.label')
        },
        input: {
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon fontSize="small" sx={{ color: tertiaryText }} />
            </InputAdornment>
          ),
          endAdornment: searchKeyword ? (
            <InputAdornment position="end">
              <IconButton
                size="small"
                edge="end"
                onClick={() => setSearchKeyword('')}
                aria-label={t('subscriptions.accessLogs.search.clear')}
              >
                <ClearIcon fontSize="small" />
              </IconButton>
            </InputAdornment>
          ) : null
        }
      }}
    />
  );

  const FilteredEmptyState = () => (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        py: 6,
        borderRadius: 2.5,
        bgcolor: nestedPanelSurface,
        border: '1px solid',
        borderColor: rowBorder,
        color: secondaryText
      }}
    >
      <SearchIcon sx={{ fontSize: 42, mb: 1.5, opacity: 0.5, color: tertiaryText }} />
      <Typography sx={{ color: secondaryText }}>{t('subscriptions.accessLogs.search.empty')}</Typography>
    </Box>
  );

  const MobileSortControls = () => (
    <Stack
      direction="row"
      spacing={1}
      sx={{
        mb: 1.5,
        minHeight: 40,
        alignItems: 'center'
      }}
    >
      <FormControl size="small" sx={mobileSortControlSx}>
        <InputLabel id="access-logs-sort-field-label">{t('subscriptions.accessLogs.sort.field')}</InputLabel>
        <Select
          labelId="access-logs-sort-field-label"
          value={sortField}
          label={t('subscriptions.accessLogs.sort.field')}
          onChange={(event) => handleSortFieldChange(event.target.value)}
        >
          {sortOptions.map((option) => (
            <MenuItem key={option.value} value={option.value}>
              {option.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <ToggleButtonGroup
        exclusive
        size="small"
        value={sortOrder}
        onChange={(_, nextOrder) => {
          if (nextOrder) setSortOrder(nextOrder);
        }}
        aria-label={t('subscriptions.accessLogs.sort.direction')}
        sx={directionToggleSx}
      >
        <ToggleButton value={SORT_ORDERS.asc} aria-label={t('subscriptions.accessLogs.sort.ascending')}>
          {t('subscriptions.accessLogs.sort.ascShort')}
        </ToggleButton>
        <ToggleButton value={SORT_ORDERS.desc} aria-label={t('subscriptions.accessLogs.sort.descending')}>
          {t('subscriptions.accessLogs.sort.descShort')}
        </ToggleButton>
      </ToggleButtonGroup>
    </Stack>
  );

  const renderIpBlock = (ip) => (
    <Box
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        maxWidth: '100%',
        px: 1,
        py: 0.5,
        borderRadius: 1.25,
        bgcolor: ipSurface,
        border: '1px solid',
        borderColor: ipBorder
      }}
    >
      <Typography
        variant="body2"
        sx={{
          fontFamily: 'monospace',
          color: palette.primary.main,
          fontWeight: 600,
          wordBreak: 'break-all'
        }}
      >
        {ip}
      </Typography>
    </Box>
  );

  const MobileLogCard = ({ log }) => (
    <Card
      sx={{
        mb: 1.5,
        borderRadius: 2.5,
        bgcolor: nestedPanelSurface,
        border: '1px solid',
        borderColor: rowBorder,
        transition: 'all 0.2s ease',
        '&:hover': {
          bgcolor: rowHoverSurface,
          borderColor: withAlpha(palette.primary.main, isDark ? 0.24 : 0.14)
        }
      }}
    >
      <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
        <Stack spacing={1}>
          <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 1 }}>
            <Box sx={{ flex: 1, minWidth: 0 }}>{renderIpBlock(log.IP)}</Box>
            <Chip
              size="small"
              label={t('subscriptions.accessLogs.count', { count: log.Count })}
              icon={<TouchAppIcon sx={{ fontSize: 14 }} />}
              sx={{
                height: 24,
                bgcolor: countChipSurface,
                color: palette.primary.main,
                border: '1px solid',
                borderColor: countChipBorder,
                '& .MuiChip-label': { px: 1 },
                flexShrink: 0
              }}
            />
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <LocationOnIcon sx={{ fontSize: 16, color: tertiaryText, flexShrink: 0 }} />
            <Typography
              variant="body2"
              sx={{
                color: secondaryText,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap'
              }}
            >
              {log.Addr || unknownSourceLabel}
            </Typography>
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <AccessTimeIcon sx={{ fontSize: 16, color: tertiaryText, flexShrink: 0 }} />
            <Typography variant="body2" sx={{ color: secondaryText }}>
              {log.Date}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );

  const DesktopTable = () => (
    <TableContainer
      sx={{
        borderRadius: 2.5,
        bgcolor: nestedPanelSurface,
        border: '1px solid',
        borderColor: rowBorder,
        overflow: 'hidden'
      }}
    >
      <Table size="small">
        <TableHead>
          <TableRow
            sx={{
              bgcolor: mutedPanelSurface,
              '& .MuiTableCell-root': {
                borderColor: rowBorder
              }
            }}
          >
            {renderSortableHeader(SORT_FIELDS.ip, t('subscriptions.accessLogs.ip'), {
              fontWeight: 600,
              minWidth: 140,
              color: secondaryText
            })}
            {renderSortableHeader(SORT_FIELDS.region, t('subscriptions.accessLogs.region'), {
              fontWeight: 600,
              minWidth: 120,
              color: secondaryText
            })}
            {renderSortableHeader(
              SORT_FIELDS.count,
              t('subscriptions.accessLogs.visits'),
              { fontWeight: 600, width: 120, minWidth: 120, color: secondaryText },
              'center'
            )}
            {renderSortableHeader(SORT_FIELDS.date, t('subscriptions.accessLogs.lastVisit'), {
              fontWeight: 600,
              minWidth: 160,
              color: secondaryText
            })}
          </TableRow>
        </TableHead>
        <TableBody>
          {sortedLogs.map((log) => (
            <TableRow
              key={log.ID}
              sx={{
                transition: 'background-color 0.2s ease',
                '&:hover': { bgcolor: rowHoverSurface },
                '& .MuiTableCell-root': {
                  borderColor: rowBorder
                }
              }}
            >
              <TableCell>{renderIpBlock(log.IP)}</TableCell>
              <TableCell>
                <Typography variant="body2" sx={{ color: secondaryText }}>
                  {log.Addr || unknownSourceLabel}
                </Typography>
              </TableCell>
              <TableCell align="center" sx={{ width: 120, minWidth: 120, whiteSpace: 'nowrap' }}>
                <Chip size="small" label={log.Count} sx={{ ...countChipSx, minWidth: 50 }} />
              </TableCell>
              <TableCell>
                <Typography variant="body2" sx={{ color: secondaryText }}>
                  {log.Date}
                </Typography>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      fullScreen={isMobile}
      slotProps={{
        paper: {
          sx: dialogPaperSx
        }
      }}
    >
      <DialogTitle sx={titleSx}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <TouchAppIcon sx={{ color: palette.primary.main }} />
          <Typography variant="h6" sx={{ color: primaryText }}>
            {title || t('subscriptions.accessLogs.title')}
          </Typography>
          {!loading && logs.length > 0 && (
            <Chip
              size="small"
              label={
                hasSearchKeyword
                  ? t('subscriptions.accessLogs.filteredTotal', { filtered: sortedLogs.length, total: logs.length })
                  : t('subscriptions.accessLogs.total', { count: logs.length })
              }
              sx={countChipSx}
            />
          )}
        </Stack>
      </DialogTitle>
      <DialogContent
        sx={{
          px: isMobile ? 1.5 : 2,
          pt: isMobile ? 2 : 2.5,
          pb: isMobile ? 1.5 : 2,
          bgcolor: dialogSurface,
          '&&': {
            pt: isMobile ? 2 : 2.5
          }
        }}
      >
        {loading ? (
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', py: 8 }}>
            <CircularProgress size={28} />
          </Box>
        ) : logs.length === 0 ? (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              py: 8,
              borderRadius: 2.5,
              bgcolor: nestedPanelSurface,
              border: '1px solid',
              borderColor: rowBorder,
              color: secondaryText
            }}
          >
            <TouchAppIcon sx={{ fontSize: 48, mb: 2, opacity: 0.5, color: tertiaryText }} />
            <Typography sx={{ color: secondaryText }}>{t('subscriptions.accessLogs.empty')}</Typography>
          </Box>
        ) : isMobile ? (
          <Box>
            {renderSearchField()}
            <MobileSortControls />
            {sortedLogs.length > 0 ? sortedLogs.map((log) => <MobileLogCard key={log.ID} log={log} />) : <FilteredEmptyState />}
          </Box>
        ) : (
          <Box>
            {renderSearchField()}
            {sortedLogs.length > 0 ? <DesktopTable /> : <FilteredEmptyState />}
          </Box>
        )}
      </DialogContent>
      <DialogActions sx={actionsSx}>
        <Button onClick={onClose} variant="outlined">
          {t('common.close')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
