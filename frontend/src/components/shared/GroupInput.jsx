import React, { useState, useEffect } from 'react';
import {
  Autocomplete,
  TextField,
  CircularProgress,
  Box,
  Chip,
  createFilterOptions,
} from '@mui/material';
import { Folder, Add } from '@mui/icons-material';

const filter = createFilterOptions();

/**
 * GroupInput component - autocomplete input for selecting or creating groups
 * Allows selecting from existing groups or typing a new group name
 *
 * @param {Object} props
 * @param {string} props.value - Current group value
 * @param {function} props.onChange - Callback when group changes
 * @param {string} props.resourceType - The resource type to fetch groups from ('servers', 'keys', 'env-variables', 'bash-scripts')
 * @param {boolean} props.disabled - Whether the input is disabled
 * @param {string} props.helperText - Helper text to display
 * @param {string} props.label - Label for the input
 */
const GroupInput = ({
  value = 'default',
  onChange,
  resourceType = 'servers',
  disabled = false,
  helperText = 'Select an existing group or type a new one',
  label = 'Group',
}) => {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [inputValue, setInputValue] = useState('');

  // Fetch groups from all resource types to get a comprehensive list
  const fetchAllGroups = async () => {
    try {
      setLoading(true);
      const resourceTypes = ['servers', 'keys', 'env-variables', 'bash-scripts'];

      const results = await Promise.all(
        resourceTypes.map(async (type) => {
          try {
            const response = await fetch(`/api/${type}/groups`);
            if (response.ok) {
              return await response.json();
            }
            return [];
          } catch {
            return [];
          }
        })
      );

      // Combine and deduplicate groups from all resource types
      const allGroups = [...new Set(results.flat())].filter(Boolean).sort();

      // Ensure 'default' is always in the list
      if (!allGroups.includes('default')) {
        allGroups.unshift('default');
      }

      setGroups(allGroups);
    } catch (err) {
      console.error('Failed to fetch groups:', err);
      setGroups(['default']);
    } finally {
      setLoading(false);
    }
  };

  // Load groups on component mount
  useEffect(() => {
    fetchAllGroups();
  }, []);

  // Handle selection change
  const handleChange = (event, newValue) => {
    if (typeof newValue === 'string') {
      // User typed and pressed enter
      onChange(newValue);
    } else if (newValue && newValue.inputValue) {
      // User selected "Add new group" option
      onChange(newValue.inputValue);
    } else if (newValue) {
      // User selected an existing option
      onChange(newValue);
    } else {
      // Cleared
      onChange('default');
    }
  };

  return (
    <Autocomplete
      value={value}
      onChange={handleChange}
      inputValue={inputValue}
      onInputChange={(event, newInputValue) => {
        setInputValue(newInputValue);
      }}
      filterOptions={(options, params) => {
        const filtered = filter(options, params);

        const { inputValue } = params;
        // Suggest creating a new group if the input doesn't match any existing group
        const isExisting = options.some(
          (option) => inputValue.toLowerCase() === option.toLowerCase()
        );
        if (inputValue !== '' && !isExisting) {
          filtered.push({
            inputValue,
            title: `Add "${inputValue}"`,
          });
        }

        return filtered;
      }}
      selectOnFocus
      clearOnBlur
      handleHomeEndKeys
      freeSolo
      options={groups}
      getOptionLabel={(option) => {
        // Value selected with enter from the input
        if (typeof option === 'string') {
          return option;
        }
        // Add new option created dynamically
        if (option.inputValue) {
          return option.inputValue;
        }
        return option;
      }}
      renderOption={(props, option) => {
        const { key, ...otherProps } = props;
        if (typeof option === 'object' && option.inputValue) {
          // "Add new group" option
          return (
            <Box
              component="li"
              key={key}
              {...otherProps}
              sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <Add fontSize="small" color="primary" />
              <span>{option.title}</span>
            </Box>
          );
        }
        // Existing group option
        return (
          <Box
            component="li"
            key={key}
            {...otherProps}
            sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
          >
            <Folder fontSize="small" color="primary" />
            <span>{option}</span>
            {option === 'default' && (
              <Chip
                label="Default"
                size="small"
                sx={{ ml: 'auto', height: 20, fontSize: '0.7rem' }}
              />
            )}
          </Box>
        );
      }}
      loading={loading}
      disabled={disabled}
      renderInput={(params) => (
        <TextField
          {...params}
          label={label}
          placeholder="Select or type a group name"
          helperText={helperText}
          margin="dense"
          InputProps={{
            ...params.InputProps,
            endAdornment: (
              <>
                {loading ? <CircularProgress color="inherit" size={20} /> : null}
                {params.InputProps.endAdornment}
              </>
            ),
          }}
        />
      )}
    />
  );
};

export default GroupInput;
