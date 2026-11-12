// ABOUTME: The chart grammar's data file: perly.y's rules, one per line, each with the alternative it cites.
// ABOUTME: One placeholder per perly.y alternative, in perly.y order; a family replaces its own with the rules.

package grammar

import (
	"slices"

	"tamarou.com/pvm/internal/lexer"
)

// perlyRules is `lhs: rhs # perly.y:N` per rule, or `lhs: rhs # perly.y:N, fitted: <what the lexer
// emits instead>`; a rule whose alternative carries a %prec ends its right-hand side with `%prec TOKEN`, as
// perly.y does. A line starting with # is a placeholder for the alternative it names. A blank line
// separates adjacent placeholders and rules (and ledger entries), so two families' neighbouring edits
// merge cleanly; only rules citing the same perly.y:N, one family's slot, may be adjacent.
const perlyRules = `
grammar: remember stmtseq # perly.y:187, fitted: GRAMPROG is a start selector, not a token

grammar: GRAMEXPR optexpr # perly.y:198

grammar: GRAMBLOCK block # perly.y:208

grammar: GRAMBARESTMT barestmt # perly.y:221

grammar: GRAMFULLSTMT fullstmt # perly.y:234

grammar: GRAMSTMTSEQ stmtseq # perly.y:247

grammar: GRAMSUBSIGNATURE subsigguts # perly.y:257

bare_statement_block: block cont # perly.y:270

bare_statement_class_declaration: KW_CLASS BAREWORD optattrlist PERLY_SEMICOLON # perly.y:278, fitted: toke.c forces an empty version token before the name, and pvm emits nothing for it
bare_statement_class_declaration: KW_CLASS BAREWORD BAREWORD optattrlist PERLY_SEMICOLON # perly.y:278, fitted: pvm emits the name and then the version, which toke.c forces as two BAREWORDs in the other order

bare_statement_class_definition: KW_CLASS BAREWORD optattrlist PERLY_BRACE_OPEN remember stmtseq PERLY_BRACE_CLOSE # perly.y:294, fitted: toke.c forces an empty version token before the name, and pvm emits nothing for it
bare_statement_class_definition: KW_CLASS BAREWORD BAREWORD optattrlist PERLY_BRACE_OPEN remember stmtseq PERLY_BRACE_CLOSE # perly.y:294, fitted: pvm emits the name and then the version, which toke.c forces as two BAREWORDs in the other order

bare_statement_default: KW_DEFAULT block # perly.y:317

bare_statement_defer: KW_DEFER mblock # perly.y:325

bare_statement_expression: sideff PERLY_SEMICOLON # perly.y:333

bare_statement_field_declaration: fielddecl PERLY_SEMICOLON # perly.y:340

bare_statement_for: KW_FOR PERLY_PAREN_OPEN remember mnexpr PERLY_SEMICOLON texpr PERLY_SEMICOLON mintro mnexpr PERLY_PAREN_CLOSE mblock # perly.y:348

bare_statement_for: KW_FOR KW_MY remember my_scalar PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock cont # perly.y:379

bare_statement_for: KW_FOR KW_MY remember PERLY_PAREN_OPEN my_list_of_itervars PERLY_PAREN_CLOSE PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock cont # perly.y:392

bare_statement_for: KW_FOR scalar PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock cont # perly.y:411

bare_statement_for: KW_FOR my_refgen remember my_var PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock cont # perly.y:423

bare_statement_for: KW_FOR REFGEN refgen_topic PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock cont # perly.y:452

bare_statement_for: KW_FOR PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock cont # perly.y:474

bare_statement_format: KW_FORMAT startformsub formname formblock # perly.y:488

bare_statement_given: KW_GIVEN PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock # perly.y:504

bare_statement_if: KW_IF PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock else # perly.y:517

bare_statement_null: PERLY_SEMICOLON # perly.y:532

bare_statement_package_declaration: KW_PACKAGE BAREWORD PERLY_SEMICOLON # perly.y:540, fitted: toke.c forces an empty version token before the name, and pvm emits nothing for it
bare_statement_package_declaration: KW_PACKAGE BAREWORD BAREWORD PERLY_SEMICOLON # perly.y:540, fitted: pvm emits the name and then the version, which toke.c forces as two BAREWORDs in the other order

bare_statement_package_definition: KW_PACKAGE BAREWORD PERLY_BRACE_OPEN remember stmtseq PERLY_BRACE_CLOSE # perly.y:556, fitted: toke.c forces an empty version token before the name, and pvm emits nothing for it
bare_statement_package_definition: KW_PACKAGE BAREWORD BAREWORD PERLY_BRACE_OPEN remember stmtseq PERLY_BRACE_CLOSE # perly.y:556, fitted: pvm emits the name and then the version, which toke.c forces as two BAREWORDs in the other order

bare_statement_phaser: PHASER startsub optsubbody # perly.y:574

bare_statement_sub_signature: sigsub_or_method_named subname startsub subattrlist optsigsubbody # perly.y:608

bare_statement_sub_traditional: KW_SUB_named subname startsub proto subattrlist optsubbody # perly.y:638, fitted: as written; the ledger marks one line once, and the two rules after it are its fits
bare_statement_sub_traditional: KW_SUB_named startsub proto subattrlist optsubbody # perly.y:638, fitted: pvm emits BEGIN, END and their kin as one word where toke.c (toke.c:8315, really_sub) emits KW_SUB_named and the name
bare_statement_sub_traditional: KW_MY KW_SUB_named subname startsub proto subattrlist optsubbody # perly.y:638, fitted: pvm emits the my, our or state of a lexical sub, which toke.c reads silently (yyl_my calls yyl_sub, toke.c:7577)

bare_statement_try_catch: KW_TRY mblock KW_CATCH remember catch_paren mblock finally # perly.y:662, fitted: catch_paren: empty (perly.y:812) is excluded, so a catch has a variable

bare_statement_unless: KW_UNLESS PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock else # perly.y:684

bare_statement_until: KW_UNTIL PERLY_PAREN_OPEN remember iexpr PERLY_PAREN_CLOSE mintro mblock cont # perly.y:699

bare_statement_utilize: KW_USE_or_NO startsub BAREWORD optlistexpr PERLY_SEMICOLON # perly.y:714, fitted: toke.c forces an empty version token before the module name, and pvm emits nothing for it
bare_statement_utilize: KW_USE_or_NO startsub BAREWORD BAREWORD optlistexpr PERLY_SEMICOLON # perly.y:714, fitted: pvm emits the module and then its version, which toke.c forces first, in the other order, as two BAREWORDs; use VERSION; and no VERSION; are the rule above with the version as its BAREWORD

bare_statement_when: KW_WHEN PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock # perly.y:732

bare_statement_while: KW_WHILE PERLY_PAREN_OPEN remember texpr PERLY_PAREN_CLOSE mintro mblock cont # perly.y:744

bare_statement_yadayada: YADAYADA PERLY_SEMICOLON # perly.y:759

sigsub_or_method_named: KW_SUB_named_sig # perly.y:768, fitted: as written; the ledger marks one line once, and the next rule is its fit
sigsub_or_method_named: KW_MY KW_SUB_named_sig # perly.y:768, fitted: pvm emits the my, our or state of a lexical sub, which toke.c reads silently (yyl_my calls yyl_sub, toke.c:7577)

sigsub_or_method_named: KW_METHOD_named # perly.y:770, fitted: as written; the ledger marks one line once, and the next rule is its fit
sigsub_or_method_named: KW_MY KW_METHOD_named # perly.y:770, fitted: pvm emits the my, our or state of a lexical method, which toke.c replaces by passing KEY_method to yyl_sub (toke.c:7580)

block: PERLY_BRACE_OPEN remember stmtseq PERLY_BRACE_CLOSE # perly.y:775

empty: # perly.y:783

formblock: PERLY_EQUAL_SIGN remember FORMRBRACK # perly.y:787, fitted: pvm emits a format's picture lines, argument lines and closing . as one FormatBody token (toke.c's PERLY_SEMICOLON FORMRBRACK formstmtseq PERLY_SEMICOLON PERLY_DOT), so the token is FORMRBRACK and formstmtseq, formline and formarg are excluded folded-into-token

remember: # perly.y:794

mblock: PERLY_BRACE_OPEN mremember stmtseq PERLY_BRACE_CLOSE # perly.y:799

mremember: # perly.y:806

# perly.y:812 catch_paren: empty

catch_paren: PERLY_PAREN_OPEN scalar PERLY_PAREN_CLOSE # perly.y:815

stmtseq: empty # perly.y:825

stmtseq: stmtseq fullstmt # perly.y:826

# perly.y:836 formstmtseq: empty

# perly.y:837 formstmtseq: formstmtseq formline

fullstmt: barestmt # perly.y:846

fullstmt: labfullstmt # perly.y:850

labfullstmt: LABEL PERLY_COLON barestmt # perly.y:854, fitted: toke.c's LABEL token includes its colon, and pvm emits the colon as a token of its own

labfullstmt: LABEL PERLY_COLON labfullstmt # perly.y:861, fitted: toke.c's LABEL token includes its colon, and pvm emits the colon as a token of its own

# perly.y:878 barestmt: PLUGSTMT

barestmt: bare_statement_block # perly.y:879

barestmt: bare_statement_class_declaration # perly.y:880

barestmt: bare_statement_class_definition # perly.y:881

barestmt: bare_statement_default # perly.y:882

barestmt: bare_statement_defer # perly.y:883

barestmt: bare_statement_expression # perly.y:884

barestmt: bare_statement_field_declaration # perly.y:885

barestmt: bare_statement_for # perly.y:886

barestmt: bare_statement_format # perly.y:887

barestmt: bare_statement_given # perly.y:888

barestmt: bare_statement_if # perly.y:889

barestmt: bare_statement_null # perly.y:890

barestmt: bare_statement_package_declaration # perly.y:891

barestmt: bare_statement_package_definition # perly.y:892

barestmt: bare_statement_phaser # perly.y:893

barestmt: bare_statement_sub_signature # perly.y:894

barestmt: bare_statement_sub_traditional # perly.y:895

barestmt: bare_statement_try_catch # perly.y:896

barestmt: bare_statement_unless # perly.y:897

barestmt: bare_statement_until # perly.y:898

barestmt: bare_statement_utilize # perly.y:899

barestmt: bare_statement_when # perly.y:900

barestmt: bare_statement_while # perly.y:901

barestmt: bare_statement_yadayada # perly.y:902

# perly.y:906 formline: THING formarg

# perly.y:924 formarg: empty

# perly.y:925 formarg: FORMLBRACK stmtseq FORMRBRACK

condition: expr # perly.y:929

# perly.y:933 sideff: error

sideff: expr # perly.y:935

sideff: expr KW_IF condition # perly.y:937

sideff: expr KW_UNLESS condition # perly.y:940

sideff: expr KW_WHILE condition # perly.y:943

sideff: expr KW_UNTIL iexpr # perly.y:945

sideff: expr KW_FOR condition # perly.y:947

sideff: expr KW_WHEN condition # perly.y:950

else: empty # perly.y:956

else: KW_ELSE mblock # perly.y:957

else: KW_ELSIF PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock else # perly.y:962

cont: empty # perly.y:973

cont: KW_CONTINUE block # perly.y:974

finally: # perly.y:979

finally: KW_FINALLY block # perly.y:981

mintro: # perly.y:986

nexpr: empty # perly.y:993

nexpr: sideff # perly.y:994

texpr: # perly.y:998

texpr: expr # perly.y:1002

iexpr: expr # perly.y:1006

mexpr: expr # perly.y:1011

mnexpr: nexpr # perly.y:1015

formname: BAREWORD # perly.y:1019

formname: empty # perly.y:1020

startsub: # perly.y:1023

startanonsub: # perly.y:1029

startanonmethod: # perly.y:1034

startformsub: # perly.y:1039

subname: BAREWORD # perly.y:1045

subname: PRIVATEREF # perly.y:1046

proto: empty # perly.y:1051

proto: PROTOTYPE # perly.y:1052

subattrlist: empty # perly.y:1057

subattrlist: COLONATTR attr attrs # perly.y:1058, fitted: the attrlist fit (perly.y:1070): each attribute is a word with its argument, if any, in parentheses
subattrlist: COLONATTR attr attrs COLONATTR # perly.y:1058, fitted: the attrlist fit (perly.y:1070): perl accepts one colon after the attributes, which toke.c leaves out of its ATTRLIST

subattrlist: COLONATTR # perly.y:1065

attrlist: COLONATTR attr attrs # perly.y:1070, fitted: pvm emits each attribute as a word after its colon with its argument, if any, inside parentheses, where toke.c delivers one ATTRLIST; one token of argument
attrlist: COLONATTR attr attrs COLONATTR # perly.y:1070, fitted: perl accepts one colon after the attributes, which toke.c leaves out of its ATTRLIST
attr: ATTRLIST # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attr: ATTRLIST PROTOTYPE # perly.y:1070, fitted: pvm emits the argument of the prototype attribute as one Prototype token, as toke.c's opaque argument
attr: ATTRLIST PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attr: ATTRLIST PERLY_PAREN_OPEN ATTRLIST PERLY_PAREN_CLOSE # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attrs: empty # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attrs: attrs COLONATTR attr # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attrs: attrs attr # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST

attrlist: COLONATTR # perly.y:1072

optattrlist: empty # perly.y:1077

optattrlist: attrlist # perly.y:1078

sigvar: # perly.y:1087

# perly.y:1089 sigvar: PRIVATEREF

sigslurpsigil: PERLY_SNAIL # perly.y:1094

sigslurpsigil: PERLY_PERCENT_SIGN # perly.y:1096

sigslurpelem: sigslurpsigil sigvar # perly.y:1100

# perly.y:1105 sigslurpelem: sigslurpsigil sigvar ASSIGNOP

# perly.y:1109 sigslurpelem: sigslurpsigil sigvar ASSIGNOP term

optcolon: # perly.y:1115

optcolon: PERLY_COLON # perly.y:1117

sigscalarelem: optcolon PERLY_DOLLAR sigvar # perly.y:1123

sigscalarelem: optcolon PERLY_DOLLAR sigvar ASSIGNOP optsigscalardefault # perly.y:1133

optsigscalardefault: # perly.y:1146

optsigscalardefault: term # perly.y:1148

sigelem: sigscalarelem # perly.y:1153

sigelem: sigslurpelem # perly.y:1155

siglist: siglist PERLY_COMMA # perly.y:1163

siglist: siglist PERLY_COMMA sigelem # perly.y:1164

siglist: sigelem %prec PREC_LOW # perly.y:1165

optsiglist: empty # perly.y:1170

optsiglist: siglist # perly.y:1171

optsubsignature: empty # perly.y:1176

optsubsignature: subsignature # perly.y:1177

subsignature: PERLY_PAREN_OPEN subsigguts PERLY_PAREN_CLOSE # perly.y:1181

subsigguts: optsiglist # perly.y:1185

optsubbody: subbody # perly.y:1218

optsubbody: PERLY_SEMICOLON # perly.y:1219

subbody: remember PERLY_BRACE_OPEN stmtseq PERLY_BRACE_CLOSE # perly.y:1224

optsigsubbody: sigsubbody # perly.y:1236

optsigsubbody: PERLY_SEMICOLON # perly.y:1237

sigsubbody: remember optsubsignature PERLY_BRACE_OPEN stmtseq PERLY_BRACE_CLOSE # perly.y:1241

expr: expr ANDOP expr # perly.y:1254

expr: expr PLUGIN_LOGICAL_AND_LOW_OP expr # perly.y:1256

expr: expr OROP expr # perly.y:1258

expr: expr PLUGIN_LOGICAL_OR_LOW_OP expr # perly.y:1260

expr: listexpr %prec PREC_LOW # perly.y:1262

listexpr: listexpr PERLY_COMMA # perly.y:1266

listexpr: listexpr PERLY_COMMA term # perly.y:1268

listexpr: term %prec PREC_LOW # perly.y:1273

listop: LSTOP indirob listexpr # perly.y:1277

listop: BLKLSTOP block listexpr # perly.y:1281

listop: FUNC PERLY_PAREN_OPEN indirob expr PERLY_PAREN_CLOSE # perly.y:1285

listop: term ARROW methodname PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE # perly.y:1289

listop: term ARROW methodname # perly.y:1295

listop: term ARROW PERLY_AMPERSAND subname PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE # perly.y:1300

listop: term ARROW PERLY_AMPERSAND subname # perly.y:1306

listop: METHCALL0 indirob optlistexpr # perly.y:1312

listop: METHCALL indirob PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE # perly.y:1318

listop: LSTOP optlistexpr # perly.y:1324

listop: FUNC PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE # perly.y:1326

listop: FUNC # perly.y:1328, fitted: pvm emits a string, command, heredoc opener or readline and its interior as one Quote, HeredocOpen or Readline token (toke.c's FUNC SUBLEXSTART optexpr SUBLEXEND), which stands for them all as FUNC

listop: LSTOPSUB startanonsub block optlistexpr %prec LSTOP # perly.y:1330

methodname: METHCALL0 # perly.y:1344

methodname: scalar # perly.y:1345

subscripted: gelem PERLY_BRACE_OPEN expr PERLY_BRACE_CLOSE # perly.y:1349, fitted: the ; before } is toke.c's, not the source's

subscripted: scalar PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1353

subscripted: term ARROW PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1356

subscripted: subscripted PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1361

subscripted: scalar PERLY_BRACE_OPEN expr PERLY_BRACE_CLOSE # perly.y:1366, fitted: the ; before } is toke.c's, not the source's

subscripted: term ARROW PERLY_BRACE_OPEN expr PERLY_BRACE_CLOSE # perly.y:1369, fitted: the ; before } is toke.c's, not the source's

subscripted: subscripted PERLY_BRACE_OPEN expr PERLY_BRACE_CLOSE # perly.y:1373, fitted: the ; before } is toke.c's, not the source's

subscripted: term ARROW PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1377

subscripted: term ARROW PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1383

subscripted: subscripted PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1391

subscripted: subscripted PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1398

subscripted: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1404

subscripted: QWLIST PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1406

subscripted: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1408

termbinop: term PLUGIN_HIGH_OP term # perly.y:1413

termbinop: term ASSIGNOP term # perly.y:1415

termbinop: term PLUGIN_ASSIGN_OP term # perly.y:1417

termbinop: term POWOP term # perly.y:1419

termbinop: term PLUGIN_POW_OP term # perly.y:1421

termbinop: term MULOP term # perly.y:1423

termbinop: term PLUGIN_MUL_OP term # perly.y:1428

termbinop: term ADDOP term # perly.y:1430

termbinop: term PLUGIN_ADD_OP term # perly.y:1432

termbinop: term SHIFTOP term # perly.y:1434

termbinop: termrelop %prec PREC_LOW # perly.y:1436

termbinop: termeqop %prec PREC_LOW # perly.y:1438

termbinop: term BITANDOP term # perly.y:1440

termbinop: term BITOROP term # perly.y:1442

termbinop: term DOTDOT term # perly.y:1444

termbinop: term ANDAND term # perly.y:1446

termbinop: term PLUGIN_LOGICAL_AND_OP term # perly.y:1448

termbinop: term OROR term # perly.y:1450

termbinop: term PLUGIN_LOGICAL_OR_OP term # perly.y:1452

termbinop: term DORDOR term # perly.y:1454

termbinop: term MATCHOP term # perly.y:1456

termbinop: term PLUGIN_LOW_OP term # perly.y:1458

termrelop: relopchain %prec PREC_LOW # perly.y:1462

termrelop: term NCRELOP term # perly.y:1464

# perly.y:1466 termrelop: termrelop NCRELOP

# perly.y:1468 termrelop: termrelop CHRELOP

termrelop: term PLUGIN_REL_OP term # perly.y:1470

relopchain: term CHRELOP term # perly.y:1474

relopchain: relopchain CHRELOP term # perly.y:1476

termeqop: eqopchain %prec PREC_LOW # perly.y:1480

termeqop: term NCEQOP term # perly.y:1482

# perly.y:1484 termeqop: termeqop NCEQOP

# perly.y:1486 termeqop: termeqop CHEQOP

eqopchain: term CHEQOP term # perly.y:1490

eqopchain: eqopchain CHEQOP term # perly.y:1492

termunop: PERLY_MINUS term %prec UMINUS # perly.y:1497

termunop: PERLY_PLUS term %prec UMINUS # perly.y:1499

termunop: PERLY_EXCLAMATION_MARK term # perly.y:1502

termunop: PERLY_TILDE term # perly.y:1504

termunop: term POSTINC # perly.y:1506

termunop: term POSTDEC # perly.y:1509

# perly.y:1512 termunop: term POSTJOIN

termunop: PREINC term # perly.y:1523

termunop: PREDEC term # perly.y:1526

anonymous: PERLY_BRACKET_OPEN optexpr PERLY_BRACKET_CLOSE # perly.y:1534

anonymous: HASHBRACK optexpr PERLY_BRACE_CLOSE %prec PERLY_PAREN_OPEN # perly.y:1536, fitted: the ; before } is toke.c's, not the source's

anonymous: KW_SUB_anon startanonsub proto subattrlist subbody %prec PERLY_PAREN_OPEN # perly.y:1538

# perly.y:1541 anonymous: KW_SUB_anon startanonsub proto subattrlist

anonymous: KW_SUB_anon_sig startanonsub subattrlist sigsubbody %prec PERLY_PAREN_OPEN # perly.y:1543

# perly.y:1546 anonymous: KW_SUB_anon_sig startanonsub subattrlist

anonymous: KW_METHOD_anon startanonmethod subattrlist sigsubbody %prec PERLY_PAREN_OPEN # perly.y:1548, fitted: as written; the ledger marks one line once, and the next rule is its fit
anonymous: KW_MY KW_METHOD_anon startanonmethod subattrlist sigsubbody %prec PERLY_PAREN_OPEN # perly.y:1548, fitted: pvm emits the my, our or state of a lexical method, which toke.c replaces by passing KEY_method to yyl_sub (toke.c:7580)

# perly.y:1553 anonymous: KW_METHOD_anon startanonmethod subattrlist

termdo: KW_DO term %prec UNIOP # perly.y:1558

termdo: KW_DO block %prec PERLY_PAREN_OPEN # perly.y:1560

term: termbinop # perly.y:1564

term: termunop # perly.y:1565

term: anonymous # perly.y:1566

term: termdo # perly.y:1567

term: term PERLY_QUESTION_MARK term PERLY_COLON term # perly.y:1568

term: REFGEN term # perly.y:1570

term: myattrterm %prec UNIOP # perly.y:1572

term: KW_LOCAL term %prec UNIOP # perly.y:1574

term: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1576

term: QWLIST # perly.y:1578

term: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1580

term: scalar %prec PERLY_PAREN_OPEN # perly.y:1582

term: star %prec PERLY_PAREN_OPEN # perly.y:1584

term: hsh %prec PERLY_PAREN_OPEN # perly.y:1586

term: ary %prec PERLY_PAREN_OPEN # perly.y:1588

term: arylen %prec PERLY_PAREN_OPEN # perly.y:1590

term: subscripted # perly.y:1592

term: sliceme PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1594

term: kvslice PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE # perly.y:1604

term: sliceme PERLY_BRACE_OPEN expr PERLY_BRACE_CLOSE # perly.y:1614, fitted: the ; before } is toke.c's, not the source's

term: kvslice PERLY_BRACE_OPEN expr PERLY_BRACE_CLOSE # perly.y:1624, fitted: the ; before } is toke.c's, not the source's

term: THING %prec PERLY_PAREN_OPEN # perly.y:1634

term: amper # perly.y:1636

term: amper PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1638

term: amper PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1641

term: NOAMP optlistexpr # perly.y:1646, fitted: toke.c emits a zero-width NOAMP and then the word as subname (toke.c:7980-7987); pvm emits the word alone, which stands for both as NOAMP

term: term ARROW CAST_STAR # perly.y:1650, fitted: pvm emits $* after -> as one Variable token, the cast terminal its star

term: term ARROW CAST_SNAIL # perly.y:1652, fitted: pvm emits @* after -> as one Variable token, the cast terminal its sigil

term: term ARROW CAST_PERCENT_SIGN # perly.y:1654, fitted: pvm emits %* after -> as one Variable token, the cast terminal its sigil

term: term ARROW PERLY_AMPERSAND # perly.y:1656, fitted: pvm emits &* after -> as one Variable token

term: term ARROW CAST_STAR %prec PERLY_PAREN_OPEN # perly.y:1659, fitted: pvm emits ** after -> as one Variable token, the cast terminal its first star

term: LOOPEX # perly.y:1661

term: LOOPEX term # perly.y:1664

term: NOTOP listexpr # perly.y:1666

term: UNIOP # perly.y:1668

term: UNIOP block # perly.y:1670

term: UNIOP term # perly.y:1672

term: KW_REQUIRE # perly.y:1674

term: KW_REQUIRE term # perly.y:1676

term: UNIOPSUB # perly.y:1678

term: UNIOPSUB term # perly.y:1680

term: FUNC0 # perly.y:1683

term: FUNC0 PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1685

term: FUNC0OP # perly.y:1687

term: FUNC0OP PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1689

term: FUNC0SUB # perly.y:1691

term: FUNC1 PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1693

term: FUNC1 PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1697

term: PMFUNC # perly.y:1699, fitted: pvm emits a pattern, substitution or transliteration and its interior, replacement included, as one Quote token (toke.c's PMFUNC SUBLEXSTART listexpr optrepl SUBLEXEND), which stands for them all as PMFUNC, so optrepl is excluded folded-into-token

term: BAREWORD # perly.y:1712

term: listop # perly.y:1713

# perly.y:1714 term: PLUGEXPR

myattrterm: KW_MY myterm attrlist # perly.y:1719

myattrterm: KW_MY myterm # perly.y:1721

myattrterm: KW_MY REFGEN myterm attrlist # perly.y:1723

myattrterm: KW_MY REFGEN term # perly.y:1725

myterm: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1730

myterm: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1732

myterm: scalar %prec PERLY_PAREN_OPEN # perly.y:1735

myterm: hsh %prec PERLY_PAREN_OPEN # perly.y:1737

myterm: ary %prec PERLY_PAREN_OPEN # perly.y:1739

fieldvar: scalar %prec PERLY_PAREN_OPEN # perly.y:1744

fieldvar: hsh %prec PERLY_PAREN_OPEN # perly.y:1749

fieldvar: ary %prec PERLY_PAREN_OPEN # perly.y:1754

fielddecl: KW_FIELD fieldvar optattrlist # perly.y:1762

fielddecl: KW_FIELD fieldvar optattrlist ASSIGNOP term # perly.y:1769

optlistexpr: empty %prec PREC_LOW # perly.y:1787

optlistexpr: listexpr %prec PREC_LOW # perly.y:1788

optexpr: empty # perly.y:1792

optexpr: expr # perly.y:1793

# perly.y:1797 optrepl: empty

# perly.y:1798 optrepl: PERLY_SLASH expr

my_scalar: scalar # perly.y:1803

list_of_itervars: list_of_itervars PERLY_COMMA # perly.y:1808

list_of_itervars: list_of_itervars PERLY_COMMA scalar # perly.y:1810

list_of_itervars: list_of_itervars PERLY_COMMA REFGEN my_var # perly.y:1814

list_of_itervars: scalar %prec PREC_LOW # perly.y:1818

list_of_itervars: REFGEN my_var %prec PREC_LOW # perly.y:1819

my_list_of_itervars: list_of_itervars # perly.y:1825

my_var: scalar # perly.y:1829

my_var: ary # perly.y:1830

my_var: hsh # perly.y:1831

refgen_topic: my_var # perly.y:1834

refgen_topic: amper # perly.y:1835

my_refgen: KW_MY REFGEN # perly.y:1838

my_refgen: REFGEN KW_MY # perly.y:1839

amper: PERLY_AMPERSAND indirob # perly.y:1842, fitted: as written; the ledger marks one line once, and the next rule is its fit
amper: PRIVATEREF # perly.y:1842, fitted: toke.c emits a zero-width PERLY_AMPERSAND and then the word as BAREWORD or PRIVATEREF before a ( (toke.c:8195-8211); pvm emits the word alone, which stands for both as PRIVATEREF, a terminal no word ending in :: carries

scalar: PERLY_DOLLAR # perly.y:1846, fitted: pvm emits the sigil and the name as one Variable token
scalar: CAST_DOLLAR indirob # perly.y:1846, fitted: pvm emits a cast's sigil as its own DerefSigil token, mapped to the cast terminal

ary: PERLY_SNAIL # perly.y:1850, fitted: pvm emits the sigil and the name as one Variable token
ary: CAST_SNAIL indirob # perly.y:1850, fitted: pvm emits a cast's sigil as its own DerefSigil token, mapped to the cast terminal

hsh: PERLY_PERCENT_SIGN # perly.y:1856, fitted: pvm emits the sigil and the name as one Variable token
hsh: CAST_PERCENT_SIGN indirob # perly.y:1856, fitted: pvm emits a cast's sigil as its own DerefSigil token, mapped to the cast terminal

arylen: DOLSHARP # perly.y:1862, fitted: pvm emits the sigil and the name as one Variable token, $# its sigil

arylen: term ARROW DOLSHARP # perly.y:1864, fitted: pvm emits $#* after -> as one Variable token

star: PERLY_STAR indirob # perly.y:1868

sliceme: ary # perly.y:1872

sliceme: term ARROW PERLY_SNAIL # perly.y:1873

kvslice: hsh # perly.y:1877

kvslice: term ARROW PERLY_PERCENT_SIGN # perly.y:1878

gelem: star # perly.y:1882

gelem: term ARROW PERLY_STAR # perly.y:1883

indirob: BAREWORD # perly.y:1888

indirob: scalar %prec PREC_LOW # perly.y:1890

indirob: block # perly.y:1892

indirob: PRIVATEREF # perly.y:1895
`

