/* ABOUTME: Toy grammar refereeing the precedence constraint: a %left/%right operator ladder. */
/* ABOUTME: build.sh builds it with bison; sentences up to 7 terminals. */
%token NUM
%right ASSIGN
%left PLUS MINUS
%left STAR
%right POW
%expect 0
%%
e	: e ASSIGN e
	| e PLUS e
	| e MINUS e
	| e STAR e
	| e POW e
	| NUM
	;
