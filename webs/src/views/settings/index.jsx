import { useState } from 'react';

// material-ui
import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';

// icons
import PersonIcon from '@mui/icons-material/Person';
import WebhookIcon from '@mui/icons-material/Webhook';
import TelegramIcon from '@mui/icons-material/Telegram';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import ProfileSettings from './components/ProfileSettings';
import WebhookSettings from './components/WebhookSettings';
import TelegramSettings from './components/TelegramSettings';

// ==============================|| Tab Panel ||============================== //

function TabPanel({ children, value, index, ...other }) {
  return (
    <div role="tabpanel" hidden={value !== index} id={`settings-tabpanel-${index}`} aria-labelledby={`settings-tab-${index}`} {...other}>
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  );
}

function a11yProps(index) {
  return {
    id: `settings-tab-${index}`,
    'aria-controls': `settings-tabpanel-${index}`
  };
}

// ==============================|| 用户中心 ||============================== //

export default function UserSettings() {
  const [tabValue, setTabValue] = useState(0);
  const [loading, setLoading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const showMessage = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  return (
    <MainCard title="用户中心">
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs
          value={tabValue}
          onChange={handleTabChange}
          aria-label="settings tabs"
          variant="scrollable"
          scrollButtons="auto"
          allowScrollButtonsMobile
          sx={{
            '& .MuiTab-root': {
              minHeight: 48,
              textTransform: 'none',
              fontSize: '0.95rem',
              fontWeight: 500
            }
          }}
        >
          <Tab icon={<PersonIcon sx={{ mr: 1 }} />} iconPosition="start" label="个人设置" {...a11yProps(0)} />
          <Tab icon={<WebhookIcon sx={{ mr: 1 }} />} iconPosition="start" label="Webhook 设置" {...a11yProps(1)} />
          <Tab
            icon={<TelegramIcon sx={{ mr: 1, color: tabValue === 2 ? '#0088cc' : 'inherit' }} />}
            iconPosition="start"
            label="Telegram 机器人"
            {...a11yProps(2)}
          />
        </Tabs>
      </Box>

      <TabPanel value={tabValue} index={0}>
        <ProfileSettings showMessage={showMessage} loading={loading} setLoading={setLoading} />
      </TabPanel>

      <TabPanel value={tabValue} index={1}>
        <WebhookSettings showMessage={showMessage} loading={loading} setLoading={setLoading} />
      </TabPanel>

      <TabPanel value={tabValue} index={2}>
        <TelegramSettings showMessage={showMessage} loading={loading} setLoading={setLoading} />
      </TabPanel>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>
    </MainCard>
  );
}
