#!/usr/bin/env node
// ABOUTME: Parse tree visualization tool for tree-sitter-typed-perl with ERROR node highlighting
// ABOUTME: Generates visual representations of parse trees with color coding and multiple output formats

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// ANSI color codes for terminal output
const colors = {
    reset: '\x1b[0m',
    bright: '\x1b[1m',
    dim: '\x1b[2m',
    red: '\x1b[31m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    magenta: '\x1b[35m',
    cyan: '\x1b[36m',
    white: '\x1b[37m',
    bgRed: '\x1b[41m',
    bgGreen: '\x1b[42m',
    bgYellow: '\x1b[43m'
};

function colorize(text, color) {
    return `${colors[color] || ''}${text}${colors.reset}`;
}

function usage() {
    console.log(colorize('Tree-sitter Typed Perl Parse Tree Visualizer', 'bright'));
    console.log('\nUsage:');
    console.log('  ./visualize_tree.js "perl code"            # Visualize inline code');
    console.log('  ./visualize_tree.js -f file.pl             # Visualize file');
    console.log('  ./visualize_tree.js -i                     # Interactive mode');
    console.log('  ./visualize_tree.js --help                 # Show this help');
    console.log('\nOutput Options:');
    console.log('  --html, -h          Generate HTML output');
    console.log('  --json, -j          Output raw JSON tree');
    console.log('  --compact, -c       Compact tree format');
    console.log('  --no-color          Disable colored output');
    console.log('  --output, -o FILE   Save output to file');
    console.log('\nVisualization Options:');
    console.log('  --errors-only, -e   Show only ERROR nodes and their context');
    console.log('  --depth, -d NUM     Limit tree depth (default: unlimited)');
    console.log('  --indent SIZE       Set indentation size (default: 2)');
    console.log('\nExamples:');
    console.log('  ./visualize_tree.js "my Int $var = 42;"');
    console.log('  ./visualize_tree.js -f test.pl --html -o output.html');
    console.log('  ./visualize_tree.js "broken syntax" --errors-only');
    process.exit(0);
}

function createTempFile(content) {
    const tempDir = '/tmp';
    const tempFile = path.join(tempDir, `visualize_tree_${Date.now()}_${Math.random().toString(36).substr(2, 9)}.pl`);
    fs.writeFileSync(tempFile, content);
    return tempFile;
}

function parseToTree(code) {
    const tempFile = createTempFile(code);

    try {
        // Get the S-expression parse tree
        // Run from the script directory to find the grammar
        const scriptDir = path.dirname(__filename);
        const result = execSync(`tree-sitter parse ${tempFile}`, {
            encoding: 'utf8',
            maxBuffer: 1024 * 1024 * 10, // 10MB buffer
            cwd: scriptDir // Run from grammar directory
        });

        return { success: true, tree: result, error: null };
    } catch (error) {
        // tree-sitter sometimes writes warnings to stderr but still produces valid output
        const stdout = error.stdout || '';
        const stderr = error.stderr || error.message || '';

        if (stdout.includes('(source_file') || stdout.includes('(ERROR')) {
            // Valid parse tree in stdout, just ignore stderr warnings
            return { success: true, tree: stdout, error: stderr };
        } else {
            return {
                success: false,
                tree: stdout,
                error: stderr
            };
        }
    } finally {
        // Cleanup temp file
        try {
            fs.unlinkSync(tempFile);
        } catch (e) {
            // Ignore cleanup errors
        }
    }
}

function parseTreeStructure(sExpression) {
    // Parse S-expression into structured tree
    const lines = sExpression.split('\n').filter(line => line.trim());
    const tree = [];
    const stack = [{ children: tree, depth: -1 }];

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        const depth = (line.length - line.trimStart().length) / 2;

        // Parse tree-sitter format: (node_type [start] - [end] ...)
        // or with labeled children: label: (node_type ...)
        let nodeType = '';
        let content = '';
        let label = '';

        // Check for labeled children (e.g., "left: (assignment_expression...")
        const labelMatch = trimmed.match(/^(\w+):\s*(.+)$/);
        if (labelMatch) {
            label = labelMatch[1];
            const nodeContent = labelMatch[2];
            const nodeMatch = nodeContent.match(/^\((\w+)(?:\s+\[[\d,\s-]+\])?(?:\s+(.*))?$/);
            if (nodeMatch) {
                nodeType = nodeMatch[1];
                content = nodeMatch[2] || '';
            }
        } else {
            // Direct node (e.g., "(source_file [0, 0] - [1, 0]")
            const nodeMatch = trimmed.match(/^\((\w+)(?:\s+\[[\d,\s-]+\])?(?:\s+(.*))?$/);
            if (nodeMatch) {
                nodeType = nodeMatch[1];
                content = nodeMatch[2] || '';
            } else if (trimmed.match(/^\)$/)) {
                // Closing parenthesis, ignore
                continue;
            } else {
                // Leaf content
                nodeType = 'content';
                content = trimmed;
            }
        }

        if (!nodeType) continue;

        const node = {
            type: nodeType,
            content: content.replace(/\)$/, '').trim() || null,
            label: label || null,
            children: [],
            depth: depth,
            isError: nodeType === 'ERROR' || nodeType === 'MISSING',
            line: line
        };

        // Find correct parent based on depth
        while (stack.length > 0 && stack[stack.length - 1].depth >= depth) {
            stack.pop();
        }

        if (stack.length === 0) {
            stack.push({ children: tree, depth: -1 });
        }

        const parent = stack[stack.length - 1];
        parent.children.push(node);
        stack.push(node);
    }

    return tree;
}

