/* ABOUTME: Toy grammar refereeing the precedence constraint's lookaheads: An empty rule's reduction beats a shift: opt reduces before LB. */
/* ABOUTME: build.sh builds it with bison; sentences up to 7 terminals. From PRECTOY's gate pass 3. */
%token VAR NUM LB RB
%nonassoc LB
%nonassoc HIGH
%expect 0
%%
s	: VAR opt LB NUM RB
	| VAR opt
	;
opt	: %prec HIGH
	| LB NUM RB
	;
