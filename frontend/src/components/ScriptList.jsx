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
  Tooltip,
  Chip,
} from '@mui/material';
import { Add, Delete, Edit, Code, Upload, ContentCopy } from '@mui/icons-material';
import SourceChip from './shared/SourceChip';
import GroupSelector from './shared/GroupSelector';
import StorageSelector from './shared/StorageSelector';
import GroupInput from './shared/GroupInput';

/**
 * ScriptList component - displays and manages bash scripts stored in the database
 */
const ScriptList = () => {
  const [scripts, setScripts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [editingScript, setEditingScript] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    content: '',
    filename: '',
    group: 'default',
    storage: 'local',
  });
  const [selectedGroup, setSelectedGroup] = useState('all');

  // Fetch scripts from API
  const fetchScripts = async () => {
    try {
      setLoading(true);
      setError(null);
      const url = selectedGroup === 'all'
        ? '/api/bash-scripts'
        : `/api/bash-scripts?group=${encodeURIComponent(selectedGroup)}`;
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error('Failed to fetch scripts');
      }

      const data = await response.json();
      setScripts(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Fetch single script with full content
  const fetchScriptContent = async (id) => {
    try {
      const response = await fetch(`/api/bash-scripts/${id}`);
      if (!response.ok) {
        throw new Error('Failed to fetch script content');
      }
      const data = await response.json();
      return data;
    } catch (err) {
      setError(err.message);
      return null;
    }
  };

  // Load scripts on component mount or when group changes
  useEffect(() => {
    fetchScripts();
  }, [selectedGroup]);

  // Handle file upload
  const handleFileUpload = (event) => {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      setFormData({
        ...formData,
        content: e.target.result,
        filename: file.name,
        name: formData.name || file.name.replace(/\.[^/.]+$/, ''), // Use filename as name if empty
      });
    };
    reader.readAsText(file);
  };

  // Copy content to clipboard
  const handleCopyContent = async (id) => {
    const script = await fetchScriptContent(id);
    if (script && script.content) {
      navigator.clipboard.writeText(script.content);
      setSuccess('Script content copied to clipboard');
      setTimeout(() => setSuccess(null), 2000);
    }
  };

  // Handle script deletion
  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this script?')) {
      return;
    }

    try {
      const response = await fetch(`/api/bash-scripts/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to delete script');
      }

      fetchScripts();
      setSuccess('Script deleted successfully');
      setTimeout(() => setSuccess(null), 2000);
    } catch (err) {
      setError(err.message);
    }
  };

  // Handle edit click
  const handleEdit = async (script) => {
    const fullScript = await fetchScriptContent(script.id);
    if (fullScript) {
      setEditingScript(script);
      setFormData({
        name: fullScript.name,
        description: fullScript.description || '',
        content: fullScript.content || '',
        filename: fullScript.filename || '',
        group: fullScript.group || 'default',
      });
      setOpenDialog(true);
    }
  };

  // Handle add click
  const handleAdd = () => {
    setEditingScript(null);
    setFormData({
      name: '',
      description: '',
      content: '',
      filename: '',
      group: 'default',
      storage: 'local',
    });
    setOpenDialog(true);
  };

  // Handle form save
  const handleSave = async () => {
    if (!formData.name || !formData.content) {
      setError('Name and content are required');
      return;
    }

    try {
      // Choose base API endpoint based on storage selection (only for new scripts)
      const baseEndpoint = !editingScript && formData.storage === 'vault'
        ? '/api/vault/bash-scripts'
        : '/api/bash-scripts';

      const url = editingScript
        ? `/api/bash-scripts/${editingScript.id}`
        : baseEndpoint;
      const method = editingScript ? 'PUT' : 'POST';

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
        throw new Error(errorText || 'Failed to save script');
      }

      setOpenDialog(false);
      fetchScripts();
      setSuccess(editingScript ? 'Script updated' : 'Script created');
      setTimeout(() => setSuccess(null), 2000);
    } catch (err) {
      setError(err.message);
    }
  };

  // Format file size for display
  const formatSize = (content) => {
    if (!content) return '0 B';
    const bytes = new Blob([content]).size;
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5" component="h2">
          Bash Scripts
        </Typography>
        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
          <GroupSelector
            resourceType="bash-scripts"
            selectedGroup={selectedGroup}
            onGroupChange={setSelectedGroup}
          />
          <Button
            variant="contained"
            startIcon={<Add />}
            onClick={handleAdd}
          >
            Add Script
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
        Script content is encrypted at rest using AES-256-GCM encryption.
      </Alert>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : scripts.length === 0 ? (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="body1" color="text.secondary">
            No scripts found. Click "Add Script" to create one.
          </Typography>
        </Paper>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Filename</TableCell>
                <TableCell>Description</TableCell>
                <TableCell>Group</TableCell>
                <TableCell>Source</TableCell>
                <TableCell>Created At</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {scripts.map((script) => (
                <TableRow key={script.id || script.name}>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Code fontSize="small" color="primary" />
                      <Typography fontWeight="bold">{script.name}</Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    {script.filename ? (
                      <Chip
                        label={script.filename}
                        size="small"
                        variant="outlined"
                        sx={{ fontFamily: 'monospace' }}
                      />
                    ) : (
                      <Typography color="text.secondary">-</Typography>
                    )}
                  </TableCell>
                  <TableCell>{script.description || '-'}</TableCell>
                  <TableCell>{script.group || 'default'}</TableCell>
                  <TableCell>
                    <SourceChip source={script.source} />
                  </TableCell>
                  <TableCell>
                    {new Date(script.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell align="right">
                    <Tooltip title="Copy script content">
                      <IconButton
                        size="small"
                        onClick={() => handleCopyContent(script.id)}
                        sx={{ mr: 1 }}
                      >
                        <ContentCopy fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <IconButton
                      color="primary"
                      onClick={() => handleEdit(script)}
                      aria-label="edit"
                      sx={{ mr: 1 }}
                      disabled={script.source === 'vault'}
                      title={script.source === 'vault' ? 'Vault items cannot be edited here' : 'Edit'}
                    >
                      <Edit />
                    </IconButton>
                    <IconButton
                      color="error"
                      onClick={() => handleDelete(script.id)}
                      aria-label="delete"
                      disabled={script.source === 'vault'}
                      title={script.source === 'vault' ? 'Vault items cannot be deleted here' : 'Delete'}
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
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>
          {editingScript ? 'Edit Script' : 'Add Script'}
        </DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Name"
            fullWidth
            variant="outlined"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="My Script Name"
            helperText="A descriptive name for the script"
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Description (optional)"
            fullWidth
            variant="outlined"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Brief description of what this script does"
            sx={{ mb: 2 }}
          />
          <GroupInput
            value={formData.group}
            onChange={(value) => setFormData({ ...formData, group: value })}
            resourceType="bash-scripts"
            helperText="Select an existing group or type a new one"
          />
          {!editingScript && (
            <Box sx={{ mb: 2 }}>
              <StorageSelector
                value={formData.storage}
                onChange={(value) => setFormData({ ...formData, storage: value })}
              />
            </Box>
          )}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
            <TextField
              margin="dense"
              label="Filename (optional)"
              variant="outlined"
              value={formData.filename}
              onChange={(e) => setFormData({ ...formData, filename: e.target.value })}
              placeholder="script.sh"
              helperText="Original filename"
              sx={{ flex: 1 }}
              inputProps={{
                style: { fontFamily: 'monospace' },
              }}
            />
            <Box>
              <input
                type="file"
                accept=".sh,.bash,text/*"
                onChange={handleFileUpload}
                style={{ display: 'none' }}
                id="script-file-upload"
              />
              <label htmlFor="script-file-upload">
                <Button
                  variant="outlined"
                  component="span"
                  startIcon={<Upload />}
                >
                  Upload File
                </Button>
              </label>
            </Box>
          </Box>
          <TextField
            margin="dense"
            label="Script Content"
            fullWidth
            variant="outlined"
            multiline
            rows={15}
            value={formData.content}
            onChange={(e) => setFormData({ ...formData, content: e.target.value })}
            placeholder="#!/bin/bash&#10;&#10;# Your script here..."
            helperText={`Script content (${formatSize(formData.content)})`}
            inputProps={{
              style: { fontFamily: 'monospace', fontSize: '0.9rem' },
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
          <Button onClick={handleSave} variant="contained">
            {editingScript ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default ScriptList;
