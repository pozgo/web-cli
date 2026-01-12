import React, { memo } from 'react';
import PropTypes from 'prop-types';
import { Alert, Typography, List, ListItem, ListItemIcon, ListItemText } from '@mui/material';
import { CheckCircle, Error as ErrorIcon } from '@mui/icons-material';

/**
 * ValidationResult - Displays validation success or error messages
 * Shows line numbers and descriptions for each error
 */
const ValidationResult = memo(({ isValid, errors = [], successMessage = 'Valid!' }) => {
  // Not validated yet
  if (isValid === null) {
    return null;
  }

  // Valid state
  if (isValid) {
    return (
      <Alert 
        severity="success" 
        icon={<CheckCircle />}
        sx={{ mb: 2 }}
      >
        {successMessage}
      </Alert>
    );
  }

  // Error state
  return (
    <Alert 
      severity="error" 
      icon={<ErrorIcon />}
      sx={{ mb: 2 }}
    >
      <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 1 }}>
        Validation Failed
      </Typography>
      {errors.length > 0 && (
        <List dense sx={{ py: 0 }}>
          {errors.map((error, index) => (
            <ListItem key={index} sx={{ py: 0.5, px: 0 }}>
              <ListItemIcon sx={{ minWidth: 28 }}>
                <ErrorIcon fontSize="small" color="error" />
              </ListItemIcon>
              <ListItemText
                primary={
                  <Typography variant="body2" component="span">
                    {error.line && error.column 
                      ? `Line ${error.line}, Column ${error.column}: `
                      : error.line 
                        ? `Line ${error.line}: `
                        : ''
                    }
                    {error.message}
                  </Typography>
                }
                secondary={error.suggestion && (
                  <Typography variant="caption" color="text.secondary">
                    Suggestion: {error.suggestion}
                  </Typography>
                )}
              />
            </ListItem>
          ))}
        </List>
      )}
    </Alert>
  );
});

ValidationResult.displayName = 'ValidationResult';

ValidationResult.propTypes = {
  isValid: PropTypes.bool,
  errors: PropTypes.arrayOf(
    PropTypes.shape({
      line: PropTypes.number,
      column: PropTypes.number,
      message: PropTypes.string.isRequired,
      suggestion: PropTypes.string,
    })
  ),
  successMessage: PropTypes.string,
};

export default ValidationResult;
