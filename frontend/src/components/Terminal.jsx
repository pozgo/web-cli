import React, { useEffect, useRef, useState, useCallback } from 'react';
import {
  Container,
  Typography,
  Box,
  Paper,
  IconButton,
  Tooltip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import {
  ArrowBack,
  Fullscreen,
  FullscreenExit,
  Refresh,
  VpnKey,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { TerminalProvider, useTerminal } from '../context/TerminalContext';
import TabBar from './TabBar';
import TerminalPane from './TerminalPane';

/**
 * TerminalContent component - Main terminal UI with multi-tab support
 * Separated from Terminal to allow useTerminal hook usage
 */
const TerminalContent = () => {
  const navigate = useNavigate();
  const containerRef = useRef(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [availableShells, setAvailableShells] = useState([]);
  const [sshKeys, setSshKeys] = useState([]);

  const {
    tabs,
    activeTabId,
    activeTab,
    maxTabs,
    setTabConnected,
    updateTabConfig,
  } = useTerminal();

  // Fetch available shells and SSH keys on mount
  useEffect(() => {
    const fetchShells = async () => {
      try {
        const response = await fetch('/api/system/shells');
        if (response.ok) {
          const data = await response.json();
          setAvailableShells(data || []);
        }
      } catch (err) {
        console.error('Failed to fetch available shells:', err);
        setAvailableShells([{ name: 'bash', path: '/bin/bash' }]);
      }
    };

    const fetchSshKeys = async () => {
      try {
        const response = await fetch('/api/keys');
        if (response.ok) {
          const data = await response.json();
          setSshKeys(data || []);
        }
      } catch (err) {
        console.error('Failed to fetch SSH keys:', err);
      }
    };

    fetchShells();
    fetchSshKeys();
  }, []);

  // Handle connection status updates
  const handleConnected = useCallback(
    (tabId) => {
      setTabConnected(tabId, true);
    },
    [setTabConnected]
  );

  const handleDisconnected = useCallback(
    (tabId) => {
      setTabConnected(tabId, false);
    },
    [setTabConnected]
  );

  // Handle shell change for active tab
  const handleShellChange = (e) => {
    if (activeTab) {
      updateTabConfig(activeTab.id, { shell: e.target.value });
      // Trigger reconnect
      window.dispatchEvent(
        new CustomEvent('terminal-reconnect', { detail: { tabId: activeTab.id } })
      );
    }
  };

  // Handle SSH key change for active tab
  const handleSshKeyChange = (e) => {
    if (activeTab) {
      updateTabConfig(activeTab.id, { sshKeyId: e.target.value });
      // Trigger reconnect
      window.dispatchEvent(
        new CustomEvent('terminal-reconnect', { detail: { tabId: activeTab.id } })
      );
    }
  };

  // Handle reconnect for active tab
  const handleReconnect = () => {
    if (activeTab) {
      window.dispatchEvent(
        new CustomEvent('terminal-reconnect', { detail: { tabId: activeTab.id } })
      );
    }
  };

  // Toggle fullscreen
  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen();
      setIsFullscreen(true);
    } else {
      document.exitFullscreen();
      setIsFullscreen(false);
    }
  };

  // Listen for fullscreen change
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () =>
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
  }, []);

  return (
    <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }} ref={containerRef}>
      {/* Header with controls */}
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 2 }}>
        <IconButton onClick={() => navigate('/')} color="primary">
          <ArrowBack />
        </IconButton>
        <Typography variant="h4" component="h1" sx={{ flexGrow: 1 }}>
          Interactive Terminal
        </Typography>

        {/* Shell selector for active tab */}
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Shell</InputLabel>
          <Select
            value={activeTab?.shell || 'bash'}
            label="Shell"
            onChange={handleShellChange}
            disabled={availableShells.length === 0}
          >
            {availableShells.map((s) => (
              <MenuItem key={s.name} value={s.name}>
                {s.name.charAt(0).toUpperCase() + s.name.slice(1)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {/* SSH Key selector for active tab */}
        <FormControl size="small" sx={{ minWidth: 180 }}>
          <InputLabel>SSH Key</InputLabel>
          <Select
            value={activeTab?.sshKeyId || ''}
            label="SSH Key"
            onChange={handleSshKeyChange}
            startAdornment={
              activeTab?.sshKeyId ? (
                <VpnKey sx={{ mr: 1, fontSize: 18, color: 'success.main' }} />
              ) : null
            }
          >
            <MenuItem value="">
              <em>None</em>
            </MenuItem>
            {sshKeys.map((key) => (
              <MenuItem key={key.id} value={key.id}>
                {key.name}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <Tooltip title="Reconnect">
          <IconButton onClick={handleReconnect} color="primary">
            <Refresh />
          </IconButton>
        </Tooltip>
        <Tooltip title={isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'}>
          <IconButton onClick={toggleFullscreen} color="primary">
            {isFullscreen ? <FullscreenExit /> : <Fullscreen />}
          </IconButton>
        </Tooltip>
      </Box>

      {/* Tab bar */}
      <TabBar />

      {/* Terminal panes container */}
      <Paper
        elevation={3}
        sx={{
          bgcolor: '#1e1e1e',
          height: isFullscreen ? 'calc(100vh - 180px)' : 'calc(100vh - 320px)',
          minHeight: '400px',
          overflow: 'hidden',
          borderRadius: '0 0 8px 8px',
          borderTop: 0,
        }}
      >
        {tabs.map((tab) => (
          <TerminalPane
            key={tab.id}
            tabId={tab.id}
            shell={tab.shell}
            sshKeyId={tab.sshKeyId}
            isActive={tab.id === activeTabId}
            onConnected={handleConnected}
            onDisconnected={handleDisconnected}
          />
        ))}
      </Paper>

      {/* Status bar */}
      <Box
        sx={{
          mt: 2,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <Typography variant="body2" color="text.secondary">
          Status:{' '}
          {activeTab?.connected ? (
            <span style={{ color: '#4caf50' }}>Connected</span>
          ) : (
            <span style={{ color: '#f44336' }}>Disconnected</span>
          )}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Tabs: {tabs.length}/{maxTabs} | Shell: {activeTab?.shell || 'bash'}
          {activeTab?.sshKeyId &&
            ` | SSH Key: ${sshKeys.find((k) => k.id === activeTab.sshKeyId)?.name || 'Selected'}`}
        </Typography>
      </Box>
    </Container>
  );
};

/**
 * Terminal component - Wrapper that provides TerminalContext
 */
const Terminal = () => {
  return (
    <TerminalProvider>
      <TerminalContent />
    </TerminalProvider>
  );
};

export default Terminal;
