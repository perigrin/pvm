/* ABOUTME: Toy grammar refereeing the precedence constraint: empty rules, one carrying %prec against a shifted LB. */
/* ABOUTME: build.sh builds it with bison; sentences up to 7 terminals. An empty rule reduces at the file's start and after VAR. */
%token VAR NUM LB RB
%nonassoc LOW
%nonassoc LB
%expect 0
%%
s	: VAR opt LB NUM RB
	| VAR opt
	| lead RB
	;
opt	: %prec LOW
	| LB NUM RB
	;
lead	:
	| NUM
	;
