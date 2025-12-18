import PropTypes from 'prop-types';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

// icons
import DeleteIcon from '@mui/icons-material/Delete';
import ClearIcon from '@mui/icons-material/Clear';

/**
 * 批量操作栏
 */
export default function BatchActions({
  selectedCount,
  totalCount,
  onSelectAll,
  onClearSelection,
  onDelete,
  onGroup,
  onSource,
  onDialerProxy,
  onTag,
  onRemoveTag
}) {
  // 是否全选（当前页全选或所有符合条件节点全选）
  const isAllSelected = selectedCount > 0 && selectedCount >= totalCount;
  const isIndeterminate = selectedCount > 0 && selectedCount < totalCount;

  return (
    <Stack
      direction="row"
      spacing={1}
      sx={{
        mb: 2,
        flexWrap: 'wrap',
        gap: 1,
        alignItems: 'center'
      }}
    >
      {/* 全选复选框 */}
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <Checkbox checked={isAllSelected} indeterminate={isIndeterminate} onChange={(e) => onSelectAll(e)} size="small" sx={{ p: 0.5 }} />
        <Typography
          variant="body2"
          sx={{
            ml: 0.5,
            cursor: 'pointer',
            userSelect: 'none',
            whiteSpace: 'nowrap'
          }}
          onClick={(e) => onSelectAll({ target: { checked: !isAllSelected } })}
        >
          全选
        </Typography>
      </Box>

      {/* 已选择数量显示 */}
      {selectedCount > 0 && (
        <>
          <Typography variant="body2" sx={{ alignSelf: 'center', whiteSpace: 'nowrap' }}>
            已选择 {selectedCount} 个节点
          </Typography>

          {/* 清除选择按钮 */}
          <Button size="small" color="inherit" startIcon={<ClearIcon />} onClick={onClearSelection} sx={{ whiteSpace: 'nowrap' }}>
            取消
          </Button>

          <Button size="small" color="error" startIcon={<DeleteIcon />} onClick={onDelete} sx={{ whiteSpace: 'nowrap' }}>
            批量删除
          </Button>
          <Button size="small" color="primary" variant="outlined" onClick={onGroup} sx={{ whiteSpace: 'nowrap' }}>
            修改分组
          </Button>
          <Button size="small" color="info" variant="outlined" onClick={onSource} sx={{ whiteSpace: 'nowrap' }}>
            修改来源
          </Button>
          <Button size="small" color="primary" variant="outlined" onClick={onDialerProxy} sx={{ whiteSpace: 'nowrap' }}>
            修改前置代理
          </Button>
          <Button size="small" color="secondary" variant="outlined" onClick={onTag} sx={{ whiteSpace: 'nowrap' }}>
            设置标签
          </Button>
          <Button size="small" color="error" variant="outlined" onClick={onRemoveTag} sx={{ whiteSpace: 'nowrap' }}>
            删除标签
          </Button>
        </>
      )}
    </Stack>
  );
}

BatchActions.propTypes = {
  selectedCount: PropTypes.number.isRequired,
  totalCount: PropTypes.number.isRequired,
  onSelectAll: PropTypes.func.isRequired,
  onClearSelection: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired,
  onGroup: PropTypes.func.isRequired,
  onSource: PropTypes.func.isRequired,
  onDialerProxy: PropTypes.func.isRequired,
  onTag: PropTypes.func.isRequired,
  onRemoveTag: PropTypes.func.isRequired
};
