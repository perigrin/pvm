/* ABOUTME: Toy grammar refereeing the precedence constraint's lookaheads: Lookahead through a chain of predictions with nullable symbols between: c inside d inside a. */
/* ABOUTME: build.sh builds it with bison; sentences up to 7 terminals. From PRECTOY's gate pass 3. */
%token X Y Z W Q R P
%left T
%left HIGH
%expect 0
%%
s	: X a T Z
	| Y a
	| R
	;
a	: d n2
	;
n	:
	| Q
	;
d	: c n m
	;
c	: W %prec HIGH
	| W T R
	;
n2	:
	| P
	;
m	:
	| Z Z
	;
