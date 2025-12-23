import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import Chip from '@mui/material/Chip';
import Checkbox from '@mui/material/Checkbox';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import SelectAllIcon from '@mui/icons-material/SelectAll';
import DeselectIcon from '@mui/icons-material/Deselect';
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
  // 全选/取消全选
  const allSelected = items.length > 0 && selectedItems.length === items.length;
  const someSelected = selectedItems.length > 0 && selectedItems.length < items.length;

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
      <Box sx={{ mb: 1, display: 'flex', gap: 1 }}>
        <Button
          size="small"
          variant={allSelected ? 'contained' : 'outlined'}
          color={allSelected ? 'primary' : 'inherit'}
          startIcon={allSelected ? <DeselectIcon /> : <SelectAllIcon />}
          onClick={handleToggleAll}
        >
          {allSelected ? '取消全选' : '全选'}
        </Button>
        {someSelected && <Chip label={`已选 ${selectedItems.length}/${items.length}`} size="small" color="primary" variant="outlined" />}
      </Box>

      {/* 拖拽列表 */}
      <DragDropContext onDragEnd={onDragEnd}>
        <Droppable droppableId="sortList">
          {(provided) => (
            <List {...provided.droppableProps} ref={provided.innerRef} dense>
              {items.map((item, index) => {
                const isSelected = selectedItems.includes(item.Name);
                return (
                  <Draggable key={item.Name} draggableId={item.Name} index={index}>
                    {(provided, snapshot) => (
                      <ListItem
                        ref={provided.innerRef}
                        {...provided.draggableProps}
                        {...provided.dragHandleProps}
                        sx={{
                          bgcolor: snapshot.isDragging ? 'action.selected' : isSelected ? 'action.hover' : 'background.paper',
                          border: '1px solid',
                          borderColor: isSelected ? 'primary.main' : 'divider',
                          borderRadius: 1,
                          mb: 0.5,
                          transition: 'all 0.2s'
                        }}
                      >
                        {/* 多选复选框 */}
                        <Checkbox
                          size="small"
                          checked={isSelected}
                          onChange={() => onToggleSelect && onToggleSelect(item.Name)}
                          sx={{ p: 0.5, mr: 0.5 }}
                        />
                        <DragIndicatorIcon sx={{ mr: 1, color: 'text.secondary' }} />
                        <Chip
                          label={item.IsGroup ? `📁 ${item.Name} (分组)` : item.Name}
                          color={item.IsGroup ? 'warning' : 'success'}
                          variant="outlined"
                          size="small"
                        />
                        {/* 显示索引 */}
                        <Chip label={`#${index + 1}`} size="small" sx={{ ml: 'auto', minWidth: 40 }} />
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
