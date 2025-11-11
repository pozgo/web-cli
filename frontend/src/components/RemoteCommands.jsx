import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Button,
  TextField,
  Paper,
  Alert,
  CircularProgress,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Checkbox,
  FormControlLabel,
  Grid,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import { PlayArrow, ArrowBack, Save } from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';

/**
 * RemoteCommands component - execute commands on remote servers via SSH
 * Provides server selection, SSH key selection, command execution with output display
 */
const RemoteCommands = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [command, setCommand] = useState('');
  const [user, setUser] = useState('root');
  const [selectedServer, setSelectedServer] = useState('');
  const [selectedSSHKey, setSelectedSSHKey] = useState('');
  const [saveAs, setSaveAs] = useState('');
  const [shouldSave, setShouldSave] = useState(false);
  const [output, setOutput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [savedCommands, setSavedCommands] = useState([]);
  const [selectedSavedCommand, setSelectedSavedCommand] = useState('');
  const [servers, setServers] = useState([]);
  const [sshKeys, setSSHKeys] = useState([]);
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
  const [sshPassword, setSSHPassword] = useState('');
  const [availableUsers, setAvailableUsers] = useState([]);

  // Fetch data on mount
  useEffect(() => {
    fetchSavedCommands();
    fetchServers();
    fetchSSHKeys();
    fetchLocalUsers();
  }, []);

  // Handle navigation state (pre-filled command from history or saved commands)
  useEffect(() => {
    if (location.state) {
      if (location.state.command) {
        setCommand(location.state.command);
      }
      if (location.state.user) {
        setUser(location.state.user);
      }
      if (location.state.server_id) {
        setSelectedServer(location.state.server_id);
      }
      if (location.state.ssh_key_id) {
        setSelectedSSHKey(location.state.ssh_key_id);
      }
    }
  }, [location]);

  const fetchSavedCommands = async () => {
    try {
      const response = await fetch('/api/saved-commands');
      if (response.ok) {
        const data = await response.json();
        // Filter to show only remote commands
        const remoteCommands = data.filter(cmd => cmd.is_remote);
        setSavedCommands(remoteCommands || []);
      }
    } catch (err) {
      console.error('Failed to fetch saved commands:', err);
    }
  };

  const fetchServers = async () => {
    try {
      const response = await fetch('/api/servers');
      if (response.ok) {
        const data = await response.json();
        setServers(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch servers:', err);
      setError('Failed to load servers. Please check Admin panel.');
    }
  };

  const fetchSSHKeys = async () => {
    try {
      const response = await fetch('/api/keys');
      if (response.ok) {
        const data = await response.json();
        setSSHKeys(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch SSH keys:', err);
      setError('Failed to load SSH keys. Please check Admin panel.');
    }
  };

  const fetchLocalUsers = async () => {
    try {
      const response = await fetch('/api/local-users');
      if (response.ok) {
        const data = await response.json();
        const users = data || [];

        // Build unique list: always include 'root', then add stored users
        const userSet = new Set();
        userSet.add('root');

        users.forEach(u => {
          if (u.name) {
            userSet.add(u.name);
          }
        });

        setAvailableUsers(Array.from(userSet));
      }
    } catch (err) {
      console.error('Failed to fetch local users:', err);
      setAvailableUsers(['root']);
    }
  };

  // Load selected saved command
  const handleLoadSavedCommand = (event) => {
    const cmdId = event.target.value;
    setSelectedSavedCommand(cmdId);

    if (cmdId) {
      const cmd = savedCommands.find((c) => c.id === parseInt(cmdId, 10));
      if (cmd) {
        setCommand(cmd.command);
        setUser(cmd.user || 'root');
        if (cmd.server_id) {
          setSelectedServer(String(cmd.server_id));
        }
        if (cmd.ssh_key_id) {
          setSelectedSSHKey(String(cmd.ssh_key_id));
        }
      }
    }
  };

  // Execute command - check if server selected first
  const handleExecute = async () => {
    if (!command.trim()) {
      setError('Command is required');
      return;
    }

    if (!selectedServer) {
      setError('Please select a server');
      return;
    }

    // If no SSH key selected, ask for password
    if (!selectedSSHKey) {
      setPasswordDialogOpen(true);
      return;
    }

    // Execute with SSH key
    executeCommand('');
  };

  // Actually execute the command with the provided password (if needed)
  const executeCommand = async (password) => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    setOutput('');

    try {
      const payload = {
        command: command.trim(),
        user: user || 'root',
        is_remote: true,
        server_id: parseInt(selectedServer, 10),
      };

      // Add SSH key if selected
      if (selectedSSHKey) {
        payload.ssh_key_id = parseInt(selectedSSHKey, 10);
      }

      // Add SSH password if provided (fallback auth or no key)
      if (password) {
        payload.ssh_password = password;
      }

      // Add saveAs if user wants to save
      if (shouldSave && saveAs.trim()) {
        payload.save_as = saveAs.trim();
      }

      const response = await fetch('/api/commands/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to execute command');
      }

      const result = await response.json();
      setOutput(result.output || '(no output)');

      if (result.exit_code === 0) {
        setSuccess(`Command executed successfully in ${result.execution_time_ms}ms`);
      } else {
        setError(`Command exited with code ${result.exit_code}`);
      }

      // If saved, refresh saved commands list
      if (shouldSave && saveAs.trim()) {
        fetchSavedCommands();
        setSaveAs('');
        setShouldSave(false);
      }
    } catch (err) {
      setError(err.message);
      setOutput('');
    } finally {
      setLoading(false);
    }
  };

  // Handle password dialog submit
  const handlePasswordSubmit = () => {
    setPasswordDialogOpen(false);
    executeCommand(sshPassword);
    setSSHPassword(''); // Clear password after use for security
  };

  // Handle password dialog cancel
  const handlePasswordCancel = () => {
    setPasswordDialogOpen(false);
    setSSHPassword('');
  };

  // Handle Enter key in command input
  const handleKeyPress = (event) => {
    if (event.key === 'Enter' && event.ctrlKey) {
      handleExecute();
    }
  };

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 4 }}>
        <Button
          startIcon={<ArrowBack />}
          onClick={() => navigate('/')}
          sx={{ mb: 2 }}
        >
          Back to Dashboard
        </Button>

        <Typography variant="h4" component="h1" sx={{ mb: 1 }}>
          Remote Commands
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          Execute bash commands on remote servers via SSH. All commands are saved to history.
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}

        {success && (
          <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess(null)}>
            {success}
          </Alert>
        )}

        <Paper sx={{ p: 3, mb: 3 }}>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <FormControl fullWidth sx={{ mb: 2 }}>
                <InputLabel>Load Saved Command</InputLabel>
                <Select
                  value={selectedSavedCommand}
                  onChange={handleLoadSavedCommand}
                  label="Load Saved Command"
                >
                  <MenuItem value="">
                    <em>None</em>
                  </MenuItem>
                  {savedCommands.map((cmd) => (
                    <MenuItem key={cmd.id} value={cmd.id}>
                      {cmd.name} - {cmd.user}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12} md={6}>
              <FormControl fullWidth required>
                <InputLabel>Select Server</InputLabel>
                <Select
                  value={selectedServer}
                  onChange={(e) => setSelectedServer(e.target.value)}
                  label="Select Server"
                  disabled={loading}
                  error={!selectedServer && error}
                >
                  <MenuItem value="">
                    <em>Choose a server...</em>
                  </MenuItem>
                  {servers.map((server) => (
                    <MenuItem key={server.id} value={server.id}>
                      {server.name || server.ip_address} ({server.ip_address}:{server.port})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>SSH Key (Optional)</InputLabel>
                <Select
                  value={selectedSSHKey}
                  onChange={(e) => setSelectedSSHKey(e.target.value)}
                  label="SSH Key (Optional)"
                  disabled={loading}
                >
                  <MenuItem value="">
                    <em>Use password authentication</em>
                  </MenuItem>
                  {sshKeys.map((key) => (
                    <MenuItem key={key.id} value={key.id}>
                      {key.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12} md={8}>
              <TextField
                fullWidth
                multiline
                rows={4}
                label="Command"
                placeholder="Enter bash command (e.g., ls -la, whoami, df -h)"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                onKeyPress={handleKeyPress}
                disabled={loading}
                helperText="Press Ctrl+Enter to execute"
              />
            </Grid>

            <Grid item xs={12} md={4}>
              <FormControl fullWidth>
                <InputLabel>Run As User</InputLabel>
                <Select
                  value={user}
                  onChange={(e) => setUser(e.target.value)}
                  label="Run As User"
                  disabled={loading}
                >
                  {availableUsers.map((username) => (
                    <MenuItem key={username} value={username}>
                      {username}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={shouldSave}
                    onChange={(e) => setShouldSave(e.target.checked)}
                    disabled={loading}
                  />
                }
                label="Save command as template for reuse"
              />
            </Grid>

            {shouldSave && (
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Save As Name"
                  placeholder="e.g., check-disk-space, list-processes"
                  value={saveAs}
                  onChange={(e) => setSaveAs(e.target.value)}
                  disabled={loading}
                />
              </Grid>
            )}

            <Grid item xs={12}>
              <Button
                variant="contained"
                size="large"
                startIcon={loading ? <CircularProgress size={20} /> : <PlayArrow />}
                onClick={handleExecute}
                disabled={loading || !command.trim() || !selectedServer}
                fullWidth
              >
                {loading ? 'Executing...' : 'Execute Command'}
              </Button>
            </Grid>
          </Grid>
        </Paper>

        {output && (
          <Paper sx={{ p: 3, backgroundColor: '#1e1e1e', color: '#d4d4d4' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <Typography variant="h6" sx={{ flexGrow: 1 }}>
                Output
              </Typography>
              <Button
                size="small"
                onClick={() => navigator.clipboard.writeText(output)}
              >
                Copy
              </Button>
            </Box>
            <Box
              component="pre"
              sx={{
                fontFamily: 'monospace',
                fontSize: '0.875rem',
                margin: 0,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}
            >
              {output}
            </Box>
          </Paper>
        )}

        <Box sx={{ mt: 3, display: 'flex', gap: 2 }}>
          <Button
            variant="outlined"
            startIcon={<Save />}
            onClick={() => navigate('/saved-commands')}
          >
            View Saved Commands
          </Button>
          <Button
            variant="outlined"
            onClick={() => navigate('/history')}
          >
            View History
          </Button>
        </Box>
      </Box>

      {/* SSH Password Dialog */}
      <Dialog open={passwordDialogOpen} onClose={handlePasswordCancel}>
        <DialogTitle>SSH Password Required</DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ mb: 2 }}>
            No SSH key selected or key authentication failed. Please enter the SSH password:
          </Typography>
          <TextField
            autoFocus
            fullWidth
            type="password"
            label="SSH Password"
            value={sshPassword}
            onChange={(e) => setSSHPassword(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter') {
                handlePasswordSubmit();
              }
            }}
          />
          <Alert severity="info" sx={{ mt: 2 }}>
            <strong>Security Note:</strong> SSH passwords are never stored in command history.
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button onClick={handlePasswordCancel}>Cancel</Button>
          <Button onClick={handlePasswordSubmit} variant="contained">
            Execute
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default RemoteCommands;
