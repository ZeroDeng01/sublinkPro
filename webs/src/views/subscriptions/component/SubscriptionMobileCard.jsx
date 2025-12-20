import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Collapse from '@mui/material/Collapse';

import MainCard from 'ui-component/cards/MainCard';
import SortableNodeList from './SortableNodeList';

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
 * 移动端订阅卡片组件
 */
export default function SubscriptionMobileCard({
  subscriptions,
  expandedRows,
  sortingSubId,
  tempSortData,
  theme,
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
  getSortedItems
}) {
  return (
    <Stack spacing={2}>
      {subscriptions.map((sub) => (
        <MainCard key={sub.ID} content={false} border shadow={theme.shadows[1]}>
          <Box p={2}>
            <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1} onClick={() => onToggleRow(sub.ID)}>
              <Stack direction="row" spacing={1} alignItems="center">
                <Chip label={sub.Name} color="primary" />
                {sortingSubId === sub.ID && <Chip label="排序中" color="warning" size="small" />}
              </Stack>
              {expandedRows[sub.ID] || sortingSubId === sub.ID ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
            </Stack>

            <Typography variant="body2" sx={{ mb: 1 }}>
              {sub.Nodes?.length || 0} 个节点, {sub.Groups?.length || 0} 个分组
            </Typography>

            <Divider sx={{ my: 1 }} />

            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Typography variant="caption" color="textSecondary">
                {sub.CreateDate}
              </Typography>
              <Stack direction="row" spacing={0}>
                <IconButton size="small" color="info" onClick={() => onPreview(sub)}>
                  <VisibilityIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" onClick={() => onEdit(sub)}>
                  <EditIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" onClick={() => onClient(sub)}>
                  <QrCode2Icon fontSize="small" />
                </IconButton>
                <IconButton size="small" onClick={() => onLogs(sub)}>
                  <HistoryIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" color="warning" onClick={() => onChainProxy(sub)}>
                  <AccountTreeIcon fontSize="small" />
                </IconButton>
                {sortingSubId !== sub.ID ? (
                  <IconButton
                    size="small"
                    color="warning"
                    onClick={(e) => {
                      e.stopPropagation();
                      onStartSort(sub);
                    }}
                  >
                    <SortIcon fontSize="small" />
                  </IconButton>
                ) : (
                  <>
                    <IconButton
                      size="small"
                      color="success"
                      onClick={(e) => {
                        e.stopPropagation();
                        onConfirmSort(sub);
                      }}
                    >
                      <CheckIcon fontSize="small" />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={(e) => {
                        e.stopPropagation();
                        onCancelSort();
                      }}
                    >
                      <CloseIcon fontSize="small" />
                    </IconButton>
                  </>
                )}
                <IconButton size="small" color="secondary" onClick={() => onCopy(sub)}>
                  <ContentCopyIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" color="error" onClick={() => onDelete(sub)}>
                  <DeleteIcon fontSize="small" />
                </IconButton>
              </Stack>
            </Stack>

            {/* Expandable Content for Sort or Details */}
            <Collapse in={expandedRows[sub.ID] || sortingSubId === sub.ID} timeout="auto" unmountOnExit>
              <Box sx={{ mt: 2 }}>
                {sortingSubId === sub.ID ? (
                  <SortableNodeList items={tempSortData} onDragEnd={onDragEnd} />
                ) : (
                  <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                    {getSortedItems(sub).map((item, idx) =>
                      item._type === 'node' ? (
                        <Chip
                          key={item._type + item.ID}
                          label={item.Name}
                          size="small"
                          variant="outlined"
                          color="success"
                          onClick={() => onCopyToClipboard(item.Link)}
                          sx={{ mb: 1 }}
                        />
                      ) : (
                        <Chip
                          key={item._type + idx}
                          label={`📁 ${item.Name}`}
                          size="small"
                          variant="outlined"
                          color="warning"
                          sx={{ mb: 1 }}
                        />
                      )
                    )}
                  </Stack>
                )}
              </Box>
            </Collapse>
          </Box>
        </MainCard>
      ))}
    </Stack>
  );
}
