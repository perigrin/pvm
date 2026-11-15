/* ABOUTME: Toy grammar refereeing the precedence constraint: two parser states after one symbol, R, only one with a conflict. */
/* ABOUTME: build.sh builds it with bison; sentences up to 6 terminals. ind's %prec PREC_LOW meets no LB shift after P R. */
%token P Q R VAR NUM LB RB
%nonassoc PREC_LOW
%left LB
%expect 0
%%
s	: P R ind LB NUM RB
	| Q R elem
	;
ind	: VAR %prec PREC_LOW
	;
elem	: VAR LB NUM RB
	| VAR
	;
