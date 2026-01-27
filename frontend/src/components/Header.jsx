import React from 'react';
import { AppBar, Toolbar, Typography, Box, IconButton, Tooltip } from '@mui/material';
import { Api } from '@mui/icons-material';
import GitHubIcon from '@mui/icons-material/GitHub';
import { useNavigate } from 'react-router-dom';
import ThemeToggle from './ThemeToggle';
import VaultIcon from './shared/VaultIcon';
import logo from "../../assets/favicon.ico"

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
            '&:hover': {
              opacity: 0.8,
            },
          }}
          onClick={() => navigate('/')}
        >
          <img src={logo} alt="logo" width="24px" style={{marginRight: '5px'}}/>
          <Typography variant="h6" component="div">
            Web CLI
          </Typography>
        </Box>
        <Box sx={{ flexGrow: 1 }} />
        <VaultIcon sx={{ mr: 1 }} />
        <Tooltip title="API Documentation">
          <IconButton
            color="inherit"
            component="a"
            href="/swagger/"
            target="_blank"
            rel="noopener noreferrer"
            sx={{ mr: 1 }}
          >
            <Api />
          </IconButton>
        </Tooltip>
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
