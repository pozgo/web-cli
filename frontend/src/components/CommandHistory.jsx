import React, { useState, useEffect } from 'react';
import Ansi from 'ansi-to-react';
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
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import { ArrowBack, Visibility, Refresh } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

/**
 * CommandHistory component - view command execution history
 */
const CommandHistory = () => {
  const navigate = useNavigate();
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedEntry, setSelectedEntry] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [filterServer, setFilterServer] = useState('all');

  useEffect(() => {
    fetchHistory();
  }, [filterServer]);

  const fetchHistory = async () => {
    try {
      setLoading(true);
      setError(null);

      let url = '/api/history?limit=50';
      if (filterServer !== 'all') {
        url += `&server=${filterServer}`;
      }

      const response = await fetch(url);

      if (!response.ok) {
        throw new Error('Failed to fetch command history');
      }

      const data = await response.json();
      setHistory(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetails = (entry) => {
    setSelectedEntry(entry);
    setOpenDialog(true);
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  const getStatusColor = (exitCode) => {
    if (exitCode === null || exitCode === undefined) return 'default';
    return exitCode === 0 ? 'success' : 'error';
  };

  const getStatusLabel = (exitCode) => {
    if (exitCode === null || exitCode === undefined) return 'Unknown';
    return exitCode === 0 ? 'Success' : `Failed (${exitCode})`;
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

        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4" component="h1">
            Command History
          </Typography>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <FormControl size="small" sx={{ minWidth: 150 }}>
              <InputLabel>Filter</InputLabel>
              <Select
                value={filterServer}
                onChange={(e) => setFilterServer(e.target.value)}
                label="Filter"
              >
                <MenuItem value="all">All Servers</MenuItem>
                <MenuItem value="local">Local Only</MenuItem>
              </Select>
            </FormControl>
            <IconButton onClick={fetchHistory} color="primary">
              <Refresh />
            </IconButton>
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
        ) : history.length === 0 ? (
          <Paper sx={{ p: 4, textAlign: 'center' }}>
            <Typography variant="body1" color="text.secondary">
              No command history found. Execute some commands to see them here.
            </Typography>
          </Paper>
        ) : (
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Command</TableCell>
                  <TableCell>Server</TableCell>
                  <TableCell>User</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Execution Time</TableCell>
                  <TableCell>Executed At</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {history.map((entry) => (
                  <TableRow key={entry.id}>
                    <TableCell>
                      <Box
                        sx={{
                          maxWidth: 300,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          fontFamily: 'monospace',
                          fontSize: '0.875rem',
                        }}
                      >
                        {entry.command}
                      </Box>
                    </TableCell>
                    <TableCell>{entry.server}</TableCell>
                    <TableCell>{entry.user || '-'}</TableCell>
                    <TableCell>
                      <Chip
                        label={getStatusLabel(entry.exit_code)}
                        color={getStatusColor(entry.exit_code)}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>{entry.execution_time_ms}ms</TableCell>
                    <TableCell>{formatDate(entry.executed_at)}</TableCell>
                    <TableCell align="right">
                      <IconButton
                        color="primary"
                        onClick={() => handleViewDetails(entry)}
                        aria-label="view details"
                      >
                        <Visibility />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Box>

      {/* Details Dialog */}
      <Dialog
        open={openDialog}
        onClose={() => setOpenDialog(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Command Execution Details</DialogTitle>
        <DialogContent>
          {selectedEntry && (
            <Box>
              <Typography variant="subtitle2" sx={{ mt: 2 }}>
                Command:
              </Typography>
              <Paper sx={{ p: 2, backgroundColor: '#f5f5f5', mb: 2 }}>
                <Typography
                  component="pre"
                  sx={{ fontFamily: 'monospace', fontSize: '0.875rem', margin: 0 }}
                >
                  {selectedEntry.command}
                </Typography>
              </Paper>

              <Typography variant="subtitle2">
                Server: {selectedEntry.server}
              </Typography>
              {selectedEntry.user && (
                <Typography variant="subtitle2">
                  User: {selectedEntry.user}
                </Typography>
              )}
              <Typography variant="subtitle2">
                Exit Code: {selectedEntry.exit_code}
              </Typography>
              <Typography variant="subtitle2">
                Execution Time: {selectedEntry.execution_time_ms}ms
              </Typography>
              <Typography variant="subtitle2">
                Executed At: {formatDate(selectedEntry.executed_at)}
              </Typography>

              <Typography variant="subtitle2" sx={{ mt: 2, mb: 1 }}>
                Output:
              </Typography>
              <Paper
                sx={{
                  p: 2,
                  backgroundColor: '#0a0a0a',
                  color: '#e0e0e0',
                  maxHeight: 400,
                  overflow: 'auto',
                }}
              >
                <Typography
                  component="pre"
                  sx={{
                    fontFamily: 'monospace',
                    fontSize: '0.875rem',
                    margin: 0,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                    // Override ANSI colors for better visibility
                    '& .ansi-blue-fg': { color: '#5c9cff !important' },
                    '& .ansi-bright-blue-fg': { color: '#6eb5ff !important' },
                    '& .ansi-red-fg': { color: '#ff6b6b !important' },
                    '& .ansi-bright-red-fg': { color: '#ff8787 !important' },
                    // Also try inline style overrides
                    '& span[style*="34m"]': { color: '#5c9cff !important' },
                    '& span[style*="94m"]': { color: '#6eb5ff !important' },
                    '& span[style*="31m"]': { color: '#ff6b6b !important' },
                    '& span[style*="91m"]': { color: '#ff8787 !important' },
                  }}
                >
                  <Ansi useClasses>{selectedEntry.output || '(no output)'}</Ansi>
                </Typography>
              </Paper>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>Close</Button>
          {selectedEntry && (
            <Button
              variant="contained"
              onClick={() => {
                navigate('/local-commands', {
                  state: { command: selectedEntry.command, user: selectedEntry.user }
                });
              }}
            >
              Run Again
            </Button>
          )}
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default CommandHistory;