function renderTerminalTree(tree, options = {}) {
    const {
        compact = false,
        errorsOnly = false,
        maxDepth = Infinity,
        indent = 2,
        noColor = false
    } = options;

    const indentStr = ' '.repeat(indent);
    let output = '';

    function renderNode(node, depth = 0) {
        if (depth > maxDepth) return '';

        const isError = node.isError || node.type === 'ERROR' || node.type === 'MISSING';
        const hasErrorDescendant = hasErrorInSubtree(node);

        // Skip non-error nodes if errors-only mode is enabled
        if (errorsOnly && !isError && !hasErrorDescendant) {
            return '';
        }

        const prefix = indentStr.repeat(depth);
        let nodeText = `${prefix}`;

        if (!compact) {
            nodeText += depth > 0 ? '├─ ' : '';
        }

        // Color coding for different node types
        let nodeTypeText = node.type;
        if (!noColor) {
            if (isError) {
                nodeTypeText = colorize(nodeTypeText, 'bgRed') + colorize(' ← ERROR', 'red');
            } else if (hasErrorDescendant) {
                nodeTypeText = colorize(nodeTypeText, 'yellow');
            } else if (node.type.includes('type') || node.type.includes('annotation')) {
                nodeTypeText = colorize(nodeTypeText, 'cyan');
            } else if (node.type === 'identifier' || node.type === 'variable') {
                nodeTypeText = colorize(nodeTypeText, 'green');
            } else if (node.type.includes('literal') || node.type.includes('string')) {
                nodeTypeText = colorize(nodeTypeText, 'magenta');
            } else {
                nodeTypeText = colorize(nodeTypeText, 'blue');
            }
        }

        if (node.label) {
            const labelText = noColor ? node.label : colorize(node.label, 'yellow');
            nodeText += `${labelText}: `;
        }

        nodeText += `(${nodeTypeText})`;

        if (node.content && !compact) {
            const contentText = noColor ? node.content : colorize(node.content, 'dim');
            nodeText += ` ${contentText}`;
        }

        nodeText += '\n';

        // Render children
        if (node.children && node.children.length > 0) {
            for (const child of node.children) {
                const childText = renderNode(child, depth + 1);
                if (childText) { // Only add if not filtered out
                    nodeText += childText;
                }
            }
        }

        return nodeText;
    }

    function hasErrorInSubtree(node) {
        if (node.isError || node.type === 'ERROR' || node.type === 'MISSING') {
            return true;
        }
        if (node.children) {
            return node.children.some(child => hasErrorInSubtree(child));
        }
        return false;
    }

    for (const node of tree) {
        output += renderNode(node);
    }

    return output;
}

