/* ABOUTME: Toy grammar refereeing the precedence constraint's lookaheads: LALR merges the state after X W and Y W; e's lookahead T comes only from X's context. Canonical LR(1) would accept Y W T Q; bison's LALR refuses it. */
/* ABOUTME: build.sh builds it with bison; sentences up to 7 terminals. From PRECTOY's gate pass 3. */
%token X Y Z W Q
%left T
%left HIGH
%expect 0
%%
s	: X g T Z
	| Y g
	;
g	: e
	| W T Q
	;
e	: W %prec HIGH
	;
