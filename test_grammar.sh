#!/bin/bash
cd /home/perigrin/dev/pvm/tree-sitter-typed-perl
echo "Testing tree-sitter grammar generation..."
tree-sitter generate
echo "Grammar generation result: $?"