// perlyPrecedence is perly.y's precedence declarations, loosest first, one per line: `%<assoc> TOKEN... # perly.y:N`.
// Recognition ignores them and every rule's %prec at waypoint 1 (spec §8): they are waypoint 2's constraint's data.
const perlyPrecedence = `
%nonassoc PREC_LOW # perly.y:150

%nonassoc LOOPEX # perly.y:151

%nonassoc PLUGIN_LOW_OP # perly.y:153

%left OROP PLUGIN_LOGICAL_OR_LOW_OP # perly.y:154

%left ANDOP PLUGIN_LOGICAL_AND_LOW_OP # perly.y:155

%right NOTOP # perly.y:156

%nonassoc LSTOP LSTOPSUB BLKLSTOP # perly.y:157

%left PERLY_COMMA # perly.y:158

%right ASSIGNOP PLUGIN_ASSIGN_OP # perly.y:159

%right PERLY_QUESTION_MARK PERLY_COLON # perly.y:160

%nonassoc DOTDOT # perly.y:161

%left OROR DORDOR PLUGIN_LOGICAL_OR_OP # perly.y:162

%left ANDAND PLUGIN_LOGICAL_AND_OP # perly.y:163

%left BITOROP # perly.y:164

%left BITANDOP # perly.y:165

%left CHEQOP NCEQOP # perly.y:166

%left CHRELOP NCRELOP # perly.y:167

%nonassoc PLUGIN_REL_OP # perly.y:168

%nonassoc UNIOP UNIOPSUB # perly.y:169

%nonassoc KW_REQUIRE # perly.y:170

%left SHIFTOP # perly.y:171

%left ADDOP PLUGIN_ADD_OP # perly.y:172

%left MULOP PLUGIN_MUL_OP # perly.y:173

%left MATCHOP # perly.y:174

%right PERLY_EXCLAMATION_MARK PERLY_TILDE UMINUS REFGEN # perly.y:175

%right POWOP PLUGIN_POW_OP # perly.y:176

%nonassoc PREINC PREDEC POSTINC POSTDEC # perly.y:177, fitted: POSTJOIN is left out: its one rule, perly.y:1512, is excluded interpolation-only, and the generator refuses a precedence token no rule uses

%nonassoc PLUGIN_HIGH_OP # perly.y:178

%left ARROW # perly.y:179

%nonassoc PERLY_PAREN_CLOSE # perly.y:180

%left PERLY_PAREN_OPEN # perly.y:181

%left PERLY_BRACKET_OPEN PERLY_BRACE_OPEN # perly.y:182
`

