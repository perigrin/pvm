/* ABOUTME: Toy grammar refereeing the precedence constraint: a ladder, %nonassoc chains, prefix and postfix operators and indirob together. */
/* ABOUTME: build.sh builds it with bison; sentences up to 5 terminals. */
%token NUM VAR SNAIL LSTOP MINUS LB RB COMMA
%nonassoc PREC_LOW
%nonassoc LSTOP
%left COMMA
%right ASSIGN
%nonassoc DOTDOT
%left PLUS
%right UMINUS
%nonassoc PREINC POSTINC
%left PAREN_OPEN
%left LB
%expect 0
%%
expr	: listop
	| term
	;
listop	: LSTOP indirob listexpr
	| LSTOP listexpr
	;
listexpr	: term
	| listexpr COMMA term
	;
term	: term ASSIGN term
	| term DOTDOT term
	| term PLUS term
	| MINUS term %prec UMINUS
	| PREINC term
	| term POSTINC
	| scalar %prec PAREN_OPEN
	| ary %prec PAREN_OPEN
	| scalar LB expr RB
	| sliceme LB expr RB
	| LB expr RB
	| LB RB
	| NUM
	;
sliceme	: ary
	;
ary	: SNAIL indirob
	;
scalar	: VAR
	;
indirob	: scalar %prec PREC_LOW
	;
