import React, { useState, useEffect } from 'react';
import Ansi from 'ansi-to-react';
import {
  Container,
  Typography,
  Box,
  Button,
  Paper,
  Alert,
  CircularProgress,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Checkbox,
  Grid,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Chip,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import { PlayArrow, ArrowBack, ExpandMore, Code, Save, BookmarkBorder, Storage, Lock } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

/**
 * LocalScripts component - execute stored bash scripts locally
 */
const LocalScripts = () => {
  const navigate = useNavigate();
  const [scripts, setScripts] = useState([]);
  const [selectedScriptId, setSelectedScriptId] = useState('');
  const [selectedScript, setSelectedScript] = useState(null);
  const [user, setUser] = useState('root');
  const [envVars, setEnvVars] = useState([]);
  const [selectedEnvVarIds, setSelectedEnvVarIds] = useState([]);
  const [output, setOutput] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingScripts, setLoadingScripts] = useState(true);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
  const [sudoPassword, setSudoPassword] = useState('');
  const [availableUsers, setAvailableUsers] = useState(['current', 'root']);
  const [currentUsername, setCurrentUsername] = useState('current');
  
  // Preset state
  const [presets, setPresets] = useState([]);
  const [selectedPresetId, setSelectedPresetId] = useState('');
  const [saveDialogOpen, setSaveDialogOpen] = useState(false);
  const [presetName, setPresetName] = useState('');
  const [presetDescription, setPresetDescription] = useState('');
  const [savingPreset, setSavingPreset] = useState(false);

  // Fetch scripts, users, env vars, and presets on mount
  useEffect(() => {
    fetchScripts();
    fetchLocalUsers();
    fetchCurrentUser();
    fetchEnvVars();
    fetchAllPresets();
  }, []);

  const fetchScripts = async () => {
    try {
      setLoadingScripts(true);
      const response = await fetch('/api/bash-scripts');
      if (response.ok) {
        const data = await response.json();
        setScripts(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch scripts:', err);
    } finally {
      setLoadingScripts(false);
    }
  };

  const fetchEnvVars = async () => {
    try {
      const response = await fetch('/api/env-variables');
      if (response.ok) {
        const data = await response.json();
        setEnvVars(data || []);
        // Select all by default - use name for Vault items, ID for SQLite items
        setSelectedEnvVarIds((data || []).map(v => v.source === 'vault' ? v.name : v.id));
      }
    } catch (err) {
      console.error('Failed to fetch env vars:', err);
    }
  };

  const fetchCurrentUser = async () => {
    try {
      const response = await fetch('/api/system/current-user');
      if (response.ok) {
        const data = await response.json();
        setCurrentUsername(data.username || 'current');
      }
    } catch (err) {
      console.error('Failed to fetch current user:', err);
    }
  };

  const fetchLocalUsers = async () => {
    try {
      const response = await fetch('/api/local-users');
      if (response.ok) {
        const data = await response.json();
        const users = data || [];
        const userSet = new Set(['current', 'root']);
        users.forEach(u => {
          if (u.name && u.name !== 'current') {
            userSet.add(u.name);
          }
        });
        setAvailableUsers(Array.from(userSet));
      }
    } catch (err) {
      console.error('Failed to fetch local users:', err);
    }
  };

  const fetchAllPresets = async () => {
    try {
      const response = await fetch('/api/script-presets');
      if (response.ok) {
        const data = await response.json();
        // Filter to only local presets (is_remote === false)
        const localPresets = (data || []).filter(p => !p.is_remote);
        setPresets(localPresets);
      }
    } catch (err) {
      console.error('Failed to fetch presets:', err);
    }
  };

  const handleLoadPreset = async (presetId) => {
    setSelectedPresetId(presetId);
    if (!presetId) return;
    
    const preset = presets.find(p => p.id === parseInt(presetId, 10));
    if (preset) {
      // Also select the script if not already selected
      if (preset.script_id && String(preset.script_id) !== selectedScriptId) {
        setSelectedScriptId(String(preset.script_id));
        // Fetch script details
        try {
          const response = await fetch(`/api/bash-scripts/${preset.script_id}`);
          if (response.ok) {
            const data = await response.json();
            setSelectedScript(data);
          }
        } catch (err) {
          console.error('Failed to fetch script details:', err);
        }
      }
      setSelectedEnvVarIds(preset.env_var_ids || []);
      if (preset.user) {
        setUser(preset.user);
      }
      setSuccess(`Loaded preset: ${preset.name}`);
    }
  };

  const handleSavePresetOpen = () => {
    setPresetName('');
    setPresetDescription('');
    setSaveDialogOpen(true);
  };

  const handleSavePresetClose = () => {
    setSaveDialogOpen(false);
    setPresetName('');
    setPresetDescription('');
  };

  const handleSavePresetSubmit = async () => {
    if (!presetName.trim()) {
      setError('Preset name is required');
      return;
    }
    if (!selectedScriptId) {
      setError('Please select a script first');
      return;
    }

    setSavingPreset(true);
    try {
      const payload = {
        name: presetName.trim(),
        description: presetDescription.trim(),
        script_id: parseInt(selectedScriptId, 10),
        env_var_ids: selectedEnvVarIds,
        is_remote: false,
        user: user,
      };

      const response = await fetch('/api/script-presets', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to save preset');
      }

      const savedPreset = await response.json();
      setSuccess(`Preset "${savedPreset.name}" saved successfully`);
      handleSavePresetClose();
      // Refresh presets list
      fetchAllPresets();
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingPreset(false);
    }
  };

  // Fetch full script content when selected
  const handleScriptSelect = async (event) => {
    const scriptId = event.target.value;
    setSelectedScriptId(scriptId);
    setSelectedPresetId('');
    setOutput('');
    setError(null);
    setSuccess(null);

    if (scriptId) {
      try {
        const response = await fetch(`/api/bash-scripts/${scriptId}`);
        if (response.ok) {
          const data = await response.json();
          setSelectedScript(data);
        }
      } catch (err) {
        console.error('Failed to fetch script details:', err);
      }
    } else {
      setSelectedScript(null);
    }
  };

  // Execute script - check if password needed first
  const handleExecute = async () => {
    if (!selectedScriptId) {
      setError('Please select a script');
      return;
    }

    if (user === 'root') {
      setPasswordDialogOpen(true);
      return;
    }

    executeScript('');
  };

  // Actually execute the script with streaming output
  const executeScript = async (password) => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    setOutput('');

    // Find the selected script object to determine its source
    const selectedScriptObj = scripts.find(s =>
      (s.source === 'vault' ? s.name === selectedScriptId : s.id === parseInt(selectedScriptId, 10))
    );

    const payload = {
      user: user || 'root',
      is_remote: false,
    };

    // For script: use name for Vault items, ID for SQLite items
    if (selectedScriptObj) {
      if (selectedScriptObj.source === 'vault') {
        payload.script_name = selectedScriptObj.name;
        payload.script_group = selectedScriptObj.group || 'default';
      } else {
        payload.script_id = selectedScriptObj.id;
      }
    }

    // For env vars: separate SQLite IDs from Vault names (with groups)
    const sqliteEnvVarIds = [];
    const vaultEnvVarNames = [];
    const vaultEnvVarGroups = [];
    selectedEnvVarIds.forEach(id => {
      const envVar = envVars.find(v => v.id === id || v.name === id);
      if (envVar) {
        if (envVar.source === 'vault') {
          vaultEnvVarNames.push(envVar.name);
          vaultEnvVarGroups.push(envVar.group || 'default');
        } else if (envVar.id) {
          sqliteEnvVarIds.push(envVar.id);
        }
      }
    });
    if (sqliteEnvVarIds.length > 0) {
      payload.env_var_ids = sqliteEnvVarIds;
    }
    if (vaultEnvVarNames.length > 0) {
      payload.env_var_names = vaultEnvVarNames;
      payload.env_var_groups = vaultEnvVarGroups;
    }

    if (password) {
      payload.sudo_password = password;
    }

    try {
      const response = await fetch('/api/bash-scripts/execute/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to execute script');
      }

      // Read the SSE stream
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let streamedOutput = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // Process complete SSE messages
        const lines = buffer.split('\n');
        buffer = lines.pop() || ''; // Keep incomplete line in buffer

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6));
              
              if (data.type === 'output') {
                streamedOutput += data.data;
                setOutput(streamedOutput);
              } else if (data.type === 'status') {
                // Status messages can be shown as temporary info
                console.log('Status:', data.data);
              } else if (data.type === 'error') {
                setError(data.data);
              } else if (data.type === 'result') {
                // Final result received
                const result = data.result;
                if (result.exit_code === 0) {
                  let msg = `Script executed successfully in ${result.execution_time_ms}ms`;
                  if (result.env_vars_count > 0) {
                    msg += ` (${result.env_vars_count} env vars injected)`;
                  }
                  setSuccess(msg);
                } else {
                  setError(`Script exited with code ${result.exit_code}`);
                }
                // Use final output from result if we didn't stream anything
                if (!streamedOutput && result.output) {
                  setOutput(result.output);
                }
              }
            } catch (parseErr) {
              console.error('Failed to parse SSE message:', parseErr);
            }
          }
        }
      }

      // If no output was streamed, show a placeholder
      if (!streamedOutput) {
        setOutput(prev => prev || '(no output)');
      }
    } catch (err) {
      setError(err.message);
      setOutput('');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordSubmit = () => {
    setPasswordDialogOpen(false);
    executeScript(sudoPassword);
    setSudoPassword('');
  };

  const handlePasswordCancel = () => {
    setPasswordDialogOpen(false);
    setSudoPassword('');
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

        <Typography variant="h4" component="h1" sx={{ mb: 1 }}>
          Run Scripts Locally
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          Execute stored bash scripts on the local server with optional environment variables.
        </Typography>

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

        <Paper sx={{ p: 3, mb: 3 }}>
          <Grid container spacing={2}>
            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>Select Script</InputLabel>
                <Select
                  value={selectedScriptId}
                  onChange={handleScriptSelect}
                  label="Select Script"
                  disabled={loading || loadingScripts}
                >
                  <MenuItem value="">
                    <em>Choose a script...</em>
                  </MenuItem>
                  {scripts.map((script) => (
                    <MenuItem key={script.source === 'vault' ? `vault-${script.name}` : script.id} value={script.source === 'vault' ? script.name : script.id}>
                      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%', gap: 1 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Code fontSize="small" />
                          {script.name}
                          {script.filename && (
                            <Chip label={script.filename} size="small" variant="outlined" />
                          )}
                        </Box>
                        {script.source && (
                          script.source === 'vault' ? (
                            <Chip icon={<Lock fontSize="small" />} label="Vault" size="small" variant="outlined" color="secondary" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                          ) : (
                            <Chip icon={<Storage fontSize="small" />} label="Local" size="small" variant="outlined" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                          )
                        )}
                      </Box>
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>Run As User</InputLabel>
                <Select
                  value={user}
                  onChange={(e) => setUser(e.target.value)}
                  label="Run As User"
                  disabled={loading}
                >
                  {availableUsers.map((username) => (
                    <MenuItem key={username} value={username}>
                      {username === 'current' ? `current (${currentUsername})` : username}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>

            <Grid item xs={12}>
              <FormControl fullWidth>
                <InputLabel>Environment Variables</InputLabel>
                <Select
                  multiple
                  value={selectedEnvVarIds}
                  onChange={(e) => setSelectedEnvVarIds(e.target.value)}
                  label="Environment Variables"
                  disabled={loading}
                  renderValue={(selected) => (
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                      {selected.length === 0 ? (
                        <Typography color="text.secondary">None selected</Typography>
                      ) : selected.length === envVars.length ? (
                        <Chip size="small" label={`All (${envVars.length})`} />
                      ) : (
                        selected.map((val) => {
                          const envVar = envVars.find(v => (v.source === 'vault' ? v.name === val : v.id === val));
                          return envVar ? (
                            <Chip key={val} size="small" label={envVar.name} />
                          ) : null;
                        })
                      )}
                    </Box>
                  )}
                >
                  {envVars.map((envVar) => {
                    const envVarValue = envVar.source === 'vault' ? envVar.name : envVar.id;
                    return (
                      <MenuItem key={envVar.source === 'vault' ? `vault-${envVar.name}` : envVar.id} value={envVarValue}>
                        <Checkbox checked={selectedEnvVarIds.includes(envVarValue)} />
                        <Box sx={{ ml: 1, display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%', gap: 1 }}>
                          <Box>
                            <Typography>{envVar.name}</Typography>
                            {envVar.description && (
                              <Typography variant="caption" color="text.secondary">
                                {envVar.description}
                              </Typography>
                            )}
                          </Box>
                          {envVar.source && (
                            envVar.source === 'vault' ? (
                              <Chip icon={<Lock fontSize="small" />} label="Vault" size="small" variant="outlined" color="secondary" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                            ) : (
                              <Chip icon={<Storage fontSize="small" />} label="Local" size="small" variant="outlined" sx={{ fontWeight: 500, height: 20, '& .MuiChip-label': { px: 0.75 } }} />
                            )
                          )}
                        </Box>
                      </MenuItem>
                    );
                  })}
                </Select>
              </FormControl>
              <Box sx={{ mt: 1, display: 'flex', gap: 1 }}>
                <Button
                  size="small"
                  onClick={() => setSelectedEnvVarIds(envVars.map(v => v.source === 'vault' ? v.name : v.id))}
                  disabled={loading || selectedEnvVarIds.length === envVars.length}
                >
                  Select All
                </Button>
                <Button
                  size="small"
                  onClick={() => setSelectedEnvVarIds([])}
                  disabled={loading || selectedEnvVarIds.length === 0}
                >
                  Select None
                </Button>
              </Box>
            </Grid>

            {/* Preset Section - Always visible */}
            <Grid item xs={12}>
              <Box sx={{ display: 'flex', gap: 2, alignItems: 'flex-start' }}>
                <FormControl sx={{ minWidth: 200, flexGrow: 1 }}>
                  <InputLabel>Load Preset</InputLabel>
                  <Select
                    value={selectedPresetId}
                    onChange={(e) => handleLoadPreset(e.target.value)}
                    label="Load Preset"
                    disabled={loading || presets.length === 0}
                  >
                    <MenuItem value="">
                      <em>{presets.length === 0 ? 'No saved presets' : 'Select a preset...'}</em>
                    </MenuItem>
                    {presets.map((preset) => {
                      const scriptName = scripts.find(s => s.id === preset.script_id)?.name;
                      return (
                        <MenuItem key={preset.id} value={preset.id}>
                          <Box>
                            <Typography>{preset.name}</Typography>
                            <Typography variant="caption" color="text.secondary">
                              {scriptName ? `Script: ${scriptName}` : ''}
                              {scriptName && preset.description ? ' • ' : ''}
                              {preset.description || ''}
                            </Typography>
                          </Box>
                        </MenuItem>
                      );
                    })}
                  </Select>
                </FormControl>
                <Button
                  variant="outlined"
                  startIcon={<Save />}
                  onClick={handleSavePresetOpen}
                  disabled={loading || !selectedScriptId}
                  sx={{ height: 56 }}
                >
                  Save Preset
                </Button>
              </Box>
            </Grid>

            {selectedScript && (
              <Grid item xs={12}>
                <Accordion>
                  <AccordionSummary expandIcon={<ExpandMore />}>
                    <Typography>Script Preview: {selectedScript.name}</Typography>
                  </AccordionSummary>
                  <AccordionDetails>
                    {selectedScript.description && (
                      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                        {selectedScript.description}
                      </Typography>
                    )}
                    <Box
                      component="pre"
                      sx={{
                        backgroundColor: '#1a1a1a',
                        color: '#e0e0e0',
                        p: 2,
                        borderRadius: 1,
                        overflow: 'auto',
                        maxHeight: 300,
                        fontFamily: 'monospace',
                        fontSize: '0.875rem',
                        margin: 0,
                      }}
                    >
                      {selectedScript.content}
                    </Box>
                  </AccordionDetails>
                </Accordion>
              </Grid>
            )}

            <Grid item xs={12}>
              <Button
                variant="contained"
                size="large"
                startIcon={loading ? <CircularProgress size={20} /> : <PlayArrow />}
                onClick={handleExecute}
                disabled={loading || !selectedScriptId}
                fullWidth
              >
                {loading ? 'Executing...' : 'Execute Script'}
              </Button>
            </Grid>
          </Grid>
        </Paper>

        {output && (
          <Paper sx={{ p: 3, backgroundColor: '#0a0a0a', color: '#e0e0e0' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <Typography variant="h6" sx={{ flexGrow: 1 }}>
                Output
              </Typography>
              <Button
                size="small"
                onClick={() => navigator.clipboard.writeText(output)}
              >
                Copy
              </Button>
            </Box>
            <Box
              component="pre"
              sx={{
                fontFamily: 'monospace',
                fontSize: '0.875rem',
                margin: 0,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}
            >
              <Ansi useClasses>{output}</Ansi>
            </Box>
          </Paper>
        )}

        <Box sx={{ mt: 3, display: 'flex', gap: 2 }}>
          <Button
            variant="outlined"
            startIcon={<Code />}
            onClick={() => navigate('/scripts')}
          >
            Manage Scripts
          </Button>
          <Button
            variant="outlined"
            onClick={() => navigate('/history')}
          >
            View History
          </Button>
        </Box>
      </Box>

      {/* Sudo Password Dialog */}
      <Dialog open={passwordDialogOpen} onClose={handlePasswordCancel}>
        <DialogTitle>Sudo Password Required</DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ mb: 2 }}>
            This script requires root privileges. Please enter your sudo password:
          </Typography>
          <TextField
            autoFocus
            fullWidth
            type="password"
            label="Sudo Password"
            value={sudoPassword}
            onChange={(e) => setSudoPassword(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter') {
                handlePasswordSubmit();
              }
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handlePasswordCancel}>Cancel</Button>
          <Button onClick={handlePasswordSubmit} variant="contained">
            Execute
          </Button>
        </DialogActions>
      </Dialog>

      {/* Save Preset Dialog */}
      <Dialog open={saveDialogOpen} onClose={handleSavePresetClose} maxWidth="sm" fullWidth>
        <DialogTitle>Save Configuration as Preset</DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ mb: 2 }}>
            Save the current environment variable selection and user settings as a reusable preset.
          </Typography>
          <TextField
            autoFocus
            fullWidth
            label="Preset Name"
            value={presetName}
            onChange={(e) => setPresetName(e.target.value)}
            sx={{ mb: 2, mt: 1 }}
            placeholder="e.g., Production Deploy"
          />
          <TextField
            fullWidth
            label="Description (optional)"
            value={presetDescription}
            onChange={(e) => setPresetDescription(e.target.value)}
            multiline
            rows={2}
            placeholder="Brief description of this preset..."
          />
          <Box sx={{ mt: 2, p: 2, bgcolor: 'action.hover', borderRadius: 1 }}>
            <Typography variant="subtitle2" gutterBottom>
              This preset will save:
            </Typography>
            <Typography variant="body2" color="text.secondary">
              • User: {user === 'current' ? `current (${currentUsername})` : user}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              • Environment Variables: {selectedEnvVarIds.length} selected
            </Typography>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleSavePresetClose}>Cancel</Button>
          <Button 
            onClick={handleSavePresetSubmit} 
            variant="contained"
            disabled={savingPreset || !presetName.trim()}
            startIcon={savingPreset ? <CircularProgress size={16} /> : <Save />}
          >
            {savingPreset ? 'Saving...' : 'Save Preset'}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default LocalScripts;
