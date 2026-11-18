import { Fragment, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Collapse from '@mui/material/Collapse';
import Tooltip from '@mui/material/Tooltip';
import { useTheme } from '@mui/material/styles';
import SortableNodeList from './SortableNodeList';
import { getSubscriptionNameChipSx } from './subscriptionNameChipStyles';
import { getNodeDisplayName } from 'utils/nodeDisplayName';

// icons
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import QrCode2Icon from '@mui/icons-material/QrCode2';
import HistoryIcon from '@mui/icons-material/History';
import SortIcon from '@mui/icons-material/Sort';
import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import VisibilityIcon from '@mui/icons-material/Visibility';
import AccountTreeIcon from '@mui/icons-material/AccountTree';

/**
 */
export default function SubscriptionTable({
  subscriptions,
  isProxyTab = false,
  airportOptions = [],
  expandedRows,
  sortingSubId,
  tempSortData,
  selectedSortItems = [],
  onToggleRow,
  onClient,
  onLogs,
  onEdit,
  onDelete,
  onCopy,
  onPreview,
  onChainProxy,
  onStartSort,
  onConfirmSort,
  onCancelSort,
  onDragEnd,
  onCopyToClipboard,
  getSortedItems,
  onToggleSortSelect,
  onSelectAllSort,
  onClearSortSelection,
  onBatchSort,
  onBatchMove
}) {
  const theme = useTheme();
  const { t } = useTranslation();

  // 代理订阅只保存机场 ID，列表展示时按 ID 反查机场名称。
  const airportNameById = useMemo(() => {
    const map = new Map();
    (airportOptions || []).forEach((airport) => {
      const id = Number(airport?.id ?? airport?.ID);
      if (Number.isFinite(id)) {
        map.set(id, airport?.name || airport?.Name || '');
      }
    });
    return map;
  }, [airportOptions]);

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell width={50} />
            <TableCell>{t('subscriptions.table.name')}</TableCell>
            <TableCell>{isProxyTab ? t('subscriptions.table.airport') : t('subscriptions.table.nodesAndGroups')}</TableCell>
            <TableCell>{t('subscriptions.table.createdAt')}</TableCell>
            <TableCell align="right" width={350}>
              {t('subscriptions.table.actions')}
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {subscriptions.map((sub) => (
            <Fragment key={sub.ID}>
              <TableRow hover>
                <TableCell>
                  {!isProxyTab && (
                    <IconButton size="small" onClick={() => onToggleRow(sub.ID)}>
                      {expandedRows[sub.ID] ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                    </IconButton>
                  )}
                </TableCell>
                <TableCell>
                  <Chip label={sub.Name} sx={getSubscriptionNameChipSx(theme)} />
                  {sortingSubId === sub.ID && <Chip label={t('subscriptions.table.sorting')} color="warning" size="small" sx={{ ml: 1 }} />}
                  {sub.Description ? (
                    <Tooltip title={sub.Description}>
                      <Typography
                        variant="caption"
                        sx={(theme) => ({
                          display: 'block',
                          mt: 0.5,
                          maxWidth: 320,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          color: theme.vars?.palette.text.secondary || theme.palette.text.secondary
                        })}
                      >
                        {sub.Description}
                      </Typography>
                    </Tooltip>
                  ) : null}
                </TableCell>
                <TableCell>
                  {isProxyTab ? (
                    <Tooltip title={airportNameById.get(Number(sub.AirportID)) || t('subscriptions.table.airportMissing')}>
                      <Typography
                        variant="body2"
                        sx={{ maxWidth: 360, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                      >
                        {airportNameById.get(Number(sub.AirportID)) || t('subscriptions.table.airportMissing')}
                      </Typography>
                    </Tooltip>
                  ) : (
                    <Typography variant="body2">
                      {[
                        t('subscriptions.table.nodeCount', { count: sub.Nodes?.length || 0 }),
                        t('subscriptions.table.groupCount', { count: sub.Groups?.length || 0 }),
                        t('subscriptions.table.airportCount', { count: sub.Airports?.length || 0 })
                      ].join(', ')}
                    </Typography>
                  )}
                </TableCell>
                <TableCell>{sub.CreateDate}</TableCell>
                <TableCell align="right">
                  <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                    {!isProxyTab && (
                      <Tooltip title={t('subscriptions.table.tooltips.preview')}>
                        <IconButton size="small" color="info" onClick={() => onPreview(sub)}>
                          <VisibilityIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}
                    <Tooltip title={t('subscriptions.table.tooltips.copy')}>
                      <IconButton size="small" color="secondary" onClick={() => onCopy(sub)}>
                        <ContentCopyIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('common.edit')}>
                      <IconButton size="small" onClick={() => onEdit(sub)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('subscriptions.table.tooltips.share')}>
                      <IconButton size="small" onClick={() => onClient(sub)}>
                        <QrCode2Icon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title={t('subscriptions.table.tooltips.accessLogs')}>
                      <IconButton size="small" onClick={() => onLogs(sub)}>
                        <HistoryIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    {!isProxyTab && (
                      <Tooltip title={t('subscriptions.table.tooltips.chainProxy')}>
                        <IconButton size="small" color="warning" onClick={() => onChainProxy(sub)}>
                          <AccountTreeIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}
                    {!isProxyTab &&
                      (sortingSubId !== sub.ID ? (
                        <Tooltip title={t('subscriptions.table.tooltips.sort')}>
                          <IconButton size="small" onClick={() => onStartSort(sub)}>
                            <SortIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      ) : (
                        <>
                          <Tooltip title={t('common.confirm')}>
                            <IconButton size="small" color="success" onClick={() => onConfirmSort(sub)}>
                              <CheckIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title={t('common.cancel')}>
                            <IconButton size="small" onClick={onCancelSort}>
                              <CloseIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </>
                      ))}
                    <Tooltip title={t('common.delete')}>
                      <IconButton size="small" color="error" onClick={() => onDelete(sub)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </Stack>
                </TableCell>
              </TableRow>
              {!isProxyTab && (
                <TableRow>
                  <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6}>
                    <Collapse in={expandedRows[sub.ID] || sortingSubId === sub.ID} timeout="auto" unmountOnExit>
                      <Box sx={{ margin: 2 }}>
                        {sortingSubId === sub.ID ? (
                          <SortableNodeList
                            items={tempSortData}
                            onDragEnd={onDragEnd}
                            selectedItems={selectedSortItems}
                            onToggleSelect={onToggleSortSelect}
                            onSelectAll={onSelectAllSort}
                            onClearSelection={onClearSortSelection}
                            onBatchSort={onBatchSort}
                            onBatchMove={onBatchMove}
                          />
                        ) : (
                          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                            {getSortedItems(sub).map((item, idx) =>
                              item._type === 'node' ? (
                                <Chip
                                  key={item._type + item.ID}
                                  label={getNodeDisplayName(item)}
                                  size="small"
                                  variant="outlined"
                                  color="success"
                                  onClick={() => onCopyToClipboard(item.Link)}
                                />
                              ) : item._type === 'airport' ? (
                                <Chip
                                  key={item._type + item.ID}
                                  label={t('subscriptions.sort.airportLabel', { name: item.Name })}
                                  size="small"
                                  variant="outlined"
                                  color="info"
                                />
                              ) : (
                                <Chip
                                  key={item._type + idx}
                                  label={t('subscriptions.sort.groupLabel', { name: item.Name })}
                                  size="small"
                                  variant="outlined"
                                  color="warning"
                                />
                              )
                            )}
                          </Stack>
                        )}
                      </Box>
                    </Collapse>
                  </TableCell>
                </TableRow>
              )}
            </Fragment>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
