import React, { useState } from 'react';
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
import StorageSelector from './shared/StorageSelector';
import GroupInput from './shared/GroupInput';

/**
 * AddKeyDialog component - dialog for adding new SSH keys
 * @param {Object} props - Component props
 * @param {boolean} props.open - Whether the dialog is open
 * @param {Function} props.onClose - Function to close the dialog
 * @param {Function} props.onKeyAdded - Callback when key is successfully added
 */
const AddKeyDialog = ({ open, onClose, onKeyAdded }) => {
  const [name, setName] = useState('');
  const [privateKey, setPrivateKey] = useState('');
  const [group, setGroup] = useState('default');
  const [storage, setStorage] = useState('local');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

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

      // Choose API endpoint based on storage selection
      const endpoint = storage === 'vault' ? '/api/vault/ssh-keys' : '/api/keys';

      const response = await fetch(endpoint, {
        method: 'POST',
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
        throw new Error(errorText || 'Failed to create SSH key');
      }

      // Success - reset form and notify parent
      setName('');
      setPrivateKey('');
      setGroup('default');
      setStorage('local');
      setError(null);
      onKeyAdded();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle dialog close
  const handleClose = () => {
    if (!loading) {
      setName('');
      setPrivateKey('');
      setGroup('default');
      setStorage('local');
      setError(null);
      onClose();
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <form onSubmit={handleSubmit}>
        <DialogTitle>Add New SSH Key</DialogTitle>
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

          <GroupInput
            value={group}
            onChange={setGroup}
            resourceType="keys"
            disabled={loading}
            helperText="Select an existing group or type a new one"
          />

          <Box sx={{ mt: 2, mb: 2 }}>
            <StorageSelector
              value={storage}
              onChange={setStorage}
              disabled={loading}
            />
          </Box>

          <Box sx={{ mt: 2 }}>
            <Alert severity="warning">
              <strong>Security Note:</strong> Private keys are encrypted with AES-256-GCM before storage. Never share your private keys.
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
            {loading ? 'Adding...' : 'Add Key'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default AddKeyDialog;
