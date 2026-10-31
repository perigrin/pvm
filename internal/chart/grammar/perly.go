// ABOUTME: The chart grammar's data file: perly.y's rules, one per line, each with the alternative it cites.
// ABOUTME: One placeholder per perly.y alternative, in perly.y order; a family replaces its own with the rules.

package grammar

// perlyRules is `lhs: rhs # perly.y:N` per rule, or `lhs: rhs # perly.y:N, fitted: <what the lexer
// emits instead>`. A line starting with # is a placeholder for the alternative it names. A blank line
// separates adjacent placeholders and rules (and ledger entries), so two families' neighbouring edits
// merge cleanly; only rules citing the same perly.y:N, one family's slot, may be adjacent.
const perlyRules = `
# perly.y:187 grammar: GRAMPROG @1 remember stmtseq

# perly.y:198 grammar: GRAMEXPR @2 optexpr

# perly.y:208 grammar: GRAMBLOCK @3 block

# perly.y:221 grammar: GRAMBARESTMT @4 barestmt

# perly.y:234 grammar: GRAMFULLSTMT @5 fullstmt

# perly.y:247 grammar: GRAMSTMTSEQ @6 stmtseq

# perly.y:257 grammar: GRAMSUBSIGNATURE @7 subsigguts

# perly.y:270 bare_statement_block: block cont

# perly.y:278 bare_statement_class_declaration: KW_CLASS BAREWORD BAREWORD optattrlist PERLY_SEMICOLON

# perly.y:294 bare_statement_class_definition: KW_CLASS BAREWORD BAREWORD optattrlist PERLY_BRACE_OPEN remember $@8 stmtseq PERLY_BRACE_CLOSE

# perly.y:317 bare_statement_default: KW_DEFAULT block

# perly.y:325 bare_statement_defer: KW_DEFER mblock

# perly.y:333 bare_statement_expression: sideff PERLY_SEMICOLON

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

# perly.y:517 bare_statement_if: KW_IF PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock else

# perly.y:532 bare_statement_null: PERLY_SEMICOLON

# perly.y:540 bare_statement_package_declaration: KW_PACKAGE BAREWORD BAREWORD PERLY_SEMICOLON

# perly.y:556 bare_statement_package_definition: KW_PACKAGE BAREWORD BAREWORD PERLY_BRACE_OPEN remember $@12 stmtseq PERLY_BRACE_CLOSE

# perly.y:574 bare_statement_phaser: PHASER startsub $@13 optsubbody

# perly.y:608 bare_statement_sub_signature: sigsub_or_method_named subname startsub $@14 subattrlist optsigsubbody

# perly.y:638 bare_statement_sub_traditional: KW_SUB_named subname startsub $@15 proto subattrlist optsubbody

# perly.y:662 bare_statement_try_catch: KW_TRY mblock KW_CATCH remember catch_paren $@16 mblock finally

# perly.y:684 bare_statement_unless: KW_UNLESS PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock else

# perly.y:699 bare_statement_until: KW_UNTIL PERLY_PAREN_OPEN remember iexpr PERLY_PAREN_CLOSE mintro mblock cont

# perly.y:714 bare_statement_utilize: KW_USE_or_NO startsub $@17 BAREWORD BAREWORD optlistexpr PERLY_SEMICOLON

# perly.y:732 bare_statement_when: KW_WHEN PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock

# perly.y:744 bare_statement_while: KW_WHILE PERLY_PAREN_OPEN remember texpr PERLY_PAREN_CLOSE mintro mblock cont

# perly.y:759 bare_statement_yadayada: YADAYADA PERLY_SEMICOLON

# perly.y:768 sigsub_or_method_named: KW_SUB_named_sig

# perly.y:770 sigsub_or_method_named: KW_METHOD_named

# perly.y:775 block: PERLY_BRACE_OPEN remember stmtseq PERLY_BRACE_CLOSE

# perly.y:783 empty: %empty

# perly.y:787 formblock: PERLY_EQUAL_SIGN remember PERLY_SEMICOLON FORMRBRACK formstmtseq PERLY_SEMICOLON PERLY_DOT

# perly.y:794 remember: %empty

# perly.y:799 mblock: PERLY_BRACE_OPEN mremember stmtseq PERLY_BRACE_CLOSE

# perly.y:806 mremember: %empty

# perly.y:812 catch_paren: empty

# perly.y:815 catch_paren: PERLY_PAREN_OPEN $@18 scalar $@19 PERLY_PAREN_CLOSE

# perly.y:825 stmtseq: empty

# perly.y:826 stmtseq: stmtseq fullstmt

# perly.y:836 formstmtseq: empty

# perly.y:837 formstmtseq: formstmtseq formline

# perly.y:846 fullstmt: barestmt

# perly.y:850 fullstmt: labfullstmt

# perly.y:854 labfullstmt: LABEL barestmt

# perly.y:861 labfullstmt: LABEL labfullstmt

# perly.y:878 barestmt: PLUGSTMT

# perly.y:879 barestmt: bare_statement_block

# perly.y:880 barestmt: bare_statement_class_declaration

# perly.y:881 barestmt: bare_statement_class_definition

# perly.y:882 barestmt: bare_statement_default

# perly.y:883 barestmt: bare_statement_defer

# perly.y:884 barestmt: bare_statement_expression

# perly.y:885 barestmt: bare_statement_field_declaration

# perly.y:886 barestmt: bare_statement_for

# perly.y:887 barestmt: bare_statement_format

# perly.y:888 barestmt: bare_statement_given

# perly.y:889 barestmt: bare_statement_if

# perly.y:890 barestmt: bare_statement_null

# perly.y:891 barestmt: bare_statement_package_declaration

# perly.y:892 barestmt: bare_statement_package_definition

# perly.y:893 barestmt: bare_statement_phaser

# perly.y:894 barestmt: bare_statement_sub_signature

# perly.y:895 barestmt: bare_statement_sub_traditional

# perly.y:896 barestmt: bare_statement_try_catch

# perly.y:897 barestmt: bare_statement_unless

# perly.y:898 barestmt: bare_statement_until

# perly.y:899 barestmt: bare_statement_utilize

# perly.y:900 barestmt: bare_statement_when

# perly.y:901 barestmt: bare_statement_while

# perly.y:902 barestmt: bare_statement_yadayada

# perly.y:906 formline: THING formarg

# perly.y:924 formarg: empty

# perly.y:925 formarg: FORMLBRACK stmtseq FORMRBRACK

# perly.y:929 condition: expr

# perly.y:933 sideff: error

# perly.y:935 sideff: expr

# perly.y:937 sideff: expr KW_IF condition

# perly.y:940 sideff: expr KW_UNLESS condition

# perly.y:943 sideff: expr KW_WHILE condition

# perly.y:945 sideff: expr KW_UNTIL iexpr

# perly.y:947 sideff: expr KW_FOR condition

# perly.y:950 sideff: expr KW_WHEN condition

# perly.y:956 else: empty

# perly.y:957 else: KW_ELSE mblock

# perly.y:962 else: KW_ELSIF PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock else

# perly.y:973 cont: empty

# perly.y:974 cont: KW_CONTINUE block

# perly.y:979 finally: %empty

# perly.y:981 finally: KW_FINALLY block

# perly.y:986 mintro: %empty

# perly.y:993 nexpr: empty

# perly.y:994 nexpr: sideff

# perly.y:998 texpr: %empty

# perly.y:1002 texpr: expr

# perly.y:1006 iexpr: expr

# perly.y:1011 mexpr: expr

# perly.y:1015 mnexpr: nexpr

# perly.y:1019 formname: BAREWORD

# perly.y:1020 formname: empty

# perly.y:1023 startsub: %empty

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

# perly.y:1070 attrlist: COLONATTR ATTRLIST

# perly.y:1072 attrlist: COLONATTR

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

# perly.y:1170 optsiglist: empty

# perly.y:1171 optsiglist: siglist

# perly.y:1176 optsubsignature: empty

# perly.y:1177 optsubsignature: subsignature

# perly.y:1181 subsignature: PERLY_PAREN_OPEN subsigguts PERLY_PAREN_CLOSE

# perly.y:1185 subsigguts: $@20 optsiglist

# perly.y:1218 optsubbody: subbody

# perly.y:1219 optsubbody: PERLY_SEMICOLON

# perly.y:1224 subbody: remember PERLY_BRACE_OPEN stmtseq PERLY_BRACE_CLOSE

# perly.y:1236 optsigsubbody: sigsubbody

# perly.y:1237 optsigsubbody: PERLY_SEMICOLON

# perly.y:1241 sigsubbody: remember optsubsignature PERLY_BRACE_OPEN $@21 stmtseq PERLY_BRACE_CLOSE

# perly.y:1254 expr: expr ANDOP expr

# perly.y:1256 expr: expr PLUGIN_LOGICAL_AND_LOW_OP expr

# perly.y:1258 expr: expr OROP expr

# perly.y:1260 expr: expr PLUGIN_LOGICAL_OR_LOW_OP expr

# perly.y:1262 expr: listexpr

# perly.y:1266 listexpr: listexpr PERLY_COMMA

# perly.y:1268 listexpr: listexpr PERLY_COMMA term

# perly.y:1273 listexpr: term

# perly.y:1277 listop: LSTOP indirob listexpr

# perly.y:1281 listop: BLKLSTOP block listexpr

# perly.y:1285 listop: FUNC PERLY_PAREN_OPEN indirob expr PERLY_PAREN_CLOSE

# perly.y:1289 listop: term ARROW methodname PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE

# perly.y:1295 listop: term ARROW methodname

# perly.y:1300 listop: term ARROW PERLY_AMPERSAND subname PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE

# perly.y:1306 listop: term ARROW PERLY_AMPERSAND subname

# perly.y:1312 listop: METHCALL0 indirob optlistexpr

# perly.y:1318 listop: METHCALL indirob PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE

# perly.y:1324 listop: LSTOP optlistexpr

# perly.y:1326 listop: FUNC PERLY_PAREN_OPEN optexpr PERLY_PAREN_CLOSE

# perly.y:1328 listop: FUNC SUBLEXSTART optexpr SUBLEXEND

# perly.y:1330 listop: LSTOPSUB startanonsub block @22 optlistexpr

# perly.y:1344 methodname: METHCALL0

# perly.y:1345 methodname: scalar

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

# perly.y:1413 termbinop: term PLUGIN_HIGH_OP term

# perly.y:1415 termbinop: term ASSIGNOP term

# perly.y:1417 termbinop: term PLUGIN_ASSIGN_OP term

# perly.y:1419 termbinop: term POWOP term

# perly.y:1421 termbinop: term PLUGIN_POW_OP term

# perly.y:1423 termbinop: term MULOP term

# perly.y:1428 termbinop: term PLUGIN_MUL_OP term

# perly.y:1430 termbinop: term ADDOP term

# perly.y:1432 termbinop: term PLUGIN_ADD_OP term

# perly.y:1434 termbinop: term SHIFTOP term

# perly.y:1436 termbinop: termrelop

# perly.y:1438 termbinop: termeqop

# perly.y:1440 termbinop: term BITANDOP term

# perly.y:1442 termbinop: term BITOROP term

# perly.y:1444 termbinop: term DOTDOT term

# perly.y:1446 termbinop: term ANDAND term

# perly.y:1448 termbinop: term PLUGIN_LOGICAL_AND_OP term

# perly.y:1450 termbinop: term OROR term

# perly.y:1452 termbinop: term PLUGIN_LOGICAL_OR_OP term

# perly.y:1454 termbinop: term DORDOR term

# perly.y:1456 termbinop: term MATCHOP term

# perly.y:1458 termbinop: term PLUGIN_LOW_OP term

# perly.y:1462 termrelop: relopchain

# perly.y:1464 termrelop: term NCRELOP term

# perly.y:1466 termrelop: termrelop NCRELOP

# perly.y:1468 termrelop: termrelop CHRELOP

# perly.y:1470 termrelop: term PLUGIN_REL_OP term

# perly.y:1474 relopchain: term CHRELOP term

# perly.y:1476 relopchain: relopchain CHRELOP term

# perly.y:1480 termeqop: eqopchain

# perly.y:1482 termeqop: term NCEQOP term

# perly.y:1484 termeqop: termeqop NCEQOP

# perly.y:1486 termeqop: termeqop CHEQOP

# perly.y:1490 eqopchain: term CHEQOP term

# perly.y:1492 eqopchain: eqopchain CHEQOP term

# perly.y:1497 termunop: PERLY_MINUS term

# perly.y:1499 termunop: PERLY_PLUS term

# perly.y:1502 termunop: PERLY_EXCLAMATION_MARK term

# perly.y:1504 termunop: PERLY_TILDE term

# perly.y:1506 termunop: term POSTINC

# perly.y:1509 termunop: term POSTDEC

# perly.y:1512 termunop: term POSTJOIN

# perly.y:1523 termunop: PREINC term

# perly.y:1526 termunop: PREDEC term

# perly.y:1534 anonymous: PERLY_BRACKET_OPEN optexpr PERLY_BRACKET_CLOSE

# perly.y:1536 anonymous: HASHBRACK optexpr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1538 anonymous: KW_SUB_anon startanonsub proto subattrlist subbody

# perly.y:1541 anonymous: KW_SUB_anon startanonsub proto subattrlist

# perly.y:1543 anonymous: KW_SUB_anon_sig startanonsub subattrlist sigsubbody

# perly.y:1546 anonymous: KW_SUB_anon_sig startanonsub subattrlist

# perly.y:1548 anonymous: KW_METHOD_anon startanonmethod subattrlist sigsubbody

# perly.y:1553 anonymous: KW_METHOD_anon startanonmethod subattrlist

# perly.y:1558 termdo: KW_DO term

# perly.y:1560 termdo: KW_DO block

# perly.y:1564 term: termbinop

# perly.y:1565 term: termunop

# perly.y:1566 term: anonymous

# perly.y:1567 term: termdo

# perly.y:1568 term: term PERLY_QUESTION_MARK term PERLY_COLON term

# perly.y:1570 term: REFGEN term

# perly.y:1572 term: myattrterm

# perly.y:1574 term: KW_LOCAL term

# perly.y:1576 term: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE

# perly.y:1578 term: QWLIST

# perly.y:1580 term: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1582 term: scalar

# perly.y:1584 term: star

# perly.y:1586 term: hsh

# perly.y:1588 term: ary

# perly.y:1590 term: arylen

# perly.y:1592 term: subscripted

# perly.y:1594 term: sliceme PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1604 term: kvslice PERLY_BRACKET_OPEN expr PERLY_BRACKET_CLOSE

# perly.y:1614 term: sliceme PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1624 term: kvslice PERLY_BRACE_OPEN expr PERLY_SEMICOLON PERLY_BRACE_CLOSE

# perly.y:1634 term: THING

# perly.y:1636 term: amper

# perly.y:1638 term: amper PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1641 term: amper PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE

# perly.y:1646 term: NOAMP subname optlistexpr

# perly.y:1650 term: term ARROW PERLY_DOLLAR PERLY_STAR

# perly.y:1652 term: term ARROW PERLY_SNAIL PERLY_STAR

# perly.y:1654 term: term ARROW PERLY_PERCENT_SIGN PERLY_STAR

# perly.y:1656 term: term ARROW PERLY_AMPERSAND PERLY_STAR

# perly.y:1659 term: term ARROW PERLY_STAR PERLY_STAR

# perly.y:1661 term: LOOPEX

# perly.y:1664 term: LOOPEX term

# perly.y:1666 term: NOTOP listexpr

# perly.y:1668 term: UNIOP

# perly.y:1670 term: UNIOP block

# perly.y:1672 term: UNIOP term

# perly.y:1674 term: KW_REQUIRE

# perly.y:1676 term: KW_REQUIRE term

# perly.y:1678 term: UNIOPSUB

# perly.y:1680 term: UNIOPSUB term

# perly.y:1683 term: FUNC0

# perly.y:1685 term: FUNC0 PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1687 term: FUNC0OP

# perly.y:1689 term: FUNC0OP PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1691 term: FUNC0SUB

# perly.y:1693 term: FUNC1 PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1697 term: FUNC1 PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE

# perly.y:1699 term: PMFUNC @23 SUBLEXSTART listexpr optrepl SUBLEXEND

# perly.y:1712 term: BAREWORD

# perly.y:1713 term: listop

# perly.y:1714 term: PLUGEXPR

# perly.y:1719 myattrterm: KW_MY myterm attrlist

# perly.y:1721 myattrterm: KW_MY myterm

# perly.y:1723 myattrterm: KW_MY REFGEN myterm attrlist

# perly.y:1725 myattrterm: KW_MY REFGEN term

# perly.y:1730 myterm: PERLY_PAREN_OPEN expr PERLY_PAREN_CLOSE

# perly.y:1732 myterm: PERLY_PAREN_OPEN PERLY_PAREN_CLOSE

# perly.y:1735 myterm: scalar

# perly.y:1737 myterm: hsh

# perly.y:1739 myterm: ary

# perly.y:1744 fieldvar: scalar

# perly.y:1749 fieldvar: hsh

# perly.y:1754 fieldvar: ary

# perly.y:1762 fielddecl: KW_FIELD fieldvar optattrlist

# perly.y:1769 fielddecl: KW_FIELD fieldvar optattrlist ASSIGNOP $@24 term

# perly.y:1787 optlistexpr: empty

# perly.y:1788 optlistexpr: listexpr

# perly.y:1792 optexpr: empty

# perly.y:1793 optexpr: expr

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

# perly.y:1842 amper: PERLY_AMPERSAND indirob

# perly.y:1846 scalar: PERLY_DOLLAR indirob

# perly.y:1850 ary: PERLY_SNAIL indirob

# perly.y:1856 hsh: PERLY_PERCENT_SIGN indirob

# perly.y:1862 arylen: DOLSHARP indirob

# perly.y:1864 arylen: term ARROW DOLSHARP PERLY_STAR

# perly.y:1868 star: PERLY_STAR indirob

# perly.y:1872 sliceme: ary

# perly.y:1873 sliceme: term ARROW PERLY_SNAIL

# perly.y:1877 kvslice: hsh

# perly.y:1878 kvslice: term ARROW PERLY_PERCENT_SIGN

# perly.y:1882 gelem: star

# perly.y:1883 gelem: term ARROW PERLY_STAR

# perly.y:1888 indirob: BAREWORD

# perly.y:1890 indirob: scalar

# perly.y:1892 indirob: block

# perly.y:1895 indirob: PRIVATEREF
`
