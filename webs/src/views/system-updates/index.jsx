import Box from '@mui/material/Box';

import { ReleaseLogPanel, StarReminderCard } from 'components/SystemUpdatePanels';

export default function SystemUpdatesPage() {
  return (
    <Box sx={{ pb: 3 }}>
      <StarReminderCard />
      <ReleaseLogPanel />
    </Box>
  );
}
