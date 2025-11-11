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
  Tabs,
  Tab,
} from '@mui/material';
import { Add, Delete, Edit, ArrowBack } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import AddKeyDialog from './AddKeyDialog';
import EditKeyDialog from './EditKeyDialog';
import ServerList from './ServerList';
import LocalUserList from './LocalUserList';

/**
 * AdminPanel component - displays and manages SSH keys and servers
 */
const AdminPanel = () => {
  const navigate = useNavigate();
  const [tabValue, setTabValue] = useState(0);
  const [keys, setKeys] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [selectedKey, setSelectedKey] = useState(null);

  // Handle tab change
  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  // Fetch SSH keys from API
  const fetchKeys = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch('/api/keys');

      if (!response.ok) {
        throw new Error('Failed to fetch SSH keys');
      }

      const data = await response.json();
      setKeys(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Load keys on component mount
  useEffect(() => {
    fetchKeys();
  }, []);

  // Handle key deletion
  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this SSH key?')) {
      return;
    }

    try {
      const response = await fetch(`/api/keys/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to delete SSH key');
      }

      // Refresh the list
      fetchKeys();
    } catch (err) {
      setError(err.message);
    }
  };

  // Handle successful key creation
  const handleKeyAdded = () => {
    setOpenDialog(false);
    fetchKeys();
  };

  // Handle edit click
  const handleEdit = (key) => {
    setSelectedKey(key);
    setOpenEditDialog(true);
  };

  // Handle successful key update
  const handleKeyUpdated = () => {
    setOpenEditDialog(false);
    setSelectedKey(null);
    fetchKeys();
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

        <Typography variant="h4" component="h1" sx={{ mb: 3 }}>
          Admin Panel
        </Typography>

        <Paper sx={{ mb: 3 }}>
          <Tabs value={tabValue} onChange={handleTabChange} aria-label="admin panel tabs">
            <Tab label="SSH Keys" />
            <Tab label="Servers" />
            <Tab label="Local Users" />
          </Tabs>
        </Paper>

        {tabValue === 0 && (
          <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
              <Typography variant="h5" component="h2">
                SSH Key Management
              </Typography>
              <Button
                variant="contained"
                startIcon={<Add />}
                onClick={() => setOpenDialog(true)}
              >
                Add New Key
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
            ) : keys.length === 0 ? (
              <Paper sx={{ p: 4, textAlign: 'center' }}>
                <Typography variant="body1" color="text.secondary">
                  No SSH keys found. Click "Add New Key" to create one.
                </Typography>
              </Paper>
            ) : (
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Name</TableCell>
                      <TableCell>Private Key</TableCell>
                      <TableCell>Created At</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {keys.map((key) => (
                      <TableRow key={key.id}>
                        <TableCell>{key.name}</TableCell>
                        <TableCell>
                          <Box
                            sx={{
                              maxWidth: 400,
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                              fontFamily: 'monospace',
                              fontSize: '0.875rem',
                            }}
                          >
                            {key.private_key}
                          </Box>
                        </TableCell>
                        <TableCell>
                          {new Date(key.created_at).toLocaleString()}
                        </TableCell>
                        <TableCell align="right">
                          <IconButton
                            color="primary"
                            onClick={() => handleEdit(key)}
                            aria-label="edit"
                            sx={{ mr: 1 }}
                          >
                            <Edit />
                          </IconButton>
                          <IconButton
                            color="error"
                            onClick={() => handleDelete(key.id)}
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
        )}

        {tabValue === 1 && <ServerList />}

        {tabValue === 2 && <LocalUserList />}
      </Box>

      <AddKeyDialog
        open={openDialog}
        onClose={() => setOpenDialog(false)}
        onKeyAdded={handleKeyAdded}
      />

      <EditKeyDialog
        open={openEditDialog}
        onClose={() => {
          setOpenEditDialog(false);
          setSelectedKey(null);
        }}
        onKeyUpdated={handleKeyUpdated}
        keyData={selectedKey}
      />
    </Container>
  );
};

export default AdminPanel;
