// ABOUTME: The chart grammar's data file: perly.y's rules, one per line, each with the alternative it cites.
// ABOUTME: One placeholder per perly.y alternative, in perly.y order; a family replaces its own with the rules.

package grammar

import (
	"slices"

	"tamarou.com/pvm/internal/lexer"
)

// perlyRules is `lhs: rhs # perly.y:N` per rule, or `lhs: rhs # perly.y:N, fitted: <what the lexer
// emits instead>`. A line starting with # is a placeholder for the alternative it names. A blank line
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

# perly.y:278 bare_statement_class_declaration: KW_CLASS BAREWORD BAREWORD optattrlist PERLY_SEMICOLON

# perly.y:294 bare_statement_class_definition: KW_CLASS BAREWORD BAREWORD optattrlist PERLY_BRACE_OPEN remember $@8 stmtseq PERLY_BRACE_CLOSE

# perly.y:317 bare_statement_default: KW_DEFAULT block

# perly.y:325 bare_statement_defer: KW_DEFER mblock

bare_statement_expression: sideff PERLY_SEMICOLON # perly.y:333

# perly.y:340 bare_statement_field_declaration: fielddecl PERLY_SEMICOLON

# perly.y:348 bare_statement_for: KW_FOR PERLY_PAREN_OPEN remember mnexpr PERLY_SEMICOLON $@9 texpr PERLY_SEMICOLON $@10 mintro mnexpr PERLY_PAREN_CLOSE mblock

# perly.y:379 bare_statement_for: KW_FOR KW_MY remember my_scalar PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock cont

# perly.y:392 bare_statement_for: KW_FOR KW_MY remember PERLY_PAREN_OPEN my_list_of_itervars PERLY_PAREN_CLOSE PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock cont

# perly.y:411 bare_statement_for: KW_FOR scalar PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock cont

# perly.y:423 bare_statement_for: KW_FOR my_refgen remember my_var @11 PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock cont

# perly.y:452 bare_statement_for: KW_FOR REFGEN refgen_topic PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock cont

# perly.y:474 bare_statement_for: KW_FOR PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock cont

# perly.y:488 bare_statement_format: KW_FORMAT startformsub formname formblock

# perly.y:504 bare_statement_given: KW_GIVEN PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock

bare_statement_if: KW_IF PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock else # perly.y:517

bare_statement_null: PERLY_SEMICOLON # perly.y:532

# perly.y:540 bare_statement_package_declaration: KW_PACKAGE BAREWORD BAREWORD PERLY_SEMICOLON

# perly.y:556 bare_statement_package_definition: KW_PACKAGE BAREWORD BAREWORD PERLY_BRACE_OPEN remember $@12 stmtseq PERLY_BRACE_CLOSE

# perly.y:574 bare_statement_phaser: PHASER startsub $@13 optsubbody

# perly.y:608 bare_statement_sub_signature: sigsub_or_method_named subname startsub $@14 subattrlist optsigsubbody

# perly.y:638 bare_statement_sub_traditional: KW_SUB_named subname startsub $@15 proto subattrlist optsubbody

# perly.y:662 bare_statement_try_catch: KW_TRY mblock KW_CATCH remember catch_paren $@16 mblock finally

bare_statement_unless: KW_UNLESS PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock else # perly.y:684

bare_statement_until: KW_UNTIL PERLY_PAREN_OPEN remember iexpr PERLY_PAREN_CLOSE mintro mblock cont # perly.y:699

bare_statement_utilize: KW_USE_or_NO startsub BAREWORD optlistexpr PERLY_SEMICOLON # perly.y:714, fitted: toke.c forces an empty version token before the module name, and pvm emits nothing for it

# perly.y:732 bare_statement_when: KW_WHEN PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock

bare_statement_while: KW_WHILE PERLY_PAREN_OPEN remember texpr PERLY_PAREN_CLOSE mintro mblock cont # perly.y:744

# perly.y:759 bare_statement_yadayada: YADAYADA PERLY_SEMICOLON

# perly.y:768 sigsub_or_method_named: KW_SUB_named_sig

# perly.y:770 sigsub_or_method_named: KW_METHOD_named

block: PERLY_BRACE_OPEN remember stmtseq PERLY_BRACE_CLOSE # perly.y:775

