import React, { useState, useEffect, Component } from 'react';
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
import { useNavigate, useLocation } from 'react-router-dom';
import AddKeyDialog from './AddKeyDialog';
import EditKeyDialog from './EditKeyDialog';
import ServerList from './ServerList';
import LocalUserList from './LocalUserList';
import EnvVariableList from './EnvVariableList';
import VaultSettings from './VaultSettings';
import SourceChip from './shared/SourceChip';
import GroupSelector from './shared/GroupSelector';

/**
 * Error Boundary for VaultSettings
 */
class VaultErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('VaultSettings error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <Alert severity="error">
          Error loading Vault Settings: {this.state.error?.message || 'Unknown error'}
        </Alert>
      );
    }
    return this.props.children;
  }
}

/**
 * AdminPanel component - displays and manages SSH keys and servers
 */
const AdminPanel = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [tabValue, setTabValue] = useState(location.state?.tab || 0);
  const [keys, setKeys] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [selectedKey, setSelectedKey] = useState(null);
  const [selectedGroup, setSelectedGroup] = useState('all');

  // Handle tab change
  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  // Fetch SSH keys from API
  const fetchKeys = async () => {
    try {
      setLoading(true);
      setError(null);
      const url = selectedGroup === 'all'
        ? '/api/keys'
        : `/api/keys?group=${encodeURIComponent(selectedGroup)}`;
      const response = await fetch(url);

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

  // Update tab when navigating with state (e.g., from VaultIcon)
  useEffect(() => {
    if (location.state?.tab !== undefined) {
      setTabValue(location.state.tab);
    }
  }, [location.state]);

  // Load keys on component mount or when group changes
  useEffect(() => {
    fetchKeys();
  }, [selectedGroup]);

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
            <Tab label="Environment Variables" />
            <Tab label="Vault Integration" />
          </Tabs>
        </Paper>

        {tabValue === 0 && (
          <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
              <Typography variant="h5" component="h2">
                SSH Key Management
              </Typography>
              <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
                <GroupSelector
                  resourceType="keys"
                  selectedGroup={selectedGroup}
                  onGroupChange={setSelectedGroup}
                />
                <Button
                  variant="contained"
                  startIcon={<Add />}
                  onClick={() => setOpenDialog(true)}
                >
                  Add New Key
                </Button>
              </Box>
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
                      <TableCell>Group</TableCell>
                      <TableCell>Source</TableCell>
                      <TableCell>Created At</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {keys.map((key) => (
                      <TableRow key={key.id || key.name}>
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
                        <TableCell>{key.group || 'default'}</TableCell>
                        <TableCell>
                          <SourceChip source={key.source} />
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
                            disabled={key.source === 'vault'}
                            title={key.source === 'vault' ? 'Vault items cannot be edited here' : 'Edit'}
                          >
                            <Edit />
                          </IconButton>
                          <IconButton
                            color="error"
                            onClick={() => handleDelete(key.id)}
                            aria-label="delete"
                            disabled={key.source === 'vault'}
                            title={key.source === 'vault' ? 'Vault items cannot be deleted here' : 'Delete'}
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

        {tabValue === 3 && <EnvVariableList />}

        {tabValue === 4 && (
          <VaultErrorBoundary>
            <VaultSettings />
          </VaultErrorBoundary>
        )}
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
