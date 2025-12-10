import PropTypes from "prop-types";

// material-ui
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import FormControlLabel from "@mui/material/FormControlLabel";
import Typography from "@mui/material/Typography";

/**
 * 删除订阅确认对话框
 */
export default function DeleteSchedulerDialog({ open, scheduler, withNodes, setWithNodes, onClose, onConfirm }) {
  if (!scheduler) return null;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>删除订阅</DialogTitle>
      <DialogContent>
        <Typography variant="body1" gutterBottom>
          确定要删除订阅 &quot;{scheduler.Name}&quot; 吗？
        </Typography>
        {(scheduler.NodeCount || 0) > 0 && (
          <>
            <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
              该订阅关联了 {scheduler.NodeCount || 0} 个节点
            </Typography>
            <FormControlLabel
              control={<Checkbox checked={withNodes} onChange={(e) => setWithNodes(e.target.checked)} />}
              label="同时删除关联的节点"
            />
            {!withNodes && (
              <Alert severity="info" sx={{ mt: 1 }}>
                保留的节点将变为手动添加的节点，不再与此订阅关联
              </Alert>
            )}
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button onClick={onConfirm} color="error" variant="contained">
          确认删除
        </Button>
      </DialogActions>
    </Dialog>
  );
}

DeleteSchedulerDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  scheduler: PropTypes.shape({
    Name: PropTypes.string,
    NodeCount: PropTypes.number
  }),
  withNodes: PropTypes.bool.isRequired,
  setWithNodes: PropTypes.func.isRequired,
  onClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired
};
