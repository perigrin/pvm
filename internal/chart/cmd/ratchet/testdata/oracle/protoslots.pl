sub ar(\@) { } sub hr(\%) { } sub br(\[@%]) { } sub pl(+) { } sub st(*) { } sub two(\@\@) { } sub cb(&\@) { }
my (@a, %h, $s);
ar(@a);
hr %h;
br(@a); br(%h);
pl(@a); pl($s);
st(*STDOUT); st(STDOUT);
two(@a, @a);
cb { 1 } @a;
&ar(\@a);
