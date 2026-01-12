import React, { useState, useRef } from 'react';
import Editor, { DiffEditor } from '@monaco-editor/react';
import yaml from 'js-yaml';
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
 * YamlValidator - YAML validation, formatting, and auto-fix tool
 * Uses Monaco Editor for syntax highlighting and js-yaml for parsing
 */
const YamlValidator = () => {
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

  // Validate YAML syntax
  const handleValidate = () => {
    if (!input.trim()) {
      setErrors([{ message: 'Please enter some YAML content to validate' }]);
      setIsValid(false);
      return;
    }

    try {
      yaml.load(input, { schema: yaml.DEFAULT_SCHEMA });
      setErrors([]);
      setIsValid(true);
      setOutput(input);
    } catch (e) {
      if (e.mark) {
        setErrors([{
          line: e.mark.line + 1,
          column: e.mark.column + 1,
          message: e.reason || e.message,
        }]);
      } else {
        setErrors([{ message: e.message }]);
      }
      setIsValid(false);
      setOutput('');
    }
  };

  // Format/beautify YAML
  const handleFormat = () => {
    if (!input.trim()) {
      showSnackbar('Please enter some YAML content first', 'warning');
      return;
    }

    try {
      const parsed = yaml.load(input);
      const formatted = yaml.dump(parsed, {
        indent: 2,
        lineWidth: -1,
        noRefs: true,
        sortKeys: false,
      });
      setOutput(formatted);
      setErrors([]);
      setIsValid(true);
      showSnackbar('YAML formatted successfully', 'success');
    } catch (e) {
      if (e.mark) {
        setErrors([{
          line: e.mark.line + 1,
          column: e.mark.column + 1,
          message: e.reason || e.message,
        }]);
      } else {
        setErrors([{ message: e.message }]);
      }
      setIsValid(false);
    }
  };

  // Auto-fix common YAML issues
  const handleAutoFix = () => {
    if (!input.trim()) {
      showSnackbar('Please enter some YAML content first', 'warning');
      return;
    }

    const originalInput = input;
    let fixed = input;
    let fixes = [];

    // Remove BOM if present
    if (fixed.charCodeAt(0) === 0xFEFF) {
      fixed = fixed.slice(1);
      fixes.push('BOM removed');
    }

    // Convert tabs to 2 spaces (YAML doesn't allow tabs for indentation)
    if (fixed.includes('\t')) {
      fixed = fixed.replace(/\t/g, '  ');
      fixes.push('tabs converted to spaces');
    }

    // Remove trailing whitespace from each line
    const beforeTrailing = fixed;
    fixed = fixed.replace(/[ \t]+$/gm, '');
    if (beforeTrailing !== fixed) {
      fixes.push('trailing whitespace removed');
    }

    // Remove multiple consecutive blank lines
    const beforeBlankLines = fixed;
    fixed = fixed.replace(/\n{3,}/g, '\n\n');
    if (beforeBlankLines !== fixed) {
      fixes.push('extra blank lines removed');
    }

    // Quote values that contain ': ' (colon-space) which confuses the parser
    // This is needed because js-yaml is strict about colons in unquoted values
    const quoteProblematicValues = (text) => {
      const lines = text.split('\n');
      const result = [];
      
      for (const line of lines) {
        // Skip empty lines and comments
        if (!line.trim() || line.trim().startsWith('#')) {
          result.push(line);
          continue;
        }
        
        // Match key: value pattern
        const match = line.match(/^(\s*)([^:]+):\s+(.+)$/);
        if (match) {
          const [, indent, key, value] = match;
          // Check if value contains ': ' and is not already quoted
          if (value.includes(': ') && !value.startsWith('"') && !value.startsWith("'")) {
            // Quote the value with double quotes, escaping any existing quotes
            const escapedValue = value.replace(/"/g, '\\"');
            result.push(`${indent}${key}: "${escapedValue}"`);
            continue;
          }
        }
        result.push(line);
      }
      
      return result.join('\n');
    };

    const beforeQuoting = fixed;
    fixed = quoteProblematicValues(fixed);
    if (beforeQuoting !== fixed) {
      fixes.push('problematic values quoted');
    }

    // Normalize indentation to multiples of 2 spaces
    // This fixes cases like 3-space indent which YAML doesn't accept
    const normalizeIndentation = (text) => {
      const lines = text.split('\n');
      const result = [];
      
      for (const line of lines) {
        // Skip empty lines - preserve as-is
        if (!line.trim()) {
          result.push(line);
          continue;
        }
        
        // Get leading whitespace
        const match = line.match(/^(\s*)/);
        const indent = match ? match[1].length : 0;
        const content = line.trimStart();
        
        // Round indent to nearest multiple of 2
        // If indent is odd, round down (e.g., 3 -> 2, 5 -> 4)
        const normalizedIndent = Math.floor(indent / 2) * 2;
        
        result.push(' '.repeat(normalizedIndent) + content);
      }
      
      return result.join('\n');
    };

    const beforeIndentNorm = fixed;
    fixed = normalizeIndentation(fixed);
    if (beforeIndentNorm !== fixed) {
      fixes.push('indentation normalized');
    }

    // Ensure trailing newline
    if (fixed.length > 0 && !fixed.endsWith('\n')) {
      fixed += '\n';
      fixes.push('trailing newline added');
    }

    // Try to parse and re-format for consistent output
    // This will also fix indentation issues by re-serializing
    try {
      const parsed = yaml.load(fixed);
      const formatted = yaml.dump(parsed, {
        indent: 2,
        lineWidth: -1,
        noRefs: true,
        sortKeys: false,
      });

      fixed = formatted;
      fixes.push('reformatted');

      setErrors([]);
      setIsValid(true);
      setOutput(formatted);

      updateEditorValue(fixed);

      // Set diff state to show changes
      if (originalInput !== fixed) {
        setDiffOriginal(originalInput);
        setDiffModified(fixed);
        setShowDiff(true);
      }

      if (fixes.length > 0) {
        showSnackbar(`Auto-fixed: ${fixes.join(', ')}. View diff below.`, 'success');
      } else {
        showSnackbar('No issues found to fix', 'info');
      }
    } catch (e) {
      // If parsing fails, show the error and the partially fixed content
      if (e.mark) {
        setErrors([{
          line: e.mark.line + 1,
          column: e.mark.column + 1,
          message: e.reason || e.message,
        }]);
      } else {
        setErrors([{ message: e.message }]);
      }
      setIsValid(false);
      setOutput('');

      // Still update the editor with partial fixes (tabs, whitespace)
      updateEditorValue(fixed);

      if (originalInput !== fixed) {
        setDiffOriginal(originalInput);
        setDiffModified(fixed);
        setShowDiff(true);
      }

      const partialFixes = fixes.filter(f => f !== 'reformatted');
      if (partialFixes.length > 0) {
        showSnackbar(`Partial fix applied: ${partialFixes.join(', ')}. Manual fixes still needed.`, 'warning');
      } else {
        showSnackbar('Could not auto-fix: ' + (e.reason || e.message), 'error');
      }
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
  };

  return (
    <ToolLayout
      title="YAML Validator"
      description="Validate, format, and auto-fix YAML documents with syntax highlighting"
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
                language="yaml"
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
              >
                Validate
              </Button>
              <Button
                variant="outlined"
                startIcon={<FormatAlignLeft />}
                onClick={handleFormat}
              >
                Format
              </Button>
              <Button
                variant="outlined"
                startIcon={<AutoFixHigh />}
                onClick={handleAutoFix}
              >
                Auto-Fix
              </Button>
              <Button
                variant="outlined"
                color="error"
                startIcon={<Clear />}
                onClick={handleClear}
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
                language="yaml"
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
              >
                Copy Output
              </Button>
              {showDiff && (
                <Button
                  variant="outlined"
                  startIcon={<CompareArrows />}
                  onClick={() => setShowDiff(!showDiff)}
                  color="info"
                >
                  Hide Diff
                </Button>
              )}
            </Box>
            <Box sx={{ mt: 2 }}>
              <ValidationResult
                isValid={isValid}
                errors={errors}
                successMessage="Valid YAML!"
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
            <IconButton onClick={() => setShowDiff(false)} size="small">
              <Close />
            </IconButton>
          </Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Red = removed, Green = added. Left side shows original, right side shows fixed version.
          </Typography>
          <Box sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
            <DiffEditor
              height="400px"
              language="yaml"
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

export default YamlValidator;
