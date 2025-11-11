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
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

/**
 * Dashboard component - main landing page with feature cards
 */
const Dashboard = () => {
  const navigate = useNavigate();

  const features = [
    {
      icon: <Terminal sx={{ fontSize: 48 }} />,
      title: 'Local Commands',
      description: 'Execute commands on your local Linux server',
      action: 'Run Local',
      onClick: () => navigate('/local-commands'),
    },
    {
      icon: <Cloud sx={{ fontSize: 48 }} />,
      title: 'Remote Execution',
      description: 'Connect and run commands on remote servers via SSH',
      action: 'Run Remote',
      onClick: () => navigate('/remote-commands'),
    },
    {
      icon: <Settings sx={{ fontSize: 48 }} />,
      title: 'Admin Panel',
      description: 'Manage SSH keys and server configurations',
      action: 'Open Admin',
      onClick: () => navigate('/admin'),
    },
    {
      icon: <Storage sx={{ fontSize: 48 }} />,
      title: 'Command History',
      description: 'View and replay previously executed commands',
      action: 'View History',
      onClick: () => navigate('/history'),
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