function renderHTMLTree(tree, code, options = {}) {
    const { errorsOnly = false, maxDepth = Infinity } = options;

    function escapeHtml(text) {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function renderNode(node, depth = 0) {
        if (depth > maxDepth) return '';

        const isError = node.isError || node.type === 'ERROR' || node.type === 'MISSING';
        const hasErrorDescendant = hasErrorInSubtree(node);

        // Skip non-error nodes if errors-only mode is enabled
        if (errorsOnly && !isError && !hasErrorDescendant) {
            return '';
        }

        let cssClass = 'node';
        if (isError) {
            cssClass += ' error';
        } else if (hasErrorDescendant) {
            cssClass += ' has-error';
        } else if (node.type.includes('type') || node.type.includes('annotation')) {
            cssClass += ' type-node';
        } else if (node.type === 'identifier' || node.type === 'variable') {
            cssClass += ' identifier';
        } else if (node.type.includes('literal') || node.type.includes('string')) {
            cssClass += ' literal';
        }

        let html = `<div class="${cssClass}" style="margin-left: ${depth * 20}px;">`;

        if (node.label) {
            html += `<span class="node-label">${escapeHtml(node.label)}:</span> `;
        }

        html += `<span class="node-type">(${escapeHtml(node.type)})</span>`;

        if (node.content) {
            html += ` <span class="node-content">${escapeHtml(node.content)}</span>`;
        }

        if (isError) {
            html += ' <span class="error-marker">← ERROR</span>';
        }

        html += '</div>';

        // Render children
        if (node.children && node.children.length > 0) {
            for (const child of node.children) {
                const childHtml = renderNode(child, depth + 1);
                if (childHtml) {
                    html += childHtml;
                }
            }
        }

        return html;
    }

    function hasErrorInSubtree(node) {
        if (node.isError || node.type === 'ERROR' || node.type === 'MISSING') {
            return true;
        }
        if (node.children) {
            return node.children.some(child => hasErrorInSubtree(child));
        }
        return false;
    }

    let bodyContent = '';
    for (const node of tree) {
        bodyContent += renderNode(node);
    }

    const errorCount = countErrors(tree);
    const errorSummary = errorCount > 0 ?
        `<div class="error-summary">⚠️ Found ${errorCount} ERROR nodes in parse tree</div>` :
        '<div class="success-summary">✅ Clean parse tree (no errors)</div>';

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Parse Tree Visualization</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background-color: #f5f5f5;
            margin: 20px;
            line-height: 1.4;
        }
        .header {
            background: #333;
            color: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .code-block {
            background: #2d2d2d;
            color: #f8f8f2;
            padding: 15px;
            border-radius: 4px;
            overflow-x: auto;
            margin: 10px 0;
            white-space: pre;
        }
        .tree-container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .node {
            margin: 2px 0;
            padding: 2px 4px;
            border-radius: 3px;
        }
        .node-type {
            font-weight: bold;
            color: #0066cc;
        }
        .node-content {
            color: #666;
            font-style: italic;
        }
        .node-label {
            color: #e65100;
            font-weight: bold;
        }
        .error {
            background-color: #ffebee;
            border-left: 4px solid #f44336;
            padding-left: 8px;
        }
        .error .node-type {
            color: #d32f2f;
            font-weight: bold;
        }
        .error-marker {
            color: #d32f2f;
            font-weight: bold;
            background: #ffcdd2;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 0.8em;
        }
        .has-error {
            background-color: #fff3e0;
            border-left: 2px solid #ff9800;
        }
        .has-error .node-type {
            color: #f57c00;
        }
        .type-node .node-type {
            color: #7b1fa2;
        }
        .identifier .node-type {
            color: #2e7d32;
        }
        .literal .node-type {
            color: #c2185b;
        }
        .error-summary, .success-summary {
            padding: 10px;
            border-radius: 4px;
            margin-bottom: 20px;
            font-weight: bold;
        }
        .error-summary {
            background: #ffebee;
            border: 1px solid #f44336;
            color: #d32f2f;
        }
        .success-summary {
            background: #e8f5e8;
            border: 1px solid #4caf50;
            color: #2e7d32;
        }
        .stats {
            background: #f5f5f5;
            padding: 10px;
            border-radius: 4px;
            margin-top: 20px;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🌳 Parse Tree Visualization</h1>
        <p>tree-sitter-typed-perl parse tree with ERROR node highlighting</p>
    </div>

    ${errorSummary}

    <h2>Input Code:</h2>
    <div class="code-block">${escapeHtml(code)}</div>

    <h2>Parse Tree:</h2>
    <div class="tree-container">
        ${bodyContent}
    </div>

    <div class="stats">
        Generated on ${new Date().toISOString()}<br>
        Total nodes: ${countNodes(tree)} | ERROR nodes: ${errorCount}
    </div>
</body>
</html>`;
}

function countErrors(tree) {
    let count = 0;
    function traverse(nodes) {
        for (const node of nodes) {
            if (node.isError || node.type === 'ERROR' || node.type === 'MISSING') {
                count++;
            }
            if (node.children) {
                traverse(node.children);
            }
        }
    }
    traverse(tree);
    return count;
}

function countNodes(tree) {
    let count = 0;
    function traverse(nodes) {
        for (const node of nodes) {
            count++;
            if (node.children) {
                traverse(node.children);
            }
        }
    }
    traverse(tree);
    return count;
}

function visualizeCode(code, options = {}) {
    console.log(colorize('🌳 Tree-sitter Parse Tree Visualization', 'bright'));
    console.log(colorize('='.repeat(50), 'cyan'));

    if (!options.noColor) {
        console.log(colorize('Input code:', 'bright'));
        console.log(colorize(code, 'green'));
        console.log('');
    }

    // Parse the code
    const parseResult = parseToTree(code);

    if (!parseResult.success) {
        console.log(colorize('❌ Parse failed:', 'red'));
        console.log(parseResult.error);
        return;
    }

    // Parse tree structure
    const tree = parseTreeStructure(parseResult.tree);
    const errorCount = countErrors(tree);
    const totalNodes = countNodes(tree);

    // Show summary
    if (errorCount > 0) {
        console.log(colorize(`⚠️  Found ${errorCount} ERROR nodes in parse tree (${totalNodes} total nodes)`, 'yellow'));
    } else {
        console.log(colorize(`✅ Clean parse tree (${totalNodes} nodes, no errors)`, 'green'));
    }
    console.log('');

    // Generate output based on format
    if (options.html) {
        const htmlOutput = renderHTMLTree(tree, code, options);

        if (options.output) {
            fs.writeFileSync(options.output, htmlOutput);
            console.log(colorize(`✅ HTML output saved to: ${options.output}`, 'green'));
        } else {
            console.log(htmlOutput);
        }
    } else if (options.json) {
        const jsonOutput = JSON.stringify(tree, null, 2);

        if (options.output) {
            fs.writeFileSync(options.output, jsonOutput);
            console.log(colorize(`✅ JSON output saved to: ${options.output}`, 'green'));
        } else {
            console.log(jsonOutput);
        }
    } else {
        // Terminal output
        const terminalOutput = renderTerminalTree(tree, options);

        if (options.output) {
            // Strip ANSI colors for file output
            const cleanOutput = terminalOutput.replace(/\x1b\[[0-9;]*m/g, '');
            fs.writeFileSync(options.output, cleanOutput);
            console.log(colorize(`✅ Terminal output saved to: ${options.output}`, 'green'));
        } else {
            console.log(colorize('Parse Tree:', 'bright'));
            console.log(colorize('-'.repeat(30), 'cyan'));
            console.log(terminalOutput);
        }
    }

    // Show error analysis if there are errors
    if (errorCount > 0 && !options.errorsOnly) {
        console.log(colorize('🔍 Error Analysis:', 'bright'));
        console.log(colorize('-'.repeat(30), 'cyan'));
        console.log(`Found ${errorCount} ERROR nodes. Use --errors-only to focus on error contexts.`);
        console.log('Consider checking:');
        console.log('  • Grammar completeness for type annotations');
        console.log('  • Scanner token recognition');
        console.log('  • Syntax variations in input code');
    }
}

function interactiveMode(options = {}) {
    const readline = require('readline');
    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout
    });

    console.log(colorize('🌳 Interactive Parse Tree Visualizer', 'bright'));
    console.log('Enter Perl code to visualize, or commands:');
    console.log('  :help    - Show help');
    console.log('  :quit    - Exit');
    console.log('  :html    - Toggle HTML output');
    console.log('  :errors  - Toggle errors-only mode');
    console.log('  :compact - Toggle compact format');
    console.log('');

    function prompt() {
        rl.question(colorize('visualize> ', 'cyan'), (input) => {
            input = input.trim();

            if (input === ':quit' || input === ':q') {
                rl.close();
                return;
            }

            if (input === ':help' || input === ':h') {
                console.log('\nCommands:');
                console.log('  :help, :h     - Show this help');
                console.log('  :quit, :q     - Exit interactive mode');
                console.log('  :html         - Toggle HTML output mode');
                console.log('  :errors, :e   - Toggle errors-only display');
                console.log('  :compact, :c  - Toggle compact tree format');
                console.log('  :status, :s   - Show current options');
                console.log('');
                prompt();
                return;
            }

            if (input === ':html') {
                options.html = !options.html;
                console.log(colorize(`HTML output: ${options.html ? 'ON' : 'OFF'}`, 'yellow'));
                prompt();
                return;
            }

            if (input === ':errors' || input === ':e') {
                options.errorsOnly = !options.errorsOnly;
                console.log(colorize(`Errors-only mode: ${options.errorsOnly ? 'ON' : 'OFF'}`, 'yellow'));
                prompt();
                return;
            }

            if (input === ':compact' || input === ':c') {
                options.compact = !options.compact;
                console.log(colorize(`Compact format: ${options.compact ? 'ON' : 'OFF'}`, 'yellow'));
                prompt();
                return;
            }

            if (input === ':status' || input === ':s') {
                console.log(colorize('Current options:', 'bright'));
                console.log(`  HTML output: ${options.html ? 'ON' : 'OFF'}`);
                console.log(`  Errors-only: ${options.errorsOnly ? 'ON' : 'OFF'}`);
                console.log(`  Compact: ${options.compact ? 'ON' : 'OFF'}`);
                prompt();
                return;
            }

            if (input === '') {
                prompt();
                return;
            }

            // Visualize the input
            visualizeCode(input, options);
            prompt();
        });
    }

    prompt();
}

// Main execution
function main() {
    const args = process.argv.slice(2);

    if (args.length === 0 || args.includes('--help')) {
        usage();
    }

    // Parse options
    const options = {
        html: args.includes('--html') || args.includes('-h'),
        json: args.includes('--json') || args.includes('-j'),
        compact: args.includes('--compact') || args.includes('-c'),
        errorsOnly: args.includes('--errors-only') || args.includes('-e'),
        noColor: args.includes('--no-color'),
        maxDepth: Infinity,
        indent: 2,
        output: null
    };

    // Parse depth option
    const depthIndex = Math.max(args.indexOf('--depth'), args.indexOf('-d'));
    if (depthIndex !== -1 && depthIndex + 1 < args.length) {
        const depth = parseInt(args[depthIndex + 1]);
        if (!isNaN(depth) && depth > 0) {
            options.maxDepth = depth;
        }
    }

    // Parse indent option
    const indentIndex = args.indexOf('--indent');
    if (indentIndex !== -1 && indentIndex + 1 < args.length) {
        const indent = parseInt(args[indentIndex + 1]);
        if (!isNaN(indent) && indent >= 0) {
            options.indent = indent;
        }
    }

    // Parse output option
    const outputIndex = Math.max(args.indexOf('--output'), args.indexOf('-o'));
    if (outputIndex !== -1 && outputIndex + 1 < args.length) {
        options.output = args[outputIndex + 1];
    }

    // Handle interactive mode
    if (args.includes('-i') || args.includes('--interactive')) {
        interactiveMode(options);
        return;
    }

    // Handle file input
    const fileIndex = Math.max(args.indexOf('-f'), args.indexOf('--file'));
    if (fileIndex !== -1) {
        if (fileIndex + 1 >= args.length) {
            console.error(colorize('Error: No file specified after -f/--file', 'red'));
            process.exit(1);
        }

        const filename = args[fileIndex + 1];
        if (!fs.existsSync(filename)) {
            console.error(colorize(`Error: File not found: ${filename}`, 'red'));
            process.exit(1);
        }

        const content = fs.readFileSync(filename, 'utf8');
        visualizeCode(content, options);
        return;
    }

    // Handle direct code input
    const codeArgs = args.filter((arg, index) => {
        // Skip flags
        if (arg.startsWith('-')) return false;

        // Skip values that follow option flags
        if (index > 0) {
            const prevArg = args[index - 1];
            if (['-d', '--depth', '--indent', '-o', '--output', '-f', '--file'].includes(prevArg)) {
                return false;
            }
        }

        return true;
    });

    if (codeArgs.length === 0) {
        console.error(colorize('Error: No code provided to visualize', 'red'));
        usage();
    }

    const code = codeArgs.join(' ');
    visualizeCode(code, options);
}

// Check if tree-sitter is available
try {
    execSync('tree-sitter --version', { stdio: 'ignore' });
} catch (error) {
    console.error(colorize('Error: tree-sitter CLI not found', 'red'));
    console.error('Please install tree-sitter CLI: npm install -g tree-sitter-cli');
    process.exit(1);
}

main();
