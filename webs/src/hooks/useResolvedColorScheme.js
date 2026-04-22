import { useMemo } from 'react';

import { useColorScheme, useTheme } from '@mui/material/styles';

const normalizeMode = (mode) => (mode === 'dark' ? 'dark' : 'light');

export default function useResolvedColorScheme() {
  const theme = useTheme();
  const colorSchemeApi = useColorScheme();

  const mode = colorSchemeApi?.mode;
  const systemMode = colorSchemeApi?.systemMode;

  const resolvedMode = useMemo(() => {
    if (mode === 'system') {
      return normalizeMode(systemMode || theme.palette.mode);
    }

    if (mode === 'light' || mode === 'dark') {
      return mode;
    }

    return normalizeMode(theme.palette.mode);
  }, [mode, systemMode, theme.palette.mode]);

  return {
    ...colorSchemeApi,
    mode,
    systemMode,
    resolvedMode,
    isDark: resolvedMode === 'dark'
  };
}
