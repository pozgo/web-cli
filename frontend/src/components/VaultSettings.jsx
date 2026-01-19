import React, { useState, useEffect } from 'react';
import {
  Typography,
  Box,
  Button,
  Paper,
  TextField,
  Alert,
  CircularProgress,
  Switch,
  FormControlLabel,
  Chip,
  Divider,
  Card,
  CardContent,
  Grid,
  InputAdornment,
  IconButton,
  Tooltip,
  Collapse,
} from '@mui/material';
import {
  Save,
  Delete,
  CheckCircle,
  Error as ErrorIcon,
  Refresh,
  Visibility,
  VisibilityOff,
  Info,
  Storage,
  VpnKey,
  Dns,
  Code,
  Settings,
} from '@mui/icons-material';

/**
 * VaultSettings component - HashiCorp Vault integration configuration
 */
const VaultSettings = () => {
  const [config, setConfig] = useState({
    address: '',
    token: '',
    namespace: '',
    mount_path: 'secret',
    enabled: false,
  });
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [showToken, setShowToken] = useState(false);
  const [hasExistingToken, setHasExistingToken] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [vaultData, setVaultData] = useState({
    sshKeys: [],
    servers: [],
    envVars: [],
    scripts: [],
  });
  const [loadingVaultData, setLoadingVaultData] = useState(false);

  // Fetch current Vault configuration
  const fetchConfig = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch('/api/vault/config');

      if (!response.ok) {
        // 404 is expected when no config exists yet
        if (response.status === 404) {
          setLoading(false);
          return;
        }
        throw new Error('Failed to fetch Vault configuration');
      }

      const text = await response.text();
      if (!text) {
        setLoading(false);
        return;
      }

      const data = JSON.parse(text);
      if (data && data.id) {
        setConfig({
          address: data.address || '',
          token: '', // Token is never returned
          namespace: data.namespace || '',
          mount_path: data.mount_path || 'secret',
          enabled: data.enabled || false,
        });
        setHasExistingToken(data.has_token || false);
      }
    } catch (err) {
      console.error('VaultSettings fetchConfig error:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Fetch Vault connection status
  const fetchStatus = async () => {
    try {
      const response = await fetch('/api/vault/status');
      if (response.ok) {
        const text = await response.text();
        if (text) {
          const data = JSON.parse(text);
          setStatus(data);
        }
      }
    } catch (err) {
      console.error('Failed to fetch Vault status:', err);
    }
  };

  // Fetch Vault data (keys, servers, etc.)
  const fetchVaultData = async () => {
    if (!status?.connected) return;

    setLoadingVaultData(true);
    try {
      const [keysRes, serversRes, envRes, scriptsRes] = await Promise.all([
        fetch('/api/vault/ssh-keys').catch(() => ({ ok: false })),
        fetch('/api/vault/servers').catch(() => ({ ok: false })),
        fetch('/api/vault/env-variables').catch(() => ({ ok: false })),
        fetch('/api/vault/scripts').catch(() => ({ ok: false })),
      ]);

      // Parse responses, ensuring we always get arrays (API might return null)
      const parseResponse = async (res) => {
        if (!res.ok) return [];
        try {
          const data = await res.json();
          return Array.isArray(data) ? data : [];
        } catch {
          return [];
        }
      };

      const sshKeys = await parseResponse(keysRes);
      const servers = await parseResponse(serversRes);
      const envVars = await parseResponse(envRes);
      const scripts = await parseResponse(scriptsRes);

      setVaultData({ sshKeys, servers, envVars, scripts });
    } catch (err) {
      console.error('Failed to fetch Vault data:', err);
    } finally {
      setLoadingVaultData(false);
    }
  };

  // Load config and status on mount
  useEffect(() => {
    fetchConfig();
    fetchStatus();
  }, []);

  // Fetch Vault data when connected
  useEffect(() => {
    if (status?.connected) {
      fetchVaultData();
    }
  }, [status?.connected]);

  // Handle form input changes
  const handleChange = (field) => (event) => {
    const value = field === 'enabled' ? event.target.checked : event.target.value;
    setConfig((prev) => ({ ...prev, [field]: value }));
  };

  // Save Vault configuration
  const handleSave = async () => {
    try {
      setSaving(true);
      setError(null);
      setSuccess(null);

      // Validate
      if (!config.address) {
        throw new Error('Vault address is required');
      }

      if (!hasExistingToken && !config.token) {
        throw new Error('Vault token is required');
      }

      const response = await fetch('/api/vault/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error(text || 'Failed to save configuration');
      }

      setSuccess('Vault configuration saved successfully');
      setHasExistingToken(true);
      setConfig((prev) => ({ ...prev, token: '' })); // Clear token from form
      fetchStatus();
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  // Test Vault connection
  const handleTestConnection = async () => {
    try {
      setTesting(true);
      setError(null);
      setSuccess(null);

      const response = await fetch('/api/vault/test', {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error('Failed to test connection');
      }

      const result = await response.json();
      setStatus(result);

      if (result.connected) {
        setSuccess('Connection successful!');
        fetchVaultData();
      } else {
        setError(result.error || 'Connection failed');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setTesting(false);
    }
  };

  // Delete Vault configuration
  const handleDelete = async () => {
    if (!window.confirm('Are you sure you want to remove the Vault configuration?')) {
      return;
    }

    try {
      setError(null);
      setSuccess(null);

      const response = await fetch('/api/vault/config', {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to delete configuration');
      }

      setConfig({
        address: '',
        token: '',
        namespace: '',
        mount_path: 'secret',
        enabled: false,
      });
      setHasExistingToken(false);
      setStatus(null);
      setVaultData({ sshKeys: [], servers: [], envVars: [], scripts: [] });
      setSuccess('Vault configuration removed');
    } catch (err) {
      setError(err.message);
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h5" component="h2" sx={{ mb: 3 }}>
        HashiCorp Vault Integration
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

      {/* Connection Status */}
      {status && (
        <Paper sx={{ p: 2, mb: 3 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Typography variant="subtitle1">Status:</Typography>
            {status.connected ? (
              <Chip
                icon={<CheckCircle />}
                label="Connected"
                color="success"
                variant="outlined"
              />
            ) : status.configured ? (
              <Chip
                icon={<ErrorIcon />}
                label={status.enabled ? 'Disconnected' : 'Disabled'}
                color={status.enabled ? 'error' : 'default'}
                variant="outlined"
              />
            ) : (
              <Chip label="Not Configured" variant="outlined" />
            )}
            {status.vault_sealed && (
              <Chip
                label="Vault Sealed"
                color="warning"
                size="small"
              />
            )}
            <IconButton
              size="small"
              onClick={fetchStatus}
              title="Refresh status"
            >
              <Refresh />
            </IconButton>
          </Box>
          {status.error && (
            <Typography variant="body2" color="error" sx={{ mt: 1 }}>
              {status.error}
            </Typography>
          )}
        </Paper>
      )}

      {/* Configuration Form */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          <TextField
            label="Vault Address"
            value={config.address}
            onChange={handleChange('address')}
            placeholder="https://vault.example.com:8200"
            fullWidth
            required
            helperText="The URL of your HashiCorp Vault server"
          />

          <TextField
            label="Vault Token"
            type={showToken ? 'text' : 'password'}
            value={config.token}
            onChange={handleChange('token')}
            placeholder={hasExistingToken ? '(token configured - enter new token to change)' : 'hvs.xxxxx'}
            fullWidth
            required={!hasExistingToken}
            helperText={hasExistingToken ? 'Leave blank to keep existing token' : 'Your Vault authentication token'}
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    onClick={() => setShowToken(!showToken)}
                    edge="end"
                  >
                    {showToken ? <VisibilityOff /> : <Visibility />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
          />

          <Button
            variant="text"
            size="small"
            startIcon={<Settings />}
            onClick={() => setShowAdvanced(!showAdvanced)}
            sx={{ alignSelf: 'flex-start' }}
          >
            {showAdvanced ? 'Hide Advanced Settings' : 'Show Advanced Settings'}
          </Button>

          <Collapse in={showAdvanced}>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
              <TextField
                label="Namespace (optional)"
                value={config.namespace}
                onChange={handleChange('namespace')}
                placeholder="admin/my-namespace"
                fullWidth
                helperText="Vault Enterprise namespace (leave empty if not using namespaces)"
              />

              <TextField
                label="Mount Path"
                value={config.mount_path}
                onChange={handleChange('mount_path')}
                placeholder="secret"
                fullWidth
                helperText="KV v2 secrets engine mount path (default: secret)"
              />
            </Box>
          </Collapse>

          <FormControlLabel
            control={
              <Switch
                checked={config.enabled}
                onChange={handleChange('enabled')}
                color="primary"
              />
            }
            label="Enable Vault Integration"
          />

          <Box sx={{ display: 'flex', gap: 2 }}>
            <Button
              variant="contained"
              startIcon={<Save />}
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? 'Saving...' : 'Save Configuration'}
            </Button>

            <Button
              variant="outlined"
              startIcon={testing ? <CircularProgress size={20} /> : <Refresh />}
              onClick={handleTestConnection}
              disabled={testing || !config.address}
            >
              {testing ? 'Testing...' : 'Test Connection'}
            </Button>

            {hasExistingToken && (
              <Button
                variant="outlined"
                color="error"
                startIcon={<Delete />}
                onClick={handleDelete}
              >
                Remove
              </Button>
            )}
          </Box>
        </Box>
      </Paper>

      {/* Vault Data Overview */}
      {status?.connected && (
        <Paper sx={{ p: 3 }}>
          <Typography variant="h6" sx={{ mb: 2 }}>
            Vault Secrets Overview
          </Typography>

          {loadingVaultData ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
              <CircularProgress size={24} />
            </Box>
          ) : (
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6} md={3}>
                <Card variant="outlined">
                  <CardContent>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <VpnKey color="primary" />
                      <Typography variant="h6">
                        {vaultData.sshKeys.length}
                      </Typography>
                    </Box>
                    <Typography variant="body2" color="text.secondary">
                      SSH Keys
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>

              <Grid item xs={12} sm={6} md={3}>
                <Card variant="outlined">
                  <CardContent>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Dns color="primary" />
                      <Typography variant="h6">
                        {vaultData.servers.length}
                      </Typography>
                    </Box>
                    <Typography variant="body2" color="text.secondary">
                      Servers
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>

              <Grid item xs={12} sm={6} md={3}>
                <Card variant="outlined">
                  <CardContent>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Storage color="primary" />
                      <Typography variant="h6">
                        {vaultData.envVars.length}
                      </Typography>
                    </Box>
                    <Typography variant="body2" color="text.secondary">
                      Env Variables
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>

              <Grid item xs={12} sm={6} md={3}>
                <Card variant="outlined">
                  <CardContent>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Code color="primary" />
                      <Typography variant="h6">
                        {vaultData.scripts.length}
                      </Typography>
                    </Box>
                    <Typography variant="body2" color="text.secondary">
                      Scripts
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>
          )}

          <Divider sx={{ my: 2 }} />

          <Alert severity="info" icon={<Info />}>
            <Typography variant="body2">
              Secrets stored in Vault will appear alongside your local SQLite data.
              Look for the <Chip label="vault" size="small" sx={{ mx: 0.5 }} /> badge
              in SSH Keys, Servers, and Environment Variables lists.
            </Typography>
          </Alert>
        </Paper>
      )}

      {/* Setup Instructions */}
      {!status?.configured && (
        <Paper sx={{ p: 3, mt: 3 }}>
          <Typography variant="h6" sx={{ mb: 2 }}>
            Getting Started with Vault
          </Typography>
          <Typography variant="body2" paragraph>
            To use HashiCorp Vault integration, you need:
          </Typography>
          <Typography component="div" variant="body2">
            <ol>
              <li>A running HashiCorp Vault server (v1.0+)</li>
              <li>KV v2 secrets engine enabled</li>
              <li>A token with read/list permissions for <code>secret/data/web-cli/*</code></li>
            </ol>
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
            See the documentation for detailed setup instructions and Vault policy examples.
          </Typography>
        </Paper>
      )}
    </Box>
  );
};

export default VaultSettings;
