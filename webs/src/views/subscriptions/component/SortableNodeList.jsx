import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { useTheme } from '@mui/material/styles';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import Chip from '@mui/material/Chip';
import Checkbox from '@mui/material/Checkbox';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import SelectAllIcon from '@mui/icons-material/SelectAll';
import DeselectIcon from '@mui/icons-material/Deselect';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';
import { getReadableTextTokens, getSurfaceTokens } from 'themes/surfaceTokens';
import { withAlpha } from 'utils/colorUtils';
import SortToolbar from './SortToolbar';

/**
 * 可拖拽排序的节点/分组列表
 * 支持多选和批量移动
 */
export default function SortableNodeList({
  items,
  onDragEnd,
  selectedItems = [],
  onToggleSelect,
  onSelectAll,
  onClearSelection,
  onBatchSort,
  onBatchMove
}) {
  const theme = useTheme();
  const { isDark } = useResolvedColorScheme();
  const { palette, mutedPanelSurface, nestedPanelSurface, panelBorder } = getSurfaceTokens(theme, isDark);
  const { primaryText, secondaryText, tertiaryText } = getReadableTextTokens(theme, isDark);

  // 全选/取消全选
  const allSelected = items.length > 0 && selectedItems.length === items.length;
  const someSelected = selectedItems.length > 0 && selectedItems.length < items.length;
  const shellInset = isDark ? `inset 0 1px 0 ${withAlpha(palette.common.white, 0.03)}` : 'none';
  const neutralBorder = withAlpha(palette.divider, isDark ? 0.66 : 0.84);
  const hoverBorder = withAlpha(palette.primary.main, isDark ? 0.28 : 0.16);
  const selectedBorder = withAlpha(palette.primary.main, isDark ? 0.42 : 0.26);
  const rowDefaultSurface = isDark ? withAlpha(palette.background.default, 0.58) : palette.background.paper;
  const rowHoverSurface = withAlpha(palette.primary.main, isDark ? 0.1 : 0.05);
  const rowSelectedSurface = withAlpha(palette.primary.main, isDark ? 0.18 : 0.1);
  const rowSelectedHoverSurface = withAlpha(palette.primary.main, isDark ? 0.24 : 0.14);
  const rowDraggingSurface = withAlpha(palette.primary.main, isDark ? 0.22 : 0.12);
  const actionSurface = withAlpha(palette.primary.main, isDark ? 0.12 : 0.06);
  const actionHoverSurface = withAlpha(palette.primary.main, isDark ? 0.18 : 0.1);
  const actionBorder = withAlpha(palette.primary.main, isDark ? 0.3 : 0.18);

  const handleToggleAll = () => {
    if (allSelected) {
      onClearSelection && onClearSelection();
    } else {
      onSelectAll && onSelectAll();
    }
  };

  return (
    <Box>
      {/* 排序工具栏 */}
      <SortToolbar
        selectedItems={selectedItems}
        onBatchSort={onBatchSort}
        onBatchMove={onBatchMove}
        onClearSelection={onClearSelection}
        totalItems={items.length}
      />

      {/* 全选/取消全选按钮 */}
      <Stack
        direction="row"
        spacing={1}
        alignItems="center"
        flexWrap="wrap"
        sx={{
          mb: 1.25,
          p: 1.25,
          borderRadius: 2.5,
          bgcolor: mutedPanelSurface,
          border: '1px solid',
          borderColor: panelBorder,
          boxShadow: shellInset
        }}
      >
        <Button
          size="small"
          variant="outlined"
          startIcon={allSelected ? <DeselectIcon /> : <SelectAllIcon />}
          onClick={handleToggleAll}
          sx={{
            borderColor: allSelected ? selectedBorder : neutralBorder,
            bgcolor: allSelected ? actionSurface : nestedPanelSurface,
            color: allSelected ? 'primary.main' : secondaryText,
            fontWeight: 700,
            '&:hover': {
              borderColor: hoverBorder,
              bgcolor: allSelected ? actionHoverSurface : rowHoverSurface
            }
          }}
        >
          {allSelected ? '取消全选' : '全选'}
        </Button>
        {someSelected && (
          <Chip
            label={`已选 ${selectedItems.length}/${items.length}`}
            size="small"
            variant="outlined"
            sx={{
              bgcolor: actionSurface,
              color: 'primary.main',
              borderColor: actionBorder,
              fontWeight: 600
            }}
          />
        )}
        {!someSelected && !allSelected && (
          <Typography variant="caption" sx={{ color: tertiaryText }}>
            共 {items.length} 项，可多选后批量排序或移动
          </Typography>
        )}
      </Stack>

      {/* 拖拽列表 */}
      <DragDropContext onDragEnd={onDragEnd}>
        <Droppable droppableId="sortList">
          {(provided) => (
            <List
              {...provided.droppableProps}
              ref={provided.innerRef}
              dense
              sx={{
                p: 1,
                borderRadius: 2.5,
                bgcolor: nestedPanelSurface,
                border: '1px solid',
                borderColor: panelBorder,
                boxShadow: shellInset
              }}
            >
              {items.map((item, index) => {
                const isSelected = selectedItems.includes(item.Name);
                const itemAccentColor = item.IsGroup ? palette.warning.main : palette.success.main;
                return (
                  <Draggable key={item.Name} draggableId={item.Name} index={index}>
                    {(provided, snapshot) => (
                      <ListItem
                        ref={provided.innerRef}
                        {...provided.draggableProps}
                        {...provided.dragHandleProps}
                        disableGutters
                        sx={{
                          px: 1.25,
                          py: 1,
                          bgcolor: snapshot.isDragging ? rowDraggingSurface : isSelected ? rowSelectedSurface : rowDefaultSurface,
                          border: '1px solid',
                          borderColor: snapshot.isDragging ? selectedBorder : isSelected ? selectedBorder : neutralBorder,
                          borderRadius: 2,
                          mb: 0.75,
                          boxShadow: snapshot.isDragging ? theme.shadows[6] : shellInset,
                          transform: snapshot.isDragging ? 'rotate(0.25deg)' : 'none',
                          transition: 'all 0.2s ease',
                          '&:hover': {
                            bgcolor: isSelected ? rowSelectedHoverSurface : rowHoverSurface,
                            borderColor: hoverBorder
                          }
                        }}
                      >
                        {/* 多选复选框 */}
                        <Checkbox
                          size="small"
                          checked={isSelected}
                          onChange={() => onToggleSelect && onToggleSelect(item.Name)}
                          sx={{ p: 0.5, mr: 0.75 }}
                        />
                        <DragIndicatorIcon sx={{ mr: 1, color: secondaryText, flexShrink: 0 }} />
                        <Chip
                          label={item.IsGroup ? `📁 ${item.Name} (分组)` : item.Name}
                          variant="outlined"
                          size="small"
                          sx={{
                            maxWidth: 'calc(100% - 72px)',
                            bgcolor: withAlpha(itemAccentColor, isDark ? 0.14 : 0.07),
                            color: itemAccentColor,
                            borderColor: withAlpha(itemAccentColor, isDark ? 0.28 : 0.18),
                            '& .MuiChip-label': {
                              display: 'block',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                              maxWidth: '100%'
                            }
                          }}
                        />
                        {/* 显示索引 */}
                        <Chip
                          label={`#${index + 1}`}
                          size="small"
                          sx={{
                            ml: 'auto',
                            minWidth: 40,
                            bgcolor: mutedPanelSurface,
                            color: primaryText,
                            border: '1px solid',
                            borderColor: neutralBorder,
                            fontWeight: 600
                          }}
                        />
                      </ListItem>
                    )}
                  </Draggable>
                );
              })}
              {provided.placeholder}
            </List>
          )}
        </Droppable>
      </DragDropContext>
    </Box>
  );
}
