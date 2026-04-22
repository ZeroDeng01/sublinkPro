import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

export default function AirportDialogSection({ title, children, surface, borderColor, titleColor }) {
  return (
    <Box
      sx={{
        p: 2,
        borderRadius: 2.5,
        bgcolor: surface,
        border: '1px solid',
        borderColor
      }}
    >
      <Typography
        variant="subtitle2"
        sx={{
          color: titleColor,
          fontWeight: 600,
          mb: 1.75,
          display: 'flex',
          alignItems: 'center',
          '&::before': {
            content: '""',
            width: 3,
            height: 16,
            bgcolor: 'primary.main',
            borderRadius: 999,
            mr: 1
          }
        }}
      >
        {title}
      </Typography>
      {children}
    </Box>
  );
}

AirportDialogSection.propTypes = {
  title: PropTypes.node.isRequired,
  children: PropTypes.node.isRequired,
  surface: PropTypes.string.isRequired,
  borderColor: PropTypes.string.isRequired,
  titleColor: PropTypes.string.isRequired
};
