/* ABOUTME: Toy grammar refereeing the precedence constraint: %prec on a rule that is not an operator, perly.y's indirob: scalar %prec PREC_LOW. */
/* ABOUTME: build.sh builds it with bison; sentences up to 6 terminals. The conflict on LB arises after LSTOP and not after SNAIL. */
%token LSTOP SNAIL VAR NUM LB RB COMMA
%nonassoc PREC_LOW
%nonassoc LSTOP
%left COMMA
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
