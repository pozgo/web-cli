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
} from '@mui/material';
import { PlayArrow, ArrowBack, Save } from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';

/**
 * LocalCommands component - execute commands on the local server
 * Provides command input, user selection, execution, and output display
 */
const LocalCommands = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [command, setCommand] = useState('');
  const [user, setUser] = useState('root');
  const [saveAs, setSaveAs] = useState('');
  const [shouldSave, setShouldSave] = useState(false);
  const [output, setOutput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [savedCommands, setSavedCommands] = useState([]);
  const [selectedSavedCommand, setSelectedSavedCommand] = useState('');
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
  const [sudoPassword, setSudoPassword] = useState('');
  const [availableUsers, setAvailableUsers] = useState([]);
  const [currentUsername, setCurrentUsername] = useState('current');

  // Fetch saved commands, local users, and current user on mount
  useEffect(() => {
    fetchSavedCommands();
    fetchLocalUsers();
    fetchCurrentUser();
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
    }
  }, [location]);

  const fetchSavedCommands = async () => {
    try {
      const response = await fetch('/api/saved-commands');
      if (response.ok) {
        const data = await response.json();
        // Filter to show only local commands (not remote)
        const localCommands = data.filter(cmd => !cmd.is_remote);
        setSavedCommands(localCommands || []);
      }
    } catch (err) {
      console.error('Failed to fetch saved commands:', err);
    }
  };

  const fetchCurrentUser = async () => {
    try {
      const response = await fetch('/api/system/current-user');
      if (response.ok) {
        const data = await response.json();
        setCurrentUsername(data.username || 'current');
      }
    } catch (err) {
      console.error('Failed to fetch current user:', err);
    }
  };

  const fetchLocalUsers = async () => {
    try {
      const response = await fetch('/api/local-users');
      if (response.ok) {
        const data = await response.json();
        const users = data || [];

        // Build unique list: always include 'current' and 'root', then add stored users
        const userSet = new Set();

        // Add current user
        userSet.add('current');

        // Add root
        userSet.add('root');

        // Add all local users from database, removing duplicates
        users.forEach(u => {
          if (u.name && u.name !== 'current') {
            userSet.add(u.name);
          }
        });

        setAvailableUsers(Array.from(userSet));
      }
    } catch (err) {
      console.error('Failed to fetch local users:', err);
      // Fallback to basic options
      setAvailableUsers(['current', 'root']);
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
      }
    }
  };

  // Execute command - check if password needed first
  const handleExecute = async () => {
    if (!command.trim()) {
      setError('Command is required');
      return;
    }

    // If running as root, ask for password first
    if (user === 'root') {
      setPasswordDialogOpen(true);
      return;
    }

    // Otherwise execute directly (current user doesn't need password)
    executeCommand('');
  };

  // Actually execute the command with the provided password
  const executeCommand = async (password) => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    setOutput('');

    try {
      const payload = {
        command: command.trim(),
        user: user || 'root',
      };

      // Add sudo password if provided
      if (password) {
        payload.sudo_password = password;
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
        throw new Error('Failed to execute command');
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
    executeCommand(sudoPassword);
    setSudoPassword(''); // Clear password after use for security
  };

  // Handle password dialog cancel
  const handlePasswordCancel = () => {
    setPasswordDialogOpen(false);
    setSudoPassword('');
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
          Local Commands
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          Execute bash commands on the local server. All commands are saved to history.
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
                      {username === 'current' ? `current (${currentUsername})` : username}
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
                disabled={loading || !command.trim()}
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

      {/* Sudo Password Dialog */}
      <Dialog open={passwordDialogOpen} onClose={handlePasswordCancel}>
        <DialogTitle>Sudo Password Required</DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ mb: 2 }}>
            This command requires root privileges. Please enter your sudo password:
          </Typography>
          <TextField
            autoFocus
            fullWidth
            type="password"
            label="Sudo Password"
            value={sudoPassword}
            onChange={(e) => setSudoPassword(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter') {
                handlePasswordSubmit();
              }
            }}
          />
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

export default LocalCommands;
