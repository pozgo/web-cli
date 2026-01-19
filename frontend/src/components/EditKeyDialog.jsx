import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Alert,
  Box,
} from '@mui/material';

/**
 * EditKeyDialog component - dialog for editing existing SSH keys
 * @param {Object} props - Component props
 * @param {boolean} props.open - Whether the dialog is open
 * @param {Function} props.onClose - Function to close the dialog
 * @param {Function} props.onKeyUpdated - Callback when key is successfully updated
 * @param {Object} props.keyData - The SSH key data to edit (id, name, private_key)
 */
const EditKeyDialog = ({ open, onClose, onKeyUpdated, keyData }) => {
  const [name, setName] = useState('');
  const [privateKey, setPrivateKey] = useState('');
  const [group, setGroup] = useState('default');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  // Pre-fill form when keyData changes
  useEffect(() => {
    if (keyData) {
      setName(keyData.name || '');
      setPrivateKey(keyData.private_key || '');
      setGroup(keyData.group || 'default');
    }
  }, [keyData]);

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validate inputs
    if (!name.trim()) {
      setError('Name is required');
      return;
    }

    if (!privateKey.trim()) {
      setError('Private key is required');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`/api/keys/${keyData.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: name.trim(),
          private_key: privateKey,  // Don't trim - SSH keys need their newlines preserved
          group: group.trim() || 'default',
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to update SSH key');
      }

      // Success - reset form and notify parent
      setError(null);
      onKeyUpdated();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle dialog close
  const handleClose = () => {
    if (!loading) {
      setError(null);
      onClose();
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <form onSubmit={handleSubmit}>
        <DialogTitle>Edit SSH Key</DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <TextField
            autoFocus
            margin="dense"
            label="Key Name"
            type="text"
            fullWidth
            variant="outlined"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., my-server-key"
            helperText="A descriptive name for this SSH key"
            required
            disabled={loading}
          />

          <TextField
            margin="dense"
            label="Private Key"
            type="text"
            fullWidth
            variant="outlined"
            multiline
            rows={6}
            value={privateKey}
            onChange={(e) => setPrivateKey(e.target.value)}
            placeholder="-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUA..."
            helperText="Paste your SSH private key here (will be encrypted)"
            required
            disabled={loading}
            sx={{
              '& textarea': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
          />

          <TextField
            margin="dense"
            label="Group"
            type="text"
            fullWidth
            variant="outlined"
            value={group}
            onChange={(e) => setGroup(e.target.value)}
            placeholder="default"
            helperText="Group for organizing SSH keys (default: default)"
            disabled={loading}
          />

          <Box sx={{ mt: 2 }}>
            <Alert severity="warning">
              <strong>Security Note:</strong> Private keys are encrypted with AES-256-GCM before storage. This will replace the existing key.
            </Alert>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} disabled={loading}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={loading}
          >
            {loading ? 'Updating...' : 'Update Key'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default EditKeyDialog;
