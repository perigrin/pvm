#include "tree_sitter/parser.h"
#include <stdlib.h>
#include <string.h>
#include <wctype.h>
#include <stdbool.h>

// Token types corresponding to externals in grammar.js
// These MUST match the exact order and index in the externals array!
enum TokenType {
  // Order from grammar.js externals starting at index 0:
  SINGLE_QUOTE,                 // 0:  $._single_quote
  DOUBLE_QUOTE,                 // 1:  $._double_quote
  BACKTICK_QUOTE,               // 2:  $._backtick_quote
  SEARCH_SLASH_QUOTE,           // 3:  $._search_slash_quote
  NO_SEARCH_SLASH_PLZ,          // 4:  $._no_search_slash_plz
  OPEN_READLINE_BRACKET,        // 5:  $._open_readline_bracket
  OPEN_FILEGLOB_BRACKET,        // 6:  $._open_fileglob_bracket
  PERLY_SEMICOLON,              // 7:  $._PERLY_SEMICOLON
  PERLY_HEREDOC,                // 8:  $._PERLY_HEREDOC
  CTRL_Z_HACK,                  // 9:  $._ctrl_z_hack
  QUOTELIKE_BEGIN_QUOTE,        // 10: $._quotelike_begin_quote
  QUOTELIKE_MIDDLE_CLOSE_QUOTE, // 11: $._quotelike_middle_close_quote
  QUOTELIKE_MIDDLE_SKIP,        // 12: $._quotelike_middle_skip
  QUOTELIKE_END_ZW,             // 13: $._quotelike_end_zw
  QUOTELIKE_END_QUOTE,          // 14: $._quotelike_end_quote
  Q_STRING_CONTENT,             // 15: $._q_string_content
  QQ_STRING_CONTENT,            // 16: $._qq_string_content
  ESCAPE_SEQUENCE,              // 17: $.escape_sequence
  ESCAPED_DELIMITER,            // 18: $.escaped_delimiter
  DOLLAR_IN_REGEXP,             // 19: $._dollar_in_regexp
  POD,                          // 20: $.pod
  GOBBLED_CONTENT,              // 21: $._gobbled_content
  ATTRIBUTE_VALUE_BEGIN,        // 22: $._attribute_value_begin
  ATTRIBUTE_VALUE,              // 23: $.attribute_value
  PROTOTYPE,                    // 24: $.prototype
  SIGNATURE_START,              // 25: $._signature_start
  HEREDOC_DELIMITER,            // 26: $._heredoc_delimiter
  COMMAND_HEREDOC_DELIMITER,    // 27: $._command_heredoc_delimiter
  HEREDOC_START,                // 28: $._heredoc_start
  HEREDOC_MIDDLE,               // 29: $._heredoc_middle
  HEREDOC_END,                  // 30: $.heredoc_end
  FAT_COMMA_AUTOQUOTED,         // 31: $._fat_comma_autoquoted
  FILETEST,                     // 32: $._filetest
  BRACE_AUTOQUOTED_TOKEN,       // 33: $._brace_autoquoted_token
  BRACE_END_ZW,                 // 34: $._brace_end_zw
  DOLLAR_IDENT_ZW,              // 35: $._dollar_ident_zw
  NO_INTERP_WHITESPACE_ZW,      // 36: $._no_interp_whitespace_zw
  NONASSOC,                     // 37: $._NONASSOC
  ERROR_TOKEN                   // 38: $._ERROR
};

// Scanner state structure
typedef struct {
  // State for tracking context
  bool in_signature;
  bool in_prototype;
  bool in_heredoc;
  bool in_quotelike;

  // Quote state
  char quote_char;
  char delimiter_stack[32];
  int delimiter_depth;

  // Context tracking
  int paren_depth;
  int brace_depth;
} ScannerState;

void *tree_sitter_typed_perl_external_scanner_create() {
  ScannerState *state = calloc(1, sizeof(ScannerState));
  return state;
}

void tree_sitter_typed_perl_external_scanner_destroy(void *payload) {
  free(payload);
}

unsigned tree_sitter_typed_perl_external_scanner_serialize(void *payload, char *buffer) {
  ScannerState *state = (ScannerState *)payload;
  if (!state) return 0;

  memcpy(buffer, state, sizeof(ScannerState));
  return sizeof(ScannerState);
}

void tree_sitter_typed_perl_external_scanner_deserialize(void *payload, const char *buffer, unsigned length) {
  ScannerState *state = (ScannerState *)payload;
  if (!state || length != sizeof(ScannerState)) return;

  memcpy(state, buffer, sizeof(ScannerState));
}

// Helper function to skip whitespace
static bool skip_whitespace(TSLexer *lexer) {
  while (iswspace(lexer->lookahead)) {
    lexer->advance(lexer, true);
  }
  return true;
}

// Check if we're in a signature context
static bool is_signature_context(TSLexer *lexer) {
  // The parser will only ask for SIGNATURE_START when it's expecting one
  // in the context of a subroutine or method declaration.
  // So if we're being asked for it, we can trust that we're in the right context.
  // The main job is just to distinguish a signature paren from other parens.

  // For signatures, we typically see:
  // sub name(...) or method name(...)
  // The parser tracks this context, so we can be liberal here.

  return true; // Trust the parser's context tracking
}

// Scan for signature start token
static bool scan_signature_start(TSLexer *lexer, ScannerState *state) {
  skip_whitespace(lexer);

  if (lexer->lookahead != '(') {
    return false;
  }

  // Check if this looks like a signature context
  if (!is_signature_context(lexer)) {
    return false;
  }

  // Mark that we found a signature start
  state->in_signature = true;
  state->paren_depth = 1;

  // Consume the opening paren
  lexer->advance(lexer, false);
  lexer->mark_end(lexer);

  return true;
}

// Scan for prototype token
static bool scan_prototype(TSLexer *lexer, ScannerState *state) {
  skip_whitespace(lexer);

  if (lexer->lookahead != '(') {
    return false;
  }

  // Look for prototype patterns like (), ($), (@), etc.
  // This is a simplified implementation
  lexer->advance(lexer, false);

  // Scan prototype content
  while (lexer->lookahead && lexer->lookahead != ')') {
    if (strchr("$@%&*;[]\\", lexer->lookahead)) {
      lexer->advance(lexer, false);
    } else if (iswspace(lexer->lookahead)) {
      lexer->advance(lexer, true);
    } else {
      return false; // Invalid prototype character
    }
  }

  if (lexer->lookahead == ')') {
    lexer->advance(lexer, false);
    lexer->mark_end(lexer);
    return true;
  }

  return false;
}

bool tree_sitter_typed_perl_external_scanner_scan(void *payload, TSLexer *lexer, const bool *valid_symbols) {
  ScannerState *state = (ScannerState *)payload;
  if (!state) return false;

  // Debug: check if we're being asked for signature tokens
  if (valid_symbols[SIGNATURE_START]) {
    if (scan_signature_start(lexer, state)) {
      lexer->result_symbol = SIGNATURE_START;
      return true;
    }
  }

  if (valid_symbols[PROTOTYPE]) {
    if (scan_prototype(lexer, state)) {
      lexer->result_symbol = PROTOTYPE;
      return true;
    }
  }

  // For now, we only implement the critical tokens needed for signature parsing
  // Other external tokens would be implemented here as needed

  return false;
}
