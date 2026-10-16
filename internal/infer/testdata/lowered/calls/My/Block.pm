package My::Block;
use Exporter 'import';
our @EXPORT = qw(block);
sub block($&) { 1 }
1;
