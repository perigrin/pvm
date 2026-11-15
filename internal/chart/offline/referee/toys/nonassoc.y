/* ABOUTME: Toy grammar refereeing the precedence constraint: %nonassoc chains, adjacent and split by a tighter operator. */
/* ABOUTME: build.sh builds it with bison; sentences up to 7 terminals. */
%token NUM
%left OR
%nonassoc DOTDOT
%nonassoc LT
%left PLUS
%expect 0
%%
e	: e OR e
	| e DOTDOT e
	| e LT e
	| e PLUS e
	| NUM
	;
