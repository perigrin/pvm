use feature 'try';
$y = 0;
L: try { $y = 1 } catch ($e) { $y = 2 }
