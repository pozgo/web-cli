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
 * EditLocalUserDialog component - dialog for editing existing local users
 * @param {Object} props - Component props
 * @param {boolean} props.open - Whether the dialog is open
 * @param {Function} props.onClose - Function to close the dialog
 * @param {Function} props.onUserUpdated - Callback when user is successfully updated
 * @param {Object} props.userData - The user data to edit (id, name)
 */
const EditLocalUserDialog = ({ open, onClose, onUserUpdated, userData }) => {
  const [name, setName] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  // Pre-fill form when userData changes
  useEffect(() => {
    if (userData) {
      setName(userData.name || '');
    }
  }, [userData]);

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validate that name is provided
    if (!name.trim()) {
      setError('Username is required');
      return;
    }

    // Validate username format (Unix username conventions)
    const usernameRegex = /^[a-z_][a-z0-9_-]*$/;
    if (!usernameRegex.test(name.trim())) {
      setError('Username must start with a letter or underscore and contain only lowercase letters, digits, underscores, or hyphens');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`/api/local-users/${userData.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: name.trim(),
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to update local user');
      }

      // Success - reset form and notify parent
      setError(null);
      onUserUpdated();
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
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <form onSubmit={handleSubmit}>
        <DialogTitle>Edit Local User</DialogTitle>
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
              <strong>Note:</strong> This updates the reference. The user must exist on your system.
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
            {loading ? 'Updating...' : 'Update User'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default EditLocalUserDialog;
