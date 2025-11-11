import React from 'react';
import { IconButton, Tooltip } from '@mui/material';
import { Brightness4, Brightness7 } from '@mui/icons-material';

/**
 * ThemeToggle component - allows users to switch between light and dark themes
 * @param {Object} props - Component props
 * @param {string} props.mode - Current theme mode ('light' or 'dark')
 * @param {Function} props.toggleTheme - Function to toggle the theme
 */
const ThemeToggle = ({ mode, toggleTheme }) => {
  return (
    <Tooltip title={`Switch to ${mode === 'dark' ? 'light' : 'dark'} mode`}>
      <IconButton onClick={toggleTheme} color="inherit" aria-label="toggle theme">
        {mode === 'dark' ? <Brightness7 /> : <Brightness4 />}
      </IconButton>
    </Tooltip>
  );
};

export default ThemeToggle;
