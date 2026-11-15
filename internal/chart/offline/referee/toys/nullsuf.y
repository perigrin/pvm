/* ABOUTME: Toy grammar refereeing the precedence constraint's lookaheads: b's lookahead T reaches it only through a's nullable suffix opt opt2. */
/* ABOUTME: build.sh builds it with bison; sentences up to 7 terminals. From PRECTOY's gate pass 3. */
%token X Y Z W Q R P
%left T
%left HIGH
%expect 0
%%
s	: X a T Z
	| Y a
	;
a	: b opt opt2
	;
opt	:
	| Q
	;
opt2	:
	| P
	;
b	: W %prec HIGH
	| W T R
	;