// parenOpen and blockOpen are the next tokens that tell toke.c's FUNC1 from UNIOP and KW_CONTINUE from FUNC0;
// fatComma turns a word before it into a bareword (yyl_keylookup checks for => before any keyword).
var (
	fatComma  = Neighbour{Kind: lexer.Operator, Text: "=>"}
	parenOpen = Neighbour{Kind: lexer.Operator, Text: "("}
	blockOpen = Neighbour{Kind: lexer.Operator, Text: "{", OpensBlock: true, Brace: BraceBlock}
	// hashOpen is a { the lexer took for an anonymous hash's.
	hashOpen = Neighbour{Kind: lexer.Operator, Text: "{", Brace: BraceHash}

	// The NotNext lists that name one of them.
	nextNotParen, nextNotBlock, nextNotFat = []Neighbour{parenOpen}, []Neighbour{blockOpen}, []Neighbour{fatComma}

	// declarators are the words after which a variable is declared and a : may start its attributes.
	declarators = []Neighbour{{Kind: lexer.Word, Text: "my"}, {Kind: lexer.Word, Text: "our"}, {Kind: lexer.Word, Text: "state"}}

	forWords = []Neighbour{{Kind: lexer.Word, Text: "for"}, {Kind: lexer.Word, Text: "foreach"}}

	// otherWord is a word no class names, as the next token: a keyword is never intuit_method's class name.
	otherWord = Neighbour{Kind: lexer.Word, Text: OtherWord}

	// sortParen is the ( right after sort, which toke.c reads as sort's own.
	sortParen = Neighbour{Kind: lexer.Operator, Text: "(", SortParen: true}

	// sortIndirob are the tokens after which toke.c reads a bareword as the sort sub's name and leaves PL_expect at
	// XTERM, a declared sub or not (yyl_just_a_word, toke.c:8158-8167, "indir obj regardless for sort"): sort, and the (
	// after it. After the other list operators that take a filehandle the bareword is one only when no sub of that
	// name is declared (`!c.cv`), which scope decides, so there it completes a term as any bareword may.
	sortIndirob = []Neighbour{{Kind: lexer.Word, Text: "sort"}, sortParen}

	// forcedName are require and the loop exits, after which toke.c forces the word to BAREWORD unless it is a keyword
	// (force_word's CHECK_KEYWORD, toke.c:7407 and LOOPX), so x there is still a word toke.c lexes itself.
	forcedName = []Neighbour{{Kind: lexer.Word, Text: "require"}, {Kind: lexer.Word, Text: "next"}, {Kind: lexer.Word, Text: "last"},
		{Kind: lexer.Word, Text: "redo"}, {Kind: lexer.Word, Text: "dump"}, {Kind: lexer.Word, Text: "goto"}}

	// doWord is do, after which toke.c forces the word to a name only before a ( and unless it is a keyword or CORE
	// (yyl_do, toke.c:7515-7528), so `do foo;` and `do foo 1;` stay scope's and `do x(1)` is a call.
	doWord = Neighbour{Kind: lexer.Word, Text: "do"}

	// forcedWord are the tokens after which toke.c makes the word that follows a name, never a call: forcedName, sort
	// and -> (force_word, toke.c:9042, 6296), & (pending_ident, toke.c:6949), a glob's * (yyl_star), and sub, method, package,
	// class, use, no and format (yyl_sub, toke.c:8859, 8379, tokenize_use, KEY_format). The scope-decided readings of a word leave them out.
	forcedWord = append(append(slices.Clone(sortIndirob), forcedName...), namedWord...)

	// namedWord are the tokens of forcedWord after which even a keyword is a name: sort's and forcedName's are not.
	namedWord = []Neighbour{{Kind: lexer.Operator, Text: "->"}, {Kind: lexer.FuncSigil}, {Kind: lexer.Operator, Text: "*", Glob: true},
		{Kind: lexer.Word, Text: "sub"}, {Kind: lexer.Word, Text: "package"}, {Kind: lexer.Word, Text: "class"}, {Kind: lexer.Word, Text: "method"}, {Kind: lexer.Word, Text: "use"}, {Kind: lexer.Word, Text: "no"},
		{Kind: lexer.Word, Text: "format"}}

	// xForced are the tokens after which x before a ( is no call: forcedWord but forcedName, which leaves a keyword alone.
	xForced = append(slices.Clone(sortIndirob), namedWord...)

	// nextUndecided are the next tokens that decide a word whatever is declared: => makes it BAREWORD (toke.c:8176)
	// and ( the zero-width & and its name (toke.c:8195-8211), so a scope-decided reading follows neither.
	nextUndecided = []Neighbour{fatComma, parenOpen}

	// callForced are the tokens after which a word before a ( is no call: forcedWord and do.
	callForced = append(slices.Clone(forcedWord), doWord)

	// keywordsAndCORE are the words yyl_do does not force before a (.
	keywordsAndCORE = append(lexer.Keywords(), "CORE")

	// keywordsButX are the keywords other than x, which toke.c's KEY_x makes a word when no operator is expected.
	keywordsButX = slices.DeleteFunc(slices.Clone(lexer.Keywords()), func(s string) bool { return s == "x" })

	// stateStarts are the tokens after which toke.c has PL_expect at XSTATE: a file's start, a `;`, a block's
	// `{`, a block's `}` and a label's colon (the neighbour flags that colon, since a bare one may be an
	// attribute list's: `my $y : for our ($x);` is legal).
	// shortcut: a do, eval or anonymous-sub block's `}` reads as a statement start because the lexer does not mark
	// what a `}` closes (01a1a422-9176-7153-8625-6535b483a4b4); upgrade when it does.
	stateStarts = []Neighbour{Boundary, {Kind: lexer.Semicolon}, blockOpen, {Kind: lexer.CloseBracket, Text: "}", OpensBlock: true},
		{Kind: lexer.Operator, Text: ":", Label: true}}

	// subscriptBrace and notHashBrace mirror ClassOf's brace decision, and the three change together: after a
	// variable, a closing bracket or `->` a `{` the lexer did not open as a block is a subscript's, and after a
	// deref sigil a block's, never an anonymous hash's. ClassOf's `)` and term-following-variable cases are not
	// listed: the lexer opens every `{` after them as a block, so no key reaches them here; nor is its name case,
	// since a neighbour cannot name every word (the subscript's { past a glob's name has its own class).
	// notHashBrace also lists a sub's `&` and a glob's `*`, after which a `{` is a block's too (yyl_ampersand's and
	// yyl_star's PREREF); ClassOf leaves those Hash, which no key reads, since only a word's Next tells the two apart.
	subscriptBrace = []Neighbour{
		{Kind: lexer.Variable}, {Kind: lexer.Operator, Text: "->"},
		{Kind: lexer.CloseBracket, Text: "]"}, {Kind: lexer.CloseBracket, Text: "}"},
		{Kind: lexer.Variable, Array: true},
	}
	notHashBrace = append(slices.Clone(subscriptBrace), derefSigil, funcSigil, globStar)

	// A word alone in a subscript's braces, after a minus or not (toke.c:6656-6679), is the Lone class's key.
	subscriptOpen  = Neighbour{Kind: lexer.Operator, Text: "{", Brace: BraceSubscript}
	subscriptClose = Neighbour{Kind: lexer.CloseBracket, Text: "}"}
	minus          = Neighbour{Kind: lexer.Operator, Text: "-"}

	// After ->, toke.c's XPOSTDEREF reads a sigil and a star as a postfix dereference (`->@*`, S_postderef), which pvm
	// emits as one Variable token; a bare sigil there opens a slice. So a variable's head classes leave -> out, and
	// classes keyed on its text after -> take it.
	arrow      = Neighbour{Kind: lexer.Operator, Text: "->"}
	afterArrow = []Neighbour{arrow}
	// globStar is a * toke.c's yyl_star reads as a glob's sigil, after which a { opens a block (PREREF, PL_expect XREF),
	// and two tokens after which, past the glob's name, a subscript's (TERM, PL_expect XOPERATOR).
	globStar = Neighbour{Kind: lexer.Operator, Text: "*", Glob: true}
	// derefSigil is the sigil of a cast and funcSigil a sub's &, a { after either of which opens a block (`${ $x }`, `&{ $c }`).
	derefSigil = Neighbour{Kind: lexer.DerefSigil}
	funcSigil  = Neighbour{Kind: lexer.FuncSigil}
	// nextNotClose is a `{` with its `}` right after it, which yyl_leftcurly reads as HASHBRACK after a sigil.
	nextNotClose = []Neighbour{subscriptClose}
)

