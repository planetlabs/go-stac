import CssBaseline from '@mui/material/CssBaseline';
import React from 'react';
import Viewer from './components/Viewer.jsx';
import {BrowserRouter, Route, Routes} from 'react-router-dom';
import {ThemeProvider} from '@mui/material/styles';
import {createRoot} from 'react-dom/client';
import {theme} from './theme.js';

function Main() {
  return (
    <ThemeProvider theme={theme}>
      <BrowserRouter basename={import.meta.env.BASE_URL}>
        <CssBaseline />
        <Routes>
          <Route path="/*" element={<Viewer />} />
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

const root = createRoot(document.getElementById('root'));
root.render(<Main />);
