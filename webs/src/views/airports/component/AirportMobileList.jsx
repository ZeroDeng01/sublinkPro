import PropTypes from 'prop-types';

// material-ui
import { useTheme, alpha } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

// icons
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import LinkIcon from '@mui/icons-material/Link';
import AccessTimeIcon from '@mui/icons-material/AccessTime';

// utils
import { formatDateTime } from '../utils';

/**
 * 机场移动端列表组件
 */
export default function AirportMobileList({ airports, onEdit, onDelete, onPull }) {
    const theme = useTheme();

    if (airports.length === 0) {
        return (
            <Box sx={{ py: 6, textAlign: 'center' }}>
                <Typography variant="body2" color="textSecondary">
                    暂无机场数据，点击上方"添加"按钮添加
                </Typography>
            </Box>
        );
    }

    return (
        <Stack spacing={2}>
            {airports.map((airport) => (
                <Card
                    key={airport.id}
                    sx={{
                        borderRadius: 3,
                        border: `1px solid ${alpha(theme.palette.divider, 0.1)}`,
                        transition: 'all 0.2s ease',
                        '&:hover': {
                            transform: 'translateY(-2px)',
                            boxShadow: theme.shadows[4]
                        }
                    }}
                >
                    <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
                        {/* 顶部：名称和状态 */}
                        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
                            <Typography variant="subtitle1" fontWeight={600}>
                                {airport.name}
                            </Typography>
                            <Chip
                                label={airport.enabled ? '启用' : '禁用'}
                                color={airport.enabled ? 'success' : 'default'}
                                size="small"
                            />
                        </Box>

                        {/* 订阅地址 */}
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                            <LinkIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
                            <Typography
                                variant="body2"
                                color="textSecondary"
                                sx={{
                                    overflow: 'hidden',
                                    textOverflow: 'ellipsis',
                                    whiteSpace: 'nowrap',
                                    flex: 1
                                }}
                            >
                                {airport.url}
                            </Typography>
                        </Box>

                        {/* 信息行 */}
                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 2 }}>
                            <Chip
                                label={`${airport.nodeCount || 0} 节点`}
                                color="primary"
                                variant="outlined"
                                size="small"
                            />
                            {airport.group && (
                                <Chip label={airport.group} variant="outlined" size="small" />
                            )}
                            <Chip
                                icon={<AccessTimeIcon sx={{ fontSize: '14px !important' }} />}
                                label={airport.cronExpr}
                                variant="outlined"
                                size="small"
                            />
                        </Box>

                        {/* 时间信息 */}
                        <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
                            <Box>
                                <Typography variant="caption" color="textSecondary">
                                    上次运行
                                </Typography>
                                <Typography variant="body2">
                                    {formatDateTime(airport.lastRunTime)}
                                </Typography>
                            </Box>
                            <Box>
                                <Typography variant="caption" color="textSecondary">
                                    下次运行
                                </Typography>
                                <Typography variant="body2">
                                    {formatDateTime(airport.nextRunTime)}
                                </Typography>
                            </Box>
                        </Box>

                        {/* 操作按钮 */}
                        <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
                            <IconButton
                                size="small"
                                onClick={() => onPull(airport)}
                                sx={{
                                    bgcolor: alpha(theme.palette.primary.main, 0.1),
                                    color: theme.palette.primary.main,
                                    '&:hover': { bgcolor: alpha(theme.palette.primary.main, 0.2) }
                                }}
                            >
                                <PlayArrowIcon fontSize="small" />
                            </IconButton>
                            <IconButton
                                size="small"
                                onClick={() => onEdit(airport)}
                                sx={{
                                    bgcolor: alpha(theme.palette.info.main, 0.1),
                                    color: theme.palette.info.main,
                                    '&:hover': { bgcolor: alpha(theme.palette.info.main, 0.2) }
                                }}
                            >
                                <EditIcon fontSize="small" />
                            </IconButton>
                            <IconButton
                                size="small"
                                onClick={() => onDelete(airport)}
                                sx={{
                                    bgcolor: alpha(theme.palette.error.main, 0.1),
                                    color: theme.palette.error.main,
                                    '&:hover': { bgcolor: alpha(theme.palette.error.main, 0.2) }
                                }}
                            >
                                <DeleteIcon fontSize="small" />
                            </IconButton>
                        </Box>
                    </CardContent>
                </Card>
            ))}
        </Stack>
    );
}

AirportMobileList.propTypes = {
    airports: PropTypes.array.isRequired,
    onEdit: PropTypes.func.isRequired,
    onDelete: PropTypes.func.isRequired,
    onPull: PropTypes.func.isRequired
};
