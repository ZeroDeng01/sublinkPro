import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';

// material-ui
import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';

// icons
import PersonIcon from '@mui/icons-material/Person';
import LanguageIcon from '@mui/icons-material/Language';
import TelegramIcon from '@mui/icons-material/Telegram';
import TuneIcon from '@mui/icons-material/Tune';
import StorageIcon from '@mui/icons-material/Storage';
import PsychologyIcon from '@mui/icons-material/Psychology';
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import ExtensionIcon from '@mui/icons-material/Extension';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import ProfileSettings from './components/ProfileSettings';
import SubscriptionAddressSettings from './components/SubscriptionAddressSettings';
import TelegramSettings from './components/TelegramSettings';
import NodeDedupSettings from './components/NodeDedupSettings';
import DatabaseMigrationSettings from './components/DatabaseMigrationSettings';
import AIAssistantSettings from './components/AIAssistantSettings';
import CloudflareTunnelSettings from './components/CloudflareTunnelSettings';
import SubStoreSettings from './components/SubStoreSettings';

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
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const [tabValue, setTabValue] = useState(() => (searchParams.get('tab') === 'ai' ? 4 : 0));
  const [loading, setLoading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const handleTabChange = (_event, newValue) => {
    setTabValue(newValue);
  };

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  const handleSnackbarClose = useCallback(() => {
    setSnackbar((prev) => ({ ...prev, open: false }));
  }, []);

  return (
    <MainCard title={t('settings.title')}>
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
          <Tab icon={<PersonIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.profile')} {...a11yProps(0)} />
          <Tab
            icon={<LanguageIcon sx={{ mr: 1 }} />}
            iconPosition="start"
            label={t('settings.tabs.subscriptionAddress')}
            {...a11yProps(1)}
          />
          <Tab
            icon={<TelegramIcon sx={{ mr: 1, color: tabValue === 2 ? '#0088cc' : 'inherit' }} />}
            iconPosition="start"
            label={t('settings.tabs.telegram')}
            {...a11yProps(2)}
          />
          <Tab icon={<TuneIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.nodeDedup')} {...a11yProps(3)} />
          <Tab icon={<PsychologyIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.aiAssistant')} {...a11yProps(4)} />
          <Tab icon={<CloudQueueIcon sx={{ mr: 1 }} />} iconPosition="start" label="Cloudflare Tunnel" {...a11yProps(5)} />
          <Tab icon={<ExtensionIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.subStore')} {...a11yProps(6)} />
          <Tab icon={<StorageIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.dataMigration')} {...a11yProps(7)} />
        </Tabs>
      </Box>

      <TabPanel value={tabValue} index={0}>
        <ProfileSettings showMessage={showMessage} loading={loading} setLoading={setLoading} />
      </TabPanel>

      <TabPanel value={tabValue} index={1}>
        <SubscriptionAddressSettings showMessage={showMessage} />
      </TabPanel>

      <TabPanel value={tabValue} index={2}>
        <TelegramSettings showMessage={showMessage} loading={loading} setLoading={setLoading} />
      </TabPanel>

      <TabPanel value={tabValue} index={3}>
        <NodeDedupSettings showMessage={showMessage} />
      </TabPanel>

      <TabPanel value={tabValue} index={4}>
        <AIAssistantSettings showMessage={showMessage} loading={loading} setLoading={setLoading} />
      </TabPanel>

      <TabPanel value={tabValue} index={5}>
        <CloudflareTunnelSettings showMessage={showMessage} />
      </TabPanel>

      <TabPanel value={tabValue} index={6}>
        <SubStoreSettings showMessage={showMessage} />
      </TabPanel>

      <TabPanel value={tabValue} index={7}>
        <DatabaseMigrationSettings showMessage={showMessage} />
      </TabPanel>

      {/* 提示消息 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleSnackbarClose}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity}>{snackbar.message}</Alert>
      </Snackbar>
    </MainCard>
  );
}
