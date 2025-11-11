import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Button,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Alert,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Chip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  FormControlLabel,
  Switch,
} from '@mui/material';
import { Add, Delete, Edit, PlayArrow, ArrowBack, Computer, Cloud } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

/**
 * SavedCommands component - manage and execute saved command templates
 */
const SavedCommands = () => {
  const navigate = useNavigate();
  const [commands, setCommands] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [editingCommand, setEditingCommand] = useState(null);
  const [servers, setServers] = useState([]);
  const [sshKeys, setSSHKeys] = useState([]);
  const [formData, setFormData] = useState({
    name: '',
    command: '',
    description: '',
    user: 'root',
    is_remote: false,
    server_id: null,
    ssh_key_id: null,
  });

  useEffect(() => {
    fetchCommands();
    fetchServers();
    fetchSSHKeys();
  }, []);

  const fetchCommands = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch('/api/saved-commands');

      if (!response.ok) {
        throw new Error('Failed to fetch saved commands');
      }

      const data = await response.json();
      setCommands(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
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
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this saved command?')) {
      return;
    }

    try {
      const response = await fetch(`/api/saved-commands/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to delete saved command');
      }

      fetchCommands();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleEdit = (cmd) => {
    setEditingCommand(cmd);
    setFormData({
      name: cmd.name,
      command: cmd.command,
      description: cmd.description || '',
      user: cmd.user || 'root',
      is_remote: cmd.is_remote || false,
      server_id: cmd.server_id || null,
      ssh_key_id: cmd.ssh_key_id || null,
    });
    setOpenDialog(true);
  };

  const handleAdd = () => {
    setEditingCommand(null);
    setFormData({
      name: '',
      command: '',
      description: '',
      user: 'root',
      is_remote: false,
      server_id: null,
      ssh_key_id: null,
    });
    setOpenDialog(true);
  };

  const handleSave = async () => {
    if (!formData.name || !formData.command) {
      setError('Name and command are required');
      return;
    }

    try {
      const url = editingCommand
        ? `/api/saved-commands/${editingCommand.id}`
        : '/api/saved-commands';
      const method = editingCommand ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        throw new Error('Failed to save command');
      }

      setOpenDialog(false);
      fetchCommands();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleExecute = (cmd) => {
    // Navigate to appropriate commands page based on type
    if (cmd.is_remote) {
      navigate('/remote-commands', {
        state: {
          command: cmd.command,
          user: cmd.user,
          server_id: cmd.server_id,
          ssh_key_id: cmd.ssh_key_id,
        },
      });
    } else {
      navigate('/local-commands', {
        state: { command: cmd.command, user: cmd.user },
      });
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

        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4" component="h1">
            Saved Commands
          </Typography>
          <Button
            variant="contained"
            startIcon={<Add />}
            onClick={handleAdd}
          >
            Add Command
          </Button>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}

        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CircularProgress />
          </Box>
        ) : commands.length === 0 ? (
          <Paper sx={{ p: 4, textAlign: 'center' }}>
            <Typography variant="body1" color="text.secondary">
              No saved commands found. Click "Add Command" to create one.
            </Typography>
          </Paper>
        ) : (
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Type</TableCell>
                  <TableCell>Command</TableCell>
                  <TableCell>User</TableCell>
                  <TableCell>Description</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {commands.map((cmd) => (
                  <TableRow key={cmd.id}>
                    <TableCell>{cmd.name}</TableCell>
                    <TableCell>
                      <Chip
                        icon={cmd.is_remote ? <Cloud /> : <Computer />}
                        label={cmd.is_remote ? 'Remote' : 'Local'}
                        color={cmd.is_remote ? 'primary' : 'default'}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Box
                        sx={{
                          maxWidth: 300,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          fontFamily: 'monospace',
                          fontSize: '0.875rem',
                        }}
                      >
                        {cmd.command}
                      </Box>
                    </TableCell>
                    <TableCell>{cmd.user}</TableCell>
                    <TableCell>{cmd.description || '-'}</TableCell>
                    <TableCell align="right">
                      <IconButton
                        color="primary"
                        onClick={() => handleExecute(cmd)}
                        aria-label="execute"
                        sx={{ mr: 1 }}
                        title="Execute"
                      >
                        <PlayArrow />
                      </IconButton>
                      <IconButton
                        color="primary"
                        onClick={() => handleEdit(cmd)}
                        aria-label="edit"
                        sx={{ mr: 1 }}
                      >
                        <Edit />
                      </IconButton>
                      <IconButton
                        color="error"
                        onClick={() => handleDelete(cmd.id)}
                        aria-label="delete"
                      >
                        <Delete />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Box>

      {/* Add/Edit Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{editingCommand ? 'Edit Command' : 'Add New Command'}</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Name"
            fullWidth
            variant="outlined"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Command"
            fullWidth
            multiline
            rows={3}
            variant="outlined"
            value={formData.command}
            onChange={(e) => setFormData({ ...formData, command: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Description (optional)"
            fullWidth
            variant="outlined"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Run As User"
            fullWidth
            variant="outlined"
            value={formData.user}
            onChange={(e) => setFormData({ ...formData, user: e.target.value })}
            sx={{ mb: 2 }}
          />

          <FormControlLabel
            control={
              <Switch
                checked={formData.is_remote}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    is_remote: e.target.checked,
                    // Reset remote-specific fields when switching to local
                    server_id: e.target.checked ? formData.server_id : null,
                    ssh_key_id: e.target.checked ? formData.ssh_key_id : null,
                  })
                }
              />
            }
            label={
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                {formData.is_remote ? <Cloud /> : <Computer />}
                <Typography>{formData.is_remote ? 'Remote Command' : 'Local Command'}</Typography>
              </Box>
            }
            sx={{ mb: 2 }}
          />

          {formData.is_remote && (
            <>
              <FormControl fullWidth sx={{ mb: 2 }}>
                <InputLabel>Server</InputLabel>
                <Select
                  value={formData.server_id || ''}
                  onChange={(e) => setFormData({ ...formData, server_id: e.target.value || null })}
                  label="Server"
                >
                  <MenuItem value="">
                    <em>None</em>
                  </MenuItem>
                  {servers.map((server) => (
                    <MenuItem key={server.id} value={server.id}>
                      {server.name || server.ip} ({server.user}@{server.ip}:{server.port})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              <FormControl fullWidth sx={{ mb: 2 }}>
                <InputLabel>SSH Key (optional)</InputLabel>
                <Select
                  value={formData.ssh_key_id || ''}
                  onChange={(e) => setFormData({ ...formData, ssh_key_id: e.target.value || null })}
                  label="SSH Key (optional)"
                >
                  <MenuItem value="">
                    <em>None (use password)</em>
                  </MenuItem>
                  {sshKeys.map((key) => (
                    <MenuItem key={key.id} value={key.id}>
                      {key.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button onClick={handleSave} variant="contained">
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default SavedCommands;