empty: # perly.y:783

# perly.y:787 formblock: PERLY_EQUAL_SIGN remember PERLY_SEMICOLON FORMRBRACK formstmtseq PERLY_SEMICOLON PERLY_DOT

remember: # perly.y:794

mblock: PERLY_BRACE_OPEN mremember stmtseq PERLY_BRACE_CLOSE # perly.y:799

mremember: # perly.y:806

# perly.y:812 catch_paren: empty

# perly.y:815 catch_paren: PERLY_PAREN_OPEN $@18 scalar $@19 PERLY_PAREN_CLOSE

stmtseq: empty # perly.y:825

stmtseq: stmtseq fullstmt # perly.y:826

# perly.y:836 formstmtseq: empty

# perly.y:837 formstmtseq: formstmtseq formline

fullstmt: barestmt # perly.y:846

# perly.y:850 fullstmt: labfullstmt

# perly.y:854 labfullstmt: LABEL barestmt

# perly.y:861 labfullstmt: LABEL labfullstmt

# perly.y:878 barestmt: PLUGSTMT

barestmt: bare_statement_block # perly.y:879

# perly.y:880 barestmt: bare_statement_class_declaration

# perly.y:881 barestmt: bare_statement_class_definition

# perly.y:882 barestmt: bare_statement_default

# perly.y:883 barestmt: bare_statement_defer

barestmt: bare_statement_expression # perly.y:884

# perly.y:885 barestmt: bare_statement_field_declaration

# perly.y:886 barestmt: bare_statement_for

# perly.y:887 barestmt: bare_statement_format

# perly.y:888 barestmt: bare_statement_given

barestmt: bare_statement_if # perly.y:889

barestmt: bare_statement_null # perly.y:890

# perly.y:891 barestmt: bare_statement_package_declaration

# perly.y:892 barestmt: bare_statement_package_definition

# perly.y:893 barestmt: bare_statement_phaser

# perly.y:894 barestmt: bare_statement_sub_signature

# perly.y:895 barestmt: bare_statement_sub_traditional

# perly.y:896 barestmt: bare_statement_try_catch

barestmt: bare_statement_unless # perly.y:897

barestmt: bare_statement_until # perly.y:898

barestmt: bare_statement_utilize # perly.y:899

# perly.y:900 barestmt: bare_statement_when

barestmt: bare_statement_while # perly.y:901

# perly.y:902 barestmt: bare_statement_yadayada

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

# perly.y:979 finally: %empty

# perly.y:981 finally: KW_FINALLY block

mintro: # perly.y:986

# perly.y:993 nexpr: empty

# perly.y:994 nexpr: sideff

texpr: # perly.y:998

texpr: expr # perly.y:1002

iexpr: expr # perly.y:1006

mexpr: expr # perly.y:1011

# perly.y:1015 mnexpr: nexpr

# perly.y:1019 formname: BAREWORD

# perly.y:1020 formname: empty

startsub: # perly.y:1023

# perly.y:1029 startanonsub: %empty

# perly.y:1034 startanonmethod: %empty

# perly.y:1039 startformsub: %empty

# perly.y:1045 subname: BAREWORD

# perly.y:1046 subname: PRIVATEREF

# perly.y:1051 proto: empty

# perly.y:1052 proto: PROTOTYPE

# perly.y:1057 subattrlist: empty

# perly.y:1058 subattrlist: COLONATTR ATTRLIST

# perly.y:1065 subattrlist: COLONATTR

attrlist: COLONATTR attr attrs # perly.y:1070, fitted: pvm emits each attribute as a word after its colon with its argument, if any, inside parentheses, where toke.c delivers one ATTRLIST; one token of argument
attrlist: COLONATTR attr attrs COLONATTR # perly.y:1070, fitted: perl accepts one colon after the attributes, which toke.c leaves out of its ATTRLIST
attr: ATTRLIST # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attr: ATTRLIST PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attr: ATTRLIST PERLY_PAREN_OPEN ATTRLIST PERLY_PAREN_CLOSE # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attrs: empty # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attrs: attrs COLONATTR attr # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST
attrs: attrs attr # perly.y:1070, fitted: the attribute words and their arguments, which toke.c delivers as one ATTRLIST

attrlist: COLONATTR # perly.y:1072

# perly.y:1077 optattrlist: empty

