import { Link as RouterLink } from 'react-router-dom';
// project imports
import useConfig from 'hooks/useConfig';

// material-ui
import Link from '@mui/material/Link';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

// ==============================|| FOOTER - AUTHENTICATION ||============================== //

export default function AuthFooter() {
  const { version } = useConfig();

  return (
    <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
      <Typography variant="caption">
        &copy; All rights reserved{' '}
        <Typography
          component={RouterLink}
          to="https://github.com/ZeroDeng01/sublinkPro"
          target="_blank"
          sx={{ textDecoration: 'none', color: 'primary.main' }}
        >
          SublinkPro {version || 'dev'}
        </Typography>
      </Typography>
      <Stack direction="row" sx={{ gap: 1.5, alignItems: 'center', justifyContent: 'space-between' }}>
        <Link
          component={RouterLink}
          to="https://github.com/ZeroDeng01/sublinkPro"
          underline="hover"
          target="_blank"
          variant="caption"
          color="text.primary"
        >
          GitHub
        </Link>
        <Link
          component={RouterLink}
          to="https://github.com/ZeroDeng01/sublinkPro/blob/master/LICENSE"
          underline="hover"
          target="_blank"
          variant="caption"
          color="text.primary"
        >
          License
        </Link>
      </Stack>
    </Stack>
  );
}
