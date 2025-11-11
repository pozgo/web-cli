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
} from '@mui/material';
import { Add, Delete, Edit } from '@mui/icons-material';
import AddServerDialog from './AddServerDialog';
import EditServerDialog from './EditServerDialog';

/**
 * ServerList component - displays and manages servers
 */
const ServerList = () => {
  const [servers, setServers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [selectedServer, setSelectedServer] = useState(null);

  // Fetch servers from API
  const fetchServers = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch('/api/servers');

      if (!response.ok) {
        throw new Error('Failed to fetch servers');
      }

      const data = await response.json();
      setServers(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Load servers on component mount
  useEffect(() => {
    fetchServers();
  }, []);

  // Handle server deletion
  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this server?')) {
      return;
    }

    try {
      const response = await fetch(`/api/servers/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to delete server');
      }

      // Refresh the list
      fetchServers();
    } catch (err) {
      setError(err.message);
    }
  };

  // Handle successful server creation
  const handleServerAdded = () => {
    setOpenDialog(false);
    fetchServers();
  };

  // Handle edit click
  const handleEdit = (server) => {
    setSelectedServer(server);
    setOpenEditDialog(true);
  };

  // Handle successful server update
  const handleServerUpdated = () => {
    setOpenEditDialog(false);
    setSelectedServer(null);
    fetchServers();
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5" component="h2">
          Server Management
        </Typography>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={() => setOpenDialog(true)}
        >
          Add New Server
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
      ) : servers.length === 0 ? (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="body1" color="text.secondary">
            No servers found. Click "Add New Server" to create one.
          </Typography>
        </Paper>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Server Name</TableCell>
                <TableCell>IP Address / Hostname</TableCell>
                <TableCell>Port</TableCell>
                <TableCell>Username</TableCell>
                <TableCell>Created At</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {servers.map((server) => (
                <TableRow key={server.id}>
                  <TableCell>{server.name || '-'}</TableCell>
                  <TableCell>{server.ip_address || '-'}</TableCell>
                  <TableCell>{server.port || 22}</TableCell>
                  <TableCell>{server.username || 'root'}</TableCell>
                  <TableCell>
                    {new Date(server.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell align="right">
                    <IconButton
                      color="primary"
                      onClick={() => handleEdit(server)}
                      aria-label="edit"
                      sx={{ mr: 1 }}
                    >
                      <Edit />
                    </IconButton>
                    <IconButton
                      color="error"
                      onClick={() => handleDelete(server.id)}
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

      <AddServerDialog
        open={openDialog}
        onClose={() => setOpenDialog(false)}
        onServerAdded={handleServerAdded}
      />

      <EditServerDialog
        open={openEditDialog}
        onClose={() => {
          setOpenEditDialog(false);
          setSelectedServer(null);
        }}
        onServerUpdated={handleServerUpdated}
        serverData={selectedServer}
      />
    </Box>
  );
};

export default ServerList;
