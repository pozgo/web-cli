import React, { useState, useMemo, useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { getTheme } from './theme/theme';
import Header from './components/Header';
import Dashboard from './components/Dashboard';
import AdminPanel from './components/AdminPanel';
import LocalCommands from './components/LocalCommands';
import RemoteCommands from './components/RemoteCommands';
import SavedCommands from './components/SavedCommands';
import CommandHistory from './components/CommandHistory';
import ScriptsPage from './components/ScriptsPage';
import LocalScripts from './components/LocalScripts';
import RemoteScripts from './components/RemoteScripts';
import Terminal from './components/Terminal';

/**
 * Main App component - handles theme management and renders the application with routing
 */
function App() {
  // Initialize theme mode from localStorage or default to 'light'
  const [mode, setMode] = useState(() => {
    const savedMode = localStorage.getItem('themeMode');
    return savedMode || 'light';
  });

  // Save theme preference to localStorage whenever it changes
  useEffect(() => {
    localStorage.setItem('themeMode', mode);
  }, [mode]);

  // Toggle between light and dark theme
  const toggleTheme = () => {
    setMode((prevMode) => (prevMode === 'light' ? 'dark' : 'light'));
  };

  // Memoize theme to avoid unnecessary recalculations
  const theme = useMemo(() => getTheme(mode), [mode]);

  return (
    <BrowserRouter>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Header mode={mode} toggleTheme={toggleTheme} />
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/admin" element={<AdminPanel />} />
          <Route path="/local-commands" element={<LocalCommands />} />
          <Route path="/remote-commands" element={<RemoteCommands />} />
          <Route path="/saved-commands" element={<SavedCommands />} />
          <Route path="/history" element={<CommandHistory />} />
          <Route path="/scripts" element={<ScriptsPage />} />
          <Route path="/local-scripts" element={<LocalScripts />} />
          <Route path="/remote-scripts" element={<RemoteScripts />} />
          <Route path="/terminal" element={<Terminal />} />
        </Routes>
      </ThemeProvider>
    </BrowserRouter>
  );
}

export default App;
