import React, { useState, useEffect } from 'react';
import {
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
  InputAdornment,
  Tooltip,
} from '@mui/material';
import { Add, Delete, Edit, Visibility, VisibilityOff, ContentCopy } from '@mui/icons-material';
import SourceChip from './shared/SourceChip';
import GroupSelector from './shared/GroupSelector';
import StorageSelector from './shared/StorageSelector';
import GroupInput from './shared/GroupInput';

/**
 * EnvVariableList component - displays and manages encrypted environment variables
 */
const EnvVariableList = () => {
  const [envVars, setEnvVars] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [editingVar, setEditingVar] = useState(null);
  const [showValues, setShowValues] = useState({});
  const [formData, setFormData] = useState({
    name: '',
    value: '',
    description: '',
    group: 'default',
    storage: 'local',
  });
  const [showFormValue, setShowFormValue] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState('all');

  // Fetch environment variables from API
  const fetchEnvVars = async () => {
    try {
      setLoading(true);
      setError(null);
      const url = selectedGroup === 'all'
        ? '/api/env-variables'
        : `/api/env-variables?group=${encodeURIComponent(selectedGroup)}`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error('Failed to fetch environment variables');
      }

      const data = await response.json();
      setEnvVars(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Fetch single env var with value revealed
  const fetchEnvVarValue = async (id) => {
    try {
      const response = await fetch(`/api/env-variables/${id}?show_value=true`);
      if (!response.ok) {
        throw new Error('Failed to fetch environment variable value');
      }
      const data = await response.json();
      return data.value;
    } catch (err) {
      setError(err.message);
      return null;
    }
  };

  // Load env vars on component mount or when group changes
  useEffect(() => {
    fetchEnvVars();
  }, [selectedGroup]);

  // Toggle value visibility for a specific variable
  const handleToggleVisibility = async (id) => {
    if (showValues[id]) {
      // Hide value
      setShowValues((prev) => ({ ...prev, [id]: null }));
    } else {
      // Fetch and show value
      const value = await fetchEnvVarValue(id);
      if (value) {
        setShowValues((prev) => ({ ...prev, [id]: value }));
      }
    }
  };

  // Copy value to clipboard
  const handleCopyValue = async (id) => {
    let value = showValues[id];
    if (!value) {
      value = await fetchEnvVarValue(id);
    }
    if (value) {
      navigator.clipboard.writeText(value);
      setSuccess('Value copied to clipboard');
      setTimeout(() => setSuccess(null), 2000);
    }
  };

  // Handle env var deletion
  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this environment variable?')) {
      return;
    }

    try {
      const response = await fetch(`/api/env-variables/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to delete environment variable');
      }

      fetchEnvVars();
      setSuccess('Environment variable deleted successfully');
      setTimeout(() => setSuccess(null), 2000);
    } catch (err) {
      setError(err.message);
    }
  };

  // Handle edit click
  const handleEdit = async (envVar) => {
    // Fetch the actual value for editing
    const value = await fetchEnvVarValue(envVar.id);
    setEditingVar(envVar);
    setFormData({
      name: envVar.name,
      value: value || '',
      description: envVar.description || '',
      group: envVar.group || 'default',
    });
    setShowFormValue(false);
    setOpenDialog(true);
  };

  // Handle add click
  const handleAdd = () => {
    setEditingVar(null);
    setFormData({
      name: '',
      value: '',
      description: '',
      group: 'default',
      storage: 'local',
    });
    setShowFormValue(false);
    setOpenDialog(true);
  };

  // Handle form save
  const handleSave = async () => {
    if (!formData.name || !formData.value) {
      setError('Name and value are required');
      return;
    }

    try {
      // Choose base API endpoint based on storage selection (only for new env vars)
      const baseEndpoint = !editingVar && formData.storage === 'vault'
        ? '/api/vault/env-variables'
        : '/api/env-variables';

      const url = editingVar
        ? `/api/env-variables/${editingVar.id}`
        : baseEndpoint;
      const method = editingVar ? 'PUT' : 'POST';

      // Don't send storage field to the API
      const { storage, ...dataToSend } = formData;

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(dataToSend),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to save environment variable');
      }

      setOpenDialog(false);
      fetchEnvVars();
      setSuccess(editingVar ? 'Environment variable updated' : 'Environment variable created');
      setTimeout(() => setSuccess(null), 2000);
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5" component="h2">
          Environment Variables
        </Typography>
        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
          <GroupSelector
            resourceType="env-variables"
            selectedGroup={selectedGroup}
            onGroupChange={setSelectedGroup}
          />
          <Button
            variant="contained"
            startIcon={<Add />}
            onClick={handleAdd}
          >
            Add Variable
          </Button>
        </Box>
      </Box>

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

      <Alert severity="info" sx={{ mb: 2 }}>
        Environment variables are encrypted at rest using AES-256-GCM encryption.
      </Alert>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : envVars.length === 0 ? (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="body1" color="text.secondary">
            No environment variables found. Click "Add Variable" to create one.
          </Typography>
        </Paper>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Value</TableCell>
                <TableCell>Description</TableCell>
                <TableCell>Group</TableCell>
                <TableCell>Source</TableCell>
                <TableCell>Created At</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {envVars.map((envVar) => (
                <TableRow key={envVar.id || envVar.name}>
                  <TableCell>
                    <Box
                      sx={{
                        fontFamily: 'monospace',
                        fontWeight: 'bold',
                      }}
                    >
                      {envVar.name}
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Box
                        sx={{
                          fontFamily: 'monospace',
                          maxWidth: 200,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {showValues[envVar.id] || '••••••••'}
                      </Box>
                      <Tooltip title={showValues[envVar.id] ? 'Hide value' : 'Show value'}>
                        <IconButton
                          size="small"
                          onClick={() => handleToggleVisibility(envVar.id)}
                        >
                          {showValues[envVar.id] ? <VisibilityOff fontSize="small" /> : <Visibility fontSize="small" />}
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Copy value">
                        <IconButton
                          size="small"
                          onClick={() => handleCopyValue(envVar.id)}
                        >
                          <ContentCopy fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                  <TableCell>{envVar.description || '-'}</TableCell>
                  <TableCell>{envVar.group || 'default'}</TableCell>
                  <TableCell>
                    <SourceChip source={envVar.source} />
                  </TableCell>
                  <TableCell>
                    {new Date(envVar.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell align="right">
                    <IconButton
                      color="primary"
                      onClick={() => handleEdit(envVar)}
                      aria-label="edit"
                      sx={{ mr: 1 }}
                      disabled={envVar.source === 'vault'}
                      title={envVar.source === 'vault' ? 'Vault items cannot be edited here' : 'Edit'}
                    >
                      <Edit />
                    </IconButton>
                    <IconButton
                      color="error"
                      onClick={() => handleDelete(envVar.id)}
                      aria-label="delete"
                      disabled={envVar.source === 'vault'}
                      title={envVar.source === 'vault' ? 'Vault items cannot be deleted here' : 'Delete'}
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

      {/* Add/Edit Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingVar ? 'Edit Environment Variable' : 'Add Environment Variable'}
        </DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Name"
            fullWidth
            variant="outlined"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, '_') })}
            placeholder="MY_VARIABLE_NAME"
            helperText="Must start with letter or underscore, contain only letters, digits, and underscores"
            sx={{ mb: 2 }}
            inputProps={{
              style: { fontFamily: 'monospace' },
            }}
          />
          <TextField
            margin="dense"
            label="Value"
            fullWidth
            variant="outlined"
            type={showFormValue ? 'text' : 'password'}
            value={formData.value}
            onChange={(e) => setFormData({ ...formData, value: e.target.value })}
            placeholder="Enter value (will be encrypted)"
            sx={{ mb: 2 }}
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    onClick={() => setShowFormValue(!showFormValue)}
                    edge="end"
                  >
                    {showFormValue ? <VisibilityOff /> : <Visibility />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
          />
          <TextField
            margin="dense"
            label="Description (optional)"
            fullWidth
            variant="outlined"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Brief description of this variable"
            sx={{ mb: 2 }}
          />
          <GroupInput
            value={formData.group}
            onChange={(value) => setFormData({ ...formData, group: value })}
            resourceType="env-variables"
            helperText="Select an existing group or type a new one"
          />
          {!editingVar && (
            <Box sx={{ mb: 2 }}>
              <StorageSelector
                value={formData.storage}
                onChange={(value) => setFormData({ ...formData, storage: value })}
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button onClick={handleSave} variant="contained">
            {editingVar ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default EnvVariableList;
