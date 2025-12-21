import { useState, useEffect } from 'react';
// import { useTheme } from '@mui/material/styles'; // Unused
import { Box } from '@mui/material';
import { getCountryStats } from 'api/total';
import NodeMap from './NodeMap';

const NodeMapPage = () => {
  const [countryStats, setCountryStats] = useState({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true);
        const res = await getCountryStats();
        setCountryStats(res.data || {});
      } catch (error) {
        console.error('Failed to fetch country stats:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchStats();
  }, []);

  return (
    <Box sx={{ height: 'calc(100vh - 120px)', width: '100%', position: 'relative' }}>
      <NodeMap data={countryStats} loading={loading} />
    </Box>
  );
};

export default NodeMapPage;
