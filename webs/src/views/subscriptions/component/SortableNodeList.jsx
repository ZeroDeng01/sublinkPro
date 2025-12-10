import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import Chip from '@mui/material/Chip';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';

/**
 * 可拖拽排序的节点/分组列表
 */
export default function SortableNodeList({ items, onDragEnd }) {
  return (
    <DragDropContext onDragEnd={onDragEnd}>
      <Droppable droppableId="sortList">
        {(provided) => (
          <List {...provided.droppableProps} ref={provided.innerRef} dense>
            {items.map((item, index) => (
              <Draggable key={item.Name} draggableId={item.Name} index={index}>
                {(provided, snapshot) => (
                  <ListItem
                    ref={provided.innerRef}
                    {...provided.draggableProps}
                    {...provided.dragHandleProps}
                    sx={{
                      bgcolor: snapshot.isDragging ? 'action.selected' : 'background.paper',
                      border: '1px solid',
                      borderColor: 'divider',
                      borderRadius: 1,
                      mb: 0.5
                    }}
                  >
                    <DragIndicatorIcon sx={{ mr: 1, color: 'text.secondary' }} />
                    <Chip
                      label={item.IsGroup ? `📁 ${item.Name} (分组)` : item.Name}
                      color={item.IsGroup ? 'warning' : 'success'}
                      variant="outlined"
                      size="small"
                    />
                  </ListItem>
                )}
              </Draggable>
            ))}
            {provided.placeholder}
          </List>
        )}
      </Droppable>
    </DragDropContext>
  );
}
