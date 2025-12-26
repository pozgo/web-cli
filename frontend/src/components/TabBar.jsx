import React, { useState, useRef, useEffect } from 'react';
import {
  Box,
  Tab,
  Tabs,
  IconButton,
  Tooltip,
  TextField,
  Typography,
} from '@mui/material';
import {
  Add as AddIcon,
  Close as CloseIcon,
  Circle as CircleIcon,
} from '@mui/icons-material';
import { useTerminal } from '../context/TerminalContext';

/**
 * TabBar component - Manages terminal tabs navigation
 * Features: Add/close tabs, rename on double-click, connection status indicators
 */
const TabBar = () => {
  const {
    tabs,
    activeTabId,
    canAddTab,
    maxTabs,
    addTab,
    closeTab,
    setActiveTab,
    renameTab,
  } = useTerminal();

  const [editingTabId, setEditingTabId] = useState(null);
  const [editValue, setEditValue] = useState('');
  const inputRef = useRef(null);

  // Focus input when editing starts
  useEffect(() => {
    if (editingTabId && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editingTabId]);

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  const handleAddTab = (e) => {
    e.stopPropagation();
    if (canAddTab) {
      addTab();
    }
  };

  const handleCloseTab = (e, tabId) => {
    e.stopPropagation();
    closeTab(tabId);
  };

  const handleDoubleClick = (tab) => {
    setEditingTabId(tab.id);
    setEditValue(tab.title);
  };

  const handleRenameSubmit = (tabId) => {
    if (editValue.trim()) {
      renameTab(tabId, editValue.trim());
    }
    setEditingTabId(null);
  };

  const handleRenameKeyDown = (e, tabId) => {
    if (e.key === 'Enter') {
      handleRenameSubmit(tabId);
    } else if (e.key === 'Escape') {
      setEditingTabId(null);
    }
  };

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        bgcolor: 'background.paper',
        borderBottom: 1,
        borderColor: 'divider',
        minHeight: 48,
      }}
    >
      <Tabs
        value={activeTabId}
        onChange={handleTabChange}
        variant="scrollable"
        scrollButtons="auto"
        sx={{
          flexGrow: 1,
          minHeight: 48,
          '& .MuiTabs-indicator': {
            height: 3,
          },
        }}
      >
        {tabs.map((tab) => (
          <Tab
            key={tab.id}
            value={tab.id}
            sx={{
              minHeight: 48,
              textTransform: 'none',
              minWidth: 120,
              maxWidth: 200,
              px: 1,
            }}
            label={
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  width: '100%',
                  gap: 1,
                }}
                onDoubleClick={(e) => {
                  e.stopPropagation();
                  handleDoubleClick(tab);
                }}
              >
                {/* Connection status indicator */}
                <Tooltip title={tab.connected ? 'Connected' : 'Disconnected'}>
                  <CircleIcon
                    sx={{
                      fontSize: 10,
                      color: tab.connected ? 'success.main' : 'error.main',
                      flexShrink: 0,
                    }}
                  />
                </Tooltip>

                {/* Tab title - editable */}
                {editingTabId === tab.id ? (
                  <TextField
                    inputRef={inputRef}
                    size="small"
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    onBlur={() => handleRenameSubmit(tab.id)}
                    onKeyDown={(e) => handleRenameKeyDown(e, tab.id)}
                    onClick={(e) => e.stopPropagation()}
                    sx={{
                      '& .MuiInputBase-input': {
                        py: 0.5,
                        px: 1,
                        fontSize: '0.875rem',
                      },
                    }}
                    inputProps={{
                      maxLength: 20,
                    }}
                  />
                ) : (
                  <Typography
                    variant="body2"
                    noWrap
                    sx={{
                      flexGrow: 1,
                      textAlign: 'left',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                    }}
                  >
                    {tab.title}
                  </Typography>
                )}

                {/* Close button */}
                {tabs.length > 1 && (
                  <IconButton
                    size="small"
                    onClick={(e) => handleCloseTab(e, tab.id)}
                    sx={{
                      p: 0.25,
                      ml: 0.5,
                      opacity: 0.6,
                      '&:hover': {
                        opacity: 1,
                        bgcolor: 'action.hover',
                      },
                    }}
                  >
                    <CloseIcon sx={{ fontSize: 16 }} />
                  </IconButton>
                )}
              </Box>
            }
          />
        ))}
      </Tabs>

      {/* Add new tab button */}
      <Tooltip
        title={
          canAddTab
            ? 'New Terminal (max ' + maxTabs + ')'
            : 'Maximum tabs reached'
        }
      >
        <span>
          <IconButton
            onClick={handleAddTab}
            disabled={!canAddTab}
            sx={{
              mx: 1,
              bgcolor: 'action.hover',
              '&:hover': {
                bgcolor: 'action.selected',
              },
            }}
          >
            <AddIcon />
          </IconButton>
        </span>
      </Tooltip>
    </Box>
  );
};

export default TabBar;
