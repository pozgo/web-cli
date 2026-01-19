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
  ListItemIcon,
  ListItemText,
  Chip,
} from '@mui/material';
import {
  ArrowBack,
  Fullscreen,
  FullscreenExit,
  Refresh,
  VpnKey,
  Storage,
  Lock,
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
        // /api/keys already returns merged local + vault keys with source field
        const response = await fetch('/api/keys');
        if (response.ok) {
          const data = await response.json();
          const keysWithCompositeId = (data || []).map((key) => {
            // Determine source: vault keys have source="vault", local have source="sqlite" or no source
            const isVault = key.source === 'vault';
            // Create composite ID: vault keys use name (no numeric id), local keys use id
            const compositeId = isVault ? `vault:${key.name}` : `local:${key.id}`;
            return {
              ...key,
              source: isVault ? 'vault' : 'local',
              compositeId,
            };
          });
          setSshKeys(keysWithCompositeId);
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
      const compositeId = e.target.value;
      updateTabConfig(activeTab.id, { sshKeyId: compositeId });
      // Trigger reconnect
      window.dispatchEvent(
        new CustomEvent('terminal-reconnect', { detail: { tabId: activeTab.id } })
      );
    }
  };

  // Helper to get the selected key object from composite ID
  const getSelectedKey = (compositeId) => {
    if (!compositeId) return null;
    return sshKeys.find((k) => k.compositeId === compositeId);
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
        <FormControl size="small" sx={{ minWidth: 220 }}>
          <InputLabel>SSH Key</InputLabel>
          <Select
            value={activeTab?.sshKeyId || ''}
            label="SSH Key"
            onChange={handleSshKeyChange}
            renderValue={(value) => {
              if (!value) return <em>None</em>;
              const key = getSelectedKey(value);
              if (!key) return value;
              return (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <VpnKey sx={{ fontSize: 18, color: 'success.main' }} />
                  <span>{key.name}</span>
                  {key.source === 'vault' ? (
                    <Lock sx={{ fontSize: 14, color: 'secondary.main' }} />
                  ) : (
                    <Storage sx={{ fontSize: 14, color: 'text.secondary' }} />
                  )}
                </Box>
              );
            }}
          >
            <MenuItem value="">
              <em>None</em>
            </MenuItem>
            {sshKeys.map((key) => (
              <MenuItem key={key.compositeId} value={key.compositeId}>
                <ListItemIcon sx={{ minWidth: 36 }}>
                  <VpnKey fontSize="small" />
                </ListItemIcon>
                <ListItemText primary={key.name} />
                <Chip
                  icon={key.source === 'vault' ? <Lock sx={{ fontSize: '14px !important' }} /> : <Storage sx={{ fontSize: '14px !important' }} />}
                  label={key.source === 'vault' ? 'Vault' : 'Local'}
                  size="small"
                  variant="outlined"
                  color={key.source === 'vault' ? 'secondary' : 'default'}
                  sx={{ ml: 1, height: 24 }}
                />
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
        <Typography variant="body2" color="text.secondary" component="div">
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <span>Tabs: {tabs.length}/{maxTabs}</span>
            <span>|</span>
            <span>Shell: {activeTab?.shell || 'bash'}</span>
            {activeTab?.sshKeyId && (() => {
              const selectedKey = getSelectedKey(activeTab.sshKeyId);
              return selectedKey ? (
                <>
                  <span>|</span>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <VpnKey sx={{ fontSize: 16, color: 'success.main' }} />
                    <span>{selectedKey.name}</span>
                    {selectedKey.source === 'vault' ? (
                      <Chip
                        icon={<Lock sx={{ fontSize: '12px !important' }} />}
                        label="Vault"
                        size="small"
                        variant="outlined"
                        color="secondary"
                        sx={{ height: 20, '& .MuiChip-label': { px: 0.5, fontSize: '0.7rem' } }}
                      />
                    ) : (
                      <Chip
                        icon={<Storage sx={{ fontSize: '12px !important' }} />}
                        label="Local"
                        size="small"
                        variant="outlined"
                        sx={{ height: 20, '& .MuiChip-label': { px: 0.5, fontSize: '0.7rem' } }}
                      />
                    )}
                  </Box>
                </>
              ) : null;
            })()}
          </Box>
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
