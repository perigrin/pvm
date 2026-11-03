use feature 'defer';
$y = 0;
if ($y) { defer { $y = 1 } $y = 2; }
