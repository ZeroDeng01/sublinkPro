import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';

// icons
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';

// utils
import { formatDateTime } from '../utils';

/**
 * 机场列表表格组件
 */
export default function AirportTable({ airports, onEdit, onDelete, onPull }) {
    const theme = useTheme();

    return (
        <TableContainer>
            <Table size="small">
                <TableHead>
                    <TableRow>
                        <TableCell>名称</TableCell>
                        <TableCell>订阅地址</TableCell>
                        <TableCell align="center">节点数量</TableCell>
                        <TableCell>上次运行</TableCell>
                        <TableCell>下次运行</TableCell>
                        <TableCell>Cron表达式</TableCell>
                        <TableCell>分组</TableCell>
                        <TableCell align="center">状态</TableCell>
                        <TableCell align="right">操作</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {airports.length === 0 ? (
                        <TableRow>
                            <TableCell colSpan={9} align="center">
                                <Typography variant="body2" color="textSecondary" sx={{ py: 4 }}>
                                    暂无机场数据，点击上方"添加机场"按钮添加
                                </Typography>
                            </TableCell>
                        </TableRow>
                    ) : (
                        airports.map((airport) => (
                            <TableRow key={airport.id} hover>
                                <TableCell>
                                    <Typography variant="body2" fontWeight={500}>
                                        {airport.name}
                                    </Typography>
                                </TableCell>
                                <TableCell>
                                    <Tooltip title={airport.url} arrow placement="top">
                                        <Typography
                                            variant="body2"
                                            sx={{
                                                maxWidth: 200,
                                                overflow: 'hidden',
                                                textOverflow: 'ellipsis',
                                                whiteSpace: 'nowrap'
                                            }}
                                        >
                                            {airport.url}
                                        </Typography>
                                    </Tooltip>
                                </TableCell>
                                <TableCell align="center">
                                    <Chip label={airport.nodeCount || 0} color="primary" variant="outlined" size="small" />
                                </TableCell>
                                <TableCell>
                                    <Typography variant="body2" color="textSecondary">
                                        {formatDateTime(airport.lastRunTime)}
                                    </Typography>
                                </TableCell>
                                <TableCell>
                                    <Typography variant="body2" color="textSecondary">
                                        {formatDateTime(airport.nextRunTime)}
                                    </Typography>
                                </TableCell>
                                <TableCell>
                                    <Typography variant="body2" fontFamily="monospace">
                                        {airport.cronExpr}
                                    </Typography>
                                </TableCell>
                                <TableCell>
                                    <Typography variant="body2">{airport.group || '-'}</Typography>
                                </TableCell>
                                <TableCell align="center">
                                    <Chip label={airport.enabled ? '启用' : '禁用'} color={airport.enabled ? 'success' : 'default'} size="small" />
                                </TableCell>
                                <TableCell align="right">
                                    <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 0.5 }}>
                                        <Tooltip title="立即拉取" arrow>
                                            <IconButton
                                                size="small"
                                                onClick={() => onPull(airport)}
                                                sx={{ color: theme.palette.primary.main }}
                                            >
                                                <PlayArrowIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title="编辑" arrow>
                                            <IconButton size="small" onClick={() => onEdit(airport)}>
                                                <EditIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title="删除" arrow>
                                            <IconButton size="small" color="error" onClick={() => onDelete(airport)}>
                                                <DeleteIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                    </Box>
                                </TableCell>
                            </TableRow>
                        ))
                    )}
                </TableBody>
            </Table>
        </TableContainer>
    );
}

AirportTable.propTypes = {
    airports: PropTypes.array.isRequired,
    onEdit: PropTypes.func.isRequired,
    onDelete: PropTypes.func.isRequired,
    onPull: PropTypes.func.isRequired
};
