import React from 'react';
import { Chip, Tooltip } from '@mui/material';
import { Storage, Lock } from '@mui/icons-material';

/**
 * SourceChip component - displays the data source (sqlite or vault)
 * @param {Object} props
 * @param {string} props.source - The data source ('sqlite' or 'vault')
 * @param {string} props.size - Chip size ('small' or 'medium'), default 'small'
 */
const SourceChip = ({ source, size = 'small' }) => {
  if (!source) return null;

  const isVault = source === 'vault';

  return (
    <Tooltip title={isVault ? 'Stored in HashiCorp Vault' : 'Stored in local database'}>
      <Chip
        icon={isVault ? <Lock fontSize="small" /> : <Storage fontSize="small" />}
        label={isVault ? 'Vault' : 'Local'}
        size={size}
        variant="outlined"
        color={isVault ? 'secondary' : 'default'}
        sx={{
          fontWeight: 500,
          '& .MuiChip-icon': {
            fontSize: '1rem',
          },
        }}
      />
    </Tooltip>
  );
};

export default SourceChip;
