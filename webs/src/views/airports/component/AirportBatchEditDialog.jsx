import PropTypes from 'prop-types';

// material-ui
import Alert from '@mui/material/Alert';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControlLabel from '@mui/material/FormControlLabel';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

// project imports
import CronExpressionGenerator from 'components/CronExpressionGenerator';

/**
 * 机场批量编辑对话框
 */
export default function AirportBatchEditDialog({
  open,
  selectedCount,
  batchForm,
  setBatchForm,
  groupOptions,
  onClose,
  onSubmit,
  submitting
}) {
  const summaryItems = [];

  if (batchForm.applyGroup) {
    summaryItems.push(`分组：${batchForm.group.trim() ? batchForm.group.trim() : '清空分组'}`);
  }
  if (batchForm.applySchedule) {
    summaryItems.push(`调度：${batchForm.cronExpr.trim() || '未设置'}`);
  }

  return (
    <Dialog open={open} onClose={submitting ? undefined : onClose} maxWidth="md" fullWidth>
      <DialogTitle>批量设置机场</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2.5}>
          <Alert severity={summaryItems.length > 0 ? 'info' : 'warning'}>
            {summaryItems.length > 0
              ? `将更新 ${selectedCount} 个机场：${summaryItems.join('；')}`
              : `已选择 ${selectedCount} 个机场，请先勾选本次要修改的字段`}
          </Alert>

          <Box>
            <FormControlLabel
              control={
                <Checkbox checked={batchForm.applyGroup} onChange={(e) => setBatchForm({ ...batchForm, applyGroup: e.target.checked })} />
              }
              label="统一设置节点分组"
            />
            <Collapse in={batchForm.applyGroup}>
              <Stack spacing={1.5} sx={{ mt: 1 }}>
                <Autocomplete
                  freeSolo
                  size="small"
                  options={groupOptions}
                  value={batchForm.group}
                  onChange={(e, newValue) => setBatchForm({ ...batchForm, group: newValue || '' })}
                  onInputChange={(e, newValue) => setBatchForm({ ...batchForm, group: newValue ?? '' })}
                  renderInput={(params) => <TextField {...params} label="节点分组" placeholder="输入或选择分组，留空表示清空分组" />}
                />
                <Typography variant="caption" color="textSecondary">
                  会同步更新这些机场已导入节点的分组，留空表示清空分组。
                </Typography>
              </Stack>
            </Collapse>
          </Box>

          <Box>
            <FormControlLabel
              control={
                <Checkbox
                  checked={batchForm.applySchedule}
                  onChange={(e) => setBatchForm({ ...batchForm, applySchedule: e.target.checked })}
                />
              }
              label="统一设置定时更新"
            />
            <Collapse in={batchForm.applySchedule}>
              <Box sx={{ mt: 1 }}>
                <CronExpressionGenerator
                  value={batchForm.cronExpr}
                  onChange={(value) => setBatchForm({ ...batchForm, cronExpr: value })}
                  label=""
                  helperText="只修改 Cron 表达式，不会改变机场当前的启用或禁用状态。"
                />
              </Box>
            </Collapse>
          </Box>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={submitting}>
          取消
        </Button>
        <Button variant="contained" onClick={onSubmit} disabled={submitting}>
          {submitting ? '保存中...' : '确认批量更新'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

AirportBatchEditDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  selectedCount: PropTypes.number.isRequired,
  batchForm: PropTypes.shape({
    applyGroup: PropTypes.bool.isRequired,
    group: PropTypes.string.isRequired,
    applySchedule: PropTypes.bool.isRequired,
    cronExpr: PropTypes.string.isRequired
  }).isRequired,
  setBatchForm: PropTypes.func.isRequired,
  groupOptions: PropTypes.array.isRequired,
  onClose: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  submitting: PropTypes.bool
};

AirportBatchEditDialog.defaultProps = {
  submitting: false
};
