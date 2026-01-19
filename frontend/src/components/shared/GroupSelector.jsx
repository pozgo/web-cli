import React, { useState, useEffect } from 'react';
import {
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  Box,
  CircularProgress,
  Typography,
} from '@mui/material';
import { Folder, FolderOpen } from '@mui/icons-material';

/**
 * GroupSelector component - allows filtering resources by group
 * @param {Object} props
 * @param {string} props.resourceType - The resource type ('servers', 'keys', 'env-variables', 'bash-scripts')
 * @param {string} props.selectedGroup - Currently selected group ('all' for all groups)
 * @param {function} props.onGroupChange - Callback when group selection changes
 * @param {string} props.size - Select size ('small' or 'medium'), default 'small'
 * @param {boolean} props.showCount - Show count of groups, default false
 */
const GroupSelector = ({
  resourceType,
  selectedGroup,
  onGroupChange,
  size = 'small',
  showCount = false,
}) => {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Fetch groups from API
  const fetchGroups = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await fetch(`/api/${resourceType}/groups`);

      if (!response.ok) {
        throw new Error('Failed to fetch groups');
      }

      const data = await response.json();
      setGroups(data || []);
    } catch (err) {
      setError(err.message);
      setGroups([]);
    } finally {
      setLoading(false);
    }
  };

  // Load groups on component mount or when resourceType changes
  useEffect(() => {
    fetchGroups();
  }, [resourceType]);

  // Handle selection change
  const handleChange = (event) => {
    onGroupChange(event.target.value);
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', alignItems: 'center', minWidth: 150 }}>
        <CircularProgress size={20} sx={{ mr: 1 }} />
        <Typography variant="body2" color="text.secondary">
          Loading groups...
        </Typography>
      </Box>
    );
  }

  if (error) {
    return (
      <Typography variant="body2" color="error">
        Failed to load groups
      </Typography>
    );
  }

  // If no groups found, don't render the selector
  if (groups.length === 0) {
    return null;
  }

  return (
    <FormControl size={size} sx={{ minWidth: 180 }}>
      <InputLabel id="group-selector-label">
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Folder fontSize="small" />
          Group
        </Box>
      </InputLabel>
      <Select
        labelId="group-selector-label"
        id="group-selector"
        value={selectedGroup}
        label="Group"
        onChange={handleChange}
        renderValue={(selected) => (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {selected === 'all' ? (
              <FolderOpen fontSize="small" color="action" />
            ) : (
              <Folder fontSize="small" color="primary" />
            )}
            <span>{selected === 'all' ? 'All Groups' : selected}</span>
          </Box>
        )}
      >
        <MenuItem value="all">
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <FolderOpen fontSize="small" color="action" />
            <span>All Groups</span>
            {showCount && (
              <Chip
                label={groups.length}
                size="small"
                sx={{ ml: 'auto', height: 20, fontSize: '0.75rem' }}
              />
            )}
          </Box>
        </MenuItem>
        {groups.map((group) => (
          <MenuItem key={group} value={group}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Folder fontSize="small" color="primary" />
              <span>{group}</span>
            </Box>
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
};

export default GroupSelector;
