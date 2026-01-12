import React, { useState, useRef } from 'react';
import Editor, { DiffEditor } from '@monaco-editor/react';
import { jsonrepair } from 'jsonrepair';
import {
  Paper,
  Grid,
  Button,
  Typography,
  Box,
  Snackbar,
  Alert,
  Collapse,
  IconButton,
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import {
  CheckCircle,
  FormatAlignLeft,
  AutoFixHigh,
  ContentCopy,
  Clear,
  CompareArrows,
  Close,
} from '@mui/icons-material';
import ToolLayout from './shared/ToolLayout';
import ValidationResult from './shared/ValidationResult';

/**
 * JsonValidator - JSON validation, formatting, and beautification tool
 * Uses Monaco Editor for syntax highlighting and native JSON.parse for validation
 */
const JsonValidator = () => {
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [errors, setErrors] = useState([]);
  const [isValid, setIsValid] = useState(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });
  const [showDiff, setShowDiff] = useState(false);
  const [diffOriginal, setDiffOriginal] = useState('');
  const [diffModified, setDiffModified] = useState('');
  const editorRef = useRef(null);

  const theme = useTheme();
  const isDarkMode = theme.palette.mode === 'dark';

  // Store editor reference for programmatic updates
  const handleEditorDidMount = (editor) => {
    editorRef.current = editor;
  };

  // Update editor value programmatically
  const updateEditorValue = (newValue) => {
    setInput(newValue);
    if (editorRef.current) {
      const model = editorRef.current.getModel();
      if (model) {
        model.setValue(newValue);
      }
    }
  };

  // Helper to extract line/column from JSON parse error
  const parseJsonError = (error, jsonString) => {
    const message = error.message;
    
    // Try to extract position from error message
    // Format: "... at position 123" or "Unexpected token ... at position 123"
    const positionMatch = message.match(/position\s+(\d+)/i);
    
    if (positionMatch) {
      const position = parseInt(positionMatch[1], 10);
      const lines = jsonString.substring(0, position).split('\n');
      const line = lines.length;
      const column = lines[lines.length - 1].length + 1;
      
      return {
        line,
        column,
        message: message,
      };
    }
    
    return { message };
  };

  // Validate JSON syntax
  const handleValidate = () => {
    if (!input.trim()) {
      setErrors([{ message: 'Please enter some JSON content to validate' }]);
      setIsValid(false);
      return;
    }

    try {
      JSON.parse(input);
      setErrors([]);
      setIsValid(true);
      setOutput(input);
    } catch (e) {
      const errorInfo = parseJsonError(e, input);
      setErrors([errorInfo]);
      setIsValid(false);
      setOutput('');
    }
  };

  // Format/beautify JSON with configurable indentation
  const handleFormat = () => {
    if (!input.trim()) {
      showSnackbar('Please enter some JSON content first', 'warning');
      return;
    }

    try {
      const parsed = JSON.parse(input);
      const formatted = JSON.stringify(parsed, null, 2);
      setOutput(formatted);
      setErrors([]);
      setIsValid(true);
      showSnackbar('JSON formatted successfully', 'success');
    } catch (e) {
      const errorInfo = parseJsonError(e, input);
      setErrors([errorInfo]);
      setIsValid(false);
    }
  };

  // Auto-fix JSON issues using jsonrepair library
  // Handles: trailing commas, missing quotes, missing commas, missing brackets,
  // comments, single quotes, unquoted keys, and structural issues
  const handleAutoFix = () => {
    if (!input.trim()) {
      showSnackbar('Please enter some JSON content first', 'warning');
      return;
    }

    try {
      const originalInput = input;

      // Use jsonrepair to fix structural and syntax issues
      const repaired = jsonrepair(input);

      // Parse and re-stringify for consistent formatting
      const parsed = JSON.parse(repaired);
      const formatted = JSON.stringify(parsed, null, 2);

      // Detect what was fixed
      const fixes = [];
      if (originalInput !== repaired) {
        if (/,\s*[}\]]/.test(originalInput)) fixes.push('trailing commas');
        if (/\/\/|\/\*/.test(originalInput)) fixes.push('comments');
        if (/'/.test(originalInput) && !/'/.test(formatted)) fixes.push('quotes');
        if (/}\s*,\s*}/.test(originalInput)) fixes.push('structural issues');
        if ((originalInput.match(/{/g) || []).length !== (originalInput.match(/}/g) || []).length) {
          fixes.push('mismatched braces');
        }
        if ((originalInput.match(/\[/g) || []).length !== (originalInput.match(/]/g) || []).length) {
          fixes.push('mismatched brackets');
        }
        // Generic fix message if we can't detect specifics
        if (fixes.length === 0) fixes.push('syntax errors');
      }

      updateEditorValue(formatted);
      setOutput(formatted);
      setErrors([]);
      setIsValid(true);

      // Set diff state to show changes
      if (originalInput !== formatted) {
        setDiffOriginal(originalInput);
        setDiffModified(formatted);
        setShowDiff(true);
      }

      if (fixes.length > 0) {
        showSnackbar(`Auto-fixed: ${fixes.join(', ')}. View diff below.`, 'success');
      } else {
        showSnackbar('JSON is already valid, formatted successfully', 'success');
      }
    } catch (repairError) {
      // jsonrepair couldn't fix it - show the error
      const errorInfo = parseJsonError(repairError, input);
      setErrors([errorInfo]);
      setIsValid(false);
      setShowDiff(false);
      showSnackbar('Could not auto-fix: ' + repairError.message, 'error');
    }
  };

  // Minify JSON (compact format)
  const handleMinify = () => {
    if (!input.trim()) {
      showSnackbar('Please enter some JSON content first', 'warning');
      return;
    }

    try {
      const parsed = JSON.parse(input);
      const minified = JSON.stringify(parsed);
      setOutput(minified);
      setErrors([]);
      setIsValid(true);
      showSnackbar('JSON minified successfully', 'success');
    } catch (e) {
      const errorInfo = parseJsonError(e, input);
      setErrors([errorInfo]);
      setIsValid(false);
    }
  };

  // Copy output to clipboard
  const handleCopy = async () => {
    if (!output) {
      showSnackbar('No output to copy', 'warning');
      return;
    }

    try {
      await navigator.clipboard.writeText(output);
      showSnackbar('Copied to clipboard!', 'success');
    } catch (err) {
      showSnackbar('Failed to copy to clipboard', 'error');
    }
  };

  // Clear all state
  const handleClear = () => {
    updateEditorValue('');
    setOutput('');
    setErrors([]);
    setIsValid(null);
    setShowDiff(false);
    setDiffOriginal('');
    setDiffModified('');
  };

  // Show snackbar notification
  const showSnackbar = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleCloseSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  // Monaco editor options
  const editorOptions = {
    minimap: { enabled: false },
    fontSize: 14,
    lineNumbers: 'on',
    scrollBeyondLastLine: false,
    wordWrap: 'on',
    automaticLayout: true,
    formatOnPaste: true,
  };

  return (
    <ToolLayout
      title="JSON Validator"
      description="Validate, format, and beautify JSON data with error detection"
    >
      <Paper sx={{ p: 3 }}>
        <Grid container spacing={3}>
          {/* Input Section */}
          <Grid item xs={12} md={6}>
            <Typography variant="h6" gutterBottom>
              Input
            </Typography>
            <Box sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
              <Editor
                height="400px"
                language="json"
                theme={isDarkMode ? 'vs-dark' : 'light'}
                value={input}
                onChange={(value) => setInput(value || '')}
                onMount={handleEditorDidMount}
                options={editorOptions}
              />
            </Box>
            <Box sx={{ mt: 2, display: 'flex', flexWrap: 'wrap', gap: 1 }}>
              <Button
                variant="contained"
                color="primary"
                startIcon={<CheckCircle />}
                onClick={handleValidate}
                aria-label="Validate JSON syntax"
              >
                Validate
              </Button>
              <Button
                variant="outlined"
                startIcon={<FormatAlignLeft />}
                onClick={handleFormat}
                aria-label="Format and beautify JSON"
              >
                Format
              </Button>
              <Button
                variant="outlined"
                onClick={handleMinify}
                aria-label="Minify JSON to single line"
              >
                Minify
              </Button>
              <Button
                variant="outlined"
                startIcon={<AutoFixHigh />}
                onClick={handleAutoFix}
                aria-label="Auto-fix common JSON issues"
              >
                Auto-Fix
              </Button>
              <Button
                variant="outlined"
                color="error"
                startIcon={<Clear />}
                onClick={handleClear}
                aria-label="Clear input and output"
              >
                Clear
              </Button>
            </Box>
          </Grid>

          {/* Output Section */}
          <Grid item xs={12} md={6}>
            <Typography variant="h6" gutterBottom>
              Output
            </Typography>
            <Box sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
              <Editor
                height="400px"
                language="json"
                theme={isDarkMode ? 'vs-dark' : 'light'}
                value={output}
                options={{ ...editorOptions, readOnly: true }}
              />
            </Box>
            <Box sx={{ mt: 2, display: 'flex', gap: 1, alignItems: 'center' }}>
              <Button
                variant="outlined"
                startIcon={<ContentCopy />}
                onClick={handleCopy}
                disabled={!output}
                aria-label="Copy output to clipboard"
              >
                Copy Output
              </Button>
              {showDiff && (
                <Button
                  variant="outlined"
                  startIcon={<CompareArrows />}
                  onClick={() => setShowDiff(!showDiff)}
                  color="info"
                  aria-label="Hide diff comparison view"
                >
                  Hide Diff
                </Button>
              )}
            </Box>
            <Box sx={{ mt: 2 }}>
              <ValidationResult
                isValid={isValid}
                errors={errors}
                successMessage="Valid JSON!"
              />
            </Box>
          </Grid>
        </Grid>
      </Paper>

      {/* Diff View - Shows changes made by Auto-Fix */}
      <Collapse in={showDiff}>
        <Paper sx={{ p: 3, mt: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography variant="h6">
              Changes Made by Auto-Fix
            </Typography>
            <IconButton onClick={() => setShowDiff(false)} size="small" aria-label="Close diff view">
              <Close />
            </IconButton>
          </Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Red = removed, Green = added. Left side shows original, right side shows fixed version.
          </Typography>
          <Box sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
            <DiffEditor
              height="400px"
              language="json"
              theme={isDarkMode ? 'vs-dark' : 'light'}
              original={diffOriginal}
              modified={diffModified}
              options={{
                readOnly: true,
                renderSideBySide: true,
                minimap: { enabled: false },
                fontSize: 14,
                scrollBeyondLastLine: false,
                wordWrap: 'on',
              }}
            />
          </Box>
        </Paper>
      </Collapse>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={handleCloseSnackbar}
          severity={snackbar.severity}
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </ToolLayout>
  );
};

export default JsonValidator;
