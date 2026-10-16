# Control: a string literal names a handle nothing opened, and perl warns.
use strict;
use warnings;

seek("FH", 0, 0);
