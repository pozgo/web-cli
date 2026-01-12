import React, { memo } from 'react';
import PropTypes from 'prop-types';
import { Container, Box, Typography, Button } from '@mui/material';
import { ArrowBack } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

/**
 * ToolLayout - Common page layout wrapper for tool pages
 * Provides consistent styling with back button, title, and description
 */
const ToolLayout = memo(({ title, description, backPath = '/tools', children }) => {
  const navigate = useNavigate();

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button
          startIcon={<ArrowBack />}
          onClick={() => navigate(backPath)}
          sx={{ mb: 2 }}
          aria-label="Navigate back to tools page"
        >
          Back to Tools
        </Button>
        <Typography variant="h4" component="h1" gutterBottom>
          {title}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {description}
        </Typography>
      </Box>
      <Box sx={{ mt: 3 }}>
        {children}
      </Box>
    </Container>
  );
});

ToolLayout.displayName = 'ToolLayout';

ToolLayout.propTypes = {
  title: PropTypes.string.isRequired,
  description: PropTypes.string,
  backPath: PropTypes.string,
  children: PropTypes.node.isRequired,
};

export default ToolLayout;
