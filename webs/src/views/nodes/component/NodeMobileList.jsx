import PropTypes from 'prop-types';

// material-ui
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

// components
import NodeCard from './NodeCard';

/**
 * 移动端节点卡片列表
 */
export default function NodeMobileList({ nodes, page, rowsPerPage, selectedNodes, onSelect, onSpeedTest, onCopy, onEdit, onDelete }) {
  const isSelected = (node) => selectedNodes.some((n) => n.ID === node.ID);
  // 后端分页：nodes 已经是当前页数据，无需客户端切片

  return (
    <Stack spacing={2}>
      {nodes.length === 0 && (
        <Typography variant="body2" color="textSecondary" align="center" sx={{ py: 3 }}>
          暂无节点
        </Typography>
      )}
      {nodes.map((node) => (
        <NodeCard
          key={node.ID}
          node={node}
          isSelected={isSelected(node)}
          onSelect={onSelect}
          onSpeedTest={onSpeedTest}
          onCopy={onCopy}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      ))}
    </Stack>
  );
}

NodeMobileList.propTypes = {
  nodes: PropTypes.array.isRequired,
  page: PropTypes.number.isRequired,
  rowsPerPage: PropTypes.number.isRequired,
  selectedNodes: PropTypes.array.isRequired,
  onSelect: PropTypes.func.isRequired,
  onSpeedTest: PropTypes.func.isRequired,
  onCopy: PropTypes.func.isRequired,
  onEdit: PropTypes.func.isRequired,
  onDelete: PropTypes.func.isRequired
};
