import PropTypes from 'prop-types';
import { createContext, useMemo, useState, useEffect } from 'react';

// project imports
import request from 'api/request';

// project imports
import config from 'config';
import { useLocalStorage } from 'hooks/useLocalStorage';

// ==============================|| CONFIG CONTEXT ||============================== //

export const ConfigContext = createContext(undefined);

// ==============================|| CONFIG PROVIDER ||============================== //

export function ConfigProvider({ children }) {
  const { state, setState, setField, resetState } = useLocalStorage('berry-config-vite-js', config);
  const [version, setVersion] = useState('');

  useEffect(() => {
    const fetchVersion = async () => {
      try {
        const res = await request({ url: '/v1/version', method: 'get' });
        if (res?.data) {
          setVersion(res.data);
        }
      } catch (error) {
        console.error('Failed to fetch version:', error);
      }
    };
    fetchVersion();
  }, []);

  const memoizedValue = useMemo(
    () => ({ state, setState, setField, resetState, version }),
    [state, setField, setState, resetState, version]
  );

  return <ConfigContext.Provider value={memoizedValue}>{children}</ConfigContext.Provider>;
}

ConfigProvider.propTypes = { children: PropTypes.node };
