#include "tree_sitter/parser.h"

enum TokenType {
  // Add token types from externals section
  SINGLE_QUOTE,
  DOUBLE_QUOTE,
  BACKTICK_QUOTE,
  SEARCH_SLASH_QUOTE,
  NO_SEARCH_SLASH_PLZ,
  OPEN_READLINE_BRACKET,
  OPEN_FILEGLOB_BRACKET,
  PERLY_SEMICOLON,
  PERLY_HEREDOC,
  CTRL_Z_HACK,
  QUOTELIKE_BEGIN_QUOTE,
  QUOTELIKE_MIDDLE_CLOSE_QUOTE,
  QUOTELIKE_MIDDLE_SKIP,
  QUOTELIKE_END_ZW,
  QUOTELIKE_END_QUOTE,
  Q_STRING_CONTENT,
};

void *tree_sitter_typed_perl_external_scanner_create() {
  return NULL;
}

void tree_sitter_typed_perl_external_scanner_destroy(void *payload) {
  // No-op
}

unsigned tree_sitter_typed_perl_external_scanner_serialize(void *payload, char *buffer) {
  return 0;
}

void tree_sitter_typed_perl_external_scanner_deserialize(void *payload, const char *buffer, unsigned length) {
  // No-op
}

bool tree_sitter_typed_perl_external_scanner_scan(void *payload, TSLexer *lexer, const bool *valid_symbols) {
  return false;
}