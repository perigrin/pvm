// ABOUTME: Built-in mappings for common Perl tools to their CPAN modules
// ABOUTME: Provides comprehensive set of well-known tool mappings with validation
package tool

// BuiltinMappings contains all built-in tool-to-module mappings
var BuiltinMappings = map[string]ToolMappingInfo{
	// Code Quality Tools
	"ack": {
		Module:      "App::Ack",
		Description: "Grep-like text finder optimized for programmers",
		Category:    "search",
		Executable:  "ack",
	},
	"perlcritic": {
		Module:      "Perl::Critic",
		Description: "Critique Perl source code for best-practices",
		Category:    "quality",
		Executable:  "perlcritic",
	},
	"perltidy": {
		Module:      "Perl-Tidy",
		Description: "Indent and reformat Perl scripts",
		Category:    "formatting",
		Executable:  "perltidy",
	},

	// CPAN Tools
	"cpanm": {
		Module:      "App::cpanminus",
		Description: "Get, unpack, build and install modules from CPAN",
		Category:    "cpan",
		Executable:  "cpanm",
	},
	"cpan-upload": {
		Module:      "CPAN::Uploader",
		Description: "Upload distributions to CPAN",
		Category:    "cpan",
		Executable:  "cpan-upload",
	},
	"cpan-audit": {
		Module:      "CPAN::Audit",
		Description: "Audit CPAN modules for known vulnerabilities",
		Category:    "security",
		Executable:  "cpan-audit",
	},
	"metacpan": {
		Module:      "MetaCPAN::Client",
		Description: "Comprehensive API for MetaCPAN",
		Category:    "cpan",
		Executable:  "metacpan",
	},

	// Testing Tools
	"prove": {
		Module:      "Test::Harness",
		Description: "Run Perl standard test scripts with statistics",
		Category:    "testing",
		Executable:  "prove",
	},

	// Build and Distribution Tools
	"fatpack": {
		Module:      "App::FatPacker",
		Description: "Pack dependencies into single file scripts",
		Category:    "packaging",
		Executable:  "fatpack",
	},
	"dzil": {
		Module:      "Dist::Zilla",
		Description: "Maximum overkill for CPAN distributions",
		Category:    "packaging",
		Executable:  "dzil",
	},
	"minil": {
		Module:      "Minilla",
		Description: "CPAN module authoring tool",
		Category:    "packaging",
		Executable:  "minil",
	},

	// Dependency Management
	"cpanfile": {
		Module:      "Module::CPANfile",
		Description: "Parse cpanfile format",
		Category:    "dependencies",
		Executable:  "cpanfile-dump",
	},
	"carton": {
		Module:      "Carton",
		Description: "Perl module dependency manager",
		Category:    "dependencies",
		Executable:  "carton",
	},

	// Web Development
	"plackup": {
		Module:      "Plack",
		Description: "Perl Superglue for Web frameworks and Web Servers",
		Category:    "web",
		Executable:  "plackup",
	},
	"morbo": {
		Module:      "Mojolicious",
		Description: "Morbo HTTP and WebSocket development server",
		Category:    "web",
		Executable:  "morbo",
	},
	"hypnotoad": {
		Module:      "Mojolicious",
		Description: "Hypnotoad HTTP and WebSocket server",
		Category:    "web",
		Executable:  "hypnotoad",
	},

	// Utility Tools
	"pmversions": {
		Module:      "Perl::Version",
		Description: "Parse and manipulate Perl version strings",
		Category:    "utility",
		Executable:  "pmversions",
	},
	"reply": {
		Module:      "Reply",
		Description: "Read-eval-print-loop for Perl",
		Category:    "utility",
		Executable:  "reply",
	},
	"re-pl": {
		Module:      "Devel::REPL",
		Description: "Modern Perl interactive shell",
		Category:    "utility",
		Executable:  "re.pl",
	},

	// Documentation Tools
	"pod2usage": {
		Module:      "Pod::Usage",
		Description: "Print usage messages from embedded POD docs",
		Category:    "documentation",
		Executable:  "pod2usage",
	},
	"podchecker": {
		Module:      "Pod::Checker",
		Description: "Check POD documents for syntax errors",
		Category:    "documentation",
		Executable:  "podchecker",
	},
	"pod2html": {
		Module:      "Pod::Html",
		Description: "Convert POD files to HTML",
		Category:    "documentation",
		Executable:  "pod2html",
	},
	"pod2man": {
		Module:      "Pod::Man",
		Description: "Convert POD data to formatted *roff input",
		Category:    "documentation",
		Executable:  "pod2man",
	},
	"pod2text": {
		Module:      "Pod::Text",
		Description: "Convert POD data to formatted ASCII text",
		Category:    "documentation",
		Executable:  "pod2text",
	},
}

// GetBuiltinMapping returns the mapping info for a built-in tool
func GetBuiltinMapping(toolName string) (ToolMappingInfo, bool) {
	mapping, exists := BuiltinMappings[toolName]
	return mapping, exists
}

// ListBuiltinTools returns all built-in tool names
func ListBuiltinTools() []string {
	tools := make([]string, 0, len(BuiltinMappings))
	for tool := range BuiltinMappings {
		tools = append(tools, tool)
	}
	return tools
}

// GetToolsByCategory returns tools in a specific category
func GetToolsByCategory(category string) []string {
	var tools []string
	for tool, info := range BuiltinMappings {
		if info.Category == category {
			tools = append(tools, tool)
		}
	}
	return tools
}

// GetAllCategories returns all available tool categories
func GetAllCategories() []string {
	categories := make(map[string]bool)
	for _, info := range BuiltinMappings {
		categories[info.Category] = true
	}

	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	return result
}