# perly.y:1078 optattrlist: attrlist

# perly.y:1087 sigvar: %empty

# perly.y:1089 sigvar: PRIVATEREF

# perly.y:1094 sigslurpsigil: PERLY_SNAIL

# perly.y:1096 sigslurpsigil: PERLY_PERCENT_SIGN

# perly.y:1100 sigslurpelem: sigslurpsigil sigvar

# perly.y:1105 sigslurpelem: sigslurpsigil sigvar ASSIGNOP

# perly.y:1109 sigslurpelem: sigslurpsigil sigvar ASSIGNOP term

# perly.y:1115 optcolon: %empty

# perly.y:1117 optcolon: PERLY_COLON

# perly.y:1123 sigscalarelem: optcolon PERLY_DOLLAR sigvar

# perly.y:1133 sigscalarelem: optcolon PERLY_DOLLAR sigvar ASSIGNOP optsigscalardefault

# perly.y:1146 optsigscalardefault: %empty

# perly.y:1148 optsigscalardefault: term

# perly.y:1153 sigelem: sigscalarelem

# perly.y:1155 sigelem: sigslurpelem

# perly.y:1163 siglist: siglist PERLY_COMMA

# perly.y:1164 siglist: siglist PERLY_COMMA sigelem

# perly.y:1165 siglist: sigelem

optsiglist: empty # perly.y:1170

# perly.y:1171 optsiglist: siglist

# perly.y:1176 optsubsignature: empty

# perly.y:1177 optsubsignature: subsignature

# perly.y:1181 subsignature: PERLY_PAREN_OPEN subsigguts PERLY_PAREN_CLOSE

subsigguts: optsiglist # perly.y:1185

# perly.y:1218 optsubbody: subbody

# perly.y:1219 optsubbody: PERLY_SEMICOLON

# perly.y:1224 subbody: remember PERLY_BRACE_OPEN stmtseq PERLY_BRACE_CLOSE

# perly.y:1236 optsigsubbody: sigsubbody

# perly.y:1237 optsigsubbody: PERLY_SEMICOLON

# perly.y:1241 sigsubbody: remember optsubsignature PERLY_BRACE_OPEN $@21 stmtseq PERLY_BRACE_CLOSE

expr: expr ANDOP expr # perly.y:1254

expr: expr PLUGIN_LOGICAL_AND_LOW_OP expr # perly.y:1256

expr: expr OROP expr # perly.y:1258

expr: expr PLUGIN_LOGICAL_OR_LOW_OP expr # perly.y:1260

expr: listexpr # perly.y:1262

listexpr: listexpr PERLY_COMMA # perly.y:1266

listexpr: listexpr PERLY_COMMA term # perly.y:1268

listexpr: term # perly.y:1273

listop: LSTOP indirob listexpr # perly.y:1277

listop: BLKLSTOP block listexpr # perly.y:1281

# perly.y:1285 listop: FUNC PERLY_PAREN_OPEN indirob expr PERLY_PAREN_CLOSE

listop: term ARROW methodname PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE # perly.y:1289

listop: term ARROW methodname # perly.y:1295

# perly.y:1300 listop: term ARROW PERLY_AMPERSAND subname PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE

# perly.y:1306 listop: term ARROW PERLY_AMPERSAND subname

# perly.y:1312 listop: METHCALL0 indirob optlistexpr

# perly.y:1318 listop: METHCALL indirob PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE

listop: LSTOP optlistexpr # perly.y:1324

listop: FUNC PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE # perly.y:1326

# perly.y:1328 listop: FUNC SUBLEXSTART optexpr SUBLEXEND

# perly.y:1330 listop: LSTOPSUB startanonsub block @22 optlistexpr

methodname: METHCALL0 # perly.y:1344

methodname: scalar # perly.y:1345

# perly.y:1349 subscripted: gelem PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1353 subscripted: scalar PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1356 subscripted: term ARROW PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1361 subscripted: subscripted PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1366 subscripted: scalar PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1369 subscripted: term ARROW PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1373 subscripted: subscripted PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1377 subscripted: term ARROW PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1383 subscripted: term ARROW PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE

# perly.y:1391 subscripted: subscripted PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE

# perly.y:1398 subscripted: subscripted PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1404 subscripted: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1406 subscripted: QWLIST PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1408 subscripted: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

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

