import React from 'react';
import {
  Container,
  Grid,
  Card,
  CardContent,
  Typography,
  Button,
  Box,
} from '@mui/material';
import { ArrowBack, Code, DataObject } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

/**
 * ToolsPage - Hub page displaying available developer tools as cards
 * Follows the same pattern as Dashboard for consistency
 */
const ToolsPage = () => {
  const navigate = useNavigate();

  const tools = [
    {
      icon: <Code sx={{ fontSize: 48 }} />,
      title: 'YAML Validator',
      description: 'Validate, format, and auto-fix YAML documents with syntax highlighting',
      action: 'Open YAML Tool',
      onClick: () => navigate('/tools/yaml'),
    },
    {
      icon: <DataObject sx={{ fontSize: 48 }} />,
      title: 'JSON Validator',
      description: 'Validate, format, and beautify JSON data with error detection',
      action: 'Open JSON Tool',
      onClick: () => navigate('/tools/json'),
    },
  ];

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button
          startIcon={<ArrowBack />}
          onClick={() => navigate('/')}
          sx={{ mb: 2 }}
        >
          Back to Dashboard
        </Button>
      </Box>

      <Box sx={{ mb: 4, textAlign: 'center' }}>
        <Typography variant="h3" component="h1" gutterBottom>
          Developer Tools
        </Typography>
        <Typography variant="body1" color="text.secondary" paragraph>
          Validate, format, and fix common data formats
        </Typography>
      </Box>

      <Grid container spacing={3}>
        {tools.map((tool, index) => (
          <Grid item xs={12} sm={6} md={6} key={index}>
            <Card
              sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                transition: 'transform 0.2s, box-shadow 0.2s',
                '&:hover': {
                  transform: 'translateY(-4px)',
                  boxShadow: 6,
                },
              }}
            >
              <CardContent sx={{ flexGrow: 1, textAlign: 'center', p: 3 }}>
                <Box sx={{ color: 'primary.main', mb: 2 }}>
                  {tool.icon}
                </Box>
                <Typography variant="h5" component="h2" gutterBottom>
                  {tool.title}
                </Typography>
                <Typography variant="body2" color="text.secondary" paragraph>
                  {tool.description}
                </Typography>
                <Button
                  variant="contained"
                  color="primary"
                  sx={{ mt: 2 }}
                  onClick={tool.onClick}
                >
                  {tool.action}
                </Button>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default ToolsPage;
