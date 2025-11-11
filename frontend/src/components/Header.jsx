import React from 'react';
import { AppBar, Toolbar, Typography, Box, IconButton, Tooltip } from '@mui/material';
import { Terminal } from '@mui/icons-material';
import GitHubIcon from '@mui/icons-material/GitHub';
import { useNavigate } from 'react-router-dom';
import ThemeToggle from './ThemeToggle';

/**
 * Header component - displays the application header with logo and theme toggle
 * @param {Object} props - Component props
 * @param {string} props.mode - Current theme mode ('light' or 'dark')
 * @param {Function} props.toggleTheme - Function to toggle the theme
 */
const Header = ({ mode, toggleTheme }) => {
  const navigate = useNavigate();

  return (
    <AppBar position="static" elevation={0}>
      <Toolbar>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            cursor: 'pointer',
            flexGrow: 1,
            '&:hover': {
              opacity: 0.8,
            },
          }}
          onClick={() => navigate('/')}
        >
          <Terminal sx={{ mr: 2 }} />
          <Typography variant="h6" component="div">
            Web CLI
          </Typography>
        </Box>
        <Tooltip title="View on GitHub">
          <IconButton
            color="inherit"
            component="a"
            href="https://github.com/pozgo/web-cli"
            target="_blank"
            rel="noopener noreferrer"
            sx={{ mr: 1 }}
          >
            <GitHubIcon />
          </IconButton>
        </Tooltip>
        <ThemeToggle mode={mode} toggleTheme={toggleTheme} />
      </Toolbar>
    </AppBar>
  );
};

export default Header;