termbinop: termrelop # perly.y:1436

termbinop: termeqop # perly.y:1438

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

termrelop: relopchain # perly.y:1462

termrelop: term NCRELOP term # perly.y:1464

# perly.y:1466 termrelop: termrelop NCRELOP

# perly.y:1468 termrelop: termrelop CHRELOP

termrelop: term PLUGIN_REL_OP term # perly.y:1470

relopchain: term CHRELOP term # perly.y:1474

relopchain: relopchain CHRELOP term # perly.y:1476

termeqop: eqopchain # perly.y:1480

termeqop: term NCEQOP term # perly.y:1482

# perly.y:1484 termeqop: termeqop NCEQOP

# perly.y:1486 termeqop: termeqop CHEQOP

eqopchain: term CHEQOP term # perly.y:1490

eqopchain: eqopchain CHEQOP term # perly.y:1492

termunop: PERLY_MINUS term # perly.y:1497

termunop: PERLY_PLUS term # perly.y:1499

termunop: PERLY_EXCLAMATION_MARK term # perly.y:1502

termunop: PERLY_TILDE term # perly.y:1504

termunop: term POSTINC # perly.y:1506

termunop: term POSTDEC # perly.y:1509

# perly.y:1512 termunop: term POSTJOIN

termunop: PREINC term # perly.y:1523

termunop: PREDEC term # perly.y:1526

anonymous: PERLY_BRACKET_OPEN optexpr PERLY_BRACKET_CLOSE # perly.y:1534

anonymous: HASHBRACK optexpr PERLY_BRACE_CLOSE # perly.y:1536, fitted: the ; before } is toke.c's, not the source's

# perly.y:1538 anonymous: KW_SUB_anon startanonsub proto subattrlist subbody

# perly.y:1541 anonymous: KW_SUB_anon startanonsub proto subattrlist

# perly.y:1543 anonymous: KW_SUB_anon_sig startanonsub subattrlist sigsubbody

# perly.y:1546 anonymous: KW_SUB_anon_sig startanonsub subattrlist

# perly.y:1548 anonymous: KW_METHOD_anon startanonmethod subattrlist sigsubbody

# perly.y:1553 anonymous: KW_METHOD_anon startanonmethod subattrlist

termdo: KW_DO term # perly.y:1558

termdo: KW_DO block # perly.y:1560

term: termbinop # perly.y:1564

term: termunop # perly.y:1565

term: anonymous # perly.y:1566

term: termdo # perly.y:1567

term: term PERLY_QUESTION_MARK term PERLY_COLON term # perly.y:1568

term: REFGEN term # perly.y:1570

term: myattrterm # perly.y:1572

term: KW_LOCAL term # perly.y:1574

term: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1576

term: QWLIST # perly.y:1578

term: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1580

term: scalar # perly.y:1582

term: star # perly.y:1584

term: hsh # perly.y:1586

term: ary # perly.y:1588

term: arylen # perly.y:1590

# perly.y:1592 term: subscripted

# perly.y:1594 term: sliceme PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1604 term: kvslice PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1614 term: sliceme PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1624 term: kvslice PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

term: THING # perly.y:1634

term: amper # perly.y:1636

term: amper PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1638

term: amper PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1641

# perly.y:1646 term: NOAMP subname optlistexpr

# perly.y:1650 term: term ARROW PERLY_DOLLAR PERLY_STAR

# perly.y:1652 term: term ARROW PERLY_SNAIL PERLY_STAR

# perly.y:1654 term: term ARROW PERLY_PERCENT_SIGN PERLY_STAR

# perly.y:1656 term: term ARROW PERLY_AMPERSAND PERLY_STAR

# perly.y:1659 term: term ARROW PERLY_STAR PERLY_STAR

term: LOOPEX # perly.y:1661

term: LOOPEX term # perly.y:1664

term: NOTOP listexpr # perly.y:1666

term: UNIOP # perly.y:1668

term: UNIOP block # perly.y:1670

term: UNIOP term # perly.y:1672

term: KW_REQUIRE # perly.y:1674

term: KW_REQUIRE term # perly.y:1676

# perly.y:1678 term: UNIOPSUB

# perly.y:1680 term: UNIOPSUB term

term: FUNC0 # perly.y:1683

term: FUNC0 PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1685

