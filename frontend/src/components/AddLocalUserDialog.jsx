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

/**
 * AddLocalUserDialog component - dialog for adding new local users
 * @param {Object} props - Component props
 * @param {boolean} props.open - Whether the dialog is open
 * @param {Function} props.onClose - Function to close the dialog
 * @param {Function} props.onUserAdded - Callback when user is successfully added
 */
const AddLocalUserDialog = ({ open, onClose, onUserAdded }) => {
  const [name, setName] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validate that name is provided
    if (!name.trim()) {
      setError('Username is required');
      return;
    }

    // Validate username format (Unix username conventions)
    // Must start with a letter or underscore, followed by letters, digits, underscores, or hyphens
    const usernameRegex = /^[a-z_][a-z0-9_-]*$/;
    if (!usernameRegex.test(name.trim())) {
      setError('Username must start with a letter or underscore and contain only lowercase letters, digits, underscores, or hyphens');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await fetch('/api/local-users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: name.trim(),
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to create local user');
      }

      // Success - reset form and notify parent
      setName('');
      setError(null);
      onUserAdded();
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
      setError(null);
      onClose();
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <form onSubmit={handleSubmit}>
        <DialogTitle>Add Local User</DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <TextField
            autoFocus
            margin="dense"
            label="Username"
            type="text"
            fullWidth
            variant="outlined"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., root, ubuntu, admin"
            helperText="Unix username for local command execution"
            disabled={loading}
          />

          <Box sx={{ mt: 2 }}>
            <Alert severity="info">
              <strong>Note:</strong> This creates a reference to an existing local system user. The user must already exist on your system.
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
            {loading ? 'Adding...' : 'Add User'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default AddLocalUserDialog;
