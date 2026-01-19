import React, { useState, useEffect, useCallback } from 'react';
import { Box, Tooltip, Typography, Chip } from '@mui/material';
import { useNavigate } from 'react-router-dom';

/**
 * HashiCorp Vault Logo SVG
 * Simplified filled vault logo for better visibility
 */
const VaultLogo = ({ color = 'currentColor', size = 20 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Outer hexagon outline */}
    <path
      d="M12 2L3 7v10l9 5 9-5V7l-9-5z"
      fill={color}
      opacity="0.2"
    />
    {/* Middle hexagon */}
    <path
      d="M12 5L5 9v6l7 4 7-4V9l-7-4z"
      fill={color}
      opacity="0.4"
    />
    {/* Inner filled hexagon */}
    <path
      d="M12 8L8 10.5v3L12 16l4-2.5v-3L12 8z"
      fill={color}
    />
  </svg>
);

/**
 * VaultIcon component - displays Vault connection status in header
 * Only visible when Vault is configured in admin panel
 * Shows green when connected, red when disconnected, orange when sealed
 */
const VaultIcon = ({ sx = {} }) => {
  const navigate = useNavigate();
  const [vaultStatus, setVaultStatus] = useState(null);

  // Fetch Vault status
  const fetchVaultStatus = useCallback(async () => {
    try {
      const response = await fetch('/api/vault/status');
      if (response.ok) {
        const data = await response.json();
        setVaultStatus(data);
      } else {
        setVaultStatus(null);
      }
    } catch (err) {
      // Silently fail - Vault might not be configured
      console.debug('Failed to fetch vault status:', err);
      setVaultStatus(null);
    }
  }, []);

  // Initial fetch and periodic refresh
  useEffect(() => {
    fetchVaultStatus();
    // Refresh every 30 seconds
    const interval = setInterval(fetchVaultStatus, 30000);
    return () => clearInterval(interval);
  }, [fetchVaultStatus]);

  // Don't render if Vault is not configured
  if (!vaultStatus || !vaultStatus.configured) {
    return null;
  }

  // Determine status color, text, and tooltip
  const getStatusProps = () => {
    if (!vaultStatus.enabled) {
      return {
        color: '#9e9e9e', // grey
        statusText: 'Disabled',
        tooltip: 'Vault integration is disabled',
      };
    }

    if (vaultStatus.vault_sealed) {
      return {
        color: '#ff9800', // orange/warning
        statusText: 'Sealed',
        tooltip: 'Vault is sealed - secrets are inaccessible',
      };
    }

    if (vaultStatus.connected) {
      return {
        color: '#4caf50', // green/success
        statusText: 'Connected',
        tooltip: `Vault connected: ${vaultStatus.address || 'HashiCorp Vault'}`,
      };
    }

    return {
      color: '#f44336', // red/error
      statusText: 'Disconnected',
      tooltip: vaultStatus.error || 'Unable to connect to Vault',
    };
  };

  const { color, statusText, tooltip } = getStatusProps();

  return (
    <Tooltip title={tooltip} arrow>
      <Chip
        icon={<VaultLogo color={color} size={20} />}
        label={statusText}
        size="small"
        onClick={() => navigate('/admin', { state: { tab: 4 } })}
        sx={{
          cursor: 'pointer',
          backgroundColor: 'transparent',
          border: `1px solid ${color}`,
          color: color,
          fontWeight: 500,
          fontSize: '0.75rem',
          height: 28,
          transition: 'all 0.2s ease',
          '& .MuiChip-icon': {
            color: color,
            marginLeft: '8px',
          },
          '&:hover': {
            backgroundColor: `${color}20`,
            transform: 'scale(1.02)',
          },
          ...sx,
        }}
      />
    </Tooltip>
  );
};

export default VaultIcon;