term: FUNC0OP # perly.y:1687

term: FUNC0OP PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1689

# perly.y:1691 term: FUNC0SUB

term: FUNC1 PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1693

term: FUNC1 PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1697

# perly.y:1699 term: PMFUNC @23 SUBLEXSTART listexpr optrepl SUBLEXEND

term: BAREWORD # perly.y:1712

term: listop # perly.y:1713

# perly.y:1714 term: PLUGEXPR

myattrterm: KW_MY myterm attrlist # perly.y:1719

myattrterm: KW_MY myterm # perly.y:1721

myattrterm: KW_MY REFGEN myterm attrlist # perly.y:1723

myattrterm: KW_MY REFGEN term # perly.y:1725

myterm: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE # perly.y:1730

myterm: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE # perly.y:1732

myterm: scalar # perly.y:1735

myterm: hsh # perly.y:1737

myterm: ary # perly.y:1739

# perly.y:1744 fieldvar: scalar

# perly.y:1749 fieldvar: hsh

# perly.y:1754 fieldvar: ary

# perly.y:1762 fielddecl: KW_FIELD fieldvar optattrlist

# perly.y:1769 fielddecl: KW_FIELD fieldvar optattrlist ASSIGNOP $@24 term

optlistexpr: empty # perly.y:1787

optlistexpr: listexpr # perly.y:1788

optexpr: empty # perly.y:1792

optexpr: expr # perly.y:1793

# perly.y:1797 optrepl: empty

# perly.y:1798 optrepl: PERLY_SLASH expr

# perly.y:1803 my_scalar: scalar

# perly.y:1808 list_of_itervars: list_of_itervars PERLY_COMMA

# perly.y:1810 list_of_itervars: list_of_itervars PERLY_COMMA scalar

# perly.y:1814 list_of_itervars: list_of_itervars PERLY_COMMA REFGEN my_var

# perly.y:1818 list_of_itervars: scalar

# perly.y:1819 list_of_itervars: REFGEN my_var

# perly.y:1825 my_list_of_itervars: list_of_itervars

# perly.y:1829 my_var: scalar

# perly.y:1830 my_var: ary

# perly.y:1831 my_var: hsh

# perly.y:1834 refgen_topic: my_var

# perly.y:1835 refgen_topic: amper

# perly.y:1838 my_refgen: KW_MY REFGEN

# perly.y:1839 my_refgen: REFGEN KW_MY

amper: PERLY_AMPERSAND BAREWORD # perly.y:1842, fitted: pvm emits & and the name as two tokens, the name standing for indirob's BAREWORD

scalar: PERLY_DOLLAR # perly.y:1846, fitted: pvm emits the sigil and the name as one Variable token

ary: PERLY_SNAIL # perly.y:1850, fitted: pvm emits the sigil and the name as one Variable token

hsh: PERLY_PERCENT_SIGN # perly.y:1856, fitted: pvm emits the sigil and the name as one Variable token

arylen: DOLSHARP # perly.y:1862, fitted: pvm emits the sigil and the name as one Variable token, $# its sigil

# perly.y:1864 arylen: term ARROW DOLSHARP PERLY_STAR

star: PERLY_STAR BAREWORD # perly.y:1868, fitted: pvm emits * and the name as two tokens, the name standing for indirob's BAREWORD

# perly.y:1872 sliceme: ary

# perly.y:1873 sliceme: term ARROW PERLY_SNAIL

# perly.y:1877 kvslice: hsh

# perly.y:1878 kvslice: term ARROW PERLY_PERCENT_SIGN

# perly.y:1882 gelem: star

# perly.y:1883 gelem: term ARROW PERLY_STAR

# perly.y:1888 indirob: BAREWORD

# perly.y:1890 indirob: scalar

indirob: block # perly.y:1892