// CastTerminals are the fitted terminals perly.y does not declare, each with the perly.y sigil it stands for: perigrin's
// ruling "Fitted cast terminal", an extension of spec §7 (see CAST_DOLLAR below). No other terminal is fitted.
var CastTerminals = map[string]string{
	"CAST_DOLLAR": "PERLY_DOLLAR", "CAST_PERCENT_SIGN": "PERLY_PERCENT_SIGN", "CAST_SNAIL": "PERLY_SNAIL", "CAST_STAR": "PERLY_STAR",
}

// perlyClasses are the token classes, each declared by the terminal it admits and so
// unmoved by a rule added or removed. One placeholder comment per token perly.y
// declares (%token, %left, %right, %nonassoc) that is lexed, in the order it first
// declares it (the GRAM* selectors, PLUGEXPR, PLUGSTMT,
// the PLUGIN_* terminals and the precedence-only PREC_LOW and UMINUS need none). A
// blank line separates neighbours, so two families' edits merge cleanly. A family
// replaces its terminals' comments with `{"TERMINAL", TokenClass{...}},`: the same
// terminal may be declared more than once, and the classes matching a token union.
var perlyClasses = []ClassDecl{
	{"PERLY_AMPERSAND", TokenClass{Kind: ptr(lexer.FuncSigil)}},
	{"PERLY_AMPERSAND", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("&*"), Prev1: &arrow}},

	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(true)}},
	// The { of `sub open {` follows a word that reads a { as a hash, which limitBlockBrace takes off the class above; the sub's
	// name two tokens back makes it a block again (toke.c yyl_sub leaves PL_expect at XATTRBLOCK, toke.c:5874).
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(true), Prev2: &Neighbour{Kind: lexer.Word, Text: "sub"}}},
	// A subscript's { is the one the lexer did not open as a block, after what ClassOf makes a subscript's.
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &subscriptBrace[0], NotPrev2: afterArrow}},
	// After -> the variable may be a cast's lone $ (`->${ … }`), and an empty pair after it is HASHBRACK as after any sigil.
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &subscriptBrace[0], Prev2: &arrow, NotNext: nextNotClose}},
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &subscriptBrace[1]}},
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &subscriptBrace[2]}},
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &subscriptBrace[3]}},
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &subscriptBrace[4]}},

	// After a word no class names toke.c's METHCALL0 leaves PL_expect at XBLOCKTERM (toke.c:8215), so the { is a block
	// whatever the lexer decided; a declared sub's would be a hash, which scope decides, so the hash reading stays.
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &otherWord}},

	// A cast's {, a &'s and a glob's (`${ $x }`, `&{ $c }`, `*{"g"}`) open a block, indirob's; past a glob's name the { is a
	// subscript's.
	// An empty pair after a sigil is toke.c's HASHBRACK (yyl_leftcurly, toke.c:6706-6715), which no rule takes after one, so
	// perl refuses `${}`; the { there admits nothing. Past a glob's name it is an empty subscript, which the grammar refuses.
	// shortcut: the empty pair's { is unmapped rather than HASHBRACK, since perl refuses every file that has one and no -DT
	// capture can see such a class; upgrade if a reviewed tdt-unseen list takes it.
	// shortcut: `${ {} }` (any sigil's, after -> too) reads as a block holding a hash, where toke.c's scan_ident takes
	// `${ {` as the variable `{`; internal/lexer emits a DerefSigil and two `{` (v0.1 01a1a9d4), upgrade when it emits
	// toke.c's token.
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &derefSigil, NotNext: nextNotClose}},
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &funcSigil, NotNext: nextNotClose}},
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev1: &globStar, NotNext: nextNotClose}},
	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), Prev2: &globStar}},

	{"PERLY_BRACE_CLOSE", TokenClass{Kind: ptr(lexer.CloseBracket), Text: ptr("}"), OpensBlock: ptr(true), Completes: true}},
	{"PERLY_BRACE_CLOSE", TokenClass{Kind: ptr(lexer.CloseBracket), Text: ptr("}"), OpensBlock: ptr(false), Completes: true}},

	{"PERLY_BRACKET_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("[")}},

	{"PERLY_BRACKET_CLOSE", TokenClass{Kind: ptr(lexer.CloseBracket), Text: ptr("]"), Completes: true}},

	{"PERLY_COMMA", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(",")}},
	{"PERLY_COMMA", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("=>")}},

	// A variable that is no plain name (`$::e`, `$1`, `$^W`, `$'`, `$_`) is not a catch variable: the action of
	// perly.y:815 declares it with my and refuses it, so two tokens after catch it is no scalar.
	{"PERLY_DOLLAR", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('$')), Global: ptr(false), NotPrev1: afterArrow, Completes: true}},
	// shortcut: this class also refuses `catch =>` before a variable that is no plain name, in any list (`foo(catch => $_)`), which perl accepts (catch is a hash key); it sees only the word two tokens back, so upgrade by letting it see => as Prev1, pinned in TestPortScopeDecidedKnownLimits.
	{"PERLY_DOLLAR", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('$')), Global: ptr(true), NotPrev1: afterArrow, NotPrev2: []Neighbour{{Kind: lexer.Word, Text: "catch"}}, Completes: true}},
	// After -> a $ variable is a method's name (`$o->$m`), $* the postfix dereference, and a lone $ (pvm's for `->${`) a cast's.
	{"PERLY_DOLLAR", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('$')), NotText: []string{"$*", "$"}, Prev1: &arrow, Completes: true}},

	// The fitted cast terminals (perigrin's ruling "Fitted cast terminal", an extension of spec §7): a cast's sigil, which pvm
	// emits as its own DerefSigil token, and the postfix dereference after ->, which it emits as one Variable token, are
	// each perly.y's sigil, but a whole variable is that sigil too, so sharing its terminal would read `$x $z` as a cast.
	{"CAST_DOLLAR", TokenClass{Kind: ptr(lexer.DerefSigil), Text: ptr("$")}},
	{"CAST_DOLLAR", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("$"), Prev1: &arrow}},

	// PERLY_DOT

	{"PERLY_EQUAL_SIGN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("=")}},

	{"PERLY_MINUS", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-")}},

	{"PERLY_PERCENT_SIGN", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('%')), NotPrev1: afterArrow, Completes: true}},
	{"PERLY_PERCENT_SIGN", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("%"), Prev1: &arrow}},

	{"CAST_PERCENT_SIGN", TokenClass{Kind: ptr(lexer.DerefSigil), Text: ptr("%")}},
	{"CAST_PERCENT_SIGN", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("%*"), Prev1: &arrow}},

	{"PERLY_PLUS", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("+")}},

	{"PERLY_SEMICOLON", TokenClass{Kind: ptr(lexer.Semicolon)}},

	// PERLY_SLASH

	{"PERLY_SNAIL", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('@')), NotPrev1: afterArrow, Completes: true}},
	{"PERLY_SNAIL", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("@"), Prev1: &arrow}},
	// A blank after -> leaves it TERM(ARROW), and a nameless @ then is PREREF(PERLY_SNAIL) (yyl_snail), so `->@ {` is `->@{`.
	{"PERLY_SNAIL", TokenClass{Kind: ptr(lexer.DerefSigil), Text: ptr("@"), Prev1: &arrow}},

	{"CAST_SNAIL", TokenClass{Kind: ptr(lexer.DerefSigil), Text: ptr("@"), NotPrev1: afterArrow}},
	{"CAST_SNAIL", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("@*"), Prev1: &arrow}},

	// shortcut: `$r->* {CODE}` is accepted, where toke.c reads `->` then a blank as TERM(ARROW) and the * as an operator;
	// the key cannot tell `*{` from `* {`, upgrade if a corpus file holds one.
	{"PERLY_STAR", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("*")}},

	// $* after -> is the star's and not the sigil's: CAST_DOLLAR would let `$r->$*{a}` read as a method named by a cast.
	{"CAST_STAR", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("$*"), Prev1: &arrow}},
	{"CAST_STAR", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("**"), Prev1: &arrow}},

	{"KW_FORMAT", TokenClass{Kind: ptr(lexer.Word), Text: ptr("format")}},

	// toke.c KEY_package (8858) is KW_PACKAGE wherever the word is a keyword.
	{"KW_PACKAGE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("package")}},

	// toke.c KEY_class (8376) is KW_CLASS wherever the word is a keyword; the chart does not track the feature.
	{"KW_CLASS", TokenClass{Kind: ptr(lexer.Word), Text: ptr("class")}},

	{"KW_LOCAL", TokenClass{Kind: ptr(lexer.Word), Text: ptr("local")}},

	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("my")}},
	// toke.c yyl_foreach croaks "Missing $ on loop variable" for `for our (` and `for state (` where
	// PL_expect is XSTATE, so each of our and state is a class unless it follows a for at a statement start.
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("our"), NotNext: nextNotParen}},
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("our"), NotPrev1: forWords}},
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("our"), NotPrev2: stateStarts}},
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("state"), NotNext: nextNotParen}},
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("state"), NotPrev1: forWords}},
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("state"), NotPrev2: stateStarts}},

	// toke.c KEY_field (8537).
	{"KW_FIELD", TokenClass{Kind: ptr(lexer.Word), Text: ptr("field")}},

	{"KW_IF", TokenClass{Kind: ptr(lexer.Word), Text: ptr("if")}},

	{"KW_ELSE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("else")}},

	{"KW_ELSIF", TokenClass{Kind: ptr(lexer.Word), Text: ptr("elsif")}},

	{"KW_UNLESS", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unless")}},

	{"KW_FOR", TokenClass{Kind: ptr(lexer.Word), Text: ptr("for")}},
	{"KW_FOR", TokenClass{Kind: ptr(lexer.Word), Text: ptr("foreach")}},

	{"KW_UNTIL", TokenClass{Kind: ptr(lexer.Word), Text: ptr("until")}},

	{"KW_WHILE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("while")}},

	{"KW_CONTINUE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("continue"), Next: &blockOpen}},

	{"KW_GIVEN", TokenClass{Kind: ptr(lexer.Word), Text: ptr("given")}},

	{"KW_WHEN", TokenClass{Kind: ptr(lexer.Word), Text: ptr("when")}},

	{"KW_DEFAULT", TokenClass{Kind: ptr(lexer.Word), Text: ptr("default")}},

	{"KW_TRY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("try")}},

	{"KW_CATCH", TokenClass{Kind: ptr(lexer.Word), Text: ptr("catch")}},

	{"KW_FINALLY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("finally")}},

	{"KW_DEFER", TokenClass{Kind: ptr(lexer.Word), Text: ptr("defer")}},

	{"KW_REQUIRE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("require")}},

	{"KW_DO", TokenClass{Kind: ptr(lexer.Word), Text: ptr("do")}},

	{"KW_USE_or_NO", TokenClass{Kind: ptr(lexer.Word), Text: ptr("use")}},
	{"KW_USE_or_NO", TokenClass{Kind: ptr(lexer.Word), Text: ptr("no")}},

	// toke.c (yyl_sub, toke.c:5874) names the sub when an identifier, :: or ' follows, and takes it as anonymous
	// otherwise (toke.c:5967, 5975). The seven words that really_sub takes by name when a statement may
	// begin (toke.c:8315) are KW_SUB_named too, and only where no name follows them: pvm emits the word
	// once, and perl refuses `BEGIN foo {}`. The _sig pair is the same word under the signatures feature
	// (toke.c:5965-5975, is_sigsub): the chart does not track the feature, so each `sub` is both of its
	// pair, and the lexer's tokens make the split (a Prototype token after the name, or parentheses).
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sub"), NextName: true}},
	// shortcut: the seven words take no prototype, as the lexer scans one only after sub NAME (01a1a44e-8f26-706b-96d4-b3e6d7f55c3f), upgrade when it scans one after them.
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("AUTOLOAD"), NotNextName: true}},
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("DESTROY"), NotNextName: true}},
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("BEGIN"), NotNextName: true}},
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("UNITCHECK"), NotNextName: true}},
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("CHECK"), NotNextName: true}},
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("INIT"), NotNextName: true}},
	{"KW_SUB_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("END"), NotNextName: true}},

	{"KW_SUB_named_sig", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sub"), NextName: true}},

	{"KW_SUB_anon", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sub"), NotNextName: true}},

	{"KW_SUB_anon_sig", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sub"), NotNextName: true}},

	// toke.c names the method as it names a sub (yyl_sub, 5874, 5906): the same split on a name that follows.
	{"KW_METHOD_named", TokenClass{Kind: ptr(lexer.Word), Text: ptr("method"), NextName: true}},

	{"KW_METHOD_anon", TokenClass{Kind: ptr(lexer.Word), Text: ptr("method"), NotNextName: true}},

	// Every word that is not a keyword; the one keyword toke.c falls through to a word when no
	// operator is expected, x (KEY_x), is listed below. A keyword is a BAREWORD only where a position
	// class says so; lt, gt, le, ge, eq, ne and cmp are never words.
	// A word two after `package` is never the BAREWORD toke.c forces for the version (force_strict_version refuses
	// it), and one two after `use` or `no` is a term, which versions.go gives back unless a number or a quote follows.
	// After sort the word is the sort sub's name, which does not complete a term (see sortIndirob).
	// Before a ( a word is the zero-width & and its name (PRIVATEREF below), except a package name (Foo::, toke.c:8098-8111,
	// safebw) and a word toke.c forces after a loop exit, require or do, which stay BAREWORD and which perl refuses there.
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), NotPrev1: sortIndirob, NotPrev2: afterWordSlot, NotNext: nextNotParen, Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prefix: ptr(byte(PackageName)), NotPrev1: sortIndirob, NotPrev2: afterWordSlot, Next: &parenOpen, Completes: true}},
	// A word before => is a string, a keyword included (yyl_fatcomma; see cutBeforeFatComma), but not CORE::kw.
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prefix: ptr(byte(0)), Next: &fatComma}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prefix: ptr(byte(PackageName)), Next: &fatComma}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prev1: &forcedName[0], Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prev1: &forcedName[1], Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prev1: &forcedName[2], Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prev1: &forcedName[3], Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prev1: &forcedName[4], Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prev1: &forcedName[5], Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: keywordsAndCORE, Prev1: &doWord, Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Text: ptr("x"), NotPrev1: sortIndirob, NotPrev2: afterWordSlot, Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: keywordsButX, Prev1: &sortIndirob[0]}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: keywordsButX, Prev1: &sortIndirob[1]}},
	// The word after a method name: intuit_method forces it to BAREWORD whatever follows it (toke.c:5138-5140), a (
	// included, where the generic BAREWORD leaves a word before ( to the call; scope decides which it is.
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), Prev1: &otherWord, Next: &parenOpen}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Number), Prev1: &Neighbour{Kind: lexer.Word, Text: "require"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('v')), Prev1: &Neighbour{Kind: lexer.Word, Text: "require"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('5')), Prev1: &Neighbour{Kind: lexer.Word, Text: "require"}}},
	// A glob's name: yyl_star runs scan_ident, then sets PL_expect to XOPERATOR (TERM(PERLY_STAR)), but
	// only for a * that is a glob sigil, which the neighbour class of the * says (Neighbour.Glob).
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: "*", Glob: true}, Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.FuncSigil}, Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "use"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "format"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "no"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "sub"}}}, // yyl_sub: any identifier after sub names it, toke.c:5874
	// toke.c 8858: KEY_package forces the name without a keyword check, as KEY_class does (8379) and yyl_sub does for a method (5874).
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "package"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "class"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "method"}}},

	// toke.c forces a word that stands alone in a subscript's braces to BAREWORD before it looks the word up
	// (yyl_leftcurly, toke.c:6656-6679), a minus before it included.
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &subscriptOpen, Next: &subscriptClose, Blanks: ptr(true), Lone: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev2: &subscriptOpen, Prev1: &minus, Next: &subscriptClose, Blanks: ptr(true), Lone: true}},

	// toke.c forces the word after -> to METHCALL0, a keyword or a quote-like operator's name included.
	{"METHCALL0", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: "->"}, Completes: true}},
	// A word before a variable or a block (the lexer's hash { too, see PERLY_BRACE_OPEN) is a method called on it
	// (toke.c:8215, unless it is a declared sub, which scope decides). Before another word, intuit_method makes it one
	// too (toke.c:5131-5137), METHCALL when a ( follows that word; that is two tokens on, so the word before a word is
	// both. Neither completes a term: an indirob follows. A package name (Foo::) is neither (safebw, toke.c:8111).
	{"METHCALL0", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), Next: &Neighbour{Kind: lexer.Variable}}},
	{"METHCALL0", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), Next: &blockOpen}},
	{"METHCALL0", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), Next: &hashOpen}},
	{"METHCALL0", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), Next: &otherWord}},

	{"METHCALL", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), Next: &otherWord}},

	// An attribute's name follows its colon, the attribute before it or its argument's parenthesis; its
	// one-token argument, never a parenthesis, follows its own.
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: ":"}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: OtherWord}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.CloseBracket, Text: ")"}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Number), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Quote), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Variable), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},
	// The same after the ( of an attribute named sort, which carries SortParen as sort's own ( does.
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &sortParen}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Number), Prev1: &sortParen}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Quote), Prev1: &sortParen}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Variable), Prev1: &sortParen}},

	// A number that toke.c reads as a version (a delimited one, see lex.versionShape) after `use MODULE` is
	// that version and no term, so it is not a THING there.
	{"THING", TokenClass{Kind: ptr(lexer.Number), Prefix: ptr(byte(0)), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Number), Prefix: ptr(byte(VersionStrict)), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Number), Prefix: ptr(byte(VersionLax)), NotPrev2: afterUse, Completes: true}},
	// After => toke.c expects a term, so a number there is one even two after a use or no that => quoted (yyl_fatcomma).
	// Not marked Completes: x after a number already follows the classes above, and Completes on a Prev1 class names a word.
	{"THING", TokenClass{Kind: ptr(lexer.Number), Prev1: &fatComma}},
	// S_sublex_start (toke.c:2595) returns THING for a constant: '…' and q (yyl_sglquote, toke.c:7252; KEY_q, toke.c:8868),
	// a "…" holding no $, @, \ or non-ASCII byte (yyl_dblquote, toke.c:7270), <<'EOF' (OP_CONST, toke.c:11712) and a readline
	// (OP_NULL, toke.c:12200). Every other "…", heredoc and <…> is FUNC (below), and the token key cannot tell them
	// apart, so those tokens are both: the accepted language is the same, as both are a whole term.
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('\'')), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('"')), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('q')), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.HeredocOpen), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Readline), Completes: true}},
	// A v-string is a constant term: toke.c scans `v1.2` (toke.c:9879-9887) and a number with a second dot (scan_num,
	// toke.c:13135-13137) and returns THING. pvm emits it as a Quote opening with v or a digit; after use MODULE it is
	// the version (versions.go), as a number is.
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('v')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('0')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('1')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('2')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('3')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('4')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('5')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('6')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('7')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('8')), NotPrev2: afterUse, Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('9')), NotPrev2: afterUse, Completes: true}},

	// S_sublex_start returns PMFUNC for a quote-like whose op the scan built: m// and // (scan_pat, toke.c:8752, 7085),
	// qr// (toke.c:8885), s/// (scan_subst, toke.c:8944), and tr/// and y/// (scan_trans, toke.c:9101).
	{"PMFUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('/')), Completes: true}},
	{"PMFUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('m')), Completes: true}},
	{"PMFUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('r')), Completes: true}},
	{"PMFUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('s')), Completes: true}},
	{"PMFUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('y')), Completes: true}},

	// A sub's name after &: yyl_ampersand forces it through pending_ident (toke.c:6949), which makes a lexical sub's
	// name PRIVATEREF (toke.c:10397) and a package sub's BAREWORD; scope tells them apart.
	{"PRIVATEREF", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.FuncSigil}, Completes: true}},
	// shortcut: a qualified name after my, our or state sub (`my sub Foo::bar {1}`) is accepted, where toke.c croaks while
	// scanning it (yyl_sub); upgrade when internal/lexer reports the croak (v0.1 01a1aa1e).
	// A sub's name is PRIVATEREF when the pad holds it (force_ident_maybe_lex, toke.c:5969), always so after my, our
	// or state (pending_ident under PL_in_my, toke.c:10363); scope decides the rest.
	{"PRIVATEREF", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "sub"}}},
	// A word before a ( is a call: toke.c emits a zero-width & and the name, BAREWORD or PRIVATEREF (toke.c:8195-8211),
	// which the word stands for as PRIVATEREF (perly.y:1842's fit), unless it is a package name or forced to a name (callForced). x is
	// one too where it is a word (KEY_x), a keyword force_word and yyl_do leave alone (xForced), as yyl_do leaves CORE.
	// shortcut: `CORE::foo(1)` is a call, where toke.c croaks "CORE::foo is not a keyword"; upgrade when internal/lexer
	// reports the croak (v0.1 01a1aa1e).
	{"PRIVATEREF", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), NotPrev1: callForced, Next: &parenOpen}},
	{"PRIVATEREF", TokenClass{Kind: ptr(lexer.Word), Text: ptr("x"), NotPrev1: xForced, Next: &parenOpen}},
	{"PRIVATEREF", TokenClass{Kind: ptr(lexer.Word), Text: ptr("CORE"), Prev1: &doWord, Next: &parenOpen}},

	{"QWLIST", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('w')), Completes: true}},

	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__FILE__"), Completes: true}},
	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__LINE__"), Completes: true}},
	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__PACKAGE__"), Completes: true}},
	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__SUB__"), Completes: true}},

	// A word no neighbour decides may be a declared sub, which scope decides (waypoint 2), so it is each reading
	// yyl_subproto gives one (toke.c:6590-6636) beside BAREWORD and NOAMP: FUNC0SUB for an empty prototype, which
	// completes a term (TERM), UNIOPSUB for a one-argument one, and LSTOPSUB for one opening with & before a {.
	// shortcut: a declared sub named x carries these only where a statement starts (addStatementX), so `$y = x {1};` is refused
	// where perl accepts it; upgrade by giving x them wherever no operator is expected, if a corpus file calls one so.
	// shortcut: a word two after use or no carries none of them, as the version classes own it; upgrade with the version family. After package or class the version slot is a BAREWORD alone, and no rule takes these terminals there.
	{"FUNC0SUB", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), NotPrev1: forcedWord, NotPrev2: afterUse, NotNext: nextUndecided, Completes: true}},

	{"UNIOPSUB", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), NotPrev1: forcedWord, NotPrev2: afterUse, NotNext: nextUndecided}},

	{"LSTOPSUB", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), NotPrev1: forcedWord, NotPrev2: afterUse, Next: &blockOpen}},
	{"LSTOPSUB", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), NotPrev1: forcedWord, NotPrev2: afterUse, Next: &hashOpen}},

	// Any word, a keyword too, at a statement start with a colon after it: yyl_keylookup tests for a label
	// before it looks a keyword up. Where a statement may start is the grammar's to say.
	// The colon is on the word's line, and the word has no package separator (yyl_keylookup scans it
	// without one and reads a line at a time).
	{"LABEL", TokenClass{Kind: ptr(lexer.Word), Global: ptr(false), LineColon: ptr(true)}},

	{"PROTOTYPE", TokenClass{Kind: ptr(lexer.Prototype)}},

	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dump")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("goto")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("last")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("next")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("redo")}},

	{"DOTDOT", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("..")}},
	{"DOTDOT", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("...")}},

	// yyl_dot at XSTATE; elsewhere the same three dots are DOTDOT, and where a statement may start is the grammar's to say.
	{"YADAYADA", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("...")}},

	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__CLASS__"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("break"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("endhostent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("endnetent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("endservent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("endprotoent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("endpwent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("endgrent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fork"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getppid"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getprotoent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpwent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gethostent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getnetent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getservent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getgrent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getlogin"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setpwent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setgrent"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("time"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("times"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("wait"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("wantarray"), Completes: true}},
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("continue"), NotNext: nextNotBlock, Completes: true}},

	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("abs"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("alarm"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chop"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chdir"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("close"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("closedir"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("caller"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chr"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("cos"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chroot"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("defined"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("delete"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dbmclose"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exists"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exit"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("eof"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exp"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("each"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fc"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fileno"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gmtime"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpgrp"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getprotobyname"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpwnam"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpwuid"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpeername"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gethostbyname"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getnetbyname"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getsockname"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getgrnam"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getgrgid"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("hex"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("int"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("keys"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lc"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lcfirst"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("length"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("localtime"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("log"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lock"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lstat"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ord"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("oct"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("prototype"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("quotemeta"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("reset"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rand"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rmdir"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readdir"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rewinddir"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ref"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chomp"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("scalar"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sethostent"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setnetent"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setservent"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setprotoent"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sin"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sleep"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sqrt"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("srand"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("stat"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("study"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tell"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("telldir"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tied"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("uc"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ucfirst"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("untie"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("values"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("write"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getc"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pop"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pos"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readline"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readpipe"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readlink"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shift"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("undef"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("umask"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("eval"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("evalbytes"), Next: &parenOpen}},
	{"FUNC1", TokenClass{Kind: ptr(lexer.Word), Text: ptr("not"), Next: &parenOpen}},

	// toke.c's lop() returns FUNC for a list operator whose next significant token is a parenthesis, LSTOP
	// otherwise (BLKLSTOP for all and any); return is OLDLOP, LSTOP either way. all and any are read as
	// keywords whether or not their feature is on. The next token, not adjacency, decides: `push (@a)` is FUNC.
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("accept"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("atan2"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("bind"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("binmode"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("bless"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("crypt"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chmod"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chown"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("connect"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("die"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dbmopen"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exec"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("formline"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fcntl"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("flock"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("grep"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpriority"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getprotobynumber"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gethostbyaddr"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getnetbyaddr"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getservbyname"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getservbyport"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getsockopt"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("glob"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("index"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ioctl"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("join"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("kill"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("link"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("listen"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("map"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("mkdir"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgctl"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgget"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgrcv"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgsnd"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("open"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("opendir"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("print"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("printf"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("push"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pack"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pipe"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rename"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rindex"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("read"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("recv"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("reverse"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("say"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("select"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("seek"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semctl"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semget"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semop"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("send"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setpgrp"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setpriority"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("seekdir"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setsockopt"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmctl"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmget"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmread"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmwrite"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shutdown"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("socket"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("socketpair"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sort"), Next: &sortParen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("split"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sprintf"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("splice"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("substr"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("system"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("symlink"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("syscall"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysopen"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysseek"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysread"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("syswrite"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tie"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("truncate"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unlink"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unpack"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("utime"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unshift"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("vec"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("warn"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("waitpid"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("all"), Next: &parenOpen}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("any"), Next: &parenOpen}},
	// S_sublex_start returns FUNC for an interpolating string: a "…" holding $, @, \ or a non-ASCII byte (toke.c:7272) and
	// every qq (OP_STRINGIFY, toke.c:8879); a command, `…` and qx (OP_BACKTICK, toke.c:7284, 8889); a heredoc but <<'EOF'
	// (OP_SCALAR, toke.c:11620); and a glob <*.c> (OP_GLOB, toke.c:12123). That FUNC is always followed by SUBLEXSTART,
	// never by a parenthesis, so none of these is the FUNC of a call (perly.y:1326).
	{"FUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('"')), NotNext: nextNotParen, Completes: true}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('Q')), NotNext: nextNotParen, Completes: true}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('`')), NotNext: nextNotParen, Completes: true}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('x')), NotNext: nextNotParen, Completes: true}},
	{"FUNC", TokenClass{Kind: ptr(lexer.HeredocOpen), NotNext: nextNotParen, Completes: true}},
	{"FUNC", TokenClass{Kind: ptr(lexer.Readline), NotNext: nextNotParen, Completes: true}},

	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("abs"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("alarm"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chop"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chdir"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("close"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("closedir"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("caller"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chr"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("cos"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chroot"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("defined"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("delete"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dbmclose"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exists"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exit"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("eof"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exp"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("each"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fc"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fileno"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gmtime"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpgrp"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getprotobyname"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpwnam"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpwuid"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpeername"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gethostbyname"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getnetbyname"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getsockname"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getgrnam"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getgrgid"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("hex"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("int"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("keys"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lc"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lcfirst"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("length"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("localtime"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("log"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lock"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lstat"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ord"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("oct"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("prototype"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("quotemeta"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("reset"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rand"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rmdir"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readdir"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rewinddir"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ref"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chomp"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("scalar"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sethostent"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setnetent"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setservent"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setprotoent"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sin"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sleep"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sqrt"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("srand"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("stat"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("study"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tell"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("telldir"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tied"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("uc"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ucfirst"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("untie"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("values"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("write"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getc"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pop"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pos"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readline"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readpipe"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readlink"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shift"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("undef"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("umask"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("eval"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("evalbytes"), NotNext: nextNotParen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-r")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-w")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-x")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-o")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-R")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-W")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-X")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-O")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-e")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-z")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-s")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-f")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-d")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-l")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-p")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-S")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-u")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-g")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-k")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-b")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-c")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-t")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-T")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-B")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-M")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-A")}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-C")}},

	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("return")}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("accept"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("atan2"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("bind"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("binmode"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("bless"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("crypt"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chmod"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chown"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("connect"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("die"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dbmopen"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exec"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("formline"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fcntl"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("flock"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("grep"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpriority"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getprotobynumber"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gethostbyaddr"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getnetbyaddr"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getservbyname"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getservbyport"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getsockopt"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("glob"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("index"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ioctl"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("join"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("kill"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("link"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("listen"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("map"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("mkdir"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgctl"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgget"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgrcv"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgsnd"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("open"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("opendir"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("print"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("printf"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("push"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pack"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pipe"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rename"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rindex"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("read"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("recv"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("reverse"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("say"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("select"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("seek"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semctl"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semget"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semop"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("send"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setpgrp"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setpriority"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("seekdir"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setsockopt"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmctl"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmget"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmread"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmwrite"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shutdown"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("socket"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("socketpair"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sort"), NotNext: []Neighbour{sortParen}}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("split"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sprintf"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("splice"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("substr"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("system"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("symlink"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("syscall"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysopen"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysseek"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysread"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("syswrite"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tie"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("truncate"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unlink"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unpack"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("utime"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unshift"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("vec"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("warn"), NotNext: nextNotParen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("waitpid"), NotNext: nextNotParen}},

	{"BLKLSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("all"), NotNext: nextNotParen}},
	{"BLKLSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("any"), NotNext: nextNotParen}},

	{"POWOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("**")}},

	{"MULOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("*")}},
	{"MULOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("/")}},
	{"MULOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("%")}},

	{"ADDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("+")}},
	{"ADDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-")}},
	{"ADDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(".")}},

	{"DOLSHARP", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('#')), NotPrev1: afterArrow, Completes: true}},
	{"DOLSHARP", TokenClass{Kind: ptr(lexer.Variable), Text: ptr("$#*"), Prev1: &arrow}},

	// A `{` the lexer did not open as a block is an anonymous hash unless it follows what ClassOf makes
	// a subscript's or a block's: a variable, a closing bracket, `->` or a deref sigil.
	{"HASHBRACK", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), NotPrev1: notHashBrace}},

	// A declared sub without a prototype yyl_subproto takes: toke.c emits a zero-width NOAMP and forces the word
	// (toke.c:7980-7987), which the word stands for as NOAMP (perly.y:1646's fit). PL_expect is XTERM after it.
	{"NOAMP", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Prefix: ptr(byte(0)), NotPrev1: forcedWord, NotPrev2: afterUse, NotNext: nextUndecided}},

	{"COLONATTR", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":")}},

	// FORMLBRACK

	// toke.c has no FormatBody: the picture lines, argument lines and closing . (toke.c:6890, yyl_rightcurly) are
	// FORMRBRACK THING FORMLBRACK … FORMRBRACK PERLY_DOT, and internal/lexer emits them as one token
	// (01a1a5b7-1bcb-7c4c-bbdc-1abedccd7e6b), which stands for all of them as FORMRBRACK.
	// shortcut: the argument lines are not parsed, so a dangling operator there (`$x +`) is accepted where perl refuses it
	// (01a1a5b7-1bcb-7c4c-bbdc-1abedccd7e6b), upgrade when the lexer emits the body's tokens.
	{"FORMRBRACK", TokenClass{Kind: ptr(lexer.FormatBody)}},

	// SUBLEXSTART

	// SUBLEXEND

	{"PHASER", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ADJUST")}}, // PHASERBLOCK(KEY_ADJUST), toke.c:8331

	{"OROP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("or"), NotNext: nextNotFat}},
	{"OROP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("xor"), NotNext: nextNotFat}},

	{"ANDOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("and"), NotNext: nextNotFat}},

	{"NOTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("not"), NotNext: nextNotParen}},

	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("**=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("+=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(".=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("*=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("/=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("%=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("&=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("|=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("^=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("<<=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(">>=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("&&=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("||=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("//=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("&.=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("|.=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("^.=")}},
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("^^=")}}, // toke.c yyl_caret: OP_XOR, toke.c:6436-6439
	{"ASSIGNOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("x=")}},

	{"PERLY_QUESTION_MARK", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("?")}},

	// A : is the attribute colon, not the ternary's, when it follows my, our or state and a variable and an
	// identifier character follows it (toke.c:6456 yyl_colon at 94e50866, PL_in_my set by yyl_my, toke.c:7572).
	// The three classes are the union that leaves that case out: not after a variable, after a variable
	// that no declarator precedes, and after one with no identifier next.
	// shortcut: two cases stay accepted that perl refuses. my ($x) : word, since two tokens back a my list's
	// ) is a parenthesised declaration's (my $x) : word, which perl accepts; and the colon that separates
	// attributes ($c ? my $x : shared : 1), since a word and a colon before it look the same in
	// $a ? $b ? 1 : c : d, which perl accepts. Upgrade both when the lexer marks that ) and the separator
	// (v0.1 lexer issue 01a1a314-e3f8-765f-abf2-159425ed764e).
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), NotPrev1: []Neighbour{{Kind: lexer.Variable}, {Kind: lexer.Variable, Array: true}}}},
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), Prev1: &Neighbour{Kind: lexer.Variable}, NotPrev2: declarators}},
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), Prev1: &Neighbour{Kind: lexer.Variable}, NotNextIdent: true}},
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), Prev1: &Neighbour{Kind: lexer.Variable, Array: true}, NotPrev2: declarators}},
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), Prev1: &Neighbour{Kind: lexer.Variable, Array: true}, NotNextIdent: true}},

	{"OROR", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("||")}},
	{"OROR", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("^^")}}, // toke.c yyl_caret: OP_XOR

	{"DORDOR", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("//")}},

	{"ANDAND", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("&&")}},

	{"BITOROP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("|")}},
	{"BITOROP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("^")}},
	{"BITOROP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("|.")}},
	{"BITOROP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("^.")}},

	{"BITANDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("&")}},
	{"BITANDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("&.")}},

	{"CHEQOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("==")}},
	{"CHEQOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("!=")}},
	{"CHEQOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("eq")}},
	{"CHEQOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ne")}},

	{"NCEQOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("<=>")}},
	{"NCEQOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("~~")}},
	{"NCEQOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("cmp")}},

	{"CHRELOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("<")}},
	{"CHRELOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(">")}},
	{"CHRELOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("<=")}},
	{"CHRELOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(">=")}},
	{"CHRELOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lt")}},
	{"CHRELOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gt")}},
	{"CHRELOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("le")}},
	{"CHRELOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ge")}},

	{"NCRELOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("isa")}},

	{"SHIFTOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("<<")}},
	{"SHIFTOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(">>")}},

	{"MATCHOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("=~")}},
	{"MATCHOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("!~")}},

	{"PERLY_EXCLAMATION_MARK", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("!")}},

	{"PERLY_TILDE", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("~")}},
	{"PERLY_TILDE", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("~.")}},

	{"REFGEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("\\")}},

	{"PREINC", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("++")}},

	{"PREDEC", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("--")}},

	{"POSTINC", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("++"), Completes: true}},

	{"POSTDEC", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("--"), Completes: true}},

	// POSTJOIN

	{"ARROW", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("->")}},

	{"PERLY_PAREN_CLOSE", TokenClass{Kind: ptr(lexer.CloseBracket), Text: ptr(")"), Completes: true}},

	{"PERLY_PAREN_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("(")}},
}

// xrefListOperators are the list operators toke.c's lop() calls with XREF, the ones a block can follow.
var xrefListOperators = []string{"print", "printf", "say", "sort", "map", "grep", "exec", "system"}

// init derives the x operator's classes. toke.c's KEY_x (toke.c:9209-9219 at 94e50866) makes x the
// repetition operator exactly when PL_expect is XOPERATOR, and a word otherwise (BAREWORD, which no
// neighbour limits). PL_expect is XOPERATOR right after a token that completes a term, so MULOP x
// follows each class marked Completes, as the neighbour of that class: a variable (toke.c's scan_ident
// callers), a number or string (scan_num, scan_str), a closing bracket (yyl_rightparen,
// yyl_rightcurly), a postfix ++ or -- (the `case '+'` and `case '-'` arms), a FUNC0 or FUNC0OP word
// (FUN0, FUN0OP), a bareword (yyl_just_a_word), a method name (force_word after ->) and a name called
// with & or after * (the same, and yyl_star's scan_ident), whatever they spell. The named, list and filetest operators wait for a term, so
// x after them is a word, as after a comma or an operator.
func init() {
	perlyClasses = addStatementWords(perlyClasses) // before x: the words complete a term as any word does
	// An x class has its neighbour either right before it (prev2 false) or two back.
	type repeat struct {
		n     Neighbour
		prev2 bool
	}
	seen := map[repeat]bool{}
	for _, d := range perlyClasses {
		c := d.Key
		if !c.Completes {
			continue
		}
		x := TokenClass{Kind: ptr(lexer.Word), Text: ptr("x")}
		if c.Prev1 != nil {
			x.Prev2 = c.Prev1 // the name is whatever word follows it
		} else {
			n, _ := c.CompletedBy()
			x.Prev1 = &n
			x.NotPrev2 = c.NotPrev1 // where the class does not match, its token completes nothing
		}
		key := repeat{prev2: x.Prev1 == nil}
		if key.prev2 {
			key.n = *x.Prev2
		} else {
			key.n = *x.Prev1
		}
		// A restricted x is always added and never deduplicates: classes union, so another class's x for the
		// same neighbour widens it rather than being dropped.
		if x.NotPrev2 == nil {
			if seen[key] {
				continue
			}
			seen[key] = true
		}
		perlyClasses = append(perlyClasses, ClassDecl{"MULOP", x})
	}
	perlyClasses = addStatementX(perlyClasses) // after x's MULOP: these complete a term the word x's BAREWORD already completes
	limitBlockBrace(perlyClasses)
	perlyClasses = cutBeforeFatComma(perlyClasses)
	addVersions() // after x and the cut: neither applies to the version classes
}

// statementWords are the keywords toke.c reads as a sub's name only where a statement starts, and as any other word
// elsewhere (yyl_just_a_word, toke.c:8313-8322).
var statementWords = []string{"AUTOLOAD", "DESTROY", "BEGIN", "UNITCHECK", "CHECK", "INIT", "END"}

// addStatementWords gives each statement word every class of a word that no class names (one whose NotText is every
// keyword), after anything but a statement start.
func addStatementWords(classes []ClassDecl) []ClassDecl {
	keywords := lexer.Keywords()
	for _, d := range classes {
		if !slices.Equal(d.Key.NotText, keywords) {
			continue
		}
		for _, w := range statementWords {
			c := d.Key
			c.Text, c.NotText, c.NotPrev1 = &w, nil, append(slices.Clone(c.NotPrev1), stateStarts...)
			classes = append(classes, ClassDecl{d.Terminal, c})
		}
	}
	return classes
}

// addStatementX gives x, where a statement starts, every class of a word that no class names and no neighbour before it
// decides: no operator is expected there, so KEY_x falls through to a word (toke.c:9209-9219) that scope may make a
// declared sub, as any word. x has BAREWORD and PRIVATEREF classes of its own.
func addStatementX(classes []ClassDecl) []ClassDecl {
	keywords := lexer.Keywords()
	for _, d := range classes {
		if !slices.Equal(d.Key.NotText, keywords) || d.Key.Prev1 != nil || d.Key.Prev2 != nil ||
			d.Terminal == "BAREWORD" || d.Terminal == "PRIVATEREF" {
			continue
		}
		for _, s := range stateStarts {
			c := d.Key
			c.Text, c.NotText, c.Prev1 = ptr("x"), nil, &s
			classes = append(classes, ClassDecl{d.Terminal, c})
		}
	}
	// The { after it is a block whatever the lexer decided, as after any word no class names (METHCALL0's XBLOCKTERM).
	for _, s := range stateStarts {
		classes = append(classes, ClassDecl{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"),
			OpensBlock: ptr(false), Prev2: &s, Prev1: &Neighbour{Kind: lexer.Word, Text: "x"}}})
	}
	return classes
}

// cutBeforeFatComma takes every class keyed on a word's text that init has made so far away before =>, x's above
// included (addVersions runs after it): toke.c makes any word there a string before it looks the word
// up (yyl_keylookup, toke.c:9332-9336; a keyword with a comment or line break between, toke.c:9440-9453; any other
// word, yyl_just_a_word, toke.c:8177), so the BAREWORD class for it stands alone. A word forced to a name (namedWord) skips that check, but no class keyed on a word's text names one.
// CORE::kw reaches yyl_key_core before the check (toke.c:9322-9326), so each class cut keeps a copy for it there.
func cutBeforeFatComma(classes []ClassDecl) []ClassDecl {
	for i, d := range classes {
		if k := d.Key; k.Kind != nil && *k.Kind == lexer.Word && k.Text != nil && !slices.Contains(k.NotNext, fatComma) {
			if k.Next == nil && k.Prefix == nil {
				k.Prefix, k.Next = ptr(byte(CoreQualified)), &fatComma
				classes = append(classes, ClassDecl{d.Terminal, k})
			}
			classes[i].Key.NotNext = append(slices.Clone(d.Key.NotNext), fatComma)
		}
	}
	return classes
}

// limitBlockBrace takes the { away from a block after a list operator whose PL_expect is XTERM: toke.c
// reads it as an anonymous hash, whatever the lexer decided, so `sprintf {$fh} $x` has no block to be
// an indirob; only the XREF operators read it as one. Only the { class that no neighbour conditions is
// cut: a class with a Prev1, Prev2, Next, NotText or NotNext of its own (as `sub open {`'s is) is a different reading of
// the { and keeps its reach.
func limitBlockBrace(classes []ClassDecl) {
	var xterm []Neighbour
	for _, d := range classes {
		if k := d.Key; d.Terminal == "LSTOP" && k.Text != nil && !slices.Contains(xrefListOperators, *k.Text) {
			xterm = append(xterm, Neighbour{Kind: lexer.Word, Text: *k.Text})
		}
	}
	for i, d := range classes {
		if k := d.Key; d.Terminal == "PERLY_BRACE_OPEN" && k.Prev1 == nil && k.Prev2 == nil && k.Next == nil &&
			k.NotText == nil && len(k.NotNext) == 0 {
			classes[i].Key.NotPrev1 = append(slices.Clone(k.NotPrev1), xterm...)
		}
	}
}
