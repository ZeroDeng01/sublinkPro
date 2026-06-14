import { useCallback, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams, useLocation } from 'react-router-dom';

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
import FilterAltIcon from '@mui/icons-material/FilterAlt';
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
import GlobalNodeProcessingSettings from './components/GlobalNodeProcessingSettings';
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
  const [searchParams, setSearchParams] = useSearchParams();
  const location = useLocation();
  const [tabValue, setTabValue] = useState(() => {
    // 只在首次加载时读取 URL 参数
    const tab = searchParams.get('tab');
    if (tab === 'ai') return 5;
    if (tab === 'globalNodeProcessing') return 4;
    // 或者从 location.state 读取
    if (location.state?.targetTab === 'globalNodeProcessing') return 4;
    if (location.state?.targetTab === 'ai') return 5;
    return 0;
  });
  const [loading, setLoading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  // 首次加载后清除 URL 参数，避免刷新时又跳回来
  useEffect(() => {
    if (searchParams.has('tab')) {
      // 清除 URL 参数，但不改变标签状态
      setSearchParams({}, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const handleTabChange = (_event, newValue) => {
    setTabValue(newValue);
  };

  const showMessage = useCallback((message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  }, []);

  const handleSnackbarClose = useCallback(() => {
    setSnackbar((prev) => ({ ...prev, open: false }));
  }, []);

  // 获取当前标签页的标题
  const getCurrentTabTitle = () => {
    const tabTitles = [
      t('settings.tabs.profile'),
      t('settings.tabs.subscriptionAddress'),
      t('settings.tabs.telegram'),
      t('settings.tabs.nodeDedup'),
      t('settings.tabs.globalNodeProcessing'),
      t('settings.tabs.aiAssistant'),
      'Cloudflare Tunnel',
      t('settings.tabs.subStore'),
      t('settings.tabs.dataMigration')
    ];
    return tabTitles[tabValue] || t('settings.title');
  };

  return (
    <MainCard title={getCurrentTabTitle()}>
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
          <Tab
            icon={<FilterAltIcon sx={{ mr: 1 }} />}
            iconPosition="start"
            label={t('settings.tabs.globalNodeProcessing')}
            {...a11yProps(4)}
          />
          <Tab icon={<PsychologyIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.aiAssistant')} {...a11yProps(5)} />
          <Tab icon={<CloudQueueIcon sx={{ mr: 1 }} />} iconPosition="start" label="Cloudflare Tunnel" {...a11yProps(6)} />
          <Tab icon={<ExtensionIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.subStore')} {...a11yProps(7)} />
          <Tab icon={<StorageIcon sx={{ mr: 1 }} />} iconPosition="start" label={t('settings.tabs.dataMigration')} {...a11yProps(8)} />
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
        <GlobalNodeProcessingSettings showMessage={showMessage} />
      </TabPanel>

      <TabPanel value={tabValue} index={5}>
        <AIAssistantSettings showMessage={showMessage} loading={loading} setLoading={setLoading} />
      </TabPanel>

      <TabPanel value={tabValue} index={6}>
        <CloudflareTunnelSettings showMessage={showMessage} />
      </TabPanel>

      <TabPanel value={tabValue} index={7}>
        <SubStoreSettings showMessage={showMessage} />
      </TabPanel>

      <TabPanel value={tabValue} index={8}>
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
