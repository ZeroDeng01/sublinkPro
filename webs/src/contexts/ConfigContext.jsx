import PropTypes from 'prop-types';
import { createContext, useMemo, useState, useEffect, useCallback } from 'react';

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
  const [features, setFeatures] = useState([]);

  useEffect(() => {
    const fetchVersion = async () => {
      try {
        const res = await request({ url: '/v1/version', method: 'get' });
        if (res?.data) {
          // 兼容新旧格式：新格式为 { version, features }，旧格式直接返回版本字符串
          if (typeof res.data === 'object') {
            setVersion(res.data.version || '');
            setFeatures(res.data.features || []);
          } else {
            setVersion(res.data);
            setFeatures([]);
          }
        }
      } catch (error) {
        console.error('Failed to fetch version:', error);
      }
    };
    fetchVersion();
  }, []);

  // 检查指定功能是否启用
  const isFeatureEnabled = useCallback((featureName) => features.includes(featureName), [features]);

  const memoizedValue = useMemo(
    () => ({ state, setState, setField, resetState, version, features, isFeatureEnabled }),
    [state, setField, setState, resetState, version, features, isFeatureEnabled]
  );

  return <ConfigContext.Provider value={memoizedValue}>{children}</ConfigContext.Provider>;
}

ConfigProvider.propTypes = { children: PropTypes.node };
