(ns funcgo.core
  (:require [instaparse.core :as insta]))

(def funcgo-parser
     (insta/parser "

SourceFile     = [ NL ] PackageClause _  { _ Expression _ NL }
PackageClause  = <'package'> <__> dotted NL  _ ImportDecl _ NL
ImportDecl     = <'import'> _ <'('>  NL { _ ImportSpec _ NL } <')'>
ImportSpec     = identifier _ dotted
<Expression>   = UnaryExpr | ShortVarDecl                       (* | Expression binary_op UnaryExpr *)
<UnaryExpr>    = PrimaryExpr                                                (* | unary_op UnaryExpr *)
PrimaryExpr    = Operand        |
                                                                         (*Conversion |
                                                                         BuiltinCall |
                                                                         PrimaryExpr Selector |
                                                                         PrimaryExpr Index |
                                                                         PrimaryExpr Slice |
                                                                         PrimaryExpr TypeAssertion |*)
	         PrimaryExpr Call
<Call>         = <'('> [ ArgumentList ] <')'>
<ArgumentList> = ExpressionList                                                      (* [ _ '...' ] *)
ExpressionList = Expression { _ <','> _ Expression }
<Operand>      = Literal | OperandName                          (*| MethodExpr | '(' Expression ')' *)
<OperandName>  = identifier                                                       (*| QualifiedIdent*)
<Literal>      = BasicLit                                            (*| CompositeLit | FunctionLit *)
<BasicLit>     = int_lit                      (*| float_lit | imaginary_lit | rune_lit | string_lit *)
ShortVarDecl   = identifier _ <':='> _ Expression
<int_lit>      = decimal_lit    (*| octal_lit | hex_lit .*)
decimal_lit = #'[1-9][0-9]*'

dotted         = identifier { <'.'> identifier }
<identifier>   = #'[\\p{L}_][\\p{L}_\\p{Digit}]*'              (* letter { letter | unicode_digit } *)
letter         = unicode_letter | '_'
unicode_letter = #'\\p{L}'
unicode_digit  = #'\\p{Digit}'
<_>            = <#'[ \\t\\x0B\\f\\r]*'>   (* optional non-newline whitespace *)
__             = #'[ \\t\\x0B\\f\\r]+'     (* non-newline whitespace *)
<NL>           = [ nl | comment ]
<nl>           = <#'\\s*[\\n;]\\s*'>       (* whitespace with at least one newline or semicolon *)
comment        = #'//[^\\n]*\\n'
"))



(defn funcgo-parse [fgo]
  (insta/transform
   {
    :SourceFile (fn [header body] (str header body "\n"))
    :PackageClause (fn [dotted  import-decl]
                     (str "(ns " dotted import-decl ")\n\n"))
    :ImportDecl (fn [ & import-specs] (apply str import-specs))
    :ImportSpec (fn [identifier dotted]
                  (str "\n  (:require [" dotted " :as " identifier "])"))
    :ShortVarDecl (fn [identifier expression]
                    (str "(def " identifier " " expression ")"))
    :PrimaryExpr (fn
                   ([operand] operand)
                   ([primary-expr call] (str "(" primary-expr " " call ")")))
    :ExpressionList (fn [expr0 & expr-rest]
                      (reduce
                       (fn [acc expr] (str acc ", " expr))
                       expr0
                       expr-rest))
    :dotted (fn [idf0 & idf-rest]
              (reduce
               (fn [acc idf] (str acc "." idf))
               idf0
               idf-rest))
    :decimal_lit (fn [s] s)}
   (funcgo-parser fgo)))