# perly.y:1895 indirob: PRIVATEREF
`

// parenOpen and blockOpen are the next tokens that tell toke.c's FUNC1 from UNIOP and KW_CONTINUE from FUNC0;
// fatComma turns a word before it into a bareword (yyl_keylookup checks for => before any keyword).
var (
	fatComma  = Neighbour{Kind: lexer.Operator, Text: "=>"}
	parenOpen = Neighbour{Kind: lexer.Operator, Text: "("}
	blockOpen = Neighbour{Kind: lexer.Operator, Text: "{", OpensBlock: true, Brace: BraceBlock}

	// declarators are the words after which a variable is declared and a : may start its attributes.
	declarators = []Neighbour{{Kind: lexer.Word, Text: "my"}, {Kind: lexer.Word, Text: "our"}, {Kind: lexer.Word, Text: "state"}}

	// notHashBrace mirrors ClassOf's brace decision, and the two change together: after a variable, a
	// closing bracket, `->` or a deref sigil a `{` the lexer did not open as a block is a subscript's or
	// a block's, never an anonymous hash's. ClassOf's `)` and term-following-variable cases are not
	// listed: the lexer opens every `{` after them as a block, so no key reaches them here.
	notHashBrace = []Neighbour{
		{Kind: lexer.Variable}, {Kind: lexer.DerefSigil}, {Kind: lexer.Operator, Text: "->"},
		{Kind: lexer.CloseBracket, Text: "]"}, {Kind: lexer.CloseBracket, Text: "}"},
	}
)

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

	{"PERLY_BRACE_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(true)}},

	{"PERLY_BRACE_CLOSE", TokenClass{Kind: ptr(lexer.CloseBracket), Text: ptr("}"), OpensBlock: ptr(true), Completes: true}},
	{"PERLY_BRACE_CLOSE", TokenClass{Kind: ptr(lexer.CloseBracket), Text: ptr("}"), OpensBlock: ptr(false), Completes: true}},

	{"PERLY_BRACKET_OPEN", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("[")}},

	{"PERLY_BRACKET_CLOSE", TokenClass{Kind: ptr(lexer.CloseBracket), Text: ptr("]")}},

	{"PERLY_COMMA", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(",")}},
	{"PERLY_COMMA", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("=>")}},

	{"PERLY_DOLLAR", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('$')), Completes: true}},

	// PERLY_DOT

	// PERLY_EQUAL_SIGN

	{"PERLY_MINUS", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-")}},

	{"PERLY_PERCENT_SIGN", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('%')), Completes: true}},

	{"PERLY_PLUS", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("+")}},

	{"PERLY_SEMICOLON", TokenClass{Kind: ptr(lexer.Semicolon)}},

	// PERLY_SLASH

	{"PERLY_SNAIL", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('@')), Completes: true}},

	{"PERLY_STAR", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("*")}},

	// KW_FORMAT

	// KW_PACKAGE

	// KW_CLASS

	{"KW_LOCAL", TokenClass{Kind: ptr(lexer.Word), Text: ptr("local")}},

	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("my")}},
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("our")}},
	{"KW_MY", TokenClass{Kind: ptr(lexer.Word), Text: ptr("state")}},

	// KW_FIELD

	{"KW_IF", TokenClass{Kind: ptr(lexer.Word), Text: ptr("if")}},

	{"KW_ELSE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("else")}},

	{"KW_ELSIF", TokenClass{Kind: ptr(lexer.Word), Text: ptr("elsif")}},

	{"KW_UNLESS", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unless")}},

	{"KW_FOR", TokenClass{Kind: ptr(lexer.Word), Text: ptr("for")}},
	{"KW_FOR", TokenClass{Kind: ptr(lexer.Word), Text: ptr("foreach")}},

	{"KW_UNTIL", TokenClass{Kind: ptr(lexer.Word), Text: ptr("until")}},

	{"KW_WHILE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("while")}},

	{"KW_CONTINUE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("continue"), Next: &blockOpen}},

	// KW_GIVEN

	{"KW_WHEN", TokenClass{Kind: ptr(lexer.Word), Text: ptr("when")}},

	// KW_DEFAULT

	// KW_TRY

	// KW_CATCH

	// KW_FINALLY

	// KW_DEFER

	{"KW_REQUIRE", TokenClass{Kind: ptr(lexer.Word), Text: ptr("require")}},

	{"KW_DO", TokenClass{Kind: ptr(lexer.Word), Text: ptr("do")}},

	{"KW_USE_or_NO", TokenClass{Kind: ptr(lexer.Word), Text: ptr("use")}},
	{"KW_USE_or_NO", TokenClass{Kind: ptr(lexer.Word), Text: ptr("no")}},

	// KW_SUB_named

	// KW_SUB_named_sig

	// KW_SUB_anon

	// KW_SUB_anon_sig

	// KW_METHOD_named

	// KW_METHOD_anon

	// Every word that is not a keyword; the one keyword toke.c falls through to a word when no
	// operator is expected, x (KEY_x), is listed below. A keyword is a BAREWORD only where a position
	// class says so; lt, gt, le, ge, eq, ne and cmp are never words.
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), NotText: lexer.Keywords(), Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Text: ptr("x"), Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Number), Prev1: &Neighbour{Kind: lexer.Word, Text: "require"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('v')), Prev1: &Neighbour{Kind: lexer.Word, Text: "require"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('5')), Prev1: &Neighbour{Kind: lexer.Word, Text: "require"}}},
	// A glob's name: yyl_star runs scan_ident, then sets PL_expect to XOPERATOR (TERM(PERLY_STAR)), but
	// only for a * that is a glob sigil, which the neighbour class of the * says (Neighbour.Glob).
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: "*", Glob: true}, Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.FuncSigil}, Completes: true}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "use"}}},
	{"BAREWORD", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: "no"}}},

	// toke.c forces the word after -> to METHCALL0, a keyword or a quote-like operator's name included.
	{"METHCALL0", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: "->"}, Completes: true}},

	// METHCALL

	// An attribute's name follows its colon, the attribute before it or its argument's parenthesis; its
	// one-token argument, never a parenthesis, follows its own.
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: ":"}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Word, Text: OtherWord}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.CloseBracket, Text: ")"}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Word), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Number), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Quote), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},
	{"ATTRLIST", TokenClass{Kind: ptr(lexer.Variable), Prev1: &Neighbour{Kind: lexer.Operator, Text: "("}}},

	{"THING", TokenClass{Kind: ptr(lexer.Number), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('\'')), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('"')), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('q')), Completes: true}},
	{"THING", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('Q')), Completes: true}},

	// PMFUNC

	// PRIVATEREF

	{"QWLIST", TokenClass{Kind: ptr(lexer.Quote), Prefix: ptr(byte('w'))}},

	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__FILE__"), Completes: true}},
	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__LINE__"), Completes: true}},
	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__PACKAGE__"), Completes: true}},
	{"FUNC0OP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("__SUB__"), Completes: true}},

	// FUNC0SUB

	// UNIOPSUB

	// LSTOPSUB

	// LABEL

	// PROTOTYPE

	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dump")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("goto")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("last")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("next")}},
	{"LOOPEX", TokenClass{Kind: ptr(lexer.Word), Text: ptr("redo")}},

	{"DOTDOT", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("..")}},
	{"DOTDOT", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("...")}},

	// YADAYADA

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
	{"FUNC0", TokenClass{Kind: ptr(lexer.Word), Text: ptr("continue"), NotNext: &blockOpen, Completes: true}},

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
	{"FUNC", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sort"), Next: &parenOpen}},
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

	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("abs"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("alarm"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chop"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chdir"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("close"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("closedir"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("caller"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chr"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("cos"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chroot"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("defined"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("delete"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dbmclose"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exists"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exit"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("eof"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exp"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("each"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fc"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fileno"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gmtime"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpgrp"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getprotobyname"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpwnam"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpwuid"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpeername"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gethostbyname"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getnetbyname"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getsockname"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getgrnam"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getgrgid"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("hex"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("int"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("keys"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lc"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lcfirst"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("length"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("localtime"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("log"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lock"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("lstat"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ord"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("oct"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("prototype"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("quotemeta"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("reset"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rand"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rmdir"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readdir"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rewinddir"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ref"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chomp"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("scalar"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sethostent"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setnetent"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setservent"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setprotoent"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sin"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sleep"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sqrt"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("srand"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("stat"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("study"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tell"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("telldir"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tied"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("uc"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ucfirst"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("untie"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("values"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("write"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getc"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pop"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pos"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readline"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readpipe"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("readlink"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shift"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("undef"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("umask"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("eval"), NotNext: &parenOpen}},
	{"UNIOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("evalbytes"), NotNext: &parenOpen}},
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
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("accept"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("atan2"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("bind"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("binmode"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("bless"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("crypt"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chmod"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("chown"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("connect"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("die"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("dbmopen"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("exec"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("formline"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("fcntl"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("flock"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("grep"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getpriority"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getprotobynumber"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("gethostbyaddr"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getnetbyaddr"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getservbyname"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getservbyport"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("getsockopt"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("glob"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("index"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("ioctl"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("join"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("kill"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("link"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("listen"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("map"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("mkdir"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgctl"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgget"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgrcv"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("msgsnd"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("open"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("opendir"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("print"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("printf"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("push"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pack"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("pipe"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rename"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("rindex"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("read"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("recv"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("reverse"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("say"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("select"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("seek"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semctl"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semget"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("semop"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("send"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setpgrp"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setpriority"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("seekdir"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("setsockopt"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmctl"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmget"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmread"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shmwrite"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("shutdown"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("socket"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("socketpair"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sort"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("split"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sprintf"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("splice"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("substr"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("system"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("symlink"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("syscall"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysopen"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysseek"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("sysread"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("syswrite"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("tie"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("truncate"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unlink"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unpack"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("utime"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("unshift"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("vec"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("warn"), NotNext: &parenOpen}},
	{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("waitpid"), NotNext: &parenOpen}},

	{"BLKLSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("all"), NotNext: &parenOpen}},
	{"BLKLSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("any"), NotNext: &parenOpen}},

	{"POWOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("**")}},

	{"MULOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("*")}},
	{"MULOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("/")}},
	{"MULOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("%")}},

	{"ADDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("+")}},
	{"ADDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("-")}},
	{"ADDOP", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(".")}},

	{"DOLSHARP", TokenClass{Kind: ptr(lexer.Variable), Prefix: ptr(byte('#')), Completes: true}},

	// A `{` the lexer did not open as a block is an anonymous hash unless it follows what ClassOf makes
	// a subscript's or a block's: a variable, a closing bracket, `->` or a deref sigil.
	{"HASHBRACK", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(false), NotPrev1: notHashBrace}},

	// NOAMP

	{"COLONATTR", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":")}},

	// FORMLBRACK

	// FORMRBRACK

	// SUBLEXSTART

	// SUBLEXEND

	// PHASER

	{"OROP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("or"), NotNext: &fatComma}},
	{"OROP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("xor"), NotNext: &fatComma}},

	{"ANDOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("and"), NotNext: &fatComma}},

	{"NOTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("not"), NotNext: &parenOpen}},

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
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), NotPrev1: []Neighbour{{Kind: lexer.Variable}}}},
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), Prev1: &Neighbour{Kind: lexer.Variable}, NotPrev2: declarators}},
	{"PERLY_COLON", TokenClass{Kind: ptr(lexer.Operator), Text: ptr(":"), Prev1: &Neighbour{Kind: lexer.Variable}, NotNextIdent: true}},

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
		}
		key := repeat{prev2: x.Prev1 == nil}
		if key.prev2 {
			key.n = *x.Prev2
		} else {
			key.n = *x.Prev1
		}
		if !seen[key] {
			seen[key] = true
			perlyClasses = append(perlyClasses, ClassDecl{"MULOP", x})
		}
	}
	limitBlockBrace(perlyClasses)
}

// limitBlockBrace takes the { away from a block after a list operator whose PL_expect is XTERM: toke.c
// reads it as an anonymous hash, whatever the lexer decided, so `sprintf {$fh} $x` has no block to be
// an indirob; only the XREF operators read it as one. Only the { class that no neighbour conditions is
// cut: a class with a Prev1, Prev2, Next, NotText or NotNext of its own (a later `sub open {`) is a different reading of
// the { and keeps its reach. shortcut: that unconditioned class also loses the { after a sub named like
// such an operator until the sub family adds its conditioned one.
func limitBlockBrace(classes []ClassDecl) {
	var xterm []Neighbour
	for _, d := range classes {
		if k := d.Key; d.Terminal == "LSTOP" && k.Text != nil && !slices.Contains(xrefListOperators, *k.Text) {
			xterm = append(xterm, Neighbour{Kind: lexer.Word, Text: *k.Text})
		}
	}
	for i, d := range classes {
		if k := d.Key; d.Terminal == "PERLY_BRACE_OPEN" && k.Prev1 == nil && k.Prev2 == nil && k.Next == nil &&
			k.NotText == nil && k.NotNext == nil {
			classes[i].Key.NotPrev1 = append(slices.Clone(k.NotPrev1), xterm...)
		}
	}
}
