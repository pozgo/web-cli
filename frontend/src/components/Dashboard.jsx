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
import {
  Terminal,
  Cloud,
  Settings,
  Storage,
  Code,
  PlayArrow,
  CloudQueue,
  Computer,
  Build,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

/**
 * Dashboard component - main landing page with feature cards
 */
const Dashboard = () => {
  const navigate = useNavigate();

  const features = [
    // Row 1: Commands
    {
      icon: <Terminal sx={{ fontSize: 48 }} />,
      title: 'Local Commands',
      description: 'Execute commands on your local Linux server',
      action: 'Run Local',
      onClick: () => navigate('/local-commands'),
    },
    {
      icon: <Cloud sx={{ fontSize: 48 }} />,
      title: 'Remote Commands',
      description: 'Connect and run commands on remote servers via SSH',
      action: 'Run Remote',
      onClick: () => navigate('/remote-commands'),
    },
    // Row 2: Script Execution
    {
      icon: <PlayArrow sx={{ fontSize: 48 }} />,
      title: 'Run Scripts Locally',
      description: 'Execute stored bash scripts on the local server with env vars',
      action: 'Run Local Script',
      onClick: () => navigate('/local-scripts'),
    },
    {
      icon: <CloudQueue sx={{ fontSize: 48 }} />,
      title: 'Run Scripts Remotely',
      description: 'Execute stored bash scripts on remote servers via SSH',
      action: 'Run Remote Script',
      onClick: () => navigate('/remote-scripts'),
    },
    // Row 3: Management
    {
      icon: <Code sx={{ fontSize: 48 }} />,
      title: 'Bash Scripts',
      description: 'Store and manage reusable bash scripts for execution',
      action: 'Manage Scripts',
      onClick: () => navigate('/scripts'),
    },
    {
      icon: <Settings sx={{ fontSize: 48 }} />,
      title: 'Admin Panel',
      description: 'Manage SSH keys, servers, and environment variables',
      action: 'Open Admin',
      onClick: () => navigate('/admin'),
    },
    // Row 4: History and Terminal
    {
      icon: <Storage sx={{ fontSize: 48 }} />,
      title: 'Command History',
      description: 'View and replay previously executed commands',
      action: 'View History',
      onClick: () => navigate('/history'),
    },
    {
      icon: <Computer sx={{ fontSize: 48 }} />,
      title: 'Interactive Terminal',
      description: 'Open a full interactive shell session in your browser',
      action: 'Open Terminal',
      onClick: () => navigate('/terminal'),
    },
    // Row 5: Developer Tools
    {
      icon: <Build sx={{ fontSize: 48 }} />,
      title: 'Tools',
      description: 'Validate and format YAML, JSON, and other data formats',
      action: 'Open Tools',
      onClick: () => navigate('/tools'),
    },
  ];

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Box sx={{ mb: 4, textAlign: 'center' }}>
        <Typography variant="h3" component="h1" gutterBottom>
          Welcome to Web CLI
        </Typography>
        <Typography variant="body1" color="text.secondary" paragraph>
          A powerful web-based interface for executing commands on local and remote Linux servers
        </Typography>
      </Box>

      <Grid container spacing={3}>
        {features.map((feature, index) => (
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
                  {feature.icon}
                </Box>
                <Typography variant="h5" component="h2" gutterBottom>
                  {feature.title}
                </Typography>
                <Typography variant="body2" color="text.secondary" paragraph>
                  {feature.description}
                </Typography>
                <Button
                  variant="contained"
                  color="primary"
                  sx={{ mt: 2 }}
                  onClick={feature.onClick}
                  disabled={!feature.onClick}
                >
                  {feature.action}
                </Button>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default Dashboard;
