# ABOUTME: Sample for tdt-capture.sh: a stub module use, a variable, interpolation, a => word and a heredoc.
# ABOUTME: Its expected capture is tdt-sample.expected; the stub lives in tdt-stubs/.
use FindBin;
use lib "$FindBin::Bin/tdt-stubs";
use TdtStub;
my $name = 'world';
my %h = (greeting => qq{hello, $name});
print <<"EOT";
$h{greeting}
EOT
sub later { require TdtLater; TdtLater::go() }
