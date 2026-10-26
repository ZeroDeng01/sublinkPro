import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import PublicIcon from '@mui/icons-material/Public';
import useResolvedColorScheme from 'hooks/useResolvedColorScheme';

// project imports
import CountryFlag from 'components/CountryFlag';
import { UNKNOWN_COUNTRY_FILTER } from 'utils/countryDisplay';
import { withAlpha } from 'utils/colorUtils';
import { getNodeColorChipSx, getNodeThemeTokens } from '../nodeTheme';

const normalizeCountry = (value) =>
  String(value || '')
    .trim()
    .toUpperCase();

/**
 * 节点国家分布面板
 * 展示当前筛选条件下各国家/地区的节点数量，点击国家标签可切换国家筛选
 */
export default function NodeCountryStats({ items = [], total = 0, activeCountries = [], onToggleCountry }) {
  const theme = useTheme();
  const { t } = useTranslation();
  const { isDark } = useResolvedColorScheme();
  const tokens = getNodeThemeTokens(theme, isDark);
  const primary = tokens.palette.primary.main;
  const activeCountrySet = new Set((activeCountries || []).map(normalizeCountry).filter(Boolean));

  return (
    <Box
      sx={{
        mb: 2,
        p: 2,
        borderRadius: 2.5,
        bgcolor: tokens.toolbarSurface,
        border: '1px solid',
        borderColor: tokens.softBorder
      }}
    >
      <Stack direction="row" spacing={1} sx={{ alignItems: 'center', flexWrap: 'wrap', rowGap: 0.75 }}>
        <Stack direction="row" spacing={0.75} sx={{ alignItems: 'center' }}>
          <PublicIcon sx={{ fontSize: 18, color: tokens.secondaryText }} />
          <Typography variant="subtitle2" sx={{ fontWeight: 600, color: tokens.primaryText }}>
            {t('nodes.countryStats.title')}
          </Typography>
        </Stack>
        <Chip
          size="small"
          label={t('nodes.countryStats.total', { count: total })}
          sx={{
            fontWeight: 600,
            bgcolor: withAlpha(tokens.palette.primary.main, isDark ? 0.2 : 0.08),
            color: isDark ? tokens.palette.primary.light : tokens.palette.primary.main,
            border: '1px solid',
            borderColor: withAlpha(tokens.palette.primary.main, isDark ? 0.34 : 0.18)
          }}
        />
      </Stack>

      {items.length === 0 ? (
        <Typography variant="body2" sx={{ mt: 1.25, color: tokens.secondaryText }}>
          {t('nodes.countryStats.empty')}
        </Typography>
      ) : (
        <Box
          sx={{
            mt: 1.25,
            display: 'flex',
            flexWrap: 'wrap',
            gap: 0.75,
            maxHeight: 132,
            overflowY: 'auto',
            pr: 0.5,
            scrollbarWidth: 'thin',
            '&::-webkit-scrollbar': { width: 6 },
            '&::-webkit-scrollbar-thumb': { borderRadius: 999, bgcolor: withAlpha(tokens.palette.primary.main, 0.32) }
          }}
        >
          {items.map((item) => {
            const country = normalizeCountry(item.country);
            const filterValue = country || UNKNOWN_COUNTRY_FILTER;
            const isActive = activeCountrySet.has(filterValue);

            return (
              <Chip
                key={filterValue}
                size="small"
                aria-pressed={isActive}
                onClick={onToggleCountry ? () => onToggleCountry(filterValue) : undefined}
                label={
                  <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.5 }}>
                    {country ? <CountryFlag code={country} height={11} /> : <PublicIcon sx={{ fontSize: 13 }} />}
                    <span>{country || t('common.unknown')}</span>
                    <Box component="span" sx={{ fontWeight: 700, opacity: 0.85 }}>
                      · {item.count}
                    </Box>
                  </Box>
                }
                sx={{
                  fontWeight: 600,
                  ...(isActive
                    ? {
                        ...getNodeColorChipSx(theme, tokens, primary, { emphasis: 'solid' }),
                        '&:hover': {
                          backgroundColor: isDark ? tokens.palette.primary.light : tokens.palette.primary.dark,
                          borderColor: primary
                        }
                      }
                    : {
                        bgcolor: tokens.fieldSurface,
                        color: tokens.primaryText,
                        border: '1px solid',
                        borderColor: tokens.softBorder,
                        '&:hover': {
                          bgcolor: tokens.hoverSurface,
                          borderColor: withAlpha(primary, isDark ? 0.34 : 0.22)
                        }
                      }),
                  transition: 'background-color 0.15s ease, border-color 0.15s ease, box-shadow 0.15s ease',
                  '&:focus-visible': { outline: `2px solid ${primary}`, outlineOffset: 2 },
                  '& .MuiChip-label': { display: 'flex', alignItems: 'center', px: 1 }
                }}
              />
            );
          })}
        </Box>
      )}
    </Box>
  );
}

NodeCountryStats.propTypes = {
  items: PropTypes.arrayOf(
    PropTypes.shape({
      country: PropTypes.string,
      count: PropTypes.number
    })
  ),
  total: PropTypes.number,
  activeCountries: PropTypes.arrayOf(PropTypes.string),
  onToggleCountry: PropTypes.func
};
