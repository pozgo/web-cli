import React, { createContext, useContext, useReducer, useEffect, useCallback } from 'react';

// Constants
const MAX_TABS = 10;
const STORAGE_KEY = 'terminal-tabs';

// Generate unique ID
const generateId = () => `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

// Action types
const actions = {
  ADD_TAB: 'ADD_TAB',
  CLOSE_TAB: 'CLOSE_TAB',
  SET_ACTIVE_TAB: 'SET_ACTIVE_TAB',
  RENAME_TAB: 'RENAME_TAB',
  SET_TAB_CONNECTED: 'SET_TAB_CONNECTED',
  UPDATE_TAB_CONFIG: 'UPDATE_TAB_CONFIG',
  RESTORE_TABS: 'RESTORE_TABS',
};

// Create default tab
const createDefaultTab = (index = 1) => ({
  id: generateId(),
  title: `Terminal ${index}`,
  shell: 'bash',
  sshKeyId: '',
  connected: false,
});

// Get initial state - check localStorage first
const getInitialState = () => {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed) && parsed.length > 0) {
        // Restore saved tabs with connected: false
        // Sanitize and validate data to handle legacy formats from older versions
        const restoredTabs = parsed
          .filter((tab) => tab && typeof tab === 'object' && tab.id)
          .map((tab) => ({
            // Ensure required properties exist with safe defaults
            id: String(tab.id),
            title: typeof tab.title === 'string' && tab.title ? tab.title : 'Terminal',
            // Ensure sshKeyId is always a string (older versions stored numeric IDs)
            sshKeyId: tab.sshKeyId != null ? String(tab.sshKeyId) : '',
            // Ensure shell is a string with default fallback
            shell: typeof tab.shell === 'string' && tab.shell ? tab.shell : 'bash',
            connected: false,
          }));
        
        // Only use restored tabs if we have valid ones
        if (restoredTabs.length > 0) {
          return {
            tabs: restoredTabs,
            activeTabId: restoredTabs[0].id,
          };
        }
      }
    }
  } catch (e) {
    console.error('Failed to restore terminal tabs:', e);
  }
  // Default: single tab
  const defaultTab = createDefaultTab(1);
  return {
    tabs: [defaultTab],
    activeTabId: defaultTab.id,
  };
};

// Reducer
function terminalReducer(state, action) {
  switch (action.type) {
    case actions.ADD_TAB: {
      if (state.tabs.length >= MAX_TABS) {
        return state;
      }
      const newTabIndex = state.tabs.length + 1;
      const newTab = {
        ...createDefaultTab(newTabIndex),
        ...action.payload,
      };
      return {
        ...state,
        tabs: [...state.tabs, newTab],
        activeTabId: newTab.id,
      };
    }

    case actions.CLOSE_TAB: {
      const { tabId } = action.payload;
      const tabIndex = state.tabs.findIndex((t) => t.id === tabId);
      if (tabIndex === -1 || state.tabs.length === 1) {
        // Can't close the last tab
        return state;
      }

      const newTabs = state.tabs.filter((t) => t.id !== tabId);
      let newActiveId = state.activeTabId;

      // If closing active tab, switch to adjacent tab
      if (state.activeTabId === tabId) {
        const newActiveIndex = Math.min(tabIndex, newTabs.length - 1);
        newActiveId = newTabs[newActiveIndex].id;
      }

      return {
        ...state,
        tabs: newTabs,
        activeTabId: newActiveId,
      };
    }

    case actions.SET_ACTIVE_TAB: {
      const { tabId } = action.payload;
      if (!state.tabs.find((t) => t.id === tabId)) {
        return state;
      }
      return {
        ...state,
        activeTabId: tabId,
      };
    }

    case actions.RENAME_TAB: {
      const { tabId, title } = action.payload;
      return {
        ...state,
        tabs: state.tabs.map((tab) =>
          tab.id === tabId ? { ...tab, title: title.trim() || tab.title } : tab
        ),
      };
    }

    case actions.SET_TAB_CONNECTED: {
      const { tabId, connected } = action.payload;
      return {
        ...state,
        tabs: state.tabs.map((tab) =>
          tab.id === tabId ? { ...tab, connected } : tab
        ),
      };
    }

    case actions.UPDATE_TAB_CONFIG: {
      const { tabId, config } = action.payload;
      return {
        ...state,
        tabs: state.tabs.map((tab) =>
          tab.id === tabId ? { ...tab, ...config } : tab
        ),
      };
    }

    case actions.RESTORE_TABS: {
      const { tabs } = action.payload;
      if (!tabs || tabs.length === 0) {
        return state;
      }
      // Reset connected state on restore
      const restoredTabs = tabs.map((tab) => ({
        ...tab,
        connected: false,
      }));
      return {
        ...state,
        tabs: restoredTabs,
        activeTabId: restoredTabs[0].id,
      };
    }

    default:
      return state;
  }
}

// Context
const TerminalContext = createContext(null);

// Provider component
export function TerminalProvider({ children }) {
  // Use lazy initialization to avoid double-loading from localStorage
  const [state, dispatch] = useReducer(terminalReducer, null, getInitialState);

  // Persist tabs to localStorage on change
  useEffect(() => {
    try {
      // Save only serializable tab data (not connected state)
      const tabsToSave = state.tabs.map(({ id, title, shell, sshKeyId }) => ({
        id,
        title,
        shell,
        sshKeyId,
      }));
      localStorage.setItem(STORAGE_KEY, JSON.stringify(tabsToSave));
    } catch (e) {
      console.error('Failed to save terminal tabs:', e);
    }
  }, [state.tabs]);

  // Action creators
  const addTab = useCallback((config = {}) => {
    dispatch({ type: actions.ADD_TAB, payload: config });
  }, []);

  const closeTab = useCallback((tabId) => {
    dispatch({ type: actions.CLOSE_TAB, payload: { tabId } });
  }, []);

  const setActiveTab = useCallback((tabId) => {
    dispatch({ type: actions.SET_ACTIVE_TAB, payload: { tabId } });
  }, []);

  const renameTab = useCallback((tabId, title) => {
    dispatch({ type: actions.RENAME_TAB, payload: { tabId, title } });
  }, []);

  const setTabConnected = useCallback((tabId, connected) => {
    dispatch({ type: actions.SET_TAB_CONNECTED, payload: { tabId, connected } });
  }, []);

  const updateTabConfig = useCallback((tabId, config) => {
    dispatch({ type: actions.UPDATE_TAB_CONFIG, payload: { tabId, config } });
  }, []);

  const value = {
    tabs: state.tabs,
    activeTabId: state.activeTabId,
    activeTab: state.tabs.find((t) => t.id === state.activeTabId),
    maxTabs: MAX_TABS,
    canAddTab: state.tabs.length < MAX_TABS,
    addTab,
    closeTab,
    setActiveTab,
    renameTab,
    setTabConnected,
    updateTabConfig,
  };

  return (
    <TerminalContext.Provider value={value}>
      {children}
    </TerminalContext.Provider>
  );
}

// Hook to use terminal context
export function useTerminal() {
  const context = useContext(TerminalContext);
  if (!context) {
    throw new Error('useTerminal must be used within a TerminalProvider');
  }
  return context;
}

export default TerminalContext;
