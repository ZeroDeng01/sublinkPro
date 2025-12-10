import PropTypes from "prop-types";

// material-ui
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";

// icons
import DeleteIcon from "@mui/icons-material/Delete";

/**
 * 批量操作栏
 */
export default function BatchActions({ selectedCount, onDelete, onGroup, onDialerProxy }) {
  if (selectedCount === 0) return null;

  return (
    <Stack direction="row" spacing={1} sx={{ mb: 2 }}>
      <Typography variant="body2" sx={{ alignSelf: "center" }}>
        已选择 {selectedCount} 个节点
      </Typography>
      <Button size="small" color="error" startIcon={<DeleteIcon />} onClick={onDelete}>
        批量删除
      </Button>
      <Button size="small" color="primary" variant="outlined" onClick={onGroup}>
        修改分组
      </Button>
      <Button size="small" color="primary" variant="outlined" onClick={onDialerProxy}>
        修改前置代理
      </Button>
    </Stack>
  );
}

BatchActions.propTypes = {
  selectedCount: PropTypes.number.isRequired,
  onDelete: PropTypes.func.isRequired,
  onGroup: PropTypes.func.isRequired,
  onDialerProxy: PropTypes.func.isRequired
};
