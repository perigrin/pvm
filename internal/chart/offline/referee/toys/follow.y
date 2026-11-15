/* ABOUTME: Toy grammar refereeing the precedence constraint: a reduction whose FOLLOW holds T where its LALR lookahead does not. */
/* ABOUTME: build.sh builds it with bison; sentences up to 6 terminals. After Y W, e: W reduces only on the end of input. */
%token X Y Z W
%left T
%left HIGH
%expect 0
%%
s	: X e T Z
	| Y e
	| Y f
	;
e	: W %prec HIGH
	;
f	: W T Z
	;
