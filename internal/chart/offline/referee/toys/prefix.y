/* ABOUTME: Toy grammar refereeing the precedence constraint: prefix operators, doubled, and a prefix and a postfix at one %nonassoc level. */
/* ABOUTME: build.sh builds it with bison; sentences up to 6 terminals. MINUS takes its level from %prec UMINUS, as perly.y's does. */
%token NUM MINUS
%left PLUS
%right BANG UMINUS
%nonassoc PREINC POSTINC
%expect 0
%%
e	: e PLUS e
	| MINUS e %prec UMINUS
	| BANG e
	| PREINC e
	| e POSTINC
	| NUM
	;
