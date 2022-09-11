import {createTheme} from '@mui/material/styles';
import {grey} from '@mui/material/colors';
import {lighten} from '@mui/material';

export const theme = createTheme({
  palette: {
    grey: {
      50: lighten(grey[50], 0.3),
    },
  },
  components: {
    // Name of the component ⚛️
    MuiButtonBase: {
      defaultProps: {
        // The props to apply
        disableRipple: true, // No more ripple, on the whole application 💣!
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          backgroundColor: 'white',
        },
      },
    },
  },
});
