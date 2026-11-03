use feature 'try';
$y = 0;
try { try { $y = 1 } catch ($e) { $y = 2 } } catch ($f) { $y = 3 }
