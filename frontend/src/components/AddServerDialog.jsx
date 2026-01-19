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
 * AddServerDialog component - dialog for adding new servers
 * @param {Object} props - Component props
 * @param {boolean} props.open - Whether the dialog is open
 * @param {Function} props.onClose - Function to close the dialog
 * @param {Function} props.onServerAdded - Callback when server is successfully added
 */
const AddServerDialog = ({ open, onClose, onServerAdded }) => {
  const [name, setName] = useState('');
  const [ipAddress, setIPAddress] = useState('');
  const [port, setPort] = useState('22');
  const [username, setUsername] = useState('root');
  const [group, setGroup] = useState('default');
  const [storage, setStorage] = useState('local');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  // Validate hostname according to RFC 1123
  const validateHostname = (hostname) => {
    if (!hostname) return true; // Empty is OK, will be validated with IP address

    // Hostname validation: alphanumeric, hyphens, dots (no spaces)
    // Must start and end with alphanumeric, labels max 63 chars, total max 253 chars
    const hostnameRegex = /^(?=.{1,253}$)(?:(?!-)[A-Za-z0-9-]{1,63}(?<!-)\.)*(?!-)[A-Za-z0-9-]{1,63}(?<!-)$/;

    if (!hostnameRegex.test(hostname)) {
      return false;
    }

    return true;
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validate that at least one field is provided
    if (!name.trim() && !ipAddress.trim()) {
      setError('At least one of Server Name or IP Address must be provided');
      return;
    }

    // Validate hostname format (no spaces, follows naming conventions)
    if (name.trim() && !validateHostname(name.trim())) {
      setError('Server name must follow hostname conventions (alphanumeric, hyphens, dots only; no spaces)');
      return;
    }

    // Validate port number
    const portNum = parseInt(port, 10);
    if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
      setError('Port must be a number between 1 and 65535');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // Choose API endpoint based on storage selection
      const endpoint = storage === 'vault' ? '/api/vault/servers' : '/api/servers';

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: name.trim() || undefined,
          ip_address: ipAddress.trim() || undefined,
          port: portNum,
          username: username.trim() || 'root',
          group: group.trim() || 'default',
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to create server');
      }

      // Success - reset form and notify parent
      setName('');
      setIPAddress('');
      setPort('22');
      setUsername('root');
      setGroup('default');
      setStorage('local');
      setError(null);
      onServerAdded();
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
      setIPAddress('');
      setPort('22');
      setUsername('root');
      setGroup('default');
      setStorage('local');
      setError(null);
      onClose();
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <form onSubmit={handleSubmit}>
        <DialogTitle>Add New Server</DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <TextField
            autoFocus
            margin="dense"
            label="Server Name"
            type="text"
            fullWidth
            variant="outlined"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., production-server"
            helperText="A descriptive name for this server"
            disabled={loading}
          />

          <TextField
            margin="dense"
            label="IP Address"
            type="text"
            fullWidth
            variant="outlined"
            value={ipAddress}
            onChange={(e) => setIPAddress(e.target.value)}
            placeholder="e.g., 192.168.1.100 or server.example.com"
            helperText="Server IP address or hostname"
            disabled={loading}
          />

          <TextField
            margin="dense"
            label="SSH Port"
            type="number"
            fullWidth
            variant="outlined"
            value={port}
            onChange={(e) => setPort(e.target.value)}
            placeholder="22"
            helperText="SSH port number (default: 22)"
            disabled={loading}
            inputProps={{
              min: 1,
              max: 65535,
            }}
          />

          <TextField
            margin="dense"
            label="SSH Username"
            type="text"
            fullWidth
            variant="outlined"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="root"
            helperText="Username for SSH connections (default: root)"
            disabled={loading}
          />

          <GroupInput
            value={group}
            onChange={setGroup}
            resourceType="servers"
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
            <Alert severity="info">
              <strong>Note:</strong> At least one of Server Name or IP Address must be provided. Server names must follow hostname conventions (no spaces).
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
            {loading ? 'Adding...' : 'Add Server'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default AddServerDialog;
