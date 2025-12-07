import React from 'react';
import { Container, Button, Box } from '@mui/material';
import { ArrowBack } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import ScriptList from './ScriptList';

/**
 * ScriptsPage component - dedicated page for managing bash scripts
 */
const ScriptsPage = () => {
  const navigate = useNavigate();

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 2 }}>
        <Button
          startIcon={<ArrowBack />}
          onClick={() => navigate('/')}
        >
          Back to Dashboard
        </Button>
      </Box>
      <ScriptList />
    </Container>
  );
};

export default ScriptsPage;
