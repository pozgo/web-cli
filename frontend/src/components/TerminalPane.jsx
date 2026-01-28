import React, { useEffect, useRef, useCallback, useState } from 'react';
import { Box } from '@mui/material';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

/**
 * TerminalPane component - Individual terminal instance with xterm.js
 * Each pane maintains its own WebSocket connection and PTY session
 */
const TerminalPane = ({
  tabId,
  shell,
  sshKeyId,
  isActive,
  onConnected,
  onDisconnected,
}) => {
  const terminalRef = useRef(null);
  const xtermRef = useRef(null);
  const fitAddonRef = useRef(null);
  const wsRef = useRef(null);
  const dataHandlerRef = useRef(null);
  const [isInitialized, setIsInitialized] = useState(false);

  // Send resize message to server
  const sendResize = useCallback(() => {
    if (
      wsRef.current &&
      wsRef.current.readyState === WebSocket.OPEN &&
      xtermRef.current
    ) {
      const dimensions = {
        type: 'resize',
        cols: xtermRef.current.cols,
        rows: xtermRef.current.rows,
      };
      wsRef.current.send(JSON.stringify(dimensions));
    }
  }, []);

  // Connect to WebSocket - stable function that reads current values from refs
  const connectWebSocket = useCallback((currentShell, currentSshKeyId) => {
    if (!xtermRef.current) return;

    const xterm = xtermRef.current;
    const fitAddon = fitAddonRef.current;

    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (dataHandlerRef.current) {
      dataHandlerRef.current.dispose();
      dataHandlerRef.current = null;
    }

    // Build WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsUrl = `${protocol}//${window.location.host}/api/terminal/ws?shell=${encodeURIComponent(currentShell)}`;

    // Handle composite SSH key ID (format: "source:id" e.g., "local:123" or "vault:keyname")
    if (currentSshKeyId) {
      // Ensure sshKeyId is a string (handle legacy numeric IDs from older versions)
      const sshKeyIdStr = String(currentSshKeyId);
      const colonIndex = sshKeyIdStr.indexOf(':');
      if (colonIndex > 0) {
        const source = sshKeyIdStr.substring(0, colonIndex);
        const keyIdentifier = sshKeyIdStr.substring(colonIndex + 1);
        wsUrl += `&sshKeyId=${encodeURIComponent(keyIdentifier)}`;
        wsUrl += `&sshKeySource=${encodeURIComponent(source)}`;
      } else {
        // Fallback for legacy format (just the ID)
        wsUrl += `&sshKeyId=${encodeURIComponent(sshKeyIdStr)}`;
      }
    }

    xterm.write('\r\n\x1b[33mConnecting to terminal...\x1b[0m\r\n');

    const ws = new WebSocket(wsUrl);
    ws.binaryType = 'arraybuffer';

    ws.onopen = () => {
      onConnected(tabId);
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
      if (event.data instanceof ArrayBuffer) {
        const text = new TextDecoder().decode(event.data);
        xterm.write(text);
      } else {
        xterm.write(event.data);
      }
    };

    ws.onerror = (event) => {
      console.error('WebSocket error:', event);
      onDisconnected(tabId);
    };

    ws.onclose = () => {
      onDisconnected(tabId);
      xterm.write('\r\n\x1b[31mDisconnected from terminal.\x1b[0m\r\n');
    };

    wsRef.current = ws;

    // Handle terminal input
    dataHandlerRef.current = xterm.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(data);
      }
    });
  }, [tabId, onConnected, onDisconnected, sendResize]);

  // Initialize terminal - only once when tab first becomes active
  useEffect(() => {
    // Don't initialize until the tab is active (visible)
    if (!isActive) return;
    // Only initialize once
    if (isInitialized || !terminalRef.current) return;

    // Create xterm instance
    const xterm = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily:
        '"MesloLGM Nerd Font", "MesloLGM NF", "Fira Code", "Cascadia Code", Menlo, Monaco, monospace',
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

    // Mark as initialized
    setIsInitialized(true);

    // Connect to WebSocket with current shell/sshKeyId
    connectWebSocket(shell, sshKeyId);
  }, [isActive, isInitialized, shell, sshKeyId, connectWebSocket]);

  // Cleanup on unmount only
  useEffect(() => {
    return () => {
      if (dataHandlerRef.current) {
        dataHandlerRef.current.dispose();
        dataHandlerRef.current = null;
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (xtermRef.current) {
        xtermRef.current.dispose();
        xtermRef.current = null;
      }
    };
  }, []);

  // Handle resize when tab becomes active or window resizes
  useEffect(() => {
    if (!isActive || !isInitialized) return;

    const handleResize = () => {
      if (fitAddonRef.current && xtermRef.current) {
        fitAddonRef.current.fit();
        sendResize();
      }
    };

    // Fit on activation
    setTimeout(handleResize, 50);

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [isActive, isInitialized, sendResize]);

  // Focus terminal when active
  useEffect(() => {
    if (isActive && xtermRef.current) {
      xtermRef.current.focus();
    }
  }, [isActive]);

  // Handle reconnect via custom event
  useEffect(() => {
    const handleReconnect = (e) => {
      if (e.detail.tabId === tabId && xtermRef.current) {
        xtermRef.current.clear();
        connectWebSocket(shell, sshKeyId);
      }
    };
    window.addEventListener('terminal-reconnect', handleReconnect);
    return () => window.removeEventListener('terminal-reconnect', handleReconnect);
  }, [tabId, shell, sshKeyId, connectWebSocket]);

  return (
    <Box
      sx={{
        display: isActive ? 'block' : 'none',
        width: '100%',
        height: '100%',
        bgcolor: '#1e1e1e',
        '& .xterm': {
          height: '100%',
          padding: '8px',
        },
        '& .xterm-viewport': {
          overflow: 'hidden !important',
        },
      }}
    >
      <Box
        ref={terminalRef}
        sx={{
          width: '100%',
          height: '100%',
        }}
      />
    </Box>
  );
};

export default TerminalPane;
