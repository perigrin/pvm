/* ABOUTME: Toy grammar refereeing the precedence constraint: the indirob toy, and LSTOP after SORT, where no conflict on LB arises. */
/* ABOUTME: build.sh builds it with bison; sentences up to 6 terminals. The state after SORT LSTOP differs from the one after LSTOP alone. */
%token LSTOP SNAIL VAR NUM LB RB COMMA SORT
%nonassoc PREC_LOW
%nonassoc LSTOP
%left COMMA
%left PAREN_OPEN
%left LB
%expect 0
%%
expr	: listop
	| term
	| SORT LSTOP indirob LB expr RB
	;
listop	: LSTOP indirob listexpr
	| LSTOP listexpr
	;
listexpr	: term
	| listexpr COMMA term
	;
term	: scalar %prec PAREN_OPEN
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
