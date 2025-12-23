import React, { useEffect, useRef, useState } from 'react';
import {
  Container,
  Typography,
  Box,
  Paper,
  Alert,
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
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

/**
 * Terminal component - Interactive web-based terminal
 * Uses xterm.js for terminal emulation and WebSocket for communication
 */
const Terminal = () => {
  const navigate = useNavigate();
  const terminalRef = useRef(null);
  const xtermRef = useRef(null);
  const fitAddonRef = useRef(null);
  const wsRef = useRef(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [shell, setShell] = useState('bash');
  const [sshKeys, setSshKeys] = useState([]);
  const [selectedSshKey, setSelectedSshKey] = useState('');
  const containerRef = useRef(null);

  // Fetch SSH keys on mount
  useEffect(() => {
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
    fetchSshKeys();
  }, []);

  // Initialize terminal
  useEffect(() => {
    if (!terminalRef.current) return;

    // Create xterm instance
    const xterm = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: '"MesloLGM Nerd Font", "MesloLGM NF", "Fira Code", "Cascadia Code", Menlo, Monaco, monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#ffffff',
        cursorAccent: '#1e1e1e',
        selection: 'rgba(255, 255, 255, 0.3)',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#ffffff',
      },
      allowProposedApi: true,
    });

    // Create fit addon
    const fitAddon = new FitAddon();
    xterm.loadAddon(fitAddon);

    // Open terminal in container
    xterm.open(terminalRef.current);
    
    // Fit to container
    setTimeout(() => {
      fitAddon.fit();
    }, 0);

    // Store refs
    xtermRef.current = xterm;
    fitAddonRef.current = fitAddon;

    // Connect to WebSocket
    connectWebSocket(xterm, fitAddon);

    // Handle window resize
    const handleResize = () => {
      if (fitAddonRef.current && xtermRef.current) {
        fitAddonRef.current.fit();
        sendResize();
      }
    };
    window.addEventListener('resize', handleResize);

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize);
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (xtermRef.current) {
        xtermRef.current.dispose();
      }
    };
  }, []);

  // Handle shell or SSH key change - reconnect
  useEffect(() => {
    if (xtermRef.current && wsRef.current && connected) {
      // Close existing connection
      wsRef.current.close();
      // Clear terminal
      xtermRef.current.clear();
      // Reconnect with new settings
      connectWebSocket(xtermRef.current, fitAddonRef.current);
    }
  }, [shell, selectedSshKey]);

  const connectWebSocket = (xterm, fitAddon) => {
    // Build WebSocket URL with shell and optional SSH key
    // Use encodeURIComponent to prevent URL injection
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsUrl = `${protocol}//${window.location.host}/api/terminal/ws?shell=${encodeURIComponent(shell)}`;
    if (selectedSshKey) {
      wsUrl += `&sshKeyId=${encodeURIComponent(selectedSshKey)}`;
    }

    setError(null);
    xterm.write('\r\n\x1b[33mConnecting to terminal...\x1b[0m\r\n');

    const ws = new WebSocket(wsUrl);
    ws.binaryType = 'arraybuffer';

    ws.onopen = () => {
      setConnected(true);
      xterm.write('\x1b[32mConnected!\x1b[0m\r\n\r\n');
      
      // Send initial size
      setTimeout(() => {
        if (fitAddon) {
          fitAddon.fit();
          sendResize();
        }
      }, 100);
    };

    ws.onmessage = (event) => {
      // Handle binary data from PTY
      if (event.data instanceof ArrayBuffer) {
        const text = new TextDecoder().decode(event.data);
        xterm.write(text);
      } else {
        xterm.write(event.data);
      }
    };

    ws.onerror = (event) => {
      console.error('WebSocket error:', event);
      setError('Connection error. Please check if the server is running.');
      setConnected(false);
    };

    ws.onclose = (event) => {
      setConnected(false);
      xterm.write('\r\n\x1b[31mDisconnected from terminal.\x1b[0m\r\n');
      if (!event.wasClean) {
        setError('Connection lost. Click refresh to reconnect.');
      }
    };

    wsRef.current = ws;

    // Handle terminal input
    xterm.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(data);
      }
    });
  };

  const sendResize = () => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN && xtermRef.current) {
      const dimensions = {
        type: 'resize',
        cols: xtermRef.current.cols,
        rows: xtermRef.current.rows,
      };
      wsRef.current.send(JSON.stringify(dimensions));
    }
  };

  const handleReconnect = () => {
    if (xtermRef.current && fitAddonRef.current) {
      xtermRef.current.clear();
      connectWebSocket(xtermRef.current, fitAddonRef.current);
    }
  };

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
      // Refit terminal after fullscreen change
      setTimeout(() => {
        if (fitAddonRef.current) {
          fitAddonRef.current.fit();
          sendResize();
        }
      }, 100);
    };
    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange);
  }, []);

  return (
    <Container 
      maxWidth="xl" 
      sx={{ mt: 4, mb: 4 }}
      ref={containerRef}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 2 }}>
        <IconButton onClick={() => navigate('/')} color="primary">
          <ArrowBack />
        </IconButton>
        <Typography variant="h4" component="h1" sx={{ flexGrow: 1 }}>
          Interactive Terminal
        </Typography>
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Shell</InputLabel>
          <Select
            value={shell}
            label="Shell"
            onChange={(e) => setShell(e.target.value)}
          >
            <MenuItem value="bash">Bash</MenuItem>
            <MenuItem value="sh">Shell</MenuItem>
            <MenuItem value="zsh">Zsh</MenuItem>
          </Select>
        </FormControl>
        <FormControl size="small" sx={{ minWidth: 180 }}>
          <InputLabel>SSH Key</InputLabel>
          <Select
            value={selectedSshKey}
            label="SSH Key"
            onChange={(e) => setSelectedSshKey(e.target.value)}
            startAdornment={selectedSshKey ? <VpnKey sx={{ mr: 1, fontSize: 18, color: 'success.main' }} /> : null}
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
        <Tooltip title={isFullscreen ? "Exit Fullscreen" : "Fullscreen"}>
          <IconButton onClick={toggleFullscreen} color="primary">
            {isFullscreen ? <FullscreenExit /> : <Fullscreen />}
          </IconButton>
        </Tooltip>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <Paper
        elevation={3}
        sx={{
          p: 1,
          bgcolor: '#1e1e1e',
          height: isFullscreen ? 'calc(100vh - 100px)' : 'calc(100vh - 250px)',
          minHeight: '400px',
          overflow: 'hidden',
          borderRadius: 2,
        }}
      >
        <Box
          ref={terminalRef}
          sx={{
            width: '100%',
            height: '100%',
            '& .xterm': {
              height: '100%',
              padding: '8px',
            },
            '& .xterm-viewport': {
              overflow: 'hidden !important',
            },
          }}
        />
      </Paper>

      <Box sx={{ mt: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          Status: {connected ? (
            <span style={{ color: '#4caf50' }}>● Connected</span>
          ) : (
            <span style={{ color: '#f44336' }}>● Disconnected</span>
          )}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Shell: {shell}
          {selectedSshKey && ` | SSH Key: ${sshKeys.find(k => k.id === selectedSshKey)?.name || 'Selected'}`}
          {selectedSshKey && ' (use: ssh user@host)'}
          {!selectedSshKey && ' | No SSH key selected'}
        </Typography>
      </Box>
    </Container>
  );
};

export default Terminal;
