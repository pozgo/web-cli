import React, { useState, useEffect } from 'react';
import Ansi from 'ansi-to-react';
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
  Chip,
} from '@mui/material';
import { PlayArrow, ArrowBack, Save, Storage, Lock } from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import GroupSelector from './shared/GroupSelector';

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
  const [baseUsers, setBaseUsers] = useState(['root']);
  const [serverGroupFilter, setServerGroupFilter] = useState('all');
  const [keyGroupFilter, setKeyGroupFilter] = useState('all');

  // Fetch data on mount
  useEffect(() => {
    fetchSavedCommands();
    fetchServers();
    fetchSSHKeys();
    fetchLocalUsers();
  }, []);

  // Update available users when server selection changes
  useEffect(() => {
    if (selectedServer && !location.state?.user) {
      // Find server by ID for SQLite items or by name for Vault items
      const server = servers.find(s =>
        (s.source === 'vault' ? s.name === selectedServer : s.id === parseInt(selectedServer, 10))
      );
      if (server && server.username) {
        // Add server's username to available users if not already present
        const userSet = new Set(baseUsers);
        userSet.add(server.username);
        setAvailableUsers(Array.from(userSet));
        // Auto-select the server's configured username
        setUser(server.username);
      } else {
        setAvailableUsers(baseUsers);
      }
    } else if (!selectedServer) {
      setAvailableUsers(baseUsers);
    }
  }, [selectedServer, servers, baseUsers]);

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

        const userArray = Array.from(userSet);
        setBaseUsers(userArray);
        setAvailableUsers(userArray);
      }
    } catch (err) {
      console.error('Failed to fetch local users:', err);
      setBaseUsers(['root']);
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
      // Find the selected server and SSH key objects
      const selectedServerObj = servers.find(s =>
        (s.source === 'vault' ? s.name === selectedServer : s.id === parseInt(selectedServer, 10))
      );
      const selectedSSHKeyObj = sshKeys.find(k =>
        (k.source === 'vault' ? k.name === selectedSSHKey : k.id === parseInt(selectedSSHKey, 10))
      );

      const payload = {
        command: command.trim(),
        user: user || 'root',
        is_remote: true,
      };

      // For server: use name for Vault items, ID for SQLite items
      if (selectedServerObj) {
        if (selectedServerObj.source === 'vault') {
          payload.server_name = selectedServerObj.name;
          payload.server_group = selectedServerObj.group || 'default';
        } else {
          payload.server_id = selectedServerObj.id;
        }
      }

      // Add SSH key if selected - use name for Vault items, ID for SQLite items
      if (selectedSSHKeyObj) {
        if (selectedSSHKeyObj.source === 'vault') {
          payload.ssh_key_name = selectedSSHKeyObj.name;
          payload.ssh_key_group = selectedSSHKeyObj.group || 'default';
        } else {
          payload.ssh_key_id = selectedSSHKeyObj.id;
        }
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

  // Filter servers and SSH keys by group
  const filteredServers = serverGroupFilter === 'all'
    ? servers
    : servers.filter(s => (s.group || 'default') === serverGroupFilter);

  const filteredSSHKeys = keyGroupFilter === 'all'
    ? sshKeys
    : sshKeys.filter(k => (k.group || 'default') === keyGroupFilter);

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
              <Box sx={{ display: 'flex', gap: 1 }}>
                <GroupSelector
                  resourceType="servers"
                  selectedGroup={serverGroupFilter}
                  onGroupChange={setServerGroupFilter}
                  size="small"
                />
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
                    {filteredServers.map((server) => (
                      <MenuItem key={server.source === 'vault' ? `vault-${server.name}` : server.id} value={server.source === 'vault' ? server.name : server.id}>
                        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%', gap: 1 }}>
                          <span>{server.name || server.ip_address} ({server.ip_address}:{server.port})</span>
                          {server.source && (
                            server.source === 'vault' ? (
                              <Chip icon={<Lock fontSize="small" />} label="Vault" size="small" variant="outlined" color="secondary" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                            ) : (
                              <Chip icon={<Storage fontSize="small" />} label="Local" size="small" variant="outlined" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                            )
                          )}
                        </Box>
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Box>
            </Grid>

            <Grid item xs={12} md={6}>
              <Box sx={{ display: 'flex', gap: 1 }}>
                <GroupSelector
                  resourceType="keys"
                  selectedGroup={keyGroupFilter}
                  onGroupChange={setKeyGroupFilter}
                  size="small"
                />
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
                    {filteredSSHKeys.map((key) => (
                      <MenuItem key={key.source === 'vault' ? `vault-${key.name}` : key.id} value={key.source === 'vault' ? key.name : key.id}>
                        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%', gap: 1 }}>
                          <span>{key.name}</span>
                          {key.source && (
                            key.source === 'vault' ? (
                              <Chip icon={<Lock fontSize="small" />} label="Vault" size="small" variant="outlined" color="secondary" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                            ) : (
                              <Chip icon={<Storage fontSize="small" />} label="Local" size="small" variant="outlined" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                            )
                          )}
                        </Box>
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Box>
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
          <Paper sx={{ p: 3, backgroundColor: '#0a0a0a', color: '#e0e0e0' }}>
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
                // Override ANSI colors for better visibility
                '& .ansi-blue-fg': { color: '#5c9cff !important' },
                '& .ansi-bright-blue-fg': { color: '#6eb5ff !important' },
                '& .ansi-red-fg': { color: '#ff6b6b !important' },
                '& .ansi-bright-red-fg': { color: '#ff8787 !important' },
                // Also try inline style overrides
                '& span[style*="34m"]': { color: '#5c9cff !important' },
                '& span[style*="94m"]': { color: '#6eb5ff !important' },
                '& span[style*="31m"]': { color: '#ff6b6b !important' },
                '& span[style*="91m"]': { color: '#ff8787 !important' },
              }}
            >
              <Ansi useClasses>{output}</Ansi>
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
