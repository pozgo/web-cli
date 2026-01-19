import React, { useState, useEffect } from 'react';
import {
  FormControl,
  FormLabel,
  RadioGroup,
  FormControlLabel,
  Radio,
  Box,
  Chip,
  Typography,
} from '@mui/material';
import { Storage, Cloud } from '@mui/icons-material';

/**
 * Vault Logo SVG for the selector
 */
const VaultLogo = ({ size = 16 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      d="M12 2L3 7v10l9 5 9-5V7l-9-5z"
      fill="currentColor"
      opacity="0.3"
    />
    <path
      d="M12 5L5 9v6l7 4 7-4V9l-7-4z"
      fill="currentColor"
      opacity="0.5"
    />
    <path
      d="M12 8L8 10.5v3L12 16l4-2.5v-3L12 8z"
      fill="currentColor"
    />
  </svg>
);

/**
 * StorageSelector component - allows choosing between local SQLite and Vault storage
 * Only shows Vault option when Vault is configured and connected
 *
 * @param {string} value - Current storage value ('local' or 'vault')
 * @param {function} onChange - Callback when storage changes
 * @param {boolean} disabled - Whether the selector is disabled
 */
const StorageSelector = ({ value = 'local', onChange, disabled = false }) => {
  const [vaultStatus, setVaultStatus] = useState(null);
  const [loading, setLoading] = useState(true);

  // Fetch Vault status to determine if Vault option should be available
  useEffect(() => {
    const fetchVaultStatus = async () => {
      try {
        const response = await fetch('/api/vault/status');
        if (response.ok) {
          const data = await response.json();
          setVaultStatus(data);
        } else {
          setVaultStatus(null);
        }
      } catch (err) {
        console.debug('Failed to fetch vault status:', err);
        setVaultStatus(null);
      } finally {
        setLoading(false);
      }
    };

    fetchVaultStatus();
  }, []);

  // Check if Vault is available (configured, enabled, and connected)
  const isVaultAvailable = vaultStatus?.configured && vaultStatus?.enabled && vaultStatus?.connected;

  // If Vault is not available and current value is 'vault', reset to 'local'
  useEffect(() => {
    if (!loading && !isVaultAvailable && value === 'vault') {
      onChange('local');
    }
  }, [loading, isVaultAvailable, value, onChange]);

  // Don't show selector if Vault is not available - just use local storage
  if (loading) {
    return null;
  }

  if (!isVaultAvailable) {
    return null;
  }

  return (
    <FormControl component="fieldset" disabled={disabled} sx={{ width: '100%' }}>
      <FormLabel component="legend" sx={{ mb: 1, fontSize: '0.875rem' }}>
        Storage Location
      </FormLabel>
      <RadioGroup
        row
        value={value}
        onChange={(e) => onChange(e.target.value)}
        sx={{ gap: 2 }}
      >
        <FormControlLabel
          value="local"
          control={<Radio size="small" />}
          label={
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Storage fontSize="small" />
              <Typography variant="body2">Local Database</Typography>
            </Box>
          }
          sx={{
            border: '1px solid',
            borderColor: value === 'local' ? 'primary.main' : 'divider',
            borderRadius: 1,
            px: 2,
            py: 0.5,
            m: 0,
            backgroundColor: value === 'local' ? 'action.selected' : 'transparent',
            transition: 'all 0.2s ease',
            '&:hover': {
              borderColor: 'primary.main',
              backgroundColor: 'action.hover',
            },
          }}
        />
        <FormControlLabel
          value="vault"
          control={<Radio size="small" />}
          label={
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <VaultLogo size={18} />
              <Typography variant="body2">HashiCorp Vault</Typography>
              <Chip
                label="Connected"
                size="small"
                color="success"
                sx={{ height: 20, fontSize: '0.7rem' }}
              />
            </Box>
          }
          sx={{
            border: '1px solid',
            borderColor: value === 'vault' ? 'primary.main' : 'divider',
            borderRadius: 1,
            px: 2,
            py: 0.5,
            m: 0,
            backgroundColor: value === 'vault' ? 'action.selected' : 'transparent',
            transition: 'all 0.2s ease',
            '&:hover': {
              borderColor: 'primary.main',
              backgroundColor: 'action.hover',
            },
          }}
        />
      </RadioGroup>
    </FormControl>
  );
};

export default StorageSelector;
