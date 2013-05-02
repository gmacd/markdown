
/*  Original C version https://github.com/jgm/peg-markdown/
 *	Copyright 2008 John MacFarlane (jgm at berkeley dot edu).
 *
 *  Modifications and translation from C into Go
 *  based on markdown_parser.leg and utility_functions.c
 *	Copyright 2010 Michael Teichgräber (mt at wmipf dot de)
 *
 *  This program is free software; you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License or the MIT
 *  license.  See LICENSE for details.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 */

package markdown

// PEG grammar and parser actions for markdown syntax.

import (
	"fmt"
	"io"
	"log"
	"strings"
)

const (
	parserIfaceVersion_16 = iota
)

// Semantic value of a parsing action.
type element struct {
	key int
	contents
	children *element
	next     *element
}

// Information (label, URL and title) for a link.
type link struct {
	label *element
	url   string
	title string
}

// Union for contents of an Element (string, list, or link).
type contents struct {
	str string
	*link
}

// Types of semantic values returned by parsers.
const (
	LIST = iota /* A generic list of values. For ordered and bullet lists, see below. */
	RAW         /* Raw markdown to be processed further */
	SPACE
	LINEBREAK
	ELLIPSIS
	EMDASH
	ENDASH
	APOSTROPHE
	SINGLEQUOTED
	DOUBLEQUOTED
	STR
	LINK
	IMAGE
	CODE
	HTML
	EMPH
	STRONG
	PLAIN
	PARA
	LISTITEM
	BULLETLIST
	ORDEREDLIST
	H1 /* Code assumes that H1..6 are in order. */
	H2
	H3
	H4
	H5
	H6
	BLOCKQUOTE
	VERBATIM
	HTMLBLOCK
	HRULE
	REFERENCE
	NOTE
  TABLE
  TABLEHEAD
  TABLEBODY
  TABLEROW
  TABLECELL
  CELLSPAN
  TABLECAPTION
  TABLELABEL
  TABLESEPARATOR
	DEFINITIONLIST
	DEFTITLE
	DEFDATA
	numVAL
)

type state struct {
	extension  Extensions
	heap       elemHeap
	tree       *element /* Results of parse. */
	references *element /* List of link references found. */
	notes      *element /* List of footnotes found. */
}


const (
	ruleDoc = iota
	ruleDocblock
	ruleBlock
	rulePara
	rulePlain
	ruleAtxInline
	ruleAtxStart
	ruleAtxHeading
	ruleSetextHeading
	ruleSetextBottom1
	ruleSetextBottom2
	ruleSetextHeading1
	ruleSetextHeading2
	ruleHeading
	ruleBlockQuote
	ruleBlockQuoteRaw
	ruleNonblankIndentedLine
	ruleVerbatimChunk
	ruleVerbatim
	ruleHorizontalRule
	ruleBullet
	ruleBulletList
	ruleListTight
	ruleListLoose
	ruleListItem
	ruleListItemTight
	ruleListBlock
	ruleListContinuationBlock
	ruleEnumerator
	ruleOrderedList
	ruleListBlockLine
	ruleHtmlBlockOpenAddress
	ruleHtmlBlockCloseAddress
	ruleHtmlBlockAddress
	ruleHtmlBlockOpenBlockquote
	ruleHtmlBlockCloseBlockquote
	ruleHtmlBlockBlockquote
	ruleHtmlBlockOpenCenter
	ruleHtmlBlockCloseCenter
	ruleHtmlBlockCenter
	ruleHtmlBlockOpenDir
	ruleHtmlBlockCloseDir
	ruleHtmlBlockDir
	ruleHtmlBlockOpenDiv
	ruleHtmlBlockCloseDiv
	ruleHtmlBlockDiv
	ruleHtmlBlockOpenDl
	ruleHtmlBlockCloseDl
	ruleHtmlBlockDl
	ruleHtmlBlockOpenFieldset
	ruleHtmlBlockCloseFieldset
	ruleHtmlBlockFieldset
	ruleHtmlBlockOpenForm
	ruleHtmlBlockCloseForm
	ruleHtmlBlockForm
	ruleHtmlBlockOpenH1
	ruleHtmlBlockCloseH1
	ruleHtmlBlockH1
	ruleHtmlBlockOpenH2
	ruleHtmlBlockCloseH2
	ruleHtmlBlockH2
	ruleHtmlBlockOpenH3
	ruleHtmlBlockCloseH3
	ruleHtmlBlockH3
	ruleHtmlBlockOpenH4
	ruleHtmlBlockCloseH4
	ruleHtmlBlockH4
	ruleHtmlBlockOpenH5
	ruleHtmlBlockCloseH5
	ruleHtmlBlockH5
	ruleHtmlBlockOpenH6
	ruleHtmlBlockCloseH6
	ruleHtmlBlockH6
	ruleHtmlBlockOpenMenu
	ruleHtmlBlockCloseMenu
	ruleHtmlBlockMenu
	ruleHtmlBlockOpenNoframes
	ruleHtmlBlockCloseNoframes
	ruleHtmlBlockNoframes
	ruleHtmlBlockOpenNoscript
	ruleHtmlBlockCloseNoscript
	ruleHtmlBlockNoscript
	ruleHtmlBlockOpenOl
	ruleHtmlBlockCloseOl
	ruleHtmlBlockOl
	ruleHtmlBlockOpenP
	ruleHtmlBlockCloseP
	ruleHtmlBlockP
	ruleHtmlBlockOpenPre
	ruleHtmlBlockClosePre
	ruleHtmlBlockPre
	ruleHtmlBlockOpenTable
	ruleHtmlBlockCloseTable
	ruleHtmlBlockTable
	ruleHtmlBlockOpenUl
	ruleHtmlBlockCloseUl
	ruleHtmlBlockUl
	ruleHtmlBlockOpenDd
	ruleHtmlBlockCloseDd
	ruleHtmlBlockDd
	ruleHtmlBlockOpenDt
	ruleHtmlBlockCloseDt
	ruleHtmlBlockDt
	ruleHtmlBlockOpenFrameset
	ruleHtmlBlockCloseFrameset
	ruleHtmlBlockFrameset
	ruleHtmlBlockOpenLi
	ruleHtmlBlockCloseLi
	ruleHtmlBlockLi
	ruleHtmlBlockOpenTbody
	ruleHtmlBlockCloseTbody
	ruleHtmlBlockTbody
	ruleHtmlBlockOpenTd
	ruleHtmlBlockCloseTd
	ruleHtmlBlockTd
	ruleHtmlBlockOpenTfoot
	ruleHtmlBlockCloseTfoot
	ruleHtmlBlockTfoot
	ruleHtmlBlockOpenTh
	ruleHtmlBlockCloseTh
	ruleHtmlBlockTh
	ruleHtmlBlockOpenThead
	ruleHtmlBlockCloseThead
	ruleHtmlBlockThead
	ruleHtmlBlockOpenTr
	ruleHtmlBlockCloseTr
	ruleHtmlBlockTr
	ruleHtmlBlockOpenScript
	ruleHtmlBlockCloseScript
	ruleHtmlBlockScript
	ruleHtmlBlockOpenHead
	ruleHtmlBlockCloseHead
	ruleHtmlBlockHead
	ruleHtmlBlockInTags
	ruleHtmlBlock
	ruleHtmlBlockSelfClosing
	ruleHtmlBlockType
	ruleStyleOpen
	ruleStyleClose
	ruleInStyleTags
	ruleStyleBlock
	ruleInlines
	ruleInline
	ruleSpace
	ruleStr
	ruleStrChunk
	ruleAposChunk
	ruleEscapedChar
	ruleEntity
	ruleEndline
	ruleNormalEndline
	ruleTerminalEndline
	ruleLineBreak
	ruleSymbol
	ruleUlOrStarLine
	ruleStarLine
	ruleUlLine
	ruleEmph
	ruleWhitespace
	ruleEmphStar
	ruleEmphUl
	ruleStrong
	ruleStrongStar
	ruleStrongUl
	ruleImage
	ruleLink
	ruleReferenceLink
	ruleReferenceLinkDouble
	ruleReferenceLinkSingle
	ruleExplicitLink
	ruleSource
	ruleSourceContents
	ruleTitle
	ruleTitleSingle
	ruleTitleDouble
	ruleAutoLink
	ruleAutoLinkUrl
	ruleAutoLinkEmail
	ruleReference
	ruleLabel
	ruleRefSrc
	ruleRefTitle
	ruleEmptyTitle
	ruleRefTitleSingle
	ruleRefTitleDouble
	ruleRefTitleParens
	ruleReferences
	ruleTicks1
	ruleTicks2
	ruleTicks3
	ruleTicks4
	ruleTicks5
	ruleCode
	ruleRawHtml
	ruleBlankLine
	ruleQuoted
	ruleHtmlAttribute
	ruleHtmlComment
	ruleHtmlTag
	ruleEof
	ruleSpacechar
	ruleNonspacechar
	ruleNewline
	ruleSp
	ruleSpnl
	ruleSpecialChar
	ruleNormalChar
	ruleAlphanumeric
	ruleAlphanumericAscii
	ruleDigit
	ruleHexEntity
	ruleDecEntity
	ruleCharEntity
	ruleNonindentSpace
	ruleIndent
	ruleIndentedLine
	ruleOptionallyIndentedLine
	ruleStartList
	ruleLine
	ruleRawLine
	ruleSkipBlock
	ruleExtendedSpecialChar
	ruleSmart
	ruleApostrophe
	ruleEllipsis
	ruleDash
	ruleEnDash
	ruleEmDash
	ruleSingleQuoteStart
	ruleSingleQuoteEnd
	ruleSingleQuoted
	ruleDoubleQuoteStart
	ruleDoubleQuoteEnd
	ruleDoubleQuoted
	ruleNoteReference
	ruleRawNoteReference
	ruleNote
	ruleInlineNote
	ruleNotes
	ruleRawNoteBlock
	ruleDefinitionList
	ruleDefinition
	ruleDListTitle
	ruleDefTight
	ruleDefLoose
	ruleDefmark
	ruleDefMarker
	ruleTable
	ruleTableBody
	ruleTableRow
	ruleTableLine
	ruleTableCell
	ruleExtendedCell
	ruleCellStr
	ruleFullCell
	ruleEmptyCell
	ruleSeparatorLine
	ruleAlignmentCell
	ruleLeftAlignWrap
	ruleLeftAlign
	ruleCenterAlignWrap
	ruleCenterAlign
	ruleRightAlignWrap
	ruleRightAlign
	ruleCellDivider
	ruleTableCaption
)

type yyParser struct {
	state
	Buffer string
	Min, Max int
	rules [266]func() bool
	ResetBuffer	func(string) string
}

func (p *yyParser) Parse(ruleId int) (err error) {
	if p.rules[ruleId]() {
		return
	}
	return p.parseErr()
}

type errPos struct {
	Line, Pos int
}

func	(e *errPos) String() string {
	return fmt.Sprintf("%d:%d", e.Line, e.Pos)
}

type unexpectedCharError struct {
	After, At	errPos
	Char	byte
}

func (e *unexpectedCharError) Error() string {
	return fmt.Sprintf("%v: unexpected character '%c'", &e.At, e.Char)
}

type unexpectedEOFError struct {
	After errPos
}

func (e *unexpectedEOFError) Error() string {
	return fmt.Sprintf("%v: unexpected end of file", &e.After)
}

func (p *yyParser) parseErr() (err error) {
	var pos, after errPos
	pos.Line = 1
	for i, c := range p.Buffer[0:] {
		if c == '\n' {
			pos.Line++
			pos.Pos = 0
		} else {
			pos.Pos++
		}
		if i == p.Min {
			if p.Min != p.Max {
				after = pos
			} else {
				break
			}
		} else if i == p.Max {
			break
		}
	}
	if p.Max >= len(p.Buffer) {
		err = &unexpectedEOFError{after}
	} else {
		err = &unexpectedCharError{after, pos, p.Buffer[p.Max]}
	}
	return
}

func (p *yyParser) Init() {
	var position int
	var yyp int
	var yy *element
	var yyval = make([]*element, 256)

	actions := [...]func(string, int){
		/* 0 Doc */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 1 Doc */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 p.tree = reverse(a) 
			yyval[yyp-1] = a
		},
		/* 2 Docblock */
		func(yytext string, _ int) {
			 p.tree = yy 
		},
		/* 3 Para */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a; yy.key = PARA 
			yyval[yyp-1] = a
		},
		/* 4 Plain */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = a; yy.key = PLAIN 
			yyval[yyp-1] = a
		},
		/* 5 AtxStart */
		func(yytext string, _ int) {
			 yy = p.mkElem(H1 + (len(yytext) - 1)) 
		},
		/* 6 AtxHeading */
		func(yytext string, _ int) {
			s := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = s
		},
		/* 7 AtxHeading */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			s := yyval[yyp-1]
			 yy = p.mkList(s.key, a)
              s = nil 
			yyval[yyp-2] = a
			yyval[yyp-1] = s
		},
		/* 8 SetextHeading1 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 9 SetextHeading1 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(H1, a) 
			yyval[yyp-1] = a
		},
		/* 10 SetextHeading2 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 11 SetextHeading2 */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(H2, a) 
			yyval[yyp-1] = a
		},
		/* 12 BlockQuote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			  yy = p.mkElem(BLOCKQUOTE)
                yy.children = a
             
			yyval[yyp-1] = a
		},
		/* 13 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 14 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 15 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(p.mkString("\n"), a) 
			yyval[yyp-1] = a
		},
		/* 16 BlockQuoteRaw */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = p.mkStringFromList(a, true)
                     yy.key = RAW
                 
			yyval[yyp-1] = a
		},
		/* 17 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(p.mkString("\n"), a) 
			yyval[yyp-1] = a
		},
		/* 18 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 19 VerbatimChunk */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkStringFromList(a, false) 
			yyval[yyp-1] = a
		},
		/* 20 Verbatim */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 21 Verbatim */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkStringFromList(a, false)
                 yy.key = VERBATIM 
			yyval[yyp-1] = a
		},
		/* 22 HorizontalRule */
		func(yytext string, _ int) {
			 yy = p.mkElem(HRULE) 
		},
		/* 23 BulletList */
		func(yytext string, _ int) {
			 yy.key = BULLETLIST 
		},
		/* 24 ListTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 25 ListTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 26 ListLoose */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			
                  li := b.children
                  li.contents.str += "\n\n"
                  a = cons(b, a)
              
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 27 ListLoose */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 28 ListItem */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 29 ListItem */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 30 ListItem */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
               raw := p.mkStringFromList(a, false)
               raw.key = RAW
               yy = p.mkElem(LISTITEM)
               yy.children = raw
            
			yyval[yyp-1] = a
		},
		/* 31 ListItemTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 32 ListItemTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 33 ListItemTight */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
               raw := p.mkStringFromList(a, false)
               raw.key = RAW
               yy = p.mkElem(LISTITEM)
               yy.children = raw
            
			yyval[yyp-1] = a
		},
		/* 34 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 35 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 36 ListBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkStringFromList(a, false) 
			yyval[yyp-1] = a
		},
		/* 37 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   if len(yytext) == 0 {
                                   a = cons(p.mkString("\001"), a) // block separator
                              } else {
                                   a = cons(p.mkString(yytext), a)
                              }
                          
			yyval[yyp-1] = a
		},
		/* 38 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 39 ListContinuationBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			  yy = p.mkStringFromList(a, false) 
			yyval[yyp-1] = a
		},
		/* 40 OrderedList */
		func(yytext string, _ int) {
			 yy.key = ORDEREDLIST 
		},
		/* 41 HtmlBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = p.mkList(LIST, nil)
                } else {
                    yy = p.mkString(yytext)
                    yy.key = HTMLBLOCK
                }
            
		},
		/* 42 StyleBlock */
		func(yytext string, _ int) {
			   if p.extension.FilterStyles {
                        yy = p.mkList(LIST, nil)
                    } else {
                        yy = p.mkString(yytext)
                        yy.key = HTMLBLOCK
                    }
                
		},
		/* 43 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-2] = c
			yyval[yyp-1] = a
		},
		/* 44 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			 a = cons(c, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 45 Inlines */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			c := yyval[yyp-2]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = c
		},
		/* 46 Space */
		func(yytext string, _ int) {
			 yy = p.mkString(" ")
          yy.key = SPACE 
		},
		/* 47 Str */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(p.mkString(yytext), a) 
			yyval[yyp-1] = a
		},
		/* 48 Str */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 49 Str */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 if a.next == nil { yy = a; } else { yy = p.mkList(LIST, a) } 
			yyval[yyp-1] = a
		},
		/* 50 StrChunk */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 51 AposChunk */
		func(yytext string, _ int) {
			 yy = p.mkElem(APOSTROPHE) 
		},
		/* 52 EscapedChar */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 53 Entity */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext); yy.key = HTML 
		},
		/* 54 NormalEndline */
		func(yytext string, _ int) {
			 yy = p.mkString("\n")
                    yy.key = SPACE 
		},
		/* 55 TerminalEndline */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 56 LineBreak */
		func(yytext string, _ int) {
			 yy = p.mkElem(LINEBREAK) 
		},
		/* 57 Symbol */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 58 UlOrStarLine */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 59 EmphStar */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 60 EmphStar */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 61 EmphStar */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = p.mkList(EMPH, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 62 EmphUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 63 EmphUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 64 EmphUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = p.mkList(EMPH, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 65 StrongStar */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 66 StrongStar */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = p.mkList(STRONG, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 67 StrongUl */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 68 StrongUl */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 yy = p.mkList(STRONG, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 69 Image */
		func(yytext string, _ int) {
				if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(p.mkString("!"), result.children)
		}
	
		},
		/* 70 ReferenceLinkDouble */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			
                           if match, found := p.findReference(b.children); found {
                               yy = p.mkLink(a.children, match.url, match.title);
                               a = nil
                               b = nil
                           } else {
                               result := p.mkElem(LIST)
                               result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), cons(p.mkString(yytext),
                                                   cons(p.mkString("["), cons(b, p.mkString("]")))))))
                               yy = result
                           }
                       
			yyval[yyp-2] = b
			yyval[yyp-1] = a
		},
		/* 71 ReferenceLinkSingle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
                           if match, found := p.findReference(a.children); found {
                               yy = p.mkLink(a.children, match.url, match.title)
                               a = nil
                           } else {
                               result := p.mkElem(LIST)
                               result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), p.mkString(yytext))));
                               yy = result
                           }
                       
			yyval[yyp-1] = a
		},
		/* 72 ExplicitLink */
		func(yytext string, _ int) {
			l := yyval[yyp-1]
			t := yyval[yyp-2]
			s := yyval[yyp-3]
			 yy = p.mkLink(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil 
			yyval[yyp-3] = s
			yyval[yyp-1] = l
			yyval[yyp-2] = t
		},
		/* 73 Source */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 74 Title */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 75 AutoLinkUrl */
		func(yytext string, _ int) {
			   yy = p.mkLink(p.mkString(yytext), yytext, "") 
		},
		/* 76 AutoLinkEmail */
		func(yytext string, _ int) {
			
                    yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")
                
		},
		/* 77 Reference */
		func(yytext string, _ int) {
			t := yyval[yyp-1]
			l := yyval[yyp-2]
			s := yyval[yyp-3]
			 yy = p.mkLink(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE 
			yyval[yyp-3] = s
			yyval[yyp-1] = t
			yyval[yyp-2] = l
		},
		/* 78 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 79 Label */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 80 RefSrc */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext)
           yy.key = HTML 
		},
		/* 81 RefTitle */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 82 References */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 83 References */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 p.references = reverse(a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 84 Code */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext); yy.key = CODE 
		},
		/* 85 RawHtml */
		func(yytext string, _ int) {
			   if p.extension.FilterHTML {
                    yy = p.mkList(LIST, nil)
                } else {
                    yy = p.mkString(yytext)
                    yy.key = HTML
                }
            
		},
		/* 86 StartList */
		func(yytext string, _ int) {
			 yy = nil 
		},
		/* 87 Line */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 88 Apostrophe */
		func(yytext string, _ int) {
			 yy = p.mkElem(APOSTROPHE) 
		},
		/* 89 Ellipsis */
		func(yytext string, _ int) {
			 yy = p.mkElem(ELLIPSIS) 
		},
		/* 90 EnDash */
		func(yytext string, _ int) {
			 yy = p.mkElem(ENDASH) 
		},
		/* 91 EmDash */
		func(yytext string, _ int) {
			 yy = p.mkElem(EMDASH) 
		},
		/* 92 SingleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 93 SingleQuoted */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 yy = p.mkList(SINGLEQUOTED, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 94 DoubleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 95 DoubleQuoted */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy = p.mkList(DOUBLEQUOTED, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 96 NoteReference */
		func(yytext string, _ int) {
			ref := yyval[yyp-1]
			
                    if match, ok := p.find_note(ref.contents.str); ok {
                        yy = p.mkElem(NOTE)
                        yy.children = match.children
                        yy.contents.str = ""
                    } else {
                        yy = p.mkString("[^"+ref.contents.str+"]")
                    }
                
			yyval[yyp-1] = ref
		},
		/* 97 RawNoteReference */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 98 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			ref := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = ref
		},
		/* 99 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			ref := yyval[yyp-2]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = ref
		},
		/* 100 Note */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			ref := yyval[yyp-2]
			   yy = p.mkList(NOTE, a)
                    yy.contents.str = ref.contents.str
                
			yyval[yyp-1] = a
			yyval[yyp-2] = ref
		},
		/* 101 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 102 InlineNote */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(NOTE, a)
                  yy.contents.str = "" 
			yyval[yyp-1] = a
		},
		/* 103 Notes */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			b := yyval[yyp-2]
			 a = cons(b, a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 104 Notes */
		func(yytext string, _ int) {
			b := yyval[yyp-2]
			a := yyval[yyp-1]
			 p.notes = reverse(a) 
			yyval[yyp-1] = a
			yyval[yyp-2] = b
		},
		/* 105 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 106 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(p.mkString(yytext), a) 
			yyval[yyp-1] = a
		},
		/* 107 RawNoteBlock */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			   yy = p.mkStringFromList(a, true)
                    yy.key = RAW
                
			yyval[yyp-1] = a
		},
		/* 108 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 109 DefinitionList */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(DEFINITIONLIST, a) 
			yyval[yyp-1] = a
		},
		/* 110 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 111 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			
			yyval[yyp-1] = a
		},
		/* 112 Definition */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(LIST, a) 
			yyval[yyp-1] = a
		},
		/* 113 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 114 DListTitle */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
				yy = p.mkList(LIST, a)
				yy.key = DEFTITLE
			
			yyval[yyp-1] = a
		},
		/* 115 Table */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 b = cons(yy, b) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 116 Table */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 yy.key = TABLEHEAD; a = cons(yy, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 117 Table */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			 append_list(yy, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 118 Table */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 119 Table */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 120 Table */
		func(yytext string, _ int) {
			a := yyval[yyp-2]
			b := yyval[yyp-1]
			 b = cons(yy, b) 
			yyval[yyp-2] = a
			yyval[yyp-1] = b
		},
		/* 121 Table */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			
        if b != nil { append_list(b,a) }
        yy = p.mkList(TABLE, a)
    
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},
		/* 122 TableBody */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 123 TableBody */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(TABLEBODY, a) 
			yyval[yyp-1] = a
		},
		/* 124 TableRow */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 125 TableRow */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(TABLEROW, a) 
			yyval[yyp-1] = a
		},
		/* 126 ExtendedCell */
		func(yytext string, _ int) {
			
        span := p.mkString(yytext)
        span.key = CELLSPAN
        span.next = yy.children
        yy.children = span
    
		},
		/* 127 CellStr */
		func(yytext string, _ int) {
			 yy = p.mkString(yytext) 
		},
		/* 128 FullCell */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 129 FullCell */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 yy = p.mkList(TABLECELL, a) 
			yyval[yyp-1] = a
		},
		/* 130 EmptyCell */
		func(yytext string, _ int) {
			 yy = p.mkElem(TABLECELL) 
		},
		/* 131 SeparatorLine */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			 a = cons(yy, a) 
			yyval[yyp-1] = a
		},
		/* 132 SeparatorLine */
		func(yytext string, _ int) {
			a := yyval[yyp-1]
			
        yy = p.mkStringFromList(a, false);
        yy.key = TABLESEPARATOR;
    
			yyval[yyp-1] = a
		},
		/* 133 LeftAlignWrap */
		func(yytext string, _ int) {
			 yy = p.mkString("L");
		},
		/* 134 LeftAlign */
		func(yytext string, _ int) {
			 yy = p.mkString("l");
		},
		/* 135 CenterAlignWrap */
		func(yytext string, _ int) {
			 yy = p.mkString("C");
		},
		/* 136 CenterAlign */
		func(yytext string, _ int) {
			 yy = p.mkString("c");
		},
		/* 137 RightAlignWrap */
		func(yytext string, _ int) {
			 yy = p.mkString("R");
		},
		/* 138 RightAlign */
		func(yytext string, _ int) {
			 yy = p.mkString("r");
		},
		/* 139 TableCaption */
		func(yytext string, _ int) {
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			c := yyval[yyp-3]
			 b = c; b.key = TABLELABEL;
			yyval[yyp-1] = b
			yyval[yyp-2] = a
			yyval[yyp-3] = c
		},
		/* 140 TableCaption */
		func(yytext string, _ int) {
			c := yyval[yyp-3]
			b := yyval[yyp-1]
			a := yyval[yyp-2]
			
    yy = a
    yy.key = TABLECAPTION
    if b != nil && b.key == TABLELABEL {
        b.next = yy.children
        yy.children = b
    }

			yyval[yyp-3] = c
			yyval[yyp-1] = b
			yyval[yyp-2] = a
		},

		/* yyPush */
		func(_ string, count int) {
			yyp += count
			if yyp >= len(yyval) {
				s := make([]*element, cap(yyval)+256)
				copy(s, yyval)
				yyval = s
			}
		},
		/* yyPop */
		func(_ string, count int) {
			yyp -= count
		},
		/* yySet */
		func(_ string, count int) {
			yyval[yyp+count] = yy
		},
	}
	const (
		yyPush = 141 + iota
		yyPop
		yySet
	)

	type thunk struct {
		action uint16
		begin, end int
	}
	var thunkPosition, begin, end int
	thunks := make([]thunk, 32)
	doarg := func(action uint16, arg int) {
		if thunkPosition == len(thunks) {
			newThunks := make([]thunk, 2*len(thunks))
			copy(newThunks, thunks)
			thunks = newThunks
		}
		t := &thunks[thunkPosition]
		thunkPosition++
		t.action = action
		if arg != 0 {
			t.begin = arg // use begin to store an argument
		} else {
			t.begin = begin
		}
		t.end = end
	}
	do := func(action uint16) {
		doarg(action, 0)
	}

	p.ResetBuffer = func(s string) (old string) {
		if position < len(p.Buffer) {
			old = p.Buffer[position:]
		}
		p.Buffer = s
		thunkPosition = 0
		position = 0
		p.Min = 0
		p.Max = 0
		end = 0
		return
	}

	commit := func(thunkPosition0 int) bool {
		if thunkPosition0 == 0 {
			s := ""
			for _, t := range thunks[:thunkPosition] {
				b := t.begin
				if b >= 0 && b <= t.end {
					s = p.Buffer[b:t.end]
				}
				magic := b
				actions[t.action](s, magic)
			}
			p.Min = position
			thunkPosition = 0
			return true
		}
		return false
	}
	matchDot := func() bool {
		if position < len(p.Buffer) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	matchChar := func(c byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] == c) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	peekChar := func(c byte) bool {
		return position < len(p.Buffer) && p.Buffer[position] == c
	}

	matchString := func(s string) bool {
		length := len(s)
		next := position + length
		if (next <= len(p.Buffer)) && p.Buffer[position] == s[0] && (p.Buffer[position:next] == s) {
			position = next
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	classes := [...][32]uint8{
	3:	{0, 0, 0, 0, 50, 232, 255, 3, 254, 255, 255, 135, 254, 255, 255, 71, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	1:	{0, 0, 0, 0, 10, 111, 0, 80, 0, 0, 0, 184, 1, 0, 0, 56, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	0:	{0, 0, 0, 0, 0, 0, 255, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	4:	{0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	7:	{0, 0, 0, 0, 0, 0, 255, 3, 126, 0, 0, 0, 126, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	2:	{0, 0, 0, 0, 0, 0, 0, 0, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	5:	{0, 0, 0, 0, 0, 0, 255, 3, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	6:	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	matchClass := func(class uint) bool {
		if (position < len(p.Buffer)) &&
			((classes[class][p.Buffer[position]>>3] & (1 << (p.Buffer[position] & 7))) != 0) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	peekClass := func(class uint) bool {
		if (position < len(p.Buffer)) &&
			((classes[class][p.Buffer[position]>>3] & (1 << (p.Buffer[position] & 7))) != 0) {
			return true
		}
		return false
	}


	p.rules = [...]func() bool{

		/* 0 Doc <- (StartList (Block { a = cons(yy, a) })* { p.tree = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l0
			}
			doarg(yySet, -1)
		l1:
			{
				position2 := position
				if !p.rules[ruleBlock]() {
					goto l2
				}
				do(0)
				goto l1
			l2:
				position = position2
			}
			do(1)
			if !(commit(thunkPosition0)) {
				goto l0
			}
			doarg(yyPop, 1)
			return true
		l0:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 1 Docblock <- (Block { p.tree = yy } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlock]() {
				goto l3
			}
			do(2)
			if !(commit(thunkPosition0)) {
				goto l3
			}
			return true
		l3:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 2 Block <- (BlankLine* (BlockQuote / Verbatim / Note / Reference / HorizontalRule / Heading / DefinitionList / OrderedList / BulletList / HtmlBlock / StyleBlock / (&{p.extension.Table} Table) / Para / Plain)) */
		func() bool {
			position0 := position
		l5:
			if !p.rules[ruleBlankLine]() {
				goto l6
			}
			goto l5
		l6:
			if !p.rules[ruleBlockQuote]() {
				goto l8
			}
			goto l7
		l8:
			if !p.rules[ruleVerbatim]() {
				goto l9
			}
			goto l7
		l9:
			if !p.rules[ruleNote]() {
				goto l10
			}
			goto l7
		l10:
			if !p.rules[ruleReference]() {
				goto l11
			}
			goto l7
		l11:
			if !p.rules[ruleHorizontalRule]() {
				goto l12
			}
			goto l7
		l12:
			if !p.rules[ruleHeading]() {
				goto l13
			}
			goto l7
		l13:
			if !p.rules[ruleDefinitionList]() {
				goto l14
			}
			goto l7
		l14:
			if !p.rules[ruleOrderedList]() {
				goto l15
			}
			goto l7
		l15:
			if !p.rules[ruleBulletList]() {
				goto l16
			}
			goto l7
		l16:
			if !p.rules[ruleHtmlBlock]() {
				goto l17
			}
			goto l7
		l17:
			if !p.rules[ruleStyleBlock]() {
				goto l18
			}
			goto l7
		l18:
			if !(p.extension.Table) {
				goto l19
			}
			if !p.rules[ruleTable]() {
				goto l19
			}
			goto l7
		l19:
			if !p.rules[rulePara]() {
				goto l20
			}
			goto l7
		l20:
			if !p.rules[rulePlain]() {
				goto l4
			}
		l7:
			return true
		l4:
			position = position0
			return false
		},
		/* 3 Para <- (NonindentSpace Inlines BlankLine+ { yy = a; yy.key = PARA }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l21
			}
			if !p.rules[ruleInlines]() {
				goto l21
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l21
			}
		l22:
			if !p.rules[ruleBlankLine]() {
				goto l23
			}
			goto l22
		l23:
			do(3)
			doarg(yyPop, 1)
			return true
		l21:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 4 Plain <- (Inlines { yy = a; yy.key = PLAIN }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleInlines]() {
				goto l24
			}
			doarg(yySet, -1)
			do(4)
			doarg(yyPop, 1)
			return true
		l24:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 5 AtxInline <- (!Newline !(Sp? '#'* Sp Newline) Inline) */
		func() bool {
			position0 := position
			if !p.rules[ruleNewline]() {
				goto l26
			}
			goto l25
		l26:
			{
				position27 := position
				if !p.rules[ruleSp]() {
					goto l28
				}
			l28:
			l30:
				if !matchChar('#') {
					goto l31
				}
				goto l30
			l31:
				if !p.rules[ruleSp]() {
					goto l27
				}
				if !p.rules[ruleNewline]() {
					goto l27
				}
				goto l25
			l27:
				position = position27
			}
			if !p.rules[ruleInline]() {
				goto l25
			}
			return true
		l25:
			position = position0
			return false
		},
		/* 6 AtxStart <- (&'#' < ('######' / '#####' / '####' / '###' / '##' / '#') > { yy = p.mkElem(H1 + (len(yytext) - 1)) }) */
		func() bool {
			position0 := position
			if !peekChar('#') {
				goto l32
			}
			begin = position
			if !matchString("######") {
				goto l34
			}
			goto l33
		l34:
			if !matchString("#####") {
				goto l35
			}
			goto l33
		l35:
			if !matchString("####") {
				goto l36
			}
			goto l33
		l36:
			if !matchString("###") {
				goto l37
			}
			goto l33
		l37:
			if !matchString("##") {
				goto l38
			}
			goto l33
		l38:
			if !matchChar('#') {
				goto l32
			}
		l33:
			end = position
			do(5)
			return true
		l32:
			position = position0
			return false
		},
		/* 7 AtxHeading <- (AtxStart Sp? StartList (AtxInline { a = cons(yy, a) })+ (Sp? '#'* Sp)? Newline { yy = p.mkList(s.key, a)
              s = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleAtxStart]() {
				goto l39
			}
			doarg(yySet, -1)
			if !p.rules[ruleSp]() {
				goto l40
			}
		l40:
			if !p.rules[ruleStartList]() {
				goto l39
			}
			doarg(yySet, -2)
			if !p.rules[ruleAtxInline]() {
				goto l39
			}
			do(6)
		l42:
			{
				position43 := position
				if !p.rules[ruleAtxInline]() {
					goto l43
				}
				do(6)
				goto l42
			l43:
				position = position43
			}
			{
				position44 := position
				if !p.rules[ruleSp]() {
					goto l46
				}
			l46:
			l48:
				if !matchChar('#') {
					goto l49
				}
				goto l48
			l49:
				if !p.rules[ruleSp]() {
					goto l44
				}
				goto l45
			l44:
				position = position44
			}
		l45:
			if !p.rules[ruleNewline]() {
				goto l39
			}
			do(7)
			doarg(yyPop, 2)
			return true
		l39:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 SetextHeading <- (SetextHeading1 / SetextHeading2) */
		func() bool {
			if !p.rules[ruleSetextHeading1]() {
				goto l52
			}
			goto l51
		l52:
			if !p.rules[ruleSetextHeading2]() {
				goto l50
			}
		l51:
			return true
		l50:
			return false
		},
		/* 9 SetextBottom1 <- ('='+ Newline) */
		func() bool {
			position0 := position
			if !matchChar('=') {
				goto l53
			}
		l54:
			if !matchChar('=') {
				goto l55
			}
			goto l54
		l55:
			if !p.rules[ruleNewline]() {
				goto l53
			}
			return true
		l53:
			position = position0
			return false
		},
		/* 10 SetextBottom2 <- ('-'+ Newline) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l56
			}
		l57:
			if !matchChar('-') {
				goto l58
			}
			goto l57
		l58:
			if !p.rules[ruleNewline]() {
				goto l56
			}
			return true
		l56:
			position = position0
			return false
		},
		/* 11 SetextHeading1 <- (&(RawLine SetextBottom1) StartList (!Endline Inline { a = cons(yy, a) })+ Sp? Newline SetextBottom1 { yy = p.mkList(H1, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position60 := position
				if !p.rules[ruleRawLine]() {
					goto l59
				}
				if !p.rules[ruleSetextBottom1]() {
					goto l59
				}
				position = position60
			}
			if !p.rules[ruleStartList]() {
				goto l59
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l63
			}
			goto l59
		l63:
			if !p.rules[ruleInline]() {
				goto l59
			}
			do(8)
		l61:
			{
				position62 := position
				if !p.rules[ruleEndline]() {
					goto l64
				}
				goto l62
			l64:
				if !p.rules[ruleInline]() {
					goto l62
				}
				do(8)
				goto l61
			l62:
				position = position62
			}
			if !p.rules[ruleSp]() {
				goto l65
			}
		l65:
			if !p.rules[ruleNewline]() {
				goto l59
			}
			if !p.rules[ruleSetextBottom1]() {
				goto l59
			}
			do(9)
			doarg(yyPop, 1)
			return true
		l59:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 12 SetextHeading2 <- (&(RawLine SetextBottom2) StartList (!Endline Inline { a = cons(yy, a) })+ Sp? Newline SetextBottom2 { yy = p.mkList(H2, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position68 := position
				if !p.rules[ruleRawLine]() {
					goto l67
				}
				if !p.rules[ruleSetextBottom2]() {
					goto l67
				}
				position = position68
			}
			if !p.rules[ruleStartList]() {
				goto l67
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l71
			}
			goto l67
		l71:
			if !p.rules[ruleInline]() {
				goto l67
			}
			do(10)
		l69:
			{
				position70 := position
				if !p.rules[ruleEndline]() {
					goto l72
				}
				goto l70
			l72:
				if !p.rules[ruleInline]() {
					goto l70
				}
				do(10)
				goto l69
			l70:
				position = position70
			}
			if !p.rules[ruleSp]() {
				goto l73
			}
		l73:
			if !p.rules[ruleNewline]() {
				goto l67
			}
			if !p.rules[ruleSetextBottom2]() {
				goto l67
			}
			do(11)
			doarg(yyPop, 1)
			return true
		l67:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 13 Heading <- (SetextHeading / AtxHeading) */
		func() bool {
			if !p.rules[ruleSetextHeading]() {
				goto l77
			}
			goto l76
		l77:
			if !p.rules[ruleAtxHeading]() {
				goto l75
			}
		l76:
			return true
		l75:
			return false
		},
		/* 14 BlockQuote <- (BlockQuoteRaw {  yy = p.mkElem(BLOCKQUOTE)
                yy.children = a
             }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleBlockQuoteRaw]() {
				goto l78
			}
			doarg(yySet, -1)
			do(12)
			doarg(yyPop, 1)
			return true
		l78:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 BlockQuoteRaw <- (StartList ('>' ' '? Line { a = cons(yy, a) } (!'>' !BlankLine Line { a = cons(yy, a) })* (BlankLine { a = cons(p.mkString("\n"), a) })*)+ {   yy = p.mkStringFromList(a, true)
                     yy.key = RAW
                 }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l79
			}
			doarg(yySet, -1)
			if !matchChar('>') {
				goto l79
			}
			matchChar(' ')
			if !p.rules[ruleLine]() {
				goto l79
			}
			do(13)
		l82:
			{
				position83, thunkPosition83 := position, thunkPosition
				if peekChar('>') {
					goto l83
				}
				if !p.rules[ruleBlankLine]() {
					goto l84
				}
				goto l83
			l84:
				if !p.rules[ruleLine]() {
					goto l83
				}
				do(14)
				goto l82
			l83:
				position, thunkPosition = position83, thunkPosition83
			}
		l85:
			{
				position86 := position
				if !p.rules[ruleBlankLine]() {
					goto l86
				}
				do(15)
				goto l85
			l86:
				position = position86
			}
		l80:
			{
				position81, thunkPosition81 := position, thunkPosition
				if !matchChar('>') {
					goto l81
				}
				matchChar(' ')
				if !p.rules[ruleLine]() {
					goto l81
				}
				do(13)
			l87:
				{
					position88, thunkPosition88 := position, thunkPosition
					if peekChar('>') {
						goto l88
					}
					if !p.rules[ruleBlankLine]() {
						goto l89
					}
					goto l88
				l89:
					if !p.rules[ruleLine]() {
						goto l88
					}
					do(14)
					goto l87
				l88:
					position, thunkPosition = position88, thunkPosition88
				}
			l90:
				{
					position91 := position
					if !p.rules[ruleBlankLine]() {
						goto l91
					}
					do(15)
					goto l90
				l91:
					position = position91
				}
				goto l80
			l81:
				position, thunkPosition = position81, thunkPosition81
			}
			do(16)
			doarg(yyPop, 1)
			return true
		l79:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 NonblankIndentedLine <- (!BlankLine IndentedLine) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l93
			}
			goto l92
		l93:
			if !p.rules[ruleIndentedLine]() {
				goto l92
			}
			return true
		l92:
			position = position0
			return false
		},
		/* 17 VerbatimChunk <- (StartList (BlankLine { a = cons(p.mkString("\n"), a) })* (NonblankIndentedLine { a = cons(yy, a) })+ { yy = p.mkStringFromList(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l94
			}
			doarg(yySet, -1)
		l95:
			{
				position96 := position
				if !p.rules[ruleBlankLine]() {
					goto l96
				}
				do(17)
				goto l95
			l96:
				position = position96
			}
			if !p.rules[ruleNonblankIndentedLine]() {
				goto l94
			}
			do(18)
		l97:
			{
				position98 := position
				if !p.rules[ruleNonblankIndentedLine]() {
					goto l98
				}
				do(18)
				goto l97
			l98:
				position = position98
			}
			do(19)
			doarg(yyPop, 1)
			return true
		l94:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 Verbatim <- (StartList (VerbatimChunk { a = cons(yy, a) })+ { yy = p.mkStringFromList(a, false)
                 yy.key = VERBATIM }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l99
			}
			doarg(yySet, -1)
			if !p.rules[ruleVerbatimChunk]() {
				goto l99
			}
			do(20)
		l100:
			{
				position101, thunkPosition101 := position, thunkPosition
				if !p.rules[ruleVerbatimChunk]() {
					goto l101
				}
				do(20)
				goto l100
			l101:
				position, thunkPosition = position101, thunkPosition101
			}
			do(21)
			doarg(yyPop, 1)
			return true
		l99:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 19 HorizontalRule <- (NonindentSpace ((&[_] ('_' Sp '_' Sp '_' (Sp '_')*)) | (&[\-] ('-' Sp '-' Sp '-' (Sp '-')*)) | (&[*] ('*' Sp '*' Sp '*' (Sp '*')*))) Sp Newline BlankLine+ { yy = p.mkElem(HRULE) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l102
			}
			{
				if position == len(p.Buffer) {
					goto l102
				}
				switch p.Buffer[position] {
				case '_':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l102
					}
					if !matchChar('_') {
						goto l102
					}
					if !p.rules[ruleSp]() {
						goto l102
					}
					if !matchChar('_') {
						goto l102
					}
				l104:
					{
						position105 := position
						if !p.rules[ruleSp]() {
							goto l105
						}
						if !matchChar('_') {
							goto l105
						}
						goto l104
					l105:
						position = position105
					}
					break
				case '-':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l102
					}
					if !matchChar('-') {
						goto l102
					}
					if !p.rules[ruleSp]() {
						goto l102
					}
					if !matchChar('-') {
						goto l102
					}
				l106:
					{
						position107 := position
						if !p.rules[ruleSp]() {
							goto l107
						}
						if !matchChar('-') {
							goto l107
						}
						goto l106
					l107:
						position = position107
					}
					break
				case '*':
					position++ // matchChar
					if !p.rules[ruleSp]() {
						goto l102
					}
					if !matchChar('*') {
						goto l102
					}
					if !p.rules[ruleSp]() {
						goto l102
					}
					if !matchChar('*') {
						goto l102
					}
				l108:
					{
						position109 := position
						if !p.rules[ruleSp]() {
							goto l109
						}
						if !matchChar('*') {
							goto l109
						}
						goto l108
					l109:
						position = position109
					}
					break
				default:
					goto l102
				}
			}
			if !p.rules[ruleSp]() {
				goto l102
			}
			if !p.rules[ruleNewline]() {
				goto l102
			}
			if !p.rules[ruleBlankLine]() {
				goto l102
			}
		l110:
			if !p.rules[ruleBlankLine]() {
				goto l111
			}
			goto l110
		l111:
			do(22)
			return true
		l102:
			position = position0
			return false
		},
		/* 20 Bullet <- (!HorizontalRule NonindentSpace ((&[\-] '-') | (&[*] '*') | (&[+] '+')) Spacechar+) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleHorizontalRule]() {
				goto l113
			}
			goto l112
		l113:
			if !p.rules[ruleNonindentSpace]() {
				goto l112
			}
			{
				if position == len(p.Buffer) {
					goto l112
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				case '*':
					position++ // matchChar
					break
				case '+':
					position++ // matchChar
					break
				default:
					goto l112
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l112
			}
		l115:
			if !p.rules[ruleSpacechar]() {
				goto l116
			}
			goto l115
		l116:
			return true
		l112:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 BulletList <- (&Bullet (ListTight / ListLoose) { yy.key = BULLETLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position118 := position
				if !p.rules[ruleBullet]() {
					goto l117
				}
				position = position118
			}
			if !p.rules[ruleListTight]() {
				goto l120
			}
			goto l119
		l120:
			if !p.rules[ruleListLoose]() {
				goto l117
			}
		l119:
			do(23)
			return true
		l117:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 22 ListTight <- (StartList (ListItemTight { a = cons(yy, a) })+ BlankLine* !((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l121
			}
			doarg(yySet, -1)
			if !p.rules[ruleListItemTight]() {
				goto l121
			}
			do(24)
		l122:
			{
				position123, thunkPosition123 := position, thunkPosition
				if !p.rules[ruleListItemTight]() {
					goto l123
				}
				do(24)
				goto l122
			l123:
				position, thunkPosition = position123, thunkPosition123
			}
		l124:
			if !p.rules[ruleBlankLine]() {
				goto l125
			}
			goto l124
		l125:
			{
				if position == len(p.Buffer) {
					goto l126
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l126
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l126
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l126
					}
				}
			}
			goto l121
		l126:
			do(25)
			doarg(yyPop, 1)
			return true
		l121:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 23 ListLoose <- (StartList (ListItem BlankLine* {
                  li := b.children
                  li.contents.str += "\n\n"
                  a = cons(b, a)
              })+ { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l128
			}
			doarg(yySet, -2)
			if !p.rules[ruleListItem]() {
				goto l128
			}
			doarg(yySet, -1)
		l131:
			if !p.rules[ruleBlankLine]() {
				goto l132
			}
			goto l131
		l132:
			do(26)
		l129:
			{
				position130, thunkPosition130 := position, thunkPosition
				if !p.rules[ruleListItem]() {
					goto l130
				}
				doarg(yySet, -1)
			l133:
				if !p.rules[ruleBlankLine]() {
					goto l134
				}
				goto l133
			l134:
				do(26)
				goto l129
			l130:
				position, thunkPosition = position130, thunkPosition130
			}
			do(27)
			doarg(yyPop, 2)
			return true
		l128:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 24 ListItem <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (ListContinuationBlock { a = cons(yy, a) })* {
               raw := p.mkStringFromList(a, false)
               raw.key = RAW
               yy = p.mkElem(LISTITEM)
               yy.children = raw
            }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				if position == len(p.Buffer) {
					goto l135
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l135
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l135
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l135
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto l135
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l135
			}
			do(28)
		l137:
			{
				position138, thunkPosition138 := position, thunkPosition
				if !p.rules[ruleListContinuationBlock]() {
					goto l138
				}
				do(29)
				goto l137
			l138:
				position, thunkPosition = position138, thunkPosition138
			}
			do(30)
			doarg(yyPop, 1)
			return true
		l135:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 ListItemTight <- (((&[:~] DefMarker) | (&[*+\-] Bullet) | (&[0-9] Enumerator)) StartList ListBlock { a = cons(yy, a) } (!BlankLine ListContinuationBlock { a = cons(yy, a) })* !ListContinuationBlock {
               raw := p.mkStringFromList(a, false)
               raw.key = RAW
               yy = p.mkElem(LISTITEM)
               yy.children = raw
            }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				if position == len(p.Buffer) {
					goto l139
				}
				switch p.Buffer[position] {
				case ':', '~':
					if !p.rules[ruleDefMarker]() {
						goto l139
					}
					break
				case '*', '+', '-':
					if !p.rules[ruleBullet]() {
						goto l139
					}
					break
				default:
					if !p.rules[ruleEnumerator]() {
						goto l139
					}
				}
			}
			if !p.rules[ruleStartList]() {
				goto l139
			}
			doarg(yySet, -1)
			if !p.rules[ruleListBlock]() {
				goto l139
			}
			do(31)
		l141:
			{
				position142, thunkPosition142 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l143
				}
				goto l142
			l143:
				if !p.rules[ruleListContinuationBlock]() {
					goto l142
				}
				do(32)
				goto l141
			l142:
				position, thunkPosition = position142, thunkPosition142
			}
			if !p.rules[ruleListContinuationBlock]() {
				goto l144
			}
			goto l139
		l144:
			do(33)
			doarg(yyPop, 1)
			return true
		l139:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 26 ListBlock <- (StartList !BlankLine Line { a = cons(yy, a) } (ListBlockLine { a = cons(yy, a) })* { yy = p.mkStringFromList(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l145
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l146
			}
			goto l145
		l146:
			if !p.rules[ruleLine]() {
				goto l145
			}
			do(34)
		l147:
			{
				position148 := position
				if !p.rules[ruleListBlockLine]() {
					goto l148
				}
				do(35)
				goto l147
			l148:
				position = position148
			}
			do(36)
			doarg(yyPop, 1)
			return true
		l145:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 ListContinuationBlock <- (StartList (< BlankLine* > {   if len(yytext) == 0 {
                                   a = cons(p.mkString("\001"), a) // block separator
                              } else {
                                   a = cons(p.mkString(yytext), a)
                              }
                          }) (Indent ListBlock { a = cons(yy, a) })+ {  yy = p.mkStringFromList(a, false) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l149
			}
			doarg(yySet, -1)
			begin = position
		l150:
			if !p.rules[ruleBlankLine]() {
				goto l151
			}
			goto l150
		l151:
			end = position
			do(37)
			if !p.rules[ruleIndent]() {
				goto l149
			}
			if !p.rules[ruleListBlock]() {
				goto l149
			}
			do(38)
		l152:
			{
				position153, thunkPosition153 := position, thunkPosition
				if !p.rules[ruleIndent]() {
					goto l153
				}
				if !p.rules[ruleListBlock]() {
					goto l153
				}
				do(38)
				goto l152
			l153:
				position, thunkPosition = position153, thunkPosition153
			}
			do(39)
			doarg(yyPop, 1)
			return true
		l149:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 Enumerator <- (NonindentSpace [0-9]+ '.' Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l154
			}
			if !matchClass(0) {
				goto l154
			}
		l155:
			if !matchClass(0) {
				goto l156
			}
			goto l155
		l156:
			if !matchChar('.') {
				goto l154
			}
			if !p.rules[ruleSpacechar]() {
				goto l154
			}
		l157:
			if !p.rules[ruleSpacechar]() {
				goto l158
			}
			goto l157
		l158:
			return true
		l154:
			position = position0
			return false
		},
		/* 29 OrderedList <- (&Enumerator (ListTight / ListLoose) { yy.key = ORDEREDLIST }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position160 := position
				if !p.rules[ruleEnumerator]() {
					goto l159
				}
				position = position160
			}
			if !p.rules[ruleListTight]() {
				goto l162
			}
			goto l161
		l162:
			if !p.rules[ruleListLoose]() {
				goto l159
			}
		l161:
			do(40)
			return true
		l159:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 ListBlockLine <- (!BlankLine !((&[:~] DefMarker) | (&[\t *+\-0-9] (Indent? ((&[*+\-] Bullet) | (&[0-9] Enumerator))))) !HorizontalRule OptionallyIndentedLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleBlankLine]() {
				goto l164
			}
			goto l163
		l164:
			{
				position165 := position
				{
					if position == len(p.Buffer) {
						goto l165
					}
					switch p.Buffer[position] {
					case ':', '~':
						if !p.rules[ruleDefMarker]() {
							goto l165
						}
						break
					default:
						if !p.rules[ruleIndent]() {
							goto l167
						}
					l167:
						{
							if position == len(p.Buffer) {
								goto l165
							}
							switch p.Buffer[position] {
							case '*', '+', '-':
								if !p.rules[ruleBullet]() {
									goto l165
								}
								break
							default:
								if !p.rules[ruleEnumerator]() {
									goto l165
								}
							}
						}
					}
				}
				goto l163
			l165:
				position = position165
			}
			if !p.rules[ruleHorizontalRule]() {
				goto l170
			}
			goto l163
		l170:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l163
			}
			return true
		l163:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 HtmlBlockOpenAddress <- ('<' Spnl ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l171
			}
			if !p.rules[ruleSpnl]() {
				goto l171
			}
			{
				if position == len(p.Buffer) {
					goto l171
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l171
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l171
					}
					break
				default:
					goto l171
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l171
			}
		l173:
			if !p.rules[ruleHtmlAttribute]() {
				goto l174
			}
			goto l173
		l174:
			if !matchChar('>') {
				goto l171
			}
			return true
		l171:
			position = position0
			return false
		},
		/* 32 HtmlBlockCloseAddress <- ('<' Spnl '/' ((&[A] 'ADDRESS') | (&[a] 'address')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l175
			}
			if !p.rules[ruleSpnl]() {
				goto l175
			}
			if !matchChar('/') {
				goto l175
			}
			{
				if position == len(p.Buffer) {
					goto l175
				}
				switch p.Buffer[position] {
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l175
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l175
					}
					break
				default:
					goto l175
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l175
			}
			if !matchChar('>') {
				goto l175
			}
			return true
		l175:
			position = position0
			return false
		},
		/* 33 HtmlBlockAddress <- (HtmlBlockOpenAddress (HtmlBlockAddress / (!HtmlBlockCloseAddress .))* HtmlBlockCloseAddress) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenAddress]() {
				goto l177
			}
		l178:
			{
				position179 := position
				if !p.rules[ruleHtmlBlockAddress]() {
					goto l181
				}
				goto l180
			l181:
				if !p.rules[ruleHtmlBlockCloseAddress]() {
					goto l182
				}
				goto l179
			l182:
				if !matchDot() {
					goto l179
				}
			l180:
				goto l178
			l179:
				position = position179
			}
			if !p.rules[ruleHtmlBlockCloseAddress]() {
				goto l177
			}
			return true
		l177:
			position = position0
			return false
		},
		/* 34 HtmlBlockOpenBlockquote <- ('<' Spnl ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l183
			}
			if !p.rules[ruleSpnl]() {
				goto l183
			}
			{
				if position == len(p.Buffer) {
					goto l183
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l183
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l183
					}
					break
				default:
					goto l183
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l183
			}
		l185:
			if !p.rules[ruleHtmlAttribute]() {
				goto l186
			}
			goto l185
		l186:
			if !matchChar('>') {
				goto l183
			}
			return true
		l183:
			position = position0
			return false
		},
		/* 35 HtmlBlockCloseBlockquote <- ('<' Spnl '/' ((&[B] 'BLOCKQUOTE') | (&[b] 'blockquote')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l187
			}
			if !p.rules[ruleSpnl]() {
				goto l187
			}
			if !matchChar('/') {
				goto l187
			}
			{
				if position == len(p.Buffer) {
					goto l187
				}
				switch p.Buffer[position] {
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l187
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l187
					}
					break
				default:
					goto l187
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l187
			}
			if !matchChar('>') {
				goto l187
			}
			return true
		l187:
			position = position0
			return false
		},
		/* 36 HtmlBlockBlockquote <- (HtmlBlockOpenBlockquote (HtmlBlockBlockquote / (!HtmlBlockCloseBlockquote .))* HtmlBlockCloseBlockquote) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenBlockquote]() {
				goto l189
			}
		l190:
			{
				position191 := position
				if !p.rules[ruleHtmlBlockBlockquote]() {
					goto l193
				}
				goto l192
			l193:
				if !p.rules[ruleHtmlBlockCloseBlockquote]() {
					goto l194
				}
				goto l191
			l194:
				if !matchDot() {
					goto l191
				}
			l192:
				goto l190
			l191:
				position = position191
			}
			if !p.rules[ruleHtmlBlockCloseBlockquote]() {
				goto l189
			}
			return true
		l189:
			position = position0
			return false
		},
		/* 37 HtmlBlockOpenCenter <- ('<' Spnl ((&[C] 'CENTER') | (&[c] 'center')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l195
			}
			if !p.rules[ruleSpnl]() {
				goto l195
			}
			{
				if position == len(p.Buffer) {
					goto l195
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l195
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l195
					}
					break
				default:
					goto l195
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l195
			}
		l197:
			if !p.rules[ruleHtmlAttribute]() {
				goto l198
			}
			goto l197
		l198:
			if !matchChar('>') {
				goto l195
			}
			return true
		l195:
			position = position0
			return false
		},
		/* 38 HtmlBlockCloseCenter <- ('<' Spnl '/' ((&[C] 'CENTER') | (&[c] 'center')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l199
			}
			if !p.rules[ruleSpnl]() {
				goto l199
			}
			if !matchChar('/') {
				goto l199
			}
			{
				if position == len(p.Buffer) {
					goto l199
				}
				switch p.Buffer[position] {
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l199
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l199
					}
					break
				default:
					goto l199
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l199
			}
			if !matchChar('>') {
				goto l199
			}
			return true
		l199:
			position = position0
			return false
		},
		/* 39 HtmlBlockCenter <- (HtmlBlockOpenCenter (HtmlBlockCenter / (!HtmlBlockCloseCenter .))* HtmlBlockCloseCenter) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenCenter]() {
				goto l201
			}
		l202:
			{
				position203 := position
				if !p.rules[ruleHtmlBlockCenter]() {
					goto l205
				}
				goto l204
			l205:
				if !p.rules[ruleHtmlBlockCloseCenter]() {
					goto l206
				}
				goto l203
			l206:
				if !matchDot() {
					goto l203
				}
			l204:
				goto l202
			l203:
				position = position203
			}
			if !p.rules[ruleHtmlBlockCloseCenter]() {
				goto l201
			}
			return true
		l201:
			position = position0
			return false
		},
		/* 40 HtmlBlockOpenDir <- ('<' Spnl ((&[D] 'DIR') | (&[d] 'dir')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l207
			}
			if !p.rules[ruleSpnl]() {
				goto l207
			}
			{
				if position == len(p.Buffer) {
					goto l207
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto l207
					}
					break
				case 'd':
					position++
					if !matchString("ir") {
						goto l207
					}
					break
				default:
					goto l207
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l207
			}
		l209:
			if !p.rules[ruleHtmlAttribute]() {
				goto l210
			}
			goto l209
		l210:
			if !matchChar('>') {
				goto l207
			}
			return true
		l207:
			position = position0
			return false
		},
		/* 41 HtmlBlockCloseDir <- ('<' Spnl '/' ((&[D] 'DIR') | (&[d] 'dir')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l211
			}
			if !p.rules[ruleSpnl]() {
				goto l211
			}
			if !matchChar('/') {
				goto l211
			}
			{
				if position == len(p.Buffer) {
					goto l211
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IR") {
						goto l211
					}
					break
				case 'd':
					position++
					if !matchString("ir") {
						goto l211
					}
					break
				default:
					goto l211
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l211
			}
			if !matchChar('>') {
				goto l211
			}
			return true
		l211:
			position = position0
			return false
		},
		/* 42 HtmlBlockDir <- (HtmlBlockOpenDir (HtmlBlockDir / (!HtmlBlockCloseDir .))* HtmlBlockCloseDir) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDir]() {
				goto l213
			}
		l214:
			{
				position215 := position
				if !p.rules[ruleHtmlBlockDir]() {
					goto l217
				}
				goto l216
			l217:
				if !p.rules[ruleHtmlBlockCloseDir]() {
					goto l218
				}
				goto l215
			l218:
				if !matchDot() {
					goto l215
				}
			l216:
				goto l214
			l215:
				position = position215
			}
			if !p.rules[ruleHtmlBlockCloseDir]() {
				goto l213
			}
			return true
		l213:
			position = position0
			return false
		},
		/* 43 HtmlBlockOpenDiv <- ('<' Spnl ((&[D] 'DIV') | (&[d] 'div')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l219
			}
			if !p.rules[ruleSpnl]() {
				goto l219
			}
			{
				if position == len(p.Buffer) {
					goto l219
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto l219
					}
					break
				case 'd':
					position++
					if !matchString("iv") {
						goto l219
					}
					break
				default:
					goto l219
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l219
			}
		l221:
			if !p.rules[ruleHtmlAttribute]() {
				goto l222
			}
			goto l221
		l222:
			if !matchChar('>') {
				goto l219
			}
			return true
		l219:
			position = position0
			return false
		},
		/* 44 HtmlBlockCloseDiv <- ('<' Spnl '/' ((&[D] 'DIV') | (&[d] 'div')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l223
			}
			if !p.rules[ruleSpnl]() {
				goto l223
			}
			if !matchChar('/') {
				goto l223
			}
			{
				if position == len(p.Buffer) {
					goto l223
				}
				switch p.Buffer[position] {
				case 'D':
					position++
					if !matchString("IV") {
						goto l223
					}
					break
				case 'd':
					position++
					if !matchString("iv") {
						goto l223
					}
					break
				default:
					goto l223
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l223
			}
			if !matchChar('>') {
				goto l223
			}
			return true
		l223:
			position = position0
			return false
		},
		/* 45 HtmlBlockDiv <- (HtmlBlockOpenDiv (HtmlBlockDiv / (!HtmlBlockCloseDiv .))* HtmlBlockCloseDiv) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDiv]() {
				goto l225
			}
		l226:
			{
				position227 := position
				if !p.rules[ruleHtmlBlockDiv]() {
					goto l229
				}
				goto l228
			l229:
				if !p.rules[ruleHtmlBlockCloseDiv]() {
					goto l230
				}
				goto l227
			l230:
				if !matchDot() {
					goto l227
				}
			l228:
				goto l226
			l227:
				position = position227
			}
			if !p.rules[ruleHtmlBlockCloseDiv]() {
				goto l225
			}
			return true
		l225:
			position = position0
			return false
		},
		/* 46 HtmlBlockOpenDl <- ('<' Spnl ((&[D] 'DL') | (&[d] 'dl')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l231
			}
			if !p.rules[ruleSpnl]() {
				goto l231
			}
			{
				if position == len(p.Buffer) {
					goto l231
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto l231
					}
					break
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto l231
					}
					break
				default:
					goto l231
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l231
			}
		l233:
			if !p.rules[ruleHtmlAttribute]() {
				goto l234
			}
			goto l233
		l234:
			if !matchChar('>') {
				goto l231
			}
			return true
		l231:
			position = position0
			return false
		},
		/* 47 HtmlBlockCloseDl <- ('<' Spnl '/' ((&[D] 'DL') | (&[d] 'dl')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l235
			}
			if !p.rules[ruleSpnl]() {
				goto l235
			}
			if !matchChar('/') {
				goto l235
			}
			{
				if position == len(p.Buffer) {
					goto l235
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DL`)
					if !matchChar('L') {
						goto l235
					}
					break
				case 'd':
					position++ // matchString(`dl`)
					if !matchChar('l') {
						goto l235
					}
					break
				default:
					goto l235
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l235
			}
			if !matchChar('>') {
				goto l235
			}
			return true
		l235:
			position = position0
			return false
		},
		/* 48 HtmlBlockDl <- (HtmlBlockOpenDl (HtmlBlockDl / (!HtmlBlockCloseDl .))* HtmlBlockCloseDl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDl]() {
				goto l237
			}
		l238:
			{
				position239 := position
				if !p.rules[ruleHtmlBlockDl]() {
					goto l241
				}
				goto l240
			l241:
				if !p.rules[ruleHtmlBlockCloseDl]() {
					goto l242
				}
				goto l239
			l242:
				if !matchDot() {
					goto l239
				}
			l240:
				goto l238
			l239:
				position = position239
			}
			if !p.rules[ruleHtmlBlockCloseDl]() {
				goto l237
			}
			return true
		l237:
			position = position0
			return false
		},
		/* 49 HtmlBlockOpenFieldset <- ('<' Spnl ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l243
			}
			if !p.rules[ruleSpnl]() {
				goto l243
			}
			{
				if position == len(p.Buffer) {
					goto l243
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto l243
					}
					break
				case 'f':
					position++
					if !matchString("ieldset") {
						goto l243
					}
					break
				default:
					goto l243
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l243
			}
		l245:
			if !p.rules[ruleHtmlAttribute]() {
				goto l246
			}
			goto l245
		l246:
			if !matchChar('>') {
				goto l243
			}
			return true
		l243:
			position = position0
			return false
		},
		/* 50 HtmlBlockCloseFieldset <- ('<' Spnl '/' ((&[F] 'FIELDSET') | (&[f] 'fieldset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l247
			}
			if !p.rules[ruleSpnl]() {
				goto l247
			}
			if !matchChar('/') {
				goto l247
			}
			{
				if position == len(p.Buffer) {
					goto l247
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("IELDSET") {
						goto l247
					}
					break
				case 'f':
					position++
					if !matchString("ieldset") {
						goto l247
					}
					break
				default:
					goto l247
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l247
			}
			if !matchChar('>') {
				goto l247
			}
			return true
		l247:
			position = position0
			return false
		},
		/* 51 HtmlBlockFieldset <- (HtmlBlockOpenFieldset (HtmlBlockFieldset / (!HtmlBlockCloseFieldset .))* HtmlBlockCloseFieldset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFieldset]() {
				goto l249
			}
		l250:
			{
				position251 := position
				if !p.rules[ruleHtmlBlockFieldset]() {
					goto l253
				}
				goto l252
			l253:
				if !p.rules[ruleHtmlBlockCloseFieldset]() {
					goto l254
				}
				goto l251
			l254:
				if !matchDot() {
					goto l251
				}
			l252:
				goto l250
			l251:
				position = position251
			}
			if !p.rules[ruleHtmlBlockCloseFieldset]() {
				goto l249
			}
			return true
		l249:
			position = position0
			return false
		},
		/* 52 HtmlBlockOpenForm <- ('<' Spnl ((&[F] 'FORM') | (&[f] 'form')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l255
			}
			if !p.rules[ruleSpnl]() {
				goto l255
			}
			{
				if position == len(p.Buffer) {
					goto l255
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto l255
					}
					break
				case 'f':
					position++
					if !matchString("orm") {
						goto l255
					}
					break
				default:
					goto l255
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l255
			}
		l257:
			if !p.rules[ruleHtmlAttribute]() {
				goto l258
			}
			goto l257
		l258:
			if !matchChar('>') {
				goto l255
			}
			return true
		l255:
			position = position0
			return false
		},
		/* 53 HtmlBlockCloseForm <- ('<' Spnl '/' ((&[F] 'FORM') | (&[f] 'form')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l259
			}
			if !p.rules[ruleSpnl]() {
				goto l259
			}
			if !matchChar('/') {
				goto l259
			}
			{
				if position == len(p.Buffer) {
					goto l259
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("ORM") {
						goto l259
					}
					break
				case 'f':
					position++
					if !matchString("orm") {
						goto l259
					}
					break
				default:
					goto l259
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l259
			}
			if !matchChar('>') {
				goto l259
			}
			return true
		l259:
			position = position0
			return false
		},
		/* 54 HtmlBlockForm <- (HtmlBlockOpenForm (HtmlBlockForm / (!HtmlBlockCloseForm .))* HtmlBlockCloseForm) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenForm]() {
				goto l261
			}
		l262:
			{
				position263 := position
				if !p.rules[ruleHtmlBlockForm]() {
					goto l265
				}
				goto l264
			l265:
				if !p.rules[ruleHtmlBlockCloseForm]() {
					goto l266
				}
				goto l263
			l266:
				if !matchDot() {
					goto l263
				}
			l264:
				goto l262
			l263:
				position = position263
			}
			if !p.rules[ruleHtmlBlockCloseForm]() {
				goto l261
			}
			return true
		l261:
			position = position0
			return false
		},
		/* 55 HtmlBlockOpenH1 <- ('<' Spnl ((&[H] 'H1') | (&[h] 'h1')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l267
			}
			if !p.rules[ruleSpnl]() {
				goto l267
			}
			{
				if position == len(p.Buffer) {
					goto l267
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto l267
					}
					break
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto l267
					}
					break
				default:
					goto l267
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l267
			}
		l269:
			if !p.rules[ruleHtmlAttribute]() {
				goto l270
			}
			goto l269
		l270:
			if !matchChar('>') {
				goto l267
			}
			return true
		l267:
			position = position0
			return false
		},
		/* 56 HtmlBlockCloseH1 <- ('<' Spnl '/' ((&[H] 'H1') | (&[h] 'h1')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l271
			}
			if !p.rules[ruleSpnl]() {
				goto l271
			}
			if !matchChar('/') {
				goto l271
			}
			{
				if position == len(p.Buffer) {
					goto l271
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H1`)
					if !matchChar('1') {
						goto l271
					}
					break
				case 'h':
					position++ // matchString(`h1`)
					if !matchChar('1') {
						goto l271
					}
					break
				default:
					goto l271
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l271
			}
			if !matchChar('>') {
				goto l271
			}
			return true
		l271:
			position = position0
			return false
		},
		/* 57 HtmlBlockH1 <- (HtmlBlockOpenH1 (HtmlBlockH1 / (!HtmlBlockCloseH1 .))* HtmlBlockCloseH1) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH1]() {
				goto l273
			}
		l274:
			{
				position275 := position
				if !p.rules[ruleHtmlBlockH1]() {
					goto l277
				}
				goto l276
			l277:
				if !p.rules[ruleHtmlBlockCloseH1]() {
					goto l278
				}
				goto l275
			l278:
				if !matchDot() {
					goto l275
				}
			l276:
				goto l274
			l275:
				position = position275
			}
			if !p.rules[ruleHtmlBlockCloseH1]() {
				goto l273
			}
			return true
		l273:
			position = position0
			return false
		},
		/* 58 HtmlBlockOpenH2 <- ('<' Spnl ((&[H] 'H2') | (&[h] 'h2')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l279
			}
			if !p.rules[ruleSpnl]() {
				goto l279
			}
			{
				if position == len(p.Buffer) {
					goto l279
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto l279
					}
					break
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto l279
					}
					break
				default:
					goto l279
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l279
			}
		l281:
			if !p.rules[ruleHtmlAttribute]() {
				goto l282
			}
			goto l281
		l282:
			if !matchChar('>') {
				goto l279
			}
			return true
		l279:
			position = position0
			return false
		},
		/* 59 HtmlBlockCloseH2 <- ('<' Spnl '/' ((&[H] 'H2') | (&[h] 'h2')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l283
			}
			if !p.rules[ruleSpnl]() {
				goto l283
			}
			if !matchChar('/') {
				goto l283
			}
			{
				if position == len(p.Buffer) {
					goto l283
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H2`)
					if !matchChar('2') {
						goto l283
					}
					break
				case 'h':
					position++ // matchString(`h2`)
					if !matchChar('2') {
						goto l283
					}
					break
				default:
					goto l283
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l283
			}
			if !matchChar('>') {
				goto l283
			}
			return true
		l283:
			position = position0
			return false
		},
		/* 60 HtmlBlockH2 <- (HtmlBlockOpenH2 (HtmlBlockH2 / (!HtmlBlockCloseH2 .))* HtmlBlockCloseH2) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH2]() {
				goto l285
			}
		l286:
			{
				position287 := position
				if !p.rules[ruleHtmlBlockH2]() {
					goto l289
				}
				goto l288
			l289:
				if !p.rules[ruleHtmlBlockCloseH2]() {
					goto l290
				}
				goto l287
			l290:
				if !matchDot() {
					goto l287
				}
			l288:
				goto l286
			l287:
				position = position287
			}
			if !p.rules[ruleHtmlBlockCloseH2]() {
				goto l285
			}
			return true
		l285:
			position = position0
			return false
		},
		/* 61 HtmlBlockOpenH3 <- ('<' Spnl ((&[H] 'H3') | (&[h] 'h3')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l291
			}
			if !p.rules[ruleSpnl]() {
				goto l291
			}
			{
				if position == len(p.Buffer) {
					goto l291
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto l291
					}
					break
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto l291
					}
					break
				default:
					goto l291
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l291
			}
		l293:
			if !p.rules[ruleHtmlAttribute]() {
				goto l294
			}
			goto l293
		l294:
			if !matchChar('>') {
				goto l291
			}
			return true
		l291:
			position = position0
			return false
		},
		/* 62 HtmlBlockCloseH3 <- ('<' Spnl '/' ((&[H] 'H3') | (&[h] 'h3')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l295
			}
			if !p.rules[ruleSpnl]() {
				goto l295
			}
			if !matchChar('/') {
				goto l295
			}
			{
				if position == len(p.Buffer) {
					goto l295
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H3`)
					if !matchChar('3') {
						goto l295
					}
					break
				case 'h':
					position++ // matchString(`h3`)
					if !matchChar('3') {
						goto l295
					}
					break
				default:
					goto l295
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l295
			}
			if !matchChar('>') {
				goto l295
			}
			return true
		l295:
			position = position0
			return false
		},
		/* 63 HtmlBlockH3 <- (HtmlBlockOpenH3 (HtmlBlockH3 / (!HtmlBlockCloseH3 .))* HtmlBlockCloseH3) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH3]() {
				goto l297
			}
		l298:
			{
				position299 := position
				if !p.rules[ruleHtmlBlockH3]() {
					goto l301
				}
				goto l300
			l301:
				if !p.rules[ruleHtmlBlockCloseH3]() {
					goto l302
				}
				goto l299
			l302:
				if !matchDot() {
					goto l299
				}
			l300:
				goto l298
			l299:
				position = position299
			}
			if !p.rules[ruleHtmlBlockCloseH3]() {
				goto l297
			}
			return true
		l297:
			position = position0
			return false
		},
		/* 64 HtmlBlockOpenH4 <- ('<' Spnl ((&[H] 'H4') | (&[h] 'h4')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l303
			}
			if !p.rules[ruleSpnl]() {
				goto l303
			}
			{
				if position == len(p.Buffer) {
					goto l303
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto l303
					}
					break
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto l303
					}
					break
				default:
					goto l303
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l303
			}
		l305:
			if !p.rules[ruleHtmlAttribute]() {
				goto l306
			}
			goto l305
		l306:
			if !matchChar('>') {
				goto l303
			}
			return true
		l303:
			position = position0
			return false
		},
		/* 65 HtmlBlockCloseH4 <- ('<' Spnl '/' ((&[H] 'H4') | (&[h] 'h4')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l307
			}
			if !p.rules[ruleSpnl]() {
				goto l307
			}
			if !matchChar('/') {
				goto l307
			}
			{
				if position == len(p.Buffer) {
					goto l307
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H4`)
					if !matchChar('4') {
						goto l307
					}
					break
				case 'h':
					position++ // matchString(`h4`)
					if !matchChar('4') {
						goto l307
					}
					break
				default:
					goto l307
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l307
			}
			if !matchChar('>') {
				goto l307
			}
			return true
		l307:
			position = position0
			return false
		},
		/* 66 HtmlBlockH4 <- (HtmlBlockOpenH4 (HtmlBlockH4 / (!HtmlBlockCloseH4 .))* HtmlBlockCloseH4) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH4]() {
				goto l309
			}
		l310:
			{
				position311 := position
				if !p.rules[ruleHtmlBlockH4]() {
					goto l313
				}
				goto l312
			l313:
				if !p.rules[ruleHtmlBlockCloseH4]() {
					goto l314
				}
				goto l311
			l314:
				if !matchDot() {
					goto l311
				}
			l312:
				goto l310
			l311:
				position = position311
			}
			if !p.rules[ruleHtmlBlockCloseH4]() {
				goto l309
			}
			return true
		l309:
			position = position0
			return false
		},
		/* 67 HtmlBlockOpenH5 <- ('<' Spnl ((&[H] 'H5') | (&[h] 'h5')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l315
			}
			if !p.rules[ruleSpnl]() {
				goto l315
			}
			{
				if position == len(p.Buffer) {
					goto l315
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto l315
					}
					break
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto l315
					}
					break
				default:
					goto l315
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l315
			}
		l317:
			if !p.rules[ruleHtmlAttribute]() {
				goto l318
			}
			goto l317
		l318:
			if !matchChar('>') {
				goto l315
			}
			return true
		l315:
			position = position0
			return false
		},
		/* 68 HtmlBlockCloseH5 <- ('<' Spnl '/' ((&[H] 'H5') | (&[h] 'h5')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l319
			}
			if !p.rules[ruleSpnl]() {
				goto l319
			}
			if !matchChar('/') {
				goto l319
			}
			{
				if position == len(p.Buffer) {
					goto l319
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H5`)
					if !matchChar('5') {
						goto l319
					}
					break
				case 'h':
					position++ // matchString(`h5`)
					if !matchChar('5') {
						goto l319
					}
					break
				default:
					goto l319
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l319
			}
			if !matchChar('>') {
				goto l319
			}
			return true
		l319:
			position = position0
			return false
		},
		/* 69 HtmlBlockH5 <- (HtmlBlockOpenH5 (HtmlBlockH5 / (!HtmlBlockCloseH5 .))* HtmlBlockCloseH5) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH5]() {
				goto l321
			}
		l322:
			{
				position323 := position
				if !p.rules[ruleHtmlBlockH5]() {
					goto l325
				}
				goto l324
			l325:
				if !p.rules[ruleHtmlBlockCloseH5]() {
					goto l326
				}
				goto l323
			l326:
				if !matchDot() {
					goto l323
				}
			l324:
				goto l322
			l323:
				position = position323
			}
			if !p.rules[ruleHtmlBlockCloseH5]() {
				goto l321
			}
			return true
		l321:
			position = position0
			return false
		},
		/* 70 HtmlBlockOpenH6 <- ('<' Spnl ((&[H] 'H6') | (&[h] 'h6')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l327
			}
			if !p.rules[ruleSpnl]() {
				goto l327
			}
			{
				if position == len(p.Buffer) {
					goto l327
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto l327
					}
					break
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto l327
					}
					break
				default:
					goto l327
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l327
			}
		l329:
			if !p.rules[ruleHtmlAttribute]() {
				goto l330
			}
			goto l329
		l330:
			if !matchChar('>') {
				goto l327
			}
			return true
		l327:
			position = position0
			return false
		},
		/* 71 HtmlBlockCloseH6 <- ('<' Spnl '/' ((&[H] 'H6') | (&[h] 'h6')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l331
			}
			if !p.rules[ruleSpnl]() {
				goto l331
			}
			if !matchChar('/') {
				goto l331
			}
			{
				if position == len(p.Buffer) {
					goto l331
				}
				switch p.Buffer[position] {
				case 'H':
					position++ // matchString(`H6`)
					if !matchChar('6') {
						goto l331
					}
					break
				case 'h':
					position++ // matchString(`h6`)
					if !matchChar('6') {
						goto l331
					}
					break
				default:
					goto l331
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l331
			}
			if !matchChar('>') {
				goto l331
			}
			return true
		l331:
			position = position0
			return false
		},
		/* 72 HtmlBlockH6 <- (HtmlBlockOpenH6 (HtmlBlockH6 / (!HtmlBlockCloseH6 .))* HtmlBlockCloseH6) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenH6]() {
				goto l333
			}
		l334:
			{
				position335 := position
				if !p.rules[ruleHtmlBlockH6]() {
					goto l337
				}
				goto l336
			l337:
				if !p.rules[ruleHtmlBlockCloseH6]() {
					goto l338
				}
				goto l335
			l338:
				if !matchDot() {
					goto l335
				}
			l336:
				goto l334
			l335:
				position = position335
			}
			if !p.rules[ruleHtmlBlockCloseH6]() {
				goto l333
			}
			return true
		l333:
			position = position0
			return false
		},
		/* 73 HtmlBlockOpenMenu <- ('<' Spnl ((&[M] 'MENU') | (&[m] 'menu')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l339
			}
			if !p.rules[ruleSpnl]() {
				goto l339
			}
			{
				if position == len(p.Buffer) {
					goto l339
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto l339
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l339
					}
					break
				default:
					goto l339
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l339
			}
		l341:
			if !p.rules[ruleHtmlAttribute]() {
				goto l342
			}
			goto l341
		l342:
			if !matchChar('>') {
				goto l339
			}
			return true
		l339:
			position = position0
			return false
		},
		/* 74 HtmlBlockCloseMenu <- ('<' Spnl '/' ((&[M] 'MENU') | (&[m] 'menu')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l343
			}
			if !p.rules[ruleSpnl]() {
				goto l343
			}
			if !matchChar('/') {
				goto l343
			}
			{
				if position == len(p.Buffer) {
					goto l343
				}
				switch p.Buffer[position] {
				case 'M':
					position++
					if !matchString("ENU") {
						goto l343
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l343
					}
					break
				default:
					goto l343
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l343
			}
			if !matchChar('>') {
				goto l343
			}
			return true
		l343:
			position = position0
			return false
		},
		/* 75 HtmlBlockMenu <- (HtmlBlockOpenMenu (HtmlBlockMenu / (!HtmlBlockCloseMenu .))* HtmlBlockCloseMenu) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenMenu]() {
				goto l345
			}
		l346:
			{
				position347 := position
				if !p.rules[ruleHtmlBlockMenu]() {
					goto l349
				}
				goto l348
			l349:
				if !p.rules[ruleHtmlBlockCloseMenu]() {
					goto l350
				}
				goto l347
			l350:
				if !matchDot() {
					goto l347
				}
			l348:
				goto l346
			l347:
				position = position347
			}
			if !p.rules[ruleHtmlBlockCloseMenu]() {
				goto l345
			}
			return true
		l345:
			position = position0
			return false
		},
		/* 76 HtmlBlockOpenNoframes <- ('<' Spnl ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l351
			}
			if !p.rules[ruleSpnl]() {
				goto l351
			}
			{
				if position == len(p.Buffer) {
					goto l351
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto l351
					}
					break
				case 'n':
					position++
					if !matchString("oframes") {
						goto l351
					}
					break
				default:
					goto l351
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l351
			}
		l353:
			if !p.rules[ruleHtmlAttribute]() {
				goto l354
			}
			goto l353
		l354:
			if !matchChar('>') {
				goto l351
			}
			return true
		l351:
			position = position0
			return false
		},
		/* 77 HtmlBlockCloseNoframes <- ('<' Spnl '/' ((&[N] 'NOFRAMES') | (&[n] 'noframes')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l355
			}
			if !p.rules[ruleSpnl]() {
				goto l355
			}
			if !matchChar('/') {
				goto l355
			}
			{
				if position == len(p.Buffer) {
					goto l355
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OFRAMES") {
						goto l355
					}
					break
				case 'n':
					position++
					if !matchString("oframes") {
						goto l355
					}
					break
				default:
					goto l355
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l355
			}
			if !matchChar('>') {
				goto l355
			}
			return true
		l355:
			position = position0
			return false
		},
		/* 78 HtmlBlockNoframes <- (HtmlBlockOpenNoframes (HtmlBlockNoframes / (!HtmlBlockCloseNoframes .))* HtmlBlockCloseNoframes) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoframes]() {
				goto l357
			}
		l358:
			{
				position359 := position
				if !p.rules[ruleHtmlBlockNoframes]() {
					goto l361
				}
				goto l360
			l361:
				if !p.rules[ruleHtmlBlockCloseNoframes]() {
					goto l362
				}
				goto l359
			l362:
				if !matchDot() {
					goto l359
				}
			l360:
				goto l358
			l359:
				position = position359
			}
			if !p.rules[ruleHtmlBlockCloseNoframes]() {
				goto l357
			}
			return true
		l357:
			position = position0
			return false
		},
		/* 79 HtmlBlockOpenNoscript <- ('<' Spnl ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l363
			}
			if !p.rules[ruleSpnl]() {
				goto l363
			}
			{
				if position == len(p.Buffer) {
					goto l363
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l363
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l363
					}
					break
				default:
					goto l363
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l363
			}
		l365:
			if !p.rules[ruleHtmlAttribute]() {
				goto l366
			}
			goto l365
		l366:
			if !matchChar('>') {
				goto l363
			}
			return true
		l363:
			position = position0
			return false
		},
		/* 80 HtmlBlockCloseNoscript <- ('<' Spnl '/' ((&[N] 'NOSCRIPT') | (&[n] 'noscript')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l367
			}
			if !p.rules[ruleSpnl]() {
				goto l367
			}
			if !matchChar('/') {
				goto l367
			}
			{
				if position == len(p.Buffer) {
					goto l367
				}
				switch p.Buffer[position] {
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l367
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l367
					}
					break
				default:
					goto l367
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l367
			}
			if !matchChar('>') {
				goto l367
			}
			return true
		l367:
			position = position0
			return false
		},
		/* 81 HtmlBlockNoscript <- (HtmlBlockOpenNoscript (HtmlBlockNoscript / (!HtmlBlockCloseNoscript .))* HtmlBlockCloseNoscript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenNoscript]() {
				goto l369
			}
		l370:
			{
				position371 := position
				if !p.rules[ruleHtmlBlockNoscript]() {
					goto l373
				}
				goto l372
			l373:
				if !p.rules[ruleHtmlBlockCloseNoscript]() {
					goto l374
				}
				goto l371
			l374:
				if !matchDot() {
					goto l371
				}
			l372:
				goto l370
			l371:
				position = position371
			}
			if !p.rules[ruleHtmlBlockCloseNoscript]() {
				goto l369
			}
			return true
		l369:
			position = position0
			return false
		},
		/* 82 HtmlBlockOpenOl <- ('<' Spnl ((&[O] 'OL') | (&[o] 'ol')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l375
			}
			if !p.rules[ruleSpnl]() {
				goto l375
			}
			{
				if position == len(p.Buffer) {
					goto l375
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l375
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l375
					}
					break
				default:
					goto l375
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l375
			}
		l377:
			if !p.rules[ruleHtmlAttribute]() {
				goto l378
			}
			goto l377
		l378:
			if !matchChar('>') {
				goto l375
			}
			return true
		l375:
			position = position0
			return false
		},
		/* 83 HtmlBlockCloseOl <- ('<' Spnl '/' ((&[O] 'OL') | (&[o] 'ol')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l379
			}
			if !p.rules[ruleSpnl]() {
				goto l379
			}
			if !matchChar('/') {
				goto l379
			}
			{
				if position == len(p.Buffer) {
					goto l379
				}
				switch p.Buffer[position] {
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l379
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l379
					}
					break
				default:
					goto l379
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l379
			}
			if !matchChar('>') {
				goto l379
			}
			return true
		l379:
			position = position0
			return false
		},
		/* 84 HtmlBlockOl <- (HtmlBlockOpenOl (HtmlBlockOl / (!HtmlBlockCloseOl .))* HtmlBlockCloseOl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenOl]() {
				goto l381
			}
		l382:
			{
				position383 := position
				if !p.rules[ruleHtmlBlockOl]() {
					goto l385
				}
				goto l384
			l385:
				if !p.rules[ruleHtmlBlockCloseOl]() {
					goto l386
				}
				goto l383
			l386:
				if !matchDot() {
					goto l383
				}
			l384:
				goto l382
			l383:
				position = position383
			}
			if !p.rules[ruleHtmlBlockCloseOl]() {
				goto l381
			}
			return true
		l381:
			position = position0
			return false
		},
		/* 85 HtmlBlockOpenP <- ('<' Spnl ((&[P] 'P') | (&[p] 'p')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l387
			}
			if !p.rules[ruleSpnl]() {
				goto l387
			}
			{
				if position == len(p.Buffer) {
					goto l387
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
					break
				case 'p':
					position++ // matchChar
					break
				default:
					goto l387
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l387
			}
		l389:
			if !p.rules[ruleHtmlAttribute]() {
				goto l390
			}
			goto l389
		l390:
			if !matchChar('>') {
				goto l387
			}
			return true
		l387:
			position = position0
			return false
		},
		/* 86 HtmlBlockCloseP <- ('<' Spnl '/' ((&[P] 'P') | (&[p] 'p')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l391
			}
			if !p.rules[ruleSpnl]() {
				goto l391
			}
			if !matchChar('/') {
				goto l391
			}
			{
				if position == len(p.Buffer) {
					goto l391
				}
				switch p.Buffer[position] {
				case 'P':
					position++ // matchChar
					break
				case 'p':
					position++ // matchChar
					break
				default:
					goto l391
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l391
			}
			if !matchChar('>') {
				goto l391
			}
			return true
		l391:
			position = position0
			return false
		},
		/* 87 HtmlBlockP <- (HtmlBlockOpenP (HtmlBlockP / (!HtmlBlockCloseP .))* HtmlBlockCloseP) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenP]() {
				goto l393
			}
		l394:
			{
				position395 := position
				if !p.rules[ruleHtmlBlockP]() {
					goto l397
				}
				goto l396
			l397:
				if !p.rules[ruleHtmlBlockCloseP]() {
					goto l398
				}
				goto l395
			l398:
				if !matchDot() {
					goto l395
				}
			l396:
				goto l394
			l395:
				position = position395
			}
			if !p.rules[ruleHtmlBlockCloseP]() {
				goto l393
			}
			return true
		l393:
			position = position0
			return false
		},
		/* 88 HtmlBlockOpenPre <- ('<' Spnl ((&[P] 'PRE') | (&[p] 'pre')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l399
			}
			if !p.rules[ruleSpnl]() {
				goto l399
			}
			{
				if position == len(p.Buffer) {
					goto l399
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto l399
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l399
					}
					break
				default:
					goto l399
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l399
			}
		l401:
			if !p.rules[ruleHtmlAttribute]() {
				goto l402
			}
			goto l401
		l402:
			if !matchChar('>') {
				goto l399
			}
			return true
		l399:
			position = position0
			return false
		},
		/* 89 HtmlBlockClosePre <- ('<' Spnl '/' ((&[P] 'PRE') | (&[p] 'pre')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l403
			}
			if !p.rules[ruleSpnl]() {
				goto l403
			}
			if !matchChar('/') {
				goto l403
			}
			{
				if position == len(p.Buffer) {
					goto l403
				}
				switch p.Buffer[position] {
				case 'P':
					position++
					if !matchString("RE") {
						goto l403
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l403
					}
					break
				default:
					goto l403
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l403
			}
			if !matchChar('>') {
				goto l403
			}
			return true
		l403:
			position = position0
			return false
		},
		/* 90 HtmlBlockPre <- (HtmlBlockOpenPre (HtmlBlockPre / (!HtmlBlockClosePre .))* HtmlBlockClosePre) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenPre]() {
				goto l405
			}
		l406:
			{
				position407 := position
				if !p.rules[ruleHtmlBlockPre]() {
					goto l409
				}
				goto l408
			l409:
				if !p.rules[ruleHtmlBlockClosePre]() {
					goto l410
				}
				goto l407
			l410:
				if !matchDot() {
					goto l407
				}
			l408:
				goto l406
			l407:
				position = position407
			}
			if !p.rules[ruleHtmlBlockClosePre]() {
				goto l405
			}
			return true
		l405:
			position = position0
			return false
		},
		/* 91 HtmlBlockOpenTable <- ('<' Spnl ((&[T] 'TABLE') | (&[t] 'table')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l411
			}
			if !p.rules[ruleSpnl]() {
				goto l411
			}
			{
				if position == len(p.Buffer) {
					goto l411
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto l411
					}
					break
				case 't':
					position++
					if !matchString("able") {
						goto l411
					}
					break
				default:
					goto l411
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l411
			}
		l413:
			if !p.rules[ruleHtmlAttribute]() {
				goto l414
			}
			goto l413
		l414:
			if !matchChar('>') {
				goto l411
			}
			return true
		l411:
			position = position0
			return false
		},
		/* 92 HtmlBlockCloseTable <- ('<' Spnl '/' ((&[T] 'TABLE') | (&[t] 'table')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l415
			}
			if !p.rules[ruleSpnl]() {
				goto l415
			}
			if !matchChar('/') {
				goto l415
			}
			{
				if position == len(p.Buffer) {
					goto l415
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("ABLE") {
						goto l415
					}
					break
				case 't':
					position++
					if !matchString("able") {
						goto l415
					}
					break
				default:
					goto l415
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l415
			}
			if !matchChar('>') {
				goto l415
			}
			return true
		l415:
			position = position0
			return false
		},
		/* 93 HtmlBlockTable <- (HtmlBlockOpenTable (HtmlBlockTable / (!HtmlBlockCloseTable .))* HtmlBlockCloseTable) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTable]() {
				goto l417
			}
		l418:
			{
				position419 := position
				if !p.rules[ruleHtmlBlockTable]() {
					goto l421
				}
				goto l420
			l421:
				if !p.rules[ruleHtmlBlockCloseTable]() {
					goto l422
				}
				goto l419
			l422:
				if !matchDot() {
					goto l419
				}
			l420:
				goto l418
			l419:
				position = position419
			}
			if !p.rules[ruleHtmlBlockCloseTable]() {
				goto l417
			}
			return true
		l417:
			position = position0
			return false
		},
		/* 94 HtmlBlockOpenUl <- ('<' Spnl ((&[U] 'UL') | (&[u] 'ul')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l423
			}
			if !p.rules[ruleSpnl]() {
				goto l423
			}
			{
				if position == len(p.Buffer) {
					goto l423
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l423
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l423
					}
					break
				default:
					goto l423
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l423
			}
		l425:
			if !p.rules[ruleHtmlAttribute]() {
				goto l426
			}
			goto l425
		l426:
			if !matchChar('>') {
				goto l423
			}
			return true
		l423:
			position = position0
			return false
		},
		/* 95 HtmlBlockCloseUl <- ('<' Spnl '/' ((&[U] 'UL') | (&[u] 'ul')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l427
			}
			if !p.rules[ruleSpnl]() {
				goto l427
			}
			if !matchChar('/') {
				goto l427
			}
			{
				if position == len(p.Buffer) {
					goto l427
				}
				switch p.Buffer[position] {
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l427
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l427
					}
					break
				default:
					goto l427
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l427
			}
			if !matchChar('>') {
				goto l427
			}
			return true
		l427:
			position = position0
			return false
		},
		/* 96 HtmlBlockUl <- (HtmlBlockOpenUl (HtmlBlockUl / (!HtmlBlockCloseUl .))* HtmlBlockCloseUl) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenUl]() {
				goto l429
			}
		l430:
			{
				position431 := position
				if !p.rules[ruleHtmlBlockUl]() {
					goto l433
				}
				goto l432
			l433:
				if !p.rules[ruleHtmlBlockCloseUl]() {
					goto l434
				}
				goto l431
			l434:
				if !matchDot() {
					goto l431
				}
			l432:
				goto l430
			l431:
				position = position431
			}
			if !p.rules[ruleHtmlBlockCloseUl]() {
				goto l429
			}
			return true
		l429:
			position = position0
			return false
		},
		/* 97 HtmlBlockOpenDd <- ('<' Spnl ((&[D] 'DD') | (&[d] 'dd')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l435
			}
			if !p.rules[ruleSpnl]() {
				goto l435
			}
			{
				if position == len(p.Buffer) {
					goto l435
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto l435
					}
					break
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto l435
					}
					break
				default:
					goto l435
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l435
			}
		l437:
			if !p.rules[ruleHtmlAttribute]() {
				goto l438
			}
			goto l437
		l438:
			if !matchChar('>') {
				goto l435
			}
			return true
		l435:
			position = position0
			return false
		},
		/* 98 HtmlBlockCloseDd <- ('<' Spnl '/' ((&[D] 'DD') | (&[d] 'dd')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l439
			}
			if !p.rules[ruleSpnl]() {
				goto l439
			}
			if !matchChar('/') {
				goto l439
			}
			{
				if position == len(p.Buffer) {
					goto l439
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DD`)
					if !matchChar('D') {
						goto l439
					}
					break
				case 'd':
					position++ // matchString(`dd`)
					if !matchChar('d') {
						goto l439
					}
					break
				default:
					goto l439
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l439
			}
			if !matchChar('>') {
				goto l439
			}
			return true
		l439:
			position = position0
			return false
		},
		/* 99 HtmlBlockDd <- (HtmlBlockOpenDd (HtmlBlockDd / (!HtmlBlockCloseDd .))* HtmlBlockCloseDd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDd]() {
				goto l441
			}
		l442:
			{
				position443 := position
				if !p.rules[ruleHtmlBlockDd]() {
					goto l445
				}
				goto l444
			l445:
				if !p.rules[ruleHtmlBlockCloseDd]() {
					goto l446
				}
				goto l443
			l446:
				if !matchDot() {
					goto l443
				}
			l444:
				goto l442
			l443:
				position = position443
			}
			if !p.rules[ruleHtmlBlockCloseDd]() {
				goto l441
			}
			return true
		l441:
			position = position0
			return false
		},
		/* 100 HtmlBlockOpenDt <- ('<' Spnl ((&[D] 'DT') | (&[d] 'dt')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l447
			}
			if !p.rules[ruleSpnl]() {
				goto l447
			}
			{
				if position == len(p.Buffer) {
					goto l447
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l447
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l447
					}
					break
				default:
					goto l447
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l447
			}
		l449:
			if !p.rules[ruleHtmlAttribute]() {
				goto l450
			}
			goto l449
		l450:
			if !matchChar('>') {
				goto l447
			}
			return true
		l447:
			position = position0
			return false
		},
		/* 101 HtmlBlockCloseDt <- ('<' Spnl '/' ((&[D] 'DT') | (&[d] 'dt')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l451
			}
			if !p.rules[ruleSpnl]() {
				goto l451
			}
			if !matchChar('/') {
				goto l451
			}
			{
				if position == len(p.Buffer) {
					goto l451
				}
				switch p.Buffer[position] {
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l451
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l451
					}
					break
				default:
					goto l451
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l451
			}
			if !matchChar('>') {
				goto l451
			}
			return true
		l451:
			position = position0
			return false
		},
		/* 102 HtmlBlockDt <- (HtmlBlockOpenDt (HtmlBlockDt / (!HtmlBlockCloseDt .))* HtmlBlockCloseDt) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenDt]() {
				goto l453
			}
		l454:
			{
				position455 := position
				if !p.rules[ruleHtmlBlockDt]() {
					goto l457
				}
				goto l456
			l457:
				if !p.rules[ruleHtmlBlockCloseDt]() {
					goto l458
				}
				goto l455
			l458:
				if !matchDot() {
					goto l455
				}
			l456:
				goto l454
			l455:
				position = position455
			}
			if !p.rules[ruleHtmlBlockCloseDt]() {
				goto l453
			}
			return true
		l453:
			position = position0
			return false
		},
		/* 103 HtmlBlockOpenFrameset <- ('<' Spnl ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l459
			}
			if !p.rules[ruleSpnl]() {
				goto l459
			}
			{
				if position == len(p.Buffer) {
					goto l459
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l459
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l459
					}
					break
				default:
					goto l459
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l459
			}
		l461:
			if !p.rules[ruleHtmlAttribute]() {
				goto l462
			}
			goto l461
		l462:
			if !matchChar('>') {
				goto l459
			}
			return true
		l459:
			position = position0
			return false
		},
		/* 104 HtmlBlockCloseFrameset <- ('<' Spnl '/' ((&[F] 'FRAMESET') | (&[f] 'frameset')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l463
			}
			if !p.rules[ruleSpnl]() {
				goto l463
			}
			if !matchChar('/') {
				goto l463
			}
			{
				if position == len(p.Buffer) {
					goto l463
				}
				switch p.Buffer[position] {
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l463
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l463
					}
					break
				default:
					goto l463
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l463
			}
			if !matchChar('>') {
				goto l463
			}
			return true
		l463:
			position = position0
			return false
		},
		/* 105 HtmlBlockFrameset <- (HtmlBlockOpenFrameset (HtmlBlockFrameset / (!HtmlBlockCloseFrameset .))* HtmlBlockCloseFrameset) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenFrameset]() {
				goto l465
			}
		l466:
			{
				position467 := position
				if !p.rules[ruleHtmlBlockFrameset]() {
					goto l469
				}
				goto l468
			l469:
				if !p.rules[ruleHtmlBlockCloseFrameset]() {
					goto l470
				}
				goto l467
			l470:
				if !matchDot() {
					goto l467
				}
			l468:
				goto l466
			l467:
				position = position467
			}
			if !p.rules[ruleHtmlBlockCloseFrameset]() {
				goto l465
			}
			return true
		l465:
			position = position0
			return false
		},
		/* 106 HtmlBlockOpenLi <- ('<' Spnl ((&[L] 'LI') | (&[l] 'li')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l471
			}
			if !p.rules[ruleSpnl]() {
				goto l471
			}
			{
				if position == len(p.Buffer) {
					goto l471
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l471
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l471
					}
					break
				default:
					goto l471
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l471
			}
		l473:
			if !p.rules[ruleHtmlAttribute]() {
				goto l474
			}
			goto l473
		l474:
			if !matchChar('>') {
				goto l471
			}
			return true
		l471:
			position = position0
			return false
		},
		/* 107 HtmlBlockCloseLi <- ('<' Spnl '/' ((&[L] 'LI') | (&[l] 'li')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l475
			}
			if !p.rules[ruleSpnl]() {
				goto l475
			}
			if !matchChar('/') {
				goto l475
			}
			{
				if position == len(p.Buffer) {
					goto l475
				}
				switch p.Buffer[position] {
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l475
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l475
					}
					break
				default:
					goto l475
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l475
			}
			if !matchChar('>') {
				goto l475
			}
			return true
		l475:
			position = position0
			return false
		},
		/* 108 HtmlBlockLi <- (HtmlBlockOpenLi (HtmlBlockLi / (!HtmlBlockCloseLi .))* HtmlBlockCloseLi) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenLi]() {
				goto l477
			}
		l478:
			{
				position479 := position
				if !p.rules[ruleHtmlBlockLi]() {
					goto l481
				}
				goto l480
			l481:
				if !p.rules[ruleHtmlBlockCloseLi]() {
					goto l482
				}
				goto l479
			l482:
				if !matchDot() {
					goto l479
				}
			l480:
				goto l478
			l479:
				position = position479
			}
			if !p.rules[ruleHtmlBlockCloseLi]() {
				goto l477
			}
			return true
		l477:
			position = position0
			return false
		},
		/* 109 HtmlBlockOpenTbody <- ('<' Spnl ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l483
			}
			if !p.rules[ruleSpnl]() {
				goto l483
			}
			{
				if position == len(p.Buffer) {
					goto l483
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto l483
					}
					break
				case 't':
					position++
					if !matchString("body") {
						goto l483
					}
					break
				default:
					goto l483
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l483
			}
		l485:
			if !p.rules[ruleHtmlAttribute]() {
				goto l486
			}
			goto l485
		l486:
			if !matchChar('>') {
				goto l483
			}
			return true
		l483:
			position = position0
			return false
		},
		/* 110 HtmlBlockCloseTbody <- ('<' Spnl '/' ((&[T] 'TBODY') | (&[t] 'tbody')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l487
			}
			if !p.rules[ruleSpnl]() {
				goto l487
			}
			if !matchChar('/') {
				goto l487
			}
			{
				if position == len(p.Buffer) {
					goto l487
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("BODY") {
						goto l487
					}
					break
				case 't':
					position++
					if !matchString("body") {
						goto l487
					}
					break
				default:
					goto l487
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l487
			}
			if !matchChar('>') {
				goto l487
			}
			return true
		l487:
			position = position0
			return false
		},
		/* 111 HtmlBlockTbody <- (HtmlBlockOpenTbody (HtmlBlockTbody / (!HtmlBlockCloseTbody .))* HtmlBlockCloseTbody) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTbody]() {
				goto l489
			}
		l490:
			{
				position491 := position
				if !p.rules[ruleHtmlBlockTbody]() {
					goto l493
				}
				goto l492
			l493:
				if !p.rules[ruleHtmlBlockCloseTbody]() {
					goto l494
				}
				goto l491
			l494:
				if !matchDot() {
					goto l491
				}
			l492:
				goto l490
			l491:
				position = position491
			}
			if !p.rules[ruleHtmlBlockCloseTbody]() {
				goto l489
			}
			return true
		l489:
			position = position0
			return false
		},
		/* 112 HtmlBlockOpenTd <- ('<' Spnl ((&[T] 'TD') | (&[t] 'td')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l495
			}
			if !p.rules[ruleSpnl]() {
				goto l495
			}
			{
				if position == len(p.Buffer) {
					goto l495
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto l495
					}
					break
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto l495
					}
					break
				default:
					goto l495
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l495
			}
		l497:
			if !p.rules[ruleHtmlAttribute]() {
				goto l498
			}
			goto l497
		l498:
			if !matchChar('>') {
				goto l495
			}
			return true
		l495:
			position = position0
			return false
		},
		/* 113 HtmlBlockCloseTd <- ('<' Spnl '/' ((&[T] 'TD') | (&[t] 'td')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l499
			}
			if !p.rules[ruleSpnl]() {
				goto l499
			}
			if !matchChar('/') {
				goto l499
			}
			{
				if position == len(p.Buffer) {
					goto l499
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TD`)
					if !matchChar('D') {
						goto l499
					}
					break
				case 't':
					position++ // matchString(`td`)
					if !matchChar('d') {
						goto l499
					}
					break
				default:
					goto l499
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l499
			}
			if !matchChar('>') {
				goto l499
			}
			return true
		l499:
			position = position0
			return false
		},
		/* 114 HtmlBlockTd <- (HtmlBlockOpenTd (HtmlBlockTd / (!HtmlBlockCloseTd .))* HtmlBlockCloseTd) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTd]() {
				goto l501
			}
		l502:
			{
				position503 := position
				if !p.rules[ruleHtmlBlockTd]() {
					goto l505
				}
				goto l504
			l505:
				if !p.rules[ruleHtmlBlockCloseTd]() {
					goto l506
				}
				goto l503
			l506:
				if !matchDot() {
					goto l503
				}
			l504:
				goto l502
			l503:
				position = position503
			}
			if !p.rules[ruleHtmlBlockCloseTd]() {
				goto l501
			}
			return true
		l501:
			position = position0
			return false
		},
		/* 115 HtmlBlockOpenTfoot <- ('<' Spnl ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l507
			}
			if !p.rules[ruleSpnl]() {
				goto l507
			}
			{
				if position == len(p.Buffer) {
					goto l507
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto l507
					}
					break
				case 't':
					position++
					if !matchString("foot") {
						goto l507
					}
					break
				default:
					goto l507
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l507
			}
		l509:
			if !p.rules[ruleHtmlAttribute]() {
				goto l510
			}
			goto l509
		l510:
			if !matchChar('>') {
				goto l507
			}
			return true
		l507:
			position = position0
			return false
		},
		/* 116 HtmlBlockCloseTfoot <- ('<' Spnl '/' ((&[T] 'TFOOT') | (&[t] 'tfoot')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l511
			}
			if !p.rules[ruleSpnl]() {
				goto l511
			}
			if !matchChar('/') {
				goto l511
			}
			{
				if position == len(p.Buffer) {
					goto l511
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("FOOT") {
						goto l511
					}
					break
				case 't':
					position++
					if !matchString("foot") {
						goto l511
					}
					break
				default:
					goto l511
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l511
			}
			if !matchChar('>') {
				goto l511
			}
			return true
		l511:
			position = position0
			return false
		},
		/* 117 HtmlBlockTfoot <- (HtmlBlockOpenTfoot (HtmlBlockTfoot / (!HtmlBlockCloseTfoot .))* HtmlBlockCloseTfoot) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTfoot]() {
				goto l513
			}
		l514:
			{
				position515 := position
				if !p.rules[ruleHtmlBlockTfoot]() {
					goto l517
				}
				goto l516
			l517:
				if !p.rules[ruleHtmlBlockCloseTfoot]() {
					goto l518
				}
				goto l515
			l518:
				if !matchDot() {
					goto l515
				}
			l516:
				goto l514
			l515:
				position = position515
			}
			if !p.rules[ruleHtmlBlockCloseTfoot]() {
				goto l513
			}
			return true
		l513:
			position = position0
			return false
		},
		/* 118 HtmlBlockOpenTh <- ('<' Spnl ((&[T] 'TH') | (&[t] 'th')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l519
			}
			if !p.rules[ruleSpnl]() {
				goto l519
			}
			{
				if position == len(p.Buffer) {
					goto l519
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto l519
					}
					break
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto l519
					}
					break
				default:
					goto l519
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l519
			}
		l521:
			if !p.rules[ruleHtmlAttribute]() {
				goto l522
			}
			goto l521
		l522:
			if !matchChar('>') {
				goto l519
			}
			return true
		l519:
			position = position0
			return false
		},
		/* 119 HtmlBlockCloseTh <- ('<' Spnl '/' ((&[T] 'TH') | (&[t] 'th')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l523
			}
			if !p.rules[ruleSpnl]() {
				goto l523
			}
			if !matchChar('/') {
				goto l523
			}
			{
				if position == len(p.Buffer) {
					goto l523
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TH`)
					if !matchChar('H') {
						goto l523
					}
					break
				case 't':
					position++ // matchString(`th`)
					if !matchChar('h') {
						goto l523
					}
					break
				default:
					goto l523
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l523
			}
			if !matchChar('>') {
				goto l523
			}
			return true
		l523:
			position = position0
			return false
		},
		/* 120 HtmlBlockTh <- (HtmlBlockOpenTh (HtmlBlockTh / (!HtmlBlockCloseTh .))* HtmlBlockCloseTh) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTh]() {
				goto l525
			}
		l526:
			{
				position527 := position
				if !p.rules[ruleHtmlBlockTh]() {
					goto l529
				}
				goto l528
			l529:
				if !p.rules[ruleHtmlBlockCloseTh]() {
					goto l530
				}
				goto l527
			l530:
				if !matchDot() {
					goto l527
				}
			l528:
				goto l526
			l527:
				position = position527
			}
			if !p.rules[ruleHtmlBlockCloseTh]() {
				goto l525
			}
			return true
		l525:
			position = position0
			return false
		},
		/* 121 HtmlBlockOpenThead <- ('<' Spnl ((&[T] 'THEAD') | (&[t] 'thead')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l531
			}
			if !p.rules[ruleSpnl]() {
				goto l531
			}
			{
				if position == len(p.Buffer) {
					goto l531
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto l531
					}
					break
				case 't':
					position++
					if !matchString("head") {
						goto l531
					}
					break
				default:
					goto l531
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l531
			}
		l533:
			if !p.rules[ruleHtmlAttribute]() {
				goto l534
			}
			goto l533
		l534:
			if !matchChar('>') {
				goto l531
			}
			return true
		l531:
			position = position0
			return false
		},
		/* 122 HtmlBlockCloseThead <- ('<' Spnl '/' ((&[T] 'THEAD') | (&[t] 'thead')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l535
			}
			if !p.rules[ruleSpnl]() {
				goto l535
			}
			if !matchChar('/') {
				goto l535
			}
			{
				if position == len(p.Buffer) {
					goto l535
				}
				switch p.Buffer[position] {
				case 'T':
					position++
					if !matchString("HEAD") {
						goto l535
					}
					break
				case 't':
					position++
					if !matchString("head") {
						goto l535
					}
					break
				default:
					goto l535
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l535
			}
			if !matchChar('>') {
				goto l535
			}
			return true
		l535:
			position = position0
			return false
		},
		/* 123 HtmlBlockThead <- (HtmlBlockOpenThead (HtmlBlockThead / (!HtmlBlockCloseThead .))* HtmlBlockCloseThead) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenThead]() {
				goto l537
			}
		l538:
			{
				position539 := position
				if !p.rules[ruleHtmlBlockThead]() {
					goto l541
				}
				goto l540
			l541:
				if !p.rules[ruleHtmlBlockCloseThead]() {
					goto l542
				}
				goto l539
			l542:
				if !matchDot() {
					goto l539
				}
			l540:
				goto l538
			l539:
				position = position539
			}
			if !p.rules[ruleHtmlBlockCloseThead]() {
				goto l537
			}
			return true
		l537:
			position = position0
			return false
		},
		/* 124 HtmlBlockOpenTr <- ('<' Spnl ((&[T] 'TR') | (&[t] 'tr')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l543
			}
			if !p.rules[ruleSpnl]() {
				goto l543
			}
			{
				if position == len(p.Buffer) {
					goto l543
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l543
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l543
					}
					break
				default:
					goto l543
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l543
			}
		l545:
			if !p.rules[ruleHtmlAttribute]() {
				goto l546
			}
			goto l545
		l546:
			if !matchChar('>') {
				goto l543
			}
			return true
		l543:
			position = position0
			return false
		},
		/* 125 HtmlBlockCloseTr <- ('<' Spnl '/' ((&[T] 'TR') | (&[t] 'tr')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l547
			}
			if !p.rules[ruleSpnl]() {
				goto l547
			}
			if !matchChar('/') {
				goto l547
			}
			{
				if position == len(p.Buffer) {
					goto l547
				}
				switch p.Buffer[position] {
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l547
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l547
					}
					break
				default:
					goto l547
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l547
			}
			if !matchChar('>') {
				goto l547
			}
			return true
		l547:
			position = position0
			return false
		},
		/* 126 HtmlBlockTr <- (HtmlBlockOpenTr (HtmlBlockTr / (!HtmlBlockCloseTr .))* HtmlBlockCloseTr) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenTr]() {
				goto l549
			}
		l550:
			{
				position551 := position
				if !p.rules[ruleHtmlBlockTr]() {
					goto l553
				}
				goto l552
			l553:
				if !p.rules[ruleHtmlBlockCloseTr]() {
					goto l554
				}
				goto l551
			l554:
				if !matchDot() {
					goto l551
				}
			l552:
				goto l550
			l551:
				position = position551
			}
			if !p.rules[ruleHtmlBlockCloseTr]() {
				goto l549
			}
			return true
		l549:
			position = position0
			return false
		},
		/* 127 HtmlBlockOpenScript <- ('<' Spnl ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l555
			}
			if !p.rules[ruleSpnl]() {
				goto l555
			}
			{
				if position == len(p.Buffer) {
					goto l555
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l555
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l555
					}
					break
				default:
					goto l555
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l555
			}
		l557:
			if !p.rules[ruleHtmlAttribute]() {
				goto l558
			}
			goto l557
		l558:
			if !matchChar('>') {
				goto l555
			}
			return true
		l555:
			position = position0
			return false
		},
		/* 128 HtmlBlockCloseScript <- ('<' Spnl '/' ((&[S] 'SCRIPT') | (&[s] 'script')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l559
			}
			if !p.rules[ruleSpnl]() {
				goto l559
			}
			if !matchChar('/') {
				goto l559
			}
			{
				if position == len(p.Buffer) {
					goto l559
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l559
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l559
					}
					break
				default:
					goto l559
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l559
			}
			if !matchChar('>') {
				goto l559
			}
			return true
		l559:
			position = position0
			return false
		},
		/* 129 HtmlBlockScript <- (HtmlBlockOpenScript (!HtmlBlockCloseScript .)* HtmlBlockCloseScript) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenScript]() {
				goto l561
			}
		l562:
			{
				position563 := position
				if !p.rules[ruleHtmlBlockCloseScript]() {
					goto l564
				}
				goto l563
			l564:
				if !matchDot() {
					goto l563
				}
				goto l562
			l563:
				position = position563
			}
			if !p.rules[ruleHtmlBlockCloseScript]() {
				goto l561
			}
			return true
		l561:
			position = position0
			return false
		},
		/* 130 HtmlBlockOpenHead <- ('<' Spnl ((&[H] 'HEAD') | (&[h] 'head')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l565
			}
			if !p.rules[ruleSpnl]() {
				goto l565
			}
			{
				if position == len(p.Buffer) {
					goto l565
				}
				switch p.Buffer[position] {
				case 'H':
					position++
					if !matchString("EAD") {
						goto l565
					}
					break
				case 'h':
					position++
					if !matchString("ead") {
						goto l565
					}
					break
				default:
					goto l565
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l565
			}
		l567:
			if !p.rules[ruleHtmlAttribute]() {
				goto l568
			}
			goto l567
		l568:
			if !matchChar('>') {
				goto l565
			}
			return true
		l565:
			position = position0
			return false
		},
		/* 131 HtmlBlockCloseHead <- ('<' Spnl '/' ((&[H] 'HEAD') | (&[h] 'head')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l569
			}
			if !p.rules[ruleSpnl]() {
				goto l569
			}
			if !matchChar('/') {
				goto l569
			}
			{
				if position == len(p.Buffer) {
					goto l569
				}
				switch p.Buffer[position] {
				case 'H':
					position++
					if !matchString("EAD") {
						goto l569
					}
					break
				case 'h':
					position++
					if !matchString("ead") {
						goto l569
					}
					break
				default:
					goto l569
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l569
			}
			if !matchChar('>') {
				goto l569
			}
			return true
		l569:
			position = position0
			return false
		},
		/* 132 HtmlBlockHead <- (HtmlBlockOpenHead (!HtmlBlockCloseHead .)* HtmlBlockCloseHead) */
		func() bool {
			position0 := position
			if !p.rules[ruleHtmlBlockOpenHead]() {
				goto l571
			}
		l572:
			{
				position573 := position
				if !p.rules[ruleHtmlBlockCloseHead]() {
					goto l574
				}
				goto l573
			l574:
				if !matchDot() {
					goto l573
				}
				goto l572
			l573:
				position = position573
			}
			if !p.rules[ruleHtmlBlockCloseHead]() {
				goto l571
			}
			return true
		l571:
			position = position0
			return false
		},
		/* 133 HtmlBlockInTags <- (HtmlBlockAddress / HtmlBlockBlockquote / HtmlBlockCenter / HtmlBlockDir / HtmlBlockDiv / HtmlBlockDl / HtmlBlockFieldset / HtmlBlockForm / HtmlBlockH1 / HtmlBlockH2 / HtmlBlockH3 / HtmlBlockH4 / HtmlBlockH5 / HtmlBlockH6 / HtmlBlockMenu / HtmlBlockNoframes / HtmlBlockNoscript / HtmlBlockOl / HtmlBlockP / HtmlBlockPre / HtmlBlockTable / HtmlBlockUl / HtmlBlockDd / HtmlBlockDt / HtmlBlockFrameset / HtmlBlockLi / HtmlBlockTbody / HtmlBlockTd / HtmlBlockTfoot / HtmlBlockTh / HtmlBlockThead / HtmlBlockTr / HtmlBlockScript / HtmlBlockHead) */
		func() bool {
			if !p.rules[ruleHtmlBlockAddress]() {
				goto l577
			}
			goto l576
		l577:
			if !p.rules[ruleHtmlBlockBlockquote]() {
				goto l578
			}
			goto l576
		l578:
			if !p.rules[ruleHtmlBlockCenter]() {
				goto l579
			}
			goto l576
		l579:
			if !p.rules[ruleHtmlBlockDir]() {
				goto l580
			}
			goto l576
		l580:
			if !p.rules[ruleHtmlBlockDiv]() {
				goto l581
			}
			goto l576
		l581:
			if !p.rules[ruleHtmlBlockDl]() {
				goto l582
			}
			goto l576
		l582:
			if !p.rules[ruleHtmlBlockFieldset]() {
				goto l583
			}
			goto l576
		l583:
			if !p.rules[ruleHtmlBlockForm]() {
				goto l584
			}
			goto l576
		l584:
			if !p.rules[ruleHtmlBlockH1]() {
				goto l585
			}
			goto l576
		l585:
			if !p.rules[ruleHtmlBlockH2]() {
				goto l586
			}
			goto l576
		l586:
			if !p.rules[ruleHtmlBlockH3]() {
				goto l587
			}
			goto l576
		l587:
			if !p.rules[ruleHtmlBlockH4]() {
				goto l588
			}
			goto l576
		l588:
			if !p.rules[ruleHtmlBlockH5]() {
				goto l589
			}
			goto l576
		l589:
			if !p.rules[ruleHtmlBlockH6]() {
				goto l590
			}
			goto l576
		l590:
			if !p.rules[ruleHtmlBlockMenu]() {
				goto l591
			}
			goto l576
		l591:
			if !p.rules[ruleHtmlBlockNoframes]() {
				goto l592
			}
			goto l576
		l592:
			if !p.rules[ruleHtmlBlockNoscript]() {
				goto l593
			}
			goto l576
		l593:
			if !p.rules[ruleHtmlBlockOl]() {
				goto l594
			}
			goto l576
		l594:
			if !p.rules[ruleHtmlBlockP]() {
				goto l595
			}
			goto l576
		l595:
			if !p.rules[ruleHtmlBlockPre]() {
				goto l596
			}
			goto l576
		l596:
			if !p.rules[ruleHtmlBlockTable]() {
				goto l597
			}
			goto l576
		l597:
			if !p.rules[ruleHtmlBlockUl]() {
				goto l598
			}
			goto l576
		l598:
			if !p.rules[ruleHtmlBlockDd]() {
				goto l599
			}
			goto l576
		l599:
			if !p.rules[ruleHtmlBlockDt]() {
				goto l600
			}
			goto l576
		l600:
			if !p.rules[ruleHtmlBlockFrameset]() {
				goto l601
			}
			goto l576
		l601:
			if !p.rules[ruleHtmlBlockLi]() {
				goto l602
			}
			goto l576
		l602:
			if !p.rules[ruleHtmlBlockTbody]() {
				goto l603
			}
			goto l576
		l603:
			if !p.rules[ruleHtmlBlockTd]() {
				goto l604
			}
			goto l576
		l604:
			if !p.rules[ruleHtmlBlockTfoot]() {
				goto l605
			}
			goto l576
		l605:
			if !p.rules[ruleHtmlBlockTh]() {
				goto l606
			}
			goto l576
		l606:
			if !p.rules[ruleHtmlBlockThead]() {
				goto l607
			}
			goto l576
		l607:
			if !p.rules[ruleHtmlBlockTr]() {
				goto l608
			}
			goto l576
		l608:
			if !p.rules[ruleHtmlBlockScript]() {
				goto l609
			}
			goto l576
		l609:
			if !p.rules[ruleHtmlBlockHead]() {
				goto l575
			}
		l576:
			return true
		l575:
			return false
		},
		/* 134 HtmlBlock <- (&'<' < (HtmlBlockInTags / HtmlComment / HtmlBlockSelfClosing) > BlankLine+ {   if p.extension.FilterHTML {
                    yy = p.mkList(LIST, nil)
                } else {
                    yy = p.mkString(yytext)
                    yy.key = HTMLBLOCK
                }
            }) */
		func() bool {
			position0 := position
			if !peekChar('<') {
				goto l610
			}
			begin = position
			if !p.rules[ruleHtmlBlockInTags]() {
				goto l612
			}
			goto l611
		l612:
			if !p.rules[ruleHtmlComment]() {
				goto l613
			}
			goto l611
		l613:
			if !p.rules[ruleHtmlBlockSelfClosing]() {
				goto l610
			}
		l611:
			end = position
			if !p.rules[ruleBlankLine]() {
				goto l610
			}
		l614:
			if !p.rules[ruleBlankLine]() {
				goto l615
			}
			goto l614
		l615:
			do(41)
			return true
		l610:
			position = position0
			return false
		},
		/* 135 HtmlBlockSelfClosing <- ('<' Spnl HtmlBlockType Spnl HtmlAttribute* '/' Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l616
			}
			if !p.rules[ruleSpnl]() {
				goto l616
			}
			if !p.rules[ruleHtmlBlockType]() {
				goto l616
			}
			if !p.rules[ruleSpnl]() {
				goto l616
			}
		l617:
			if !p.rules[ruleHtmlAttribute]() {
				goto l618
			}
			goto l617
		l618:
			if !matchChar('/') {
				goto l616
			}
			if !p.rules[ruleSpnl]() {
				goto l616
			}
			if !matchChar('>') {
				goto l616
			}
			return true
		l616:
			position = position0
			return false
		},
		/* 136 HtmlBlockType <- ('dir' / 'div' / 'dl' / 'fieldset' / 'form' / 'h1' / 'h2' / 'h3' / 'h4' / 'h5' / 'h6' / 'noframes' / 'p' / 'table' / 'dd' / 'tbody' / 'td' / 'tfoot' / 'th' / 'thead' / 'DIR' / 'DIV' / 'DL' / 'FIELDSET' / 'FORM' / 'H1' / 'H2' / 'H3' / 'H4' / 'H5' / 'H6' / 'NOFRAMES' / 'P' / 'TABLE' / 'DD' / 'TBODY' / 'TD' / 'TFOOT' / 'TH' / 'THEAD' / ((&[S] 'SCRIPT') | (&[T] 'TR') | (&[L] 'LI') | (&[F] 'FRAMESET') | (&[D] 'DT') | (&[U] 'UL') | (&[P] 'PRE') | (&[O] 'OL') | (&[N] 'NOSCRIPT') | (&[M] 'MENU') | (&[I] 'ISINDEX') | (&[H] 'HR') | (&[C] 'CENTER') | (&[B] 'BLOCKQUOTE') | (&[A] 'ADDRESS') | (&[s] 'script') | (&[t] 'tr') | (&[l] 'li') | (&[f] 'frameset') | (&[d] 'dt') | (&[u] 'ul') | (&[p] 'pre') | (&[o] 'ol') | (&[n] 'noscript') | (&[m] 'menu') | (&[i] 'isindex') | (&[h] 'hr') | (&[c] 'center') | (&[b] 'blockquote') | (&[a] 'address'))) */
		func() bool {
			if !matchString("dir") {
				goto l621
			}
			goto l620
		l621:
			if !matchString("div") {
				goto l622
			}
			goto l620
		l622:
			if !matchString("dl") {
				goto l623
			}
			goto l620
		l623:
			if !matchString("fieldset") {
				goto l624
			}
			goto l620
		l624:
			if !matchString("form") {
				goto l625
			}
			goto l620
		l625:
			if !matchString("h1") {
				goto l626
			}
			goto l620
		l626:
			if !matchString("h2") {
				goto l627
			}
			goto l620
		l627:
			if !matchString("h3") {
				goto l628
			}
			goto l620
		l628:
			if !matchString("h4") {
				goto l629
			}
			goto l620
		l629:
			if !matchString("h5") {
				goto l630
			}
			goto l620
		l630:
			if !matchString("h6") {
				goto l631
			}
			goto l620
		l631:
			if !matchString("noframes") {
				goto l632
			}
			goto l620
		l632:
			if !matchChar('p') {
				goto l633
			}
			goto l620
		l633:
			if !matchString("table") {
				goto l634
			}
			goto l620
		l634:
			if !matchString("dd") {
				goto l635
			}
			goto l620
		l635:
			if !matchString("tbody") {
				goto l636
			}
			goto l620
		l636:
			if !matchString("td") {
				goto l637
			}
			goto l620
		l637:
			if !matchString("tfoot") {
				goto l638
			}
			goto l620
		l638:
			if !matchString("th") {
				goto l639
			}
			goto l620
		l639:
			if !matchString("thead") {
				goto l640
			}
			goto l620
		l640:
			if !matchString("DIR") {
				goto l641
			}
			goto l620
		l641:
			if !matchString("DIV") {
				goto l642
			}
			goto l620
		l642:
			if !matchString("DL") {
				goto l643
			}
			goto l620
		l643:
			if !matchString("FIELDSET") {
				goto l644
			}
			goto l620
		l644:
			if !matchString("FORM") {
				goto l645
			}
			goto l620
		l645:
			if !matchString("H1") {
				goto l646
			}
			goto l620
		l646:
			if !matchString("H2") {
				goto l647
			}
			goto l620
		l647:
			if !matchString("H3") {
				goto l648
			}
			goto l620
		l648:
			if !matchString("H4") {
				goto l649
			}
			goto l620
		l649:
			if !matchString("H5") {
				goto l650
			}
			goto l620
		l650:
			if !matchString("H6") {
				goto l651
			}
			goto l620
		l651:
			if !matchString("NOFRAMES") {
				goto l652
			}
			goto l620
		l652:
			if !matchChar('P') {
				goto l653
			}
			goto l620
		l653:
			if !matchString("TABLE") {
				goto l654
			}
			goto l620
		l654:
			if !matchString("DD") {
				goto l655
			}
			goto l620
		l655:
			if !matchString("TBODY") {
				goto l656
			}
			goto l620
		l656:
			if !matchString("TD") {
				goto l657
			}
			goto l620
		l657:
			if !matchString("TFOOT") {
				goto l658
			}
			goto l620
		l658:
			if !matchString("TH") {
				goto l659
			}
			goto l620
		l659:
			if !matchString("THEAD") {
				goto l660
			}
			goto l620
		l660:
			{
				if position == len(p.Buffer) {
					goto l619
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("CRIPT") {
						goto l619
					}
					break
				case 'T':
					position++ // matchString(`TR`)
					if !matchChar('R') {
						goto l619
					}
					break
				case 'L':
					position++ // matchString(`LI`)
					if !matchChar('I') {
						goto l619
					}
					break
				case 'F':
					position++
					if !matchString("RAMESET") {
						goto l619
					}
					break
				case 'D':
					position++ // matchString(`DT`)
					if !matchChar('T') {
						goto l619
					}
					break
				case 'U':
					position++ // matchString(`UL`)
					if !matchChar('L') {
						goto l619
					}
					break
				case 'P':
					position++
					if !matchString("RE") {
						goto l619
					}
					break
				case 'O':
					position++ // matchString(`OL`)
					if !matchChar('L') {
						goto l619
					}
					break
				case 'N':
					position++
					if !matchString("OSCRIPT") {
						goto l619
					}
					break
				case 'M':
					position++
					if !matchString("ENU") {
						goto l619
					}
					break
				case 'I':
					position++
					if !matchString("SINDEX") {
						goto l619
					}
					break
				case 'H':
					position++ // matchString(`HR`)
					if !matchChar('R') {
						goto l619
					}
					break
				case 'C':
					position++
					if !matchString("ENTER") {
						goto l619
					}
					break
				case 'B':
					position++
					if !matchString("LOCKQUOTE") {
						goto l619
					}
					break
				case 'A':
					position++
					if !matchString("DDRESS") {
						goto l619
					}
					break
				case 's':
					position++
					if !matchString("cript") {
						goto l619
					}
					break
				case 't':
					position++ // matchString(`tr`)
					if !matchChar('r') {
						goto l619
					}
					break
				case 'l':
					position++ // matchString(`li`)
					if !matchChar('i') {
						goto l619
					}
					break
				case 'f':
					position++
					if !matchString("rameset") {
						goto l619
					}
					break
				case 'd':
					position++ // matchString(`dt`)
					if !matchChar('t') {
						goto l619
					}
					break
				case 'u':
					position++ // matchString(`ul`)
					if !matchChar('l') {
						goto l619
					}
					break
				case 'p':
					position++
					if !matchString("re") {
						goto l619
					}
					break
				case 'o':
					position++ // matchString(`ol`)
					if !matchChar('l') {
						goto l619
					}
					break
				case 'n':
					position++
					if !matchString("oscript") {
						goto l619
					}
					break
				case 'm':
					position++
					if !matchString("enu") {
						goto l619
					}
					break
				case 'i':
					position++
					if !matchString("sindex") {
						goto l619
					}
					break
				case 'h':
					position++ // matchString(`hr`)
					if !matchChar('r') {
						goto l619
					}
					break
				case 'c':
					position++
					if !matchString("enter") {
						goto l619
					}
					break
				case 'b':
					position++
					if !matchString("lockquote") {
						goto l619
					}
					break
				case 'a':
					position++
					if !matchString("ddress") {
						goto l619
					}
					break
				default:
					goto l619
				}
			}
		l620:
			return true
		l619:
			return false
		},
		/* 137 StyleOpen <- ('<' Spnl ((&[S] 'STYLE') | (&[s] 'style')) Spnl HtmlAttribute* '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l662
			}
			if !p.rules[ruleSpnl]() {
				goto l662
			}
			{
				if position == len(p.Buffer) {
					goto l662
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l662
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l662
					}
					break
				default:
					goto l662
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l662
			}
		l664:
			if !p.rules[ruleHtmlAttribute]() {
				goto l665
			}
			goto l664
		l665:
			if !matchChar('>') {
				goto l662
			}
			return true
		l662:
			position = position0
			return false
		},
		/* 138 StyleClose <- ('<' Spnl '/' ((&[S] 'STYLE') | (&[s] 'style')) Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l666
			}
			if !p.rules[ruleSpnl]() {
				goto l666
			}
			if !matchChar('/') {
				goto l666
			}
			{
				if position == len(p.Buffer) {
					goto l666
				}
				switch p.Buffer[position] {
				case 'S':
					position++
					if !matchString("TYLE") {
						goto l666
					}
					break
				case 's':
					position++
					if !matchString("tyle") {
						goto l666
					}
					break
				default:
					goto l666
				}
			}
			if !p.rules[ruleSpnl]() {
				goto l666
			}
			if !matchChar('>') {
				goto l666
			}
			return true
		l666:
			position = position0
			return false
		},
		/* 139 InStyleTags <- (StyleOpen (!StyleClose .)* StyleClose) */
		func() bool {
			position0 := position
			if !p.rules[ruleStyleOpen]() {
				goto l668
			}
		l669:
			{
				position670 := position
				if !p.rules[ruleStyleClose]() {
					goto l671
				}
				goto l670
			l671:
				if !matchDot() {
					goto l670
				}
				goto l669
			l670:
				position = position670
			}
			if !p.rules[ruleStyleClose]() {
				goto l668
			}
			return true
		l668:
			position = position0
			return false
		},
		/* 140 StyleBlock <- (< InStyleTags > BlankLine* {   if p.extension.FilterStyles {
                        yy = p.mkList(LIST, nil)
                    } else {
                        yy = p.mkString(yytext)
                        yy.key = HTMLBLOCK
                    }
                }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleInStyleTags]() {
				goto l672
			}
			end = position
		l673:
			if !p.rules[ruleBlankLine]() {
				goto l674
			}
			goto l673
		l674:
			do(42)
			return true
		l672:
			position = position0
			return false
		},
		/* 141 Inlines <- (StartList ((!Endline Inline { a = cons(yy, a) }) / (Endline &Inline { a = cons(c, a) }))+ Endline? { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l675
			}
			doarg(yySet, -1)
			{
				position678 := position
				if !p.rules[ruleEndline]() {
					goto l680
				}
				goto l679
			l680:
				if !p.rules[ruleInline]() {
					goto l679
				}
				do(43)
				goto l678
			l679:
				position = position678
				if !p.rules[ruleEndline]() {
					goto l675
				}
				doarg(yySet, -2)
				{
					position681 := position
					if !p.rules[ruleInline]() {
						goto l675
					}
					position = position681
				}
				do(44)
			}
		l678:
		l676:
			{
				position677, thunkPosition677 := position, thunkPosition
				{
					position682 := position
					if !p.rules[ruleEndline]() {
						goto l684
					}
					goto l683
				l684:
					if !p.rules[ruleInline]() {
						goto l683
					}
					do(43)
					goto l682
				l683:
					position = position682
					if !p.rules[ruleEndline]() {
						goto l677
					}
					doarg(yySet, -2)
					{
						position685 := position
						if !p.rules[ruleInline]() {
							goto l677
						}
						position = position685
					}
					do(44)
				}
			l682:
				goto l676
			l677:
				position, thunkPosition = position677, thunkPosition677
			}
			if !p.rules[ruleEndline]() {
				goto l686
			}
		l686:
			do(45)
			doarg(yyPop, 2)
			return true
		l675:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 142 Inline <- (Str / Endline / UlOrStarLine / Space / Strong / Emph / Image / Link / NoteReference / InlineNote / Code / RawHtml / Entity / EscapedChar / Smart / Symbol) */
		func() bool {
			if !p.rules[ruleStr]() {
				goto l690
			}
			goto l689
		l690:
			if !p.rules[ruleEndline]() {
				goto l691
			}
			goto l689
		l691:
			if !p.rules[ruleUlOrStarLine]() {
				goto l692
			}
			goto l689
		l692:
			if !p.rules[ruleSpace]() {
				goto l693
			}
			goto l689
		l693:
			if !p.rules[ruleStrong]() {
				goto l694
			}
			goto l689
		l694:
			if !p.rules[ruleEmph]() {
				goto l695
			}
			goto l689
		l695:
			if !p.rules[ruleImage]() {
				goto l696
			}
			goto l689
		l696:
			if !p.rules[ruleLink]() {
				goto l697
			}
			goto l689
		l697:
			if !p.rules[ruleNoteReference]() {
				goto l698
			}
			goto l689
		l698:
			if !p.rules[ruleInlineNote]() {
				goto l699
			}
			goto l689
		l699:
			if !p.rules[ruleCode]() {
				goto l700
			}
			goto l689
		l700:
			if !p.rules[ruleRawHtml]() {
				goto l701
			}
			goto l689
		l701:
			if !p.rules[ruleEntity]() {
				goto l702
			}
			goto l689
		l702:
			if !p.rules[ruleEscapedChar]() {
				goto l703
			}
			goto l689
		l703:
			if !p.rules[ruleSmart]() {
				goto l704
			}
			goto l689
		l704:
			if !p.rules[ruleSymbol]() {
				goto l688
			}
		l689:
			return true
		l688:
			return false
		},
		/* 143 Space <- (Spacechar+ { yy = p.mkString(" ")
          yy.key = SPACE }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l705
			}
		l706:
			if !p.rules[ruleSpacechar]() {
				goto l707
			}
			goto l706
		l707:
			do(46)
			return true
		l705:
			position = position0
			return false
		},
		/* 144 Str <- (StartList < NormalChar+ > { a = cons(p.mkString(yytext), a) } (StrChunk { a = cons(yy, a) })* { if a.next == nil { yy = a; } else { yy = p.mkList(LIST, a) } }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l708
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleNormalChar]() {
				goto l708
			}
		l709:
			if !p.rules[ruleNormalChar]() {
				goto l710
			}
			goto l709
		l710:
			end = position
			do(47)
		l711:
			{
				position712, thunkPosition712 := position, thunkPosition
				if !p.rules[ruleStrChunk]() {
					goto l712
				}
				do(48)
				goto l711
			l712:
				position, thunkPosition = position712, thunkPosition712
			}
			do(49)
			doarg(yyPop, 1)
			return true
		l708:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 145 StrChunk <- ((< (NormalChar / ('_'+ &Alphanumeric))+ > { yy = p.mkString(yytext) }) / AposChunk) */
		func() bool {
			position0 := position
			{
				position714 := position
				begin = position
				if !p.rules[ruleNormalChar]() {
					goto l719
				}
				goto l718
			l719:
				if !matchChar('_') {
					goto l715
				}
			l720:
				if !matchChar('_') {
					goto l721
				}
				goto l720
			l721:
				{
					position722 := position
					if !p.rules[ruleAlphanumeric]() {
						goto l715
					}
					position = position722
				}
			l718:
			l716:
				{
					position717 := position
					if !p.rules[ruleNormalChar]() {
						goto l724
					}
					goto l723
				l724:
					if !matchChar('_') {
						goto l717
					}
				l725:
					if !matchChar('_') {
						goto l726
					}
					goto l725
				l726:
					{
						position727 := position
						if !p.rules[ruleAlphanumeric]() {
							goto l717
						}
						position = position727
					}
				l723:
					goto l716
				l717:
					position = position717
				}
				end = position
				do(50)
				goto l714
			l715:
				position = position714
				if !p.rules[ruleAposChunk]() {
					goto l713
				}
			}
		l714:
			return true
		l713:
			position = position0
			return false
		},
		/* 146 AposChunk <- (&{p.extension.Smart} '\'' &Alphanumeric { yy = p.mkElem(APOSTROPHE) }) */
		func() bool {
			position0 := position
			if !(p.extension.Smart) {
				goto l728
			}
			if !matchChar('\'') {
				goto l728
			}
			{
				position729 := position
				if !p.rules[ruleAlphanumeric]() {
					goto l728
				}
				position = position729
			}
			do(51)
			return true
		l728:
			position = position0
			return false
		},
		/* 147 EscapedChar <- ('\\' !Newline < [-\\`|*_{}[\]()#+.!><] > { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !matchChar('\\') {
				goto l730
			}
			if !p.rules[ruleNewline]() {
				goto l731
			}
			goto l730
		l731:
			begin = position
			if !matchClass(1) {
				goto l730
			}
			end = position
			do(52)
			return true
		l730:
			position = position0
			return false
		},
		/* 148 Entity <- ((HexEntity / DecEntity / CharEntity) { yy = p.mkString(yytext); yy.key = HTML }) */
		func() bool {
			position0 := position
			if !p.rules[ruleHexEntity]() {
				goto l734
			}
			goto l733
		l734:
			if !p.rules[ruleDecEntity]() {
				goto l735
			}
			goto l733
		l735:
			if !p.rules[ruleCharEntity]() {
				goto l732
			}
		l733:
			do(53)
			return true
		l732:
			position = position0
			return false
		},
		/* 149 Endline <- (LineBreak / TerminalEndline / NormalEndline) */
		func() bool {
			if !p.rules[ruleLineBreak]() {
				goto l738
			}
			goto l737
		l738:
			if !p.rules[ruleTerminalEndline]() {
				goto l739
			}
			goto l737
		l739:
			if !p.rules[ruleNormalEndline]() {
				goto l736
			}
		l737:
			return true
		l736:
			return false
		},
		/* 150 NormalEndline <- (Sp Newline !BlankLine !'>' !AtxStart !(Line ((&[\-] '-'+) | (&[=] '='+)) Newline) { yy = p.mkString("\n")
                    yy.key = SPACE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l740
			}
			if !p.rules[ruleNewline]() {
				goto l740
			}
			if !p.rules[ruleBlankLine]() {
				goto l741
			}
			goto l740
		l741:
			if peekChar('>') {
				goto l740
			}
			if !p.rules[ruleAtxStart]() {
				goto l742
			}
			goto l740
		l742:
			{
				position743, thunkPosition743 := position, thunkPosition
				if !p.rules[ruleLine]() {
					goto l743
				}
				{
					if position == len(p.Buffer) {
						goto l743
					}
					switch p.Buffer[position] {
					case '-':
						if !matchChar('-') {
							goto l743
						}
					l745:
						if !matchChar('-') {
							goto l746
						}
						goto l745
					l746:
						break
					case '=':
						if !matchChar('=') {
							goto l743
						}
					l747:
						if !matchChar('=') {
							goto l748
						}
						goto l747
					l748:
						break
					default:
						goto l743
					}
				}
				if !p.rules[ruleNewline]() {
					goto l743
				}
				goto l740
			l743:
				position, thunkPosition = position743, thunkPosition743
			}
			do(54)
			return true
		l740:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 151 TerminalEndline <- (Sp Newline !. { yy = nil }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l749
			}
			if !p.rules[ruleNewline]() {
				goto l749
			}
			if (position < len(p.Buffer)) {
				goto l749
			}
			do(55)
			return true
		l749:
			position = position0
			return false
		},
		/* 152 LineBreak <- ('  ' NormalEndline { yy = p.mkElem(LINEBREAK) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("  ") {
				goto l750
			}
			if !p.rules[ruleNormalEndline]() {
				goto l750
			}
			do(56)
			return true
		l750:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 153 Symbol <- (< SpecialChar > { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleSpecialChar]() {
				goto l751
			}
			end = position
			do(57)
			return true
		l751:
			position = position0
			return false
		},
		/* 154 UlOrStarLine <- ((UlLine / StarLine) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleUlLine]() {
				goto l754
			}
			goto l753
		l754:
			if !p.rules[ruleStarLine]() {
				goto l752
			}
		l753:
			do(58)
			return true
		l752:
			position = position0
			return false
		},
		/* 155 StarLine <- ((&[*] (< '****' '*'* >)) | (&[\t ] (< Spacechar '*'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l755
				}
				switch p.Buffer[position] {
				case '*':
					begin = position
					if !matchString("****") {
						goto l755
					}
				l757:
					if !matchChar('*') {
						goto l758
					}
					goto l757
				l758:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l755
					}
					if !matchChar('*') {
						goto l755
					}
				l759:
					if !matchChar('*') {
						goto l760
					}
					goto l759
				l760:
					{
						position761 := position
						if !p.rules[ruleSpacechar]() {
							goto l755
						}
						position = position761
					}
					end = position
					break
				default:
					goto l755
				}
			}
			return true
		l755:
			position = position0
			return false
		},
		/* 156 UlLine <- ((&[_] (< '____' '_'* >)) | (&[\t ] (< Spacechar '_'+ &Spacechar >))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l762
				}
				switch p.Buffer[position] {
				case '_':
					begin = position
					if !matchString("____") {
						goto l762
					}
				l764:
					if !matchChar('_') {
						goto l765
					}
					goto l764
				l765:
					end = position
					break
				case '\t', ' ':
					begin = position
					if !p.rules[ruleSpacechar]() {
						goto l762
					}
					if !matchChar('_') {
						goto l762
					}
				l766:
					if !matchChar('_') {
						goto l767
					}
					goto l766
				l767:
					{
						position768 := position
						if !p.rules[ruleSpacechar]() {
							goto l762
						}
						position = position768
					}
					end = position
					break
				default:
					goto l762
				}
			}
			return true
		l762:
			position = position0
			return false
		},
		/* 157 Emph <- ((&[_] EmphUl) | (&[*] EmphStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l769
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleEmphUl]() {
						goto l769
					}
					break
				case '*':
					if !p.rules[ruleEmphStar]() {
						goto l769
					}
					break
				default:
					goto l769
				}
			}
			return true
		l769:
			return false
		},
		/* 158 Whitespace <- ((&[\n\r] Newline) | (&[\t ] Spacechar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l771
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l771
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l771
					}
					break
				default:
					goto l771
				}
			}
			return true
		l771:
			return false
		},
		/* 159 EmphStar <- ('*' !Whitespace StartList ((!'*' Inline { a = cons(b, a) }) / (StrongStar { a = cons(b, a) }))+ '*' { yy = p.mkList(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('*') {
				goto l773
			}
			if !p.rules[ruleWhitespace]() {
				goto l774
			}
			goto l773
		l774:
			if !p.rules[ruleStartList]() {
				goto l773
			}
			doarg(yySet, -2)
			{
				position777, thunkPosition777 := position, thunkPosition
				if peekChar('*') {
					goto l778
				}
				if !p.rules[ruleInline]() {
					goto l778
				}
				doarg(yySet, -1)
				do(59)
				goto l777
			l778:
				position, thunkPosition = position777, thunkPosition777
				if !p.rules[ruleStrongStar]() {
					goto l773
				}
				doarg(yySet, -1)
				do(60)
			}
		l777:
		l775:
			{
				position776, thunkPosition776 := position, thunkPosition
				{
					position779, thunkPosition779 := position, thunkPosition
					if peekChar('*') {
						goto l780
					}
					if !p.rules[ruleInline]() {
						goto l780
					}
					doarg(yySet, -1)
					do(59)
					goto l779
				l780:
					position, thunkPosition = position779, thunkPosition779
					if !p.rules[ruleStrongStar]() {
						goto l776
					}
					doarg(yySet, -1)
					do(60)
				}
			l779:
				goto l775
			l776:
				position, thunkPosition = position776, thunkPosition776
			}
			if !matchChar('*') {
				goto l773
			}
			do(61)
			doarg(yyPop, 2)
			return true
		l773:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 160 EmphUl <- ('_' !Whitespace StartList ((!'_' Inline { a = cons(b, a) }) / (StrongUl { a = cons(b, a) }))+ '_' { yy = p.mkList(EMPH, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('_') {
				goto l781
			}
			if !p.rules[ruleWhitespace]() {
				goto l782
			}
			goto l781
		l782:
			if !p.rules[ruleStartList]() {
				goto l781
			}
			doarg(yySet, -2)
			{
				position785, thunkPosition785 := position, thunkPosition
				if peekChar('_') {
					goto l786
				}
				if !p.rules[ruleInline]() {
					goto l786
				}
				doarg(yySet, -1)
				do(62)
				goto l785
			l786:
				position, thunkPosition = position785, thunkPosition785
				if !p.rules[ruleStrongUl]() {
					goto l781
				}
				doarg(yySet, -1)
				do(63)
			}
		l785:
		l783:
			{
				position784, thunkPosition784 := position, thunkPosition
				{
					position787, thunkPosition787 := position, thunkPosition
					if peekChar('_') {
						goto l788
					}
					if !p.rules[ruleInline]() {
						goto l788
					}
					doarg(yySet, -1)
					do(62)
					goto l787
				l788:
					position, thunkPosition = position787, thunkPosition787
					if !p.rules[ruleStrongUl]() {
						goto l784
					}
					doarg(yySet, -1)
					do(63)
				}
			l787:
				goto l783
			l784:
				position, thunkPosition = position784, thunkPosition784
			}
			if !matchChar('_') {
				goto l781
			}
			do(64)
			doarg(yyPop, 2)
			return true
		l781:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 161 Strong <- ((&[_] StrongUl) | (&[*] StrongStar)) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l789
				}
				switch p.Buffer[position] {
				case '_':
					if !p.rules[ruleStrongUl]() {
						goto l789
					}
					break
				case '*':
					if !p.rules[ruleStrongStar]() {
						goto l789
					}
					break
				default:
					goto l789
				}
			}
			return true
		l789:
			return false
		},
		/* 162 StrongStar <- ('**' !Whitespace StartList (!'**' Inline { a = cons(b, a) })+ '**' { yy = p.mkList(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchString("**") {
				goto l791
			}
			if !p.rules[ruleWhitespace]() {
				goto l792
			}
			goto l791
		l792:
			if !p.rules[ruleStartList]() {
				goto l791
			}
			doarg(yySet, -2)
			if !matchString("**") {
				goto l795
			}
			goto l791
		l795:
			if !p.rules[ruleInline]() {
				goto l791
			}
			doarg(yySet, -1)
			do(65)
		l793:
			{
				position794, thunkPosition794 := position, thunkPosition
				if !matchString("**") {
					goto l796
				}
				goto l794
			l796:
				if !p.rules[ruleInline]() {
					goto l794
				}
				doarg(yySet, -1)
				do(65)
				goto l793
			l794:
				position, thunkPosition = position794, thunkPosition794
			}
			if !matchString("**") {
				goto l791
			}
			do(66)
			doarg(yyPop, 2)
			return true
		l791:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 163 StrongUl <- ('__' !Whitespace StartList (!'__' Inline { a = cons(b, a) })+ '__' { yy = p.mkList(STRONG, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchString("__") {
				goto l797
			}
			if !p.rules[ruleWhitespace]() {
				goto l798
			}
			goto l797
		l798:
			if !p.rules[ruleStartList]() {
				goto l797
			}
			doarg(yySet, -2)
			if !matchString("__") {
				goto l801
			}
			goto l797
		l801:
			if !p.rules[ruleInline]() {
				goto l797
			}
			doarg(yySet, -1)
			do(67)
		l799:
			{
				position800, thunkPosition800 := position, thunkPosition
				if !matchString("__") {
					goto l802
				}
				goto l800
			l802:
				if !p.rules[ruleInline]() {
					goto l800
				}
				doarg(yySet, -1)
				do(67)
				goto l799
			l800:
				position, thunkPosition = position800, thunkPosition800
			}
			if !matchString("__") {
				goto l797
			}
			do(68)
			doarg(yyPop, 2)
			return true
		l797:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 164 Image <- ('!' (ExplicitLink / ReferenceLink) {	if yy.key == LINK {
			yy.key = IMAGE
		} else {
			result := yy
			yy.children = cons(p.mkString("!"), result.children)
		}
	}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto l803
			}
			if !p.rules[ruleExplicitLink]() {
				goto l805
			}
			goto l804
		l805:
			if !p.rules[ruleReferenceLink]() {
				goto l803
			}
		l804:
			do(69)
			return true
		l803:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 165 Link <- (ExplicitLink / ReferenceLink / AutoLink) */
		func() bool {
			if !p.rules[ruleExplicitLink]() {
				goto l808
			}
			goto l807
		l808:
			if !p.rules[ruleReferenceLink]() {
				goto l809
			}
			goto l807
		l809:
			if !p.rules[ruleAutoLink]() {
				goto l806
			}
		l807:
			return true
		l806:
			return false
		},
		/* 166 ReferenceLink <- (ReferenceLinkDouble / ReferenceLinkSingle) */
		func() bool {
			if !p.rules[ruleReferenceLinkDouble]() {
				goto l812
			}
			goto l811
		l812:
			if !p.rules[ruleReferenceLinkSingle]() {
				goto l810
			}
		l811:
			return true
		l810:
			return false
		},
		/* 167 ReferenceLinkDouble <- (Label < Spnl > !'[]' Label {
                           if match, found := p.findReference(b.children); found {
                               yy = p.mkLink(a.children, match.url, match.title);
                               a = nil
                               b = nil
                           } else {
                               result := p.mkElem(LIST)
                               result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), cons(p.mkString(yytext),
                                                   cons(p.mkString("["), cons(b, p.mkString("]")))))))
                               yy = result
                           }
                       }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleLabel]() {
				goto l813
			}
			doarg(yySet, -1)
			begin = position
			if !p.rules[ruleSpnl]() {
				goto l813
			}
			end = position
			if !matchString("[]") {
				goto l814
			}
			goto l813
		l814:
			if !p.rules[ruleLabel]() {
				goto l813
			}
			doarg(yySet, -2)
			do(70)
			doarg(yyPop, 2)
			return true
		l813:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 168 ReferenceLinkSingle <- (Label < (Spnl '[]')? > {
                           if match, found := p.findReference(a.children); found {
                               yy = p.mkLink(a.children, match.url, match.title)
                               a = nil
                           } else {
                               result := p.mkElem(LIST)
                               result.children = cons(p.mkString("["), cons(a, cons(p.mkString("]"), p.mkString(yytext))));
                               yy = result
                           }
                       }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleLabel]() {
				goto l815
			}
			doarg(yySet, -1)
			begin = position
			{
				position816 := position
				if !p.rules[ruleSpnl]() {
					goto l816
				}
				if !matchString("[]") {
					goto l816
				}
				goto l817
			l816:
				position = position816
			}
		l817:
			end = position
			do(71)
			doarg(yyPop, 1)
			return true
		l815:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 169 ExplicitLink <- (Label '(' Sp Source Spnl Title Sp ')' { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
                  s = nil
                  t = nil
                  l = nil }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleLabel]() {
				goto l818
			}
			doarg(yySet, -1)
			if !matchChar('(') {
				goto l818
			}
			if !p.rules[ruleSp]() {
				goto l818
			}
			if !p.rules[ruleSource]() {
				goto l818
			}
			doarg(yySet, -3)
			if !p.rules[ruleSpnl]() {
				goto l818
			}
			if !p.rules[ruleTitle]() {
				goto l818
			}
			doarg(yySet, -2)
			if !p.rules[ruleSp]() {
				goto l818
			}
			if !matchChar(')') {
				goto l818
			}
			do(72)
			doarg(yyPop, 3)
			return true
		l818:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 170 Source <- ((('<' < SourceContents > '>') / (< SourceContents >)) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			{
				position820 := position
				if !matchChar('<') {
					goto l821
				}
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l821
				}
				end = position
				if !matchChar('>') {
					goto l821
				}
				goto l820
			l821:
				position = position820
				begin = position
				if !p.rules[ruleSourceContents]() {
					goto l819
				}
				end = position
			}
		l820:
			do(73)
			return true
		l819:
			position = position0
			return false
		},
		/* 171 SourceContents <- ((!'(' !')' !'>' Nonspacechar)+ / ('(' SourceContents ')'))* */
		func() bool {
		l823:
			{
				position824 := position
				if position == len(p.Buffer) {
					goto l826
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l826
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l826
					}
				}
			l827:
				if position == len(p.Buffer) {
					goto l828
				}
				switch p.Buffer[position] {
				case '(', ')', '>':
					goto l828
				default:
					if !p.rules[ruleNonspacechar]() {
						goto l828
					}
				}
				goto l827
			l828:
				goto l825
			l826:
				if !matchChar('(') {
					goto l824
				}
				if !p.rules[ruleSourceContents]() {
					goto l824
				}
				if !matchChar(')') {
					goto l824
				}
			l825:
				goto l823
			l824:
				position = position824
			}
			return true
		},
		/* 172 Title <- ((TitleSingle / TitleDouble / (< '' >)) { yy = p.mkString(yytext) }) */
		func() bool {
			if !p.rules[ruleTitleSingle]() {
				goto l831
			}
			goto l830
		l831:
			if !p.rules[ruleTitleDouble]() {
				goto l832
			}
			goto l830
		l832:
			begin = position
			end = position
		l830:
			do(74)
			return true
		},
		/* 173 TitleSingle <- ('\'' < (!('\'' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '\'') */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l833
			}
			begin = position
		l834:
			{
				position835 := position
				{
					position836 := position
					if !matchChar('\'') {
						goto l836
					}
					if !p.rules[ruleSp]() {
						goto l836
					}
					{
						if position == len(p.Buffer) {
							goto l836
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l836
							}
							break
						default:
							goto l836
						}
					}
					goto l835
				l836:
					position = position836
				}
				if !matchDot() {
					goto l835
				}
				goto l834
			l835:
				position = position835
			}
			end = position
			if !matchChar('\'') {
				goto l833
			}
			return true
		l833:
			position = position0
			return false
		},
		/* 174 TitleDouble <- ('"' < (!('"' Sp ((&[)] ')') | (&[\n\r] Newline))) .)* > '"') */
		func() bool {
			position0 := position
			if !matchChar('"') {
				goto l838
			}
			begin = position
		l839:
			{
				position840 := position
				{
					position841 := position
					if !matchChar('"') {
						goto l841
					}
					if !p.rules[ruleSp]() {
						goto l841
					}
					{
						if position == len(p.Buffer) {
							goto l841
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l841
							}
							break
						default:
							goto l841
						}
					}
					goto l840
				l841:
					position = position841
				}
				if !matchDot() {
					goto l840
				}
				goto l839
			l840:
				position = position840
			}
			end = position
			if !matchChar('"') {
				goto l838
			}
			return true
		l838:
			position = position0
			return false
		},
		/* 175 AutoLink <- (AutoLinkUrl / AutoLinkEmail) */
		func() bool {
			if !p.rules[ruleAutoLinkUrl]() {
				goto l845
			}
			goto l844
		l845:
			if !p.rules[ruleAutoLinkEmail]() {
				goto l843
			}
		l844:
			return true
		l843:
			return false
		},
		/* 176 AutoLinkUrl <- ('<' < [A-Za-z]+ '://' (!Newline !'>' .)+ > '>' {   yy = p.mkLink(p.mkString(yytext), yytext, "") }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l846
			}
			begin = position
			if !matchClass(2) {
				goto l846
			}
		l847:
			if !matchClass(2) {
				goto l848
			}
			goto l847
		l848:
			if !matchString("://") {
				goto l846
			}
			if !p.rules[ruleNewline]() {
				goto l851
			}
			goto l846
		l851:
			if peekChar('>') {
				goto l846
			}
			if !matchDot() {
				goto l846
			}
		l849:
			{
				position850 := position
				if !p.rules[ruleNewline]() {
					goto l852
				}
				goto l850
			l852:
				if peekChar('>') {
					goto l850
				}
				if !matchDot() {
					goto l850
				}
				goto l849
			l850:
				position = position850
			}
			end = position
			if !matchChar('>') {
				goto l846
			}
			do(75)
			return true
		l846:
			position = position0
			return false
		},
		/* 177 AutoLinkEmail <- ('<' 'mailto:'? < [-A-Za-z0-9+_./!%~$]+ '@' (!Newline !'>' .)+ > '>' {
                    yy = p.mkLink(p.mkString(yytext), "mailto:"+yytext, "")
                }) */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l853
			}
			if !matchString("mailto:") {
				goto l854
			}
		l854:
			begin = position
			if !matchClass(3) {
				goto l853
			}
		l856:
			if !matchClass(3) {
				goto l857
			}
			goto l856
		l857:
			if !matchChar('@') {
				goto l853
			}
			if !p.rules[ruleNewline]() {
				goto l860
			}
			goto l853
		l860:
			if peekChar('>') {
				goto l853
			}
			if !matchDot() {
				goto l853
			}
		l858:
			{
				position859 := position
				if !p.rules[ruleNewline]() {
					goto l861
				}
				goto l859
			l861:
				if peekChar('>') {
					goto l859
				}
				if !matchDot() {
					goto l859
				}
				goto l858
			l859:
				position = position859
			}
			end = position
			if !matchChar('>') {
				goto l853
			}
			do(76)
			return true
		l853:
			position = position0
			return false
		},
		/* 178 Reference <- (NonindentSpace !'[]' Label ':' Spnl RefSrc RefTitle BlankLine+ { yy = p.mkLink(l.children, s.contents.str, t.contents.str)
              s = nil
              t = nil
              l = nil
              yy.key = REFERENCE }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleNonindentSpace]() {
				goto l862
			}
			if !matchString("[]") {
				goto l863
			}
			goto l862
		l863:
			if !p.rules[ruleLabel]() {
				goto l862
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l862
			}
			if !p.rules[ruleSpnl]() {
				goto l862
			}
			if !p.rules[ruleRefSrc]() {
				goto l862
			}
			doarg(yySet, -3)
			if !p.rules[ruleRefTitle]() {
				goto l862
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l862
			}
		l864:
			if !p.rules[ruleBlankLine]() {
				goto l865
			}
			goto l864
		l865:
			do(77)
			doarg(yyPop, 3)
			return true
		l862:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 179 Label <- ('[' ((!'^' &{p.extension.Notes}) / (&. &{!p.extension.Notes})) StartList (!']' Inline { a = cons(yy, a) })* ']' { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l866
			}
			if peekChar('^') {
				goto l868
			}
			if !(p.extension.Notes) {
				goto l868
			}
			goto l867
		l868:
			if !(position < len(p.Buffer)) {
				goto l866
			}
			if !(!p.extension.Notes) {
				goto l866
			}
		l867:
			if !p.rules[ruleStartList]() {
				goto l866
			}
			doarg(yySet, -1)
		l869:
			{
				position870 := position
				if peekChar(']') {
					goto l870
				}
				if !p.rules[ruleInline]() {
					goto l870
				}
				do(78)
				goto l869
			l870:
				position = position870
			}
			if !matchChar(']') {
				goto l866
			}
			do(79)
			doarg(yyPop, 1)
			return true
		l866:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 180 RefSrc <- (< Nonspacechar+ > { yy = p.mkString(yytext)
           yy.key = HTML }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleNonspacechar]() {
				goto l871
			}
		l872:
			if !p.rules[ruleNonspacechar]() {
				goto l873
			}
			goto l872
		l873:
			end = position
			do(80)
			return true
		l871:
			position = position0
			return false
		},
		/* 181 RefTitle <- ((RefTitleSingle / RefTitleDouble / RefTitleParens / EmptyTitle) { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRefTitleSingle]() {
				goto l876
			}
			goto l875
		l876:
			if !p.rules[ruleRefTitleDouble]() {
				goto l877
			}
			goto l875
		l877:
			if !p.rules[ruleRefTitleParens]() {
				goto l878
			}
			goto l875
		l878:
			if !p.rules[ruleEmptyTitle]() {
				goto l874
			}
		l875:
			do(81)
			return true
		l874:
			position = position0
			return false
		},
		/* 182 EmptyTitle <- (< '' >) */
		func() bool {
			begin = position
			end = position
			return true
		},
		/* 183 RefTitleSingle <- (Spnl '\'' < (!((&[\'] ('\'' Sp Newline)) | (&[\n\r] Newline)) .)* > '\'') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l880
			}
			if !matchChar('\'') {
				goto l880
			}
			begin = position
		l881:
			{
				position882 := position
				{
					position883 := position
					{
						if position == len(p.Buffer) {
							goto l883
						}
						switch p.Buffer[position] {
						case '\'':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l883
							}
							if !p.rules[ruleNewline]() {
								goto l883
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l883
							}
							break
						default:
							goto l883
						}
					}
					goto l882
				l883:
					position = position883
				}
				if !matchDot() {
					goto l882
				}
				goto l881
			l882:
				position = position882
			}
			end = position
			if !matchChar('\'') {
				goto l880
			}
			return true
		l880:
			position = position0
			return false
		},
		/* 184 RefTitleDouble <- (Spnl '"' < (!((&[\"] ('"' Sp Newline)) | (&[\n\r] Newline)) .)* > '"') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l885
			}
			if !matchChar('"') {
				goto l885
			}
			begin = position
		l886:
			{
				position887 := position
				{
					position888 := position
					{
						if position == len(p.Buffer) {
							goto l888
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l888
							}
							if !p.rules[ruleNewline]() {
								goto l888
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l888
							}
							break
						default:
							goto l888
						}
					}
					goto l887
				l888:
					position = position888
				}
				if !matchDot() {
					goto l887
				}
				goto l886
			l887:
				position = position887
			}
			end = position
			if !matchChar('"') {
				goto l885
			}
			return true
		l885:
			position = position0
			return false
		},
		/* 185 RefTitleParens <- (Spnl '(' < (!((&[)] (')' Sp Newline)) | (&[\n\r] Newline)) .)* > ')') */
		func() bool {
			position0 := position
			if !p.rules[ruleSpnl]() {
				goto l890
			}
			if !matchChar('(') {
				goto l890
			}
			begin = position
		l891:
			{
				position892 := position
				{
					position893 := position
					{
						if position == len(p.Buffer) {
							goto l893
						}
						switch p.Buffer[position] {
						case ')':
							position++ // matchChar
							if !p.rules[ruleSp]() {
								goto l893
							}
							if !p.rules[ruleNewline]() {
								goto l893
							}
							break
						case '\n', '\r':
							if !p.rules[ruleNewline]() {
								goto l893
							}
							break
						default:
							goto l893
						}
					}
					goto l892
				l893:
					position = position893
				}
				if !matchDot() {
					goto l892
				}
				goto l891
			l892:
				position = position892
			}
			end = position
			if !matchChar(')') {
				goto l890
			}
			return true
		l890:
			position = position0
			return false
		},
		/* 186 References <- (StartList ((Reference { a = cons(b, a) }) / SkipBlock)* { p.references = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l895
			}
			doarg(yySet, -2)
		l896:
			{
				position897, thunkPosition897 := position, thunkPosition
				{
					position898, thunkPosition898 := position, thunkPosition
					if !p.rules[ruleReference]() {
						goto l899
					}
					doarg(yySet, -1)
					do(82)
					goto l898
				l899:
					position, thunkPosition = position898, thunkPosition898
					if !p.rules[ruleSkipBlock]() {
						goto l897
					}
				}
			l898:
				goto l896
			l897:
				position, thunkPosition = position897, thunkPosition897
			}
			do(83)
			if !(commit(thunkPosition0)) {
				goto l895
			}
			doarg(yyPop, 2)
			return true
		l895:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 187 Ticks1 <- ('`' !'`') */
		func() bool {
			position0 := position
			if !matchChar('`') {
				goto l900
			}
			if peekChar('`') {
				goto l900
			}
			return true
		l900:
			position = position0
			return false
		},
		/* 188 Ticks2 <- ('``' !'`') */
		func() bool {
			position0 := position
			if !matchString("``") {
				goto l901
			}
			if peekChar('`') {
				goto l901
			}
			return true
		l901:
			position = position0
			return false
		},
		/* 189 Ticks3 <- ('```' !'`') */
		func() bool {
			position0 := position
			if !matchString("```") {
				goto l902
			}
			if peekChar('`') {
				goto l902
			}
			return true
		l902:
			position = position0
			return false
		},
		/* 190 Ticks4 <- ('````' !'`') */
		func() bool {
			position0 := position
			if !matchString("````") {
				goto l903
			}
			if peekChar('`') {
				goto l903
			}
			return true
		l903:
			position = position0
			return false
		},
		/* 191 Ticks5 <- ('`````' !'`') */
		func() bool {
			position0 := position
			if !matchString("`````") {
				goto l904
			}
			if peekChar('`') {
				goto l904
			}
			return true
		l904:
			position = position0
			return false
		},
		/* 192 Code <- (((Ticks1 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks1 '`'+)) | (&[\t\n\r ] (!(Sp Ticks1) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks1) / (Ticks2 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks2 '`'+)) | (&[\t\n\r ] (!(Sp Ticks2) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks2) / (Ticks3 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks3 '`'+)) | (&[\t\n\r ] (!(Sp Ticks3) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks3) / (Ticks4 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks4 '`'+)) | (&[\t\n\r ] (!(Sp Ticks4) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks4) / (Ticks5 Sp < ((!'`' Nonspacechar)+ / ((&[`] (!Ticks5 '`'+)) | (&[\t\n\r ] (!(Sp Ticks5) ((&[\n\r] (Newline !BlankLine)) | (&[\t ] Spacechar))))))+ > Sp Ticks5)) { yy = p.mkString(yytext); yy.key = CODE }) */
		func() bool {
			position0 := position
			{
				position906 := position
				if !p.rules[ruleTicks1]() {
					goto l907
				}
				if !p.rules[ruleSp]() {
					goto l907
				}
				begin = position
				if peekChar('`') {
					goto l911
				}
				if !p.rules[ruleNonspacechar]() {
					goto l911
				}
			l912:
				if peekChar('`') {
					goto l913
				}
				if !p.rules[ruleNonspacechar]() {
					goto l913
				}
				goto l912
			l913:
				goto l910
			l911:
				{
					if position == len(p.Buffer) {
						goto l907
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks1]() {
							goto l915
						}
						goto l907
					l915:
						if !matchChar('`') {
							goto l907
						}
					l916:
						if !matchChar('`') {
							goto l917
						}
						goto l916
					l917:
						break
					default:
						{
							position918 := position
							if !p.rules[ruleSp]() {
								goto l918
							}
							if !p.rules[ruleTicks1]() {
								goto l918
							}
							goto l907
						l918:
							position = position918
						}
						{
							if position == len(p.Buffer) {
								goto l907
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l907
								}
								if !p.rules[ruleBlankLine]() {
									goto l920
								}
								goto l907
							l920:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l907
								}
								break
							default:
								goto l907
							}
						}
					}
				}
			l910:
			l908:
				{
					position909 := position
					if peekChar('`') {
						goto l922
					}
					if !p.rules[ruleNonspacechar]() {
						goto l922
					}
				l923:
					if peekChar('`') {
						goto l924
					}
					if !p.rules[ruleNonspacechar]() {
						goto l924
					}
					goto l923
				l924:
					goto l921
				l922:
					{
						if position == len(p.Buffer) {
							goto l909
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks1]() {
								goto l926
							}
							goto l909
						l926:
							if !matchChar('`') {
								goto l909
							}
						l927:
							if !matchChar('`') {
								goto l928
							}
							goto l927
						l928:
							break
						default:
							{
								position929 := position
								if !p.rules[ruleSp]() {
									goto l929
								}
								if !p.rules[ruleTicks1]() {
									goto l929
								}
								goto l909
							l929:
								position = position929
							}
							{
								if position == len(p.Buffer) {
									goto l909
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l909
									}
									if !p.rules[ruleBlankLine]() {
										goto l931
									}
									goto l909
								l931:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l909
									}
									break
								default:
									goto l909
								}
							}
						}
					}
				l921:
					goto l908
				l909:
					position = position909
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l907
				}
				if !p.rules[ruleTicks1]() {
					goto l907
				}
				goto l906
			l907:
				position = position906
				if !p.rules[ruleTicks2]() {
					goto l932
				}
				if !p.rules[ruleSp]() {
					goto l932
				}
				begin = position
				if peekChar('`') {
					goto l936
				}
				if !p.rules[ruleNonspacechar]() {
					goto l936
				}
			l937:
				if peekChar('`') {
					goto l938
				}
				if !p.rules[ruleNonspacechar]() {
					goto l938
				}
				goto l937
			l938:
				goto l935
			l936:
				{
					if position == len(p.Buffer) {
						goto l932
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks2]() {
							goto l940
						}
						goto l932
					l940:
						if !matchChar('`') {
							goto l932
						}
					l941:
						if !matchChar('`') {
							goto l942
						}
						goto l941
					l942:
						break
					default:
						{
							position943 := position
							if !p.rules[ruleSp]() {
								goto l943
							}
							if !p.rules[ruleTicks2]() {
								goto l943
							}
							goto l932
						l943:
							position = position943
						}
						{
							if position == len(p.Buffer) {
								goto l932
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l932
								}
								if !p.rules[ruleBlankLine]() {
									goto l945
								}
								goto l932
							l945:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l932
								}
								break
							default:
								goto l932
							}
						}
					}
				}
			l935:
			l933:
				{
					position934 := position
					if peekChar('`') {
						goto l947
					}
					if !p.rules[ruleNonspacechar]() {
						goto l947
					}
				l948:
					if peekChar('`') {
						goto l949
					}
					if !p.rules[ruleNonspacechar]() {
						goto l949
					}
					goto l948
				l949:
					goto l946
				l947:
					{
						if position == len(p.Buffer) {
							goto l934
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks2]() {
								goto l951
							}
							goto l934
						l951:
							if !matchChar('`') {
								goto l934
							}
						l952:
							if !matchChar('`') {
								goto l953
							}
							goto l952
						l953:
							break
						default:
							{
								position954 := position
								if !p.rules[ruleSp]() {
									goto l954
								}
								if !p.rules[ruleTicks2]() {
									goto l954
								}
								goto l934
							l954:
								position = position954
							}
							{
								if position == len(p.Buffer) {
									goto l934
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l934
									}
									if !p.rules[ruleBlankLine]() {
										goto l956
									}
									goto l934
								l956:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l934
									}
									break
								default:
									goto l934
								}
							}
						}
					}
				l946:
					goto l933
				l934:
					position = position934
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l932
				}
				if !p.rules[ruleTicks2]() {
					goto l932
				}
				goto l906
			l932:
				position = position906
				if !p.rules[ruleTicks3]() {
					goto l957
				}
				if !p.rules[ruleSp]() {
					goto l957
				}
				begin = position
				if peekChar('`') {
					goto l961
				}
				if !p.rules[ruleNonspacechar]() {
					goto l961
				}
			l962:
				if peekChar('`') {
					goto l963
				}
				if !p.rules[ruleNonspacechar]() {
					goto l963
				}
				goto l962
			l963:
				goto l960
			l961:
				{
					if position == len(p.Buffer) {
						goto l957
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks3]() {
							goto l965
						}
						goto l957
					l965:
						if !matchChar('`') {
							goto l957
						}
					l966:
						if !matchChar('`') {
							goto l967
						}
						goto l966
					l967:
						break
					default:
						{
							position968 := position
							if !p.rules[ruleSp]() {
								goto l968
							}
							if !p.rules[ruleTicks3]() {
								goto l968
							}
							goto l957
						l968:
							position = position968
						}
						{
							if position == len(p.Buffer) {
								goto l957
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l957
								}
								if !p.rules[ruleBlankLine]() {
									goto l970
								}
								goto l957
							l970:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l957
								}
								break
							default:
								goto l957
							}
						}
					}
				}
			l960:
			l958:
				{
					position959 := position
					if peekChar('`') {
						goto l972
					}
					if !p.rules[ruleNonspacechar]() {
						goto l972
					}
				l973:
					if peekChar('`') {
						goto l974
					}
					if !p.rules[ruleNonspacechar]() {
						goto l974
					}
					goto l973
				l974:
					goto l971
				l972:
					{
						if position == len(p.Buffer) {
							goto l959
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks3]() {
								goto l976
							}
							goto l959
						l976:
							if !matchChar('`') {
								goto l959
							}
						l977:
							if !matchChar('`') {
								goto l978
							}
							goto l977
						l978:
							break
						default:
							{
								position979 := position
								if !p.rules[ruleSp]() {
									goto l979
								}
								if !p.rules[ruleTicks3]() {
									goto l979
								}
								goto l959
							l979:
								position = position979
							}
							{
								if position == len(p.Buffer) {
									goto l959
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l959
									}
									if !p.rules[ruleBlankLine]() {
										goto l981
									}
									goto l959
								l981:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l959
									}
									break
								default:
									goto l959
								}
							}
						}
					}
				l971:
					goto l958
				l959:
					position = position959
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l957
				}
				if !p.rules[ruleTicks3]() {
					goto l957
				}
				goto l906
			l957:
				position = position906
				if !p.rules[ruleTicks4]() {
					goto l982
				}
				if !p.rules[ruleSp]() {
					goto l982
				}
				begin = position
				if peekChar('`') {
					goto l986
				}
				if !p.rules[ruleNonspacechar]() {
					goto l986
				}
			l987:
				if peekChar('`') {
					goto l988
				}
				if !p.rules[ruleNonspacechar]() {
					goto l988
				}
				goto l987
			l988:
				goto l985
			l986:
				{
					if position == len(p.Buffer) {
						goto l982
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks4]() {
							goto l990
						}
						goto l982
					l990:
						if !matchChar('`') {
							goto l982
						}
					l991:
						if !matchChar('`') {
							goto l992
						}
						goto l991
					l992:
						break
					default:
						{
							position993 := position
							if !p.rules[ruleSp]() {
								goto l993
							}
							if !p.rules[ruleTicks4]() {
								goto l993
							}
							goto l982
						l993:
							position = position993
						}
						{
							if position == len(p.Buffer) {
								goto l982
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l982
								}
								if !p.rules[ruleBlankLine]() {
									goto l995
								}
								goto l982
							l995:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l982
								}
								break
							default:
								goto l982
							}
						}
					}
				}
			l985:
			l983:
				{
					position984 := position
					if peekChar('`') {
						goto l997
					}
					if !p.rules[ruleNonspacechar]() {
						goto l997
					}
				l998:
					if peekChar('`') {
						goto l999
					}
					if !p.rules[ruleNonspacechar]() {
						goto l999
					}
					goto l998
				l999:
					goto l996
				l997:
					{
						if position == len(p.Buffer) {
							goto l984
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks4]() {
								goto l1001
							}
							goto l984
						l1001:
							if !matchChar('`') {
								goto l984
							}
						l1002:
							if !matchChar('`') {
								goto l1003
							}
							goto l1002
						l1003:
							break
						default:
							{
								position1004 := position
								if !p.rules[ruleSp]() {
									goto l1004
								}
								if !p.rules[ruleTicks4]() {
									goto l1004
								}
								goto l984
							l1004:
								position = position1004
							}
							{
								if position == len(p.Buffer) {
									goto l984
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l984
									}
									if !p.rules[ruleBlankLine]() {
										goto l1006
									}
									goto l984
								l1006:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l984
									}
									break
								default:
									goto l984
								}
							}
						}
					}
				l996:
					goto l983
				l984:
					position = position984
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l982
				}
				if !p.rules[ruleTicks4]() {
					goto l982
				}
				goto l906
			l982:
				position = position906
				if !p.rules[ruleTicks5]() {
					goto l905
				}
				if !p.rules[ruleSp]() {
					goto l905
				}
				begin = position
				if peekChar('`') {
					goto l1010
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1010
				}
			l1011:
				if peekChar('`') {
					goto l1012
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1012
				}
				goto l1011
			l1012:
				goto l1009
			l1010:
				{
					if position == len(p.Buffer) {
						goto l905
					}
					switch p.Buffer[position] {
					case '`':
						if !p.rules[ruleTicks5]() {
							goto l1014
						}
						goto l905
					l1014:
						if !matchChar('`') {
							goto l905
						}
					l1015:
						if !matchChar('`') {
							goto l1016
						}
						goto l1015
					l1016:
						break
					default:
						{
							position1017 := position
							if !p.rules[ruleSp]() {
								goto l1017
							}
							if !p.rules[ruleTicks5]() {
								goto l1017
							}
							goto l905
						l1017:
							position = position1017
						}
						{
							if position == len(p.Buffer) {
								goto l905
							}
							switch p.Buffer[position] {
							case '\n', '\r':
								if !p.rules[ruleNewline]() {
									goto l905
								}
								if !p.rules[ruleBlankLine]() {
									goto l1019
								}
								goto l905
							l1019:
								break
							case '\t', ' ':
								if !p.rules[ruleSpacechar]() {
									goto l905
								}
								break
							default:
								goto l905
							}
						}
					}
				}
			l1009:
			l1007:
				{
					position1008 := position
					if peekChar('`') {
						goto l1021
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1021
					}
				l1022:
					if peekChar('`') {
						goto l1023
					}
					if !p.rules[ruleNonspacechar]() {
						goto l1023
					}
					goto l1022
				l1023:
					goto l1020
				l1021:
					{
						if position == len(p.Buffer) {
							goto l1008
						}
						switch p.Buffer[position] {
						case '`':
							if !p.rules[ruleTicks5]() {
								goto l1025
							}
							goto l1008
						l1025:
							if !matchChar('`') {
								goto l1008
							}
						l1026:
							if !matchChar('`') {
								goto l1027
							}
							goto l1026
						l1027:
							break
						default:
							{
								position1028 := position
								if !p.rules[ruleSp]() {
									goto l1028
								}
								if !p.rules[ruleTicks5]() {
									goto l1028
								}
								goto l1008
							l1028:
								position = position1028
							}
							{
								if position == len(p.Buffer) {
									goto l1008
								}
								switch p.Buffer[position] {
								case '\n', '\r':
									if !p.rules[ruleNewline]() {
										goto l1008
									}
									if !p.rules[ruleBlankLine]() {
										goto l1030
									}
									goto l1008
								l1030:
									break
								case '\t', ' ':
									if !p.rules[ruleSpacechar]() {
										goto l1008
									}
									break
								default:
									goto l1008
								}
							}
						}
					}
				l1020:
					goto l1007
				l1008:
					position = position1008
				}
				end = position
				if !p.rules[ruleSp]() {
					goto l905
				}
				if !p.rules[ruleTicks5]() {
					goto l905
				}
			}
		l906:
			do(84)
			return true
		l905:
			position = position0
			return false
		},
		/* 193 RawHtml <- (< (HtmlComment / HtmlBlockScript / HtmlTag) > {   if p.extension.FilterHTML {
                    yy = p.mkList(LIST, nil)
                } else {
                    yy = p.mkString(yytext)
                    yy.key = HTML
                }
            }) */
		func() bool {
			position0 := position
			begin = position
			if !p.rules[ruleHtmlComment]() {
				goto l1033
			}
			goto l1032
		l1033:
			if !p.rules[ruleHtmlBlockScript]() {
				goto l1034
			}
			goto l1032
		l1034:
			if !p.rules[ruleHtmlTag]() {
				goto l1031
			}
		l1032:
			end = position
			do(85)
			return true
		l1031:
			position = position0
			return false
		},
		/* 194 BlankLine <- (Sp Newline) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1035
			}
			if !p.rules[ruleNewline]() {
				goto l1035
			}
			return true
		l1035:
			position = position0
			return false
		},
		/* 195 Quoted <- ((&[\'] ('\'' (!'\'' .)* '\'')) | (&[\"] ('"' (!'"' .)* '"'))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1036
				}
				switch p.Buffer[position] {
				case '\'':
					position++ // matchChar
				l1038:
					if position == len(p.Buffer) {
						goto l1039
					}
					switch p.Buffer[position] {
					case '\'':
						goto l1039
					default:
						position++
					}
					goto l1038
				l1039:
					if !matchChar('\'') {
						goto l1036
					}
					break
				case '"':
					position++ // matchChar
				l1040:
					if position == len(p.Buffer) {
						goto l1041
					}
					switch p.Buffer[position] {
					case '"':
						goto l1041
					default:
						position++
					}
					goto l1040
				l1041:
					if !matchChar('"') {
						goto l1036
					}
					break
				default:
					goto l1036
				}
			}
			return true
		l1036:
			position = position0
			return false
		},
		/* 196 HtmlAttribute <- (((&[\-] '-') | (&[0-9A-Za-z] [A-Za-z0-9]))+ Spnl ('=' Spnl (Quoted / (!'>' Nonspacechar)+))? Spnl) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1042
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(5) {
						goto l1042
					}
				}
			}
		l1043:
			{
				if position == len(p.Buffer) {
					goto l1044
				}
				switch p.Buffer[position] {
				case '-':
					position++ // matchChar
					break
				default:
					if !matchClass(5) {
						goto l1044
					}
				}
			}
			goto l1043
		l1044:
			if !p.rules[ruleSpnl]() {
				goto l1042
			}
			{
				position1047 := position
				if !matchChar('=') {
					goto l1047
				}
				if !p.rules[ruleSpnl]() {
					goto l1047
				}
				if !p.rules[ruleQuoted]() {
					goto l1050
				}
				goto l1049
			l1050:
				if peekChar('>') {
					goto l1047
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1047
				}
			l1051:
				if peekChar('>') {
					goto l1052
				}
				if !p.rules[ruleNonspacechar]() {
					goto l1052
				}
				goto l1051
			l1052:
			l1049:
				goto l1048
			l1047:
				position = position1047
			}
		l1048:
			if !p.rules[ruleSpnl]() {
				goto l1042
			}
			return true
		l1042:
			position = position0
			return false
		},
		/* 197 HtmlComment <- ('<!--' (!'-->' .)* '-->') */
		func() bool {
			position0 := position
			if !matchString("<!--") {
				goto l1053
			}
		l1054:
			{
				position1055 := position
				if !matchString("-->") {
					goto l1056
				}
				goto l1055
			l1056:
				if !matchDot() {
					goto l1055
				}
				goto l1054
			l1055:
				position = position1055
			}
			if !matchString("-->") {
				goto l1053
			}
			return true
		l1053:
			position = position0
			return false
		},
		/* 198 HtmlTag <- ('<' Spnl '/'? [A-Za-z0-9]+ Spnl HtmlAttribute* '/'? Spnl '>') */
		func() bool {
			position0 := position
			if !matchChar('<') {
				goto l1057
			}
			if !p.rules[ruleSpnl]() {
				goto l1057
			}
			matchChar('/')
			if !matchClass(5) {
				goto l1057
			}
		l1058:
			if !matchClass(5) {
				goto l1059
			}
			goto l1058
		l1059:
			if !p.rules[ruleSpnl]() {
				goto l1057
			}
		l1060:
			if !p.rules[ruleHtmlAttribute]() {
				goto l1061
			}
			goto l1060
		l1061:
			matchChar('/')
			if !p.rules[ruleSpnl]() {
				goto l1057
			}
			if !matchChar('>') {
				goto l1057
			}
			return true
		l1057:
			position = position0
			return false
		},
		/* 199 Eof <- !. */
		func() bool {
			if (position < len(p.Buffer)) {
				goto l1062
			}
			return true
		l1062:
			return false
		},
		/* 200 Spacechar <- ((&[\t] '\t') | (&[ ] ' ')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1063
				}
				switch p.Buffer[position] {
				case '\t':
					position++ // matchChar
					break
				case ' ':
					position++ // matchChar
					break
				default:
					goto l1063
				}
			}
			return true
		l1063:
			return false
		},
		/* 201 Nonspacechar <- (!Spacechar !Newline .) */
		func() bool {
			position0 := position
			if !p.rules[ruleSpacechar]() {
				goto l1066
			}
			goto l1065
		l1066:
			if !p.rules[ruleNewline]() {
				goto l1067
			}
			goto l1065
		l1067:
			if !matchDot() {
				goto l1065
			}
			return true
		l1065:
			position = position0
			return false
		},
		/* 202 Newline <- ((&[\r] ('\r' '\n'?)) | (&[\n] '\n')) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1068
				}
				switch p.Buffer[position] {
				case '\r':
					position++ // matchChar
					matchChar('\n')
					break
				case '\n':
					position++ // matchChar
					break
				default:
					goto l1068
				}
			}
			return true
		l1068:
			position = position0
			return false
		},
		/* 203 Sp <- Spacechar* */
		func() bool {
		l1071:
			if !p.rules[ruleSpacechar]() {
				goto l1072
			}
			goto l1071
		l1072:
			return true
		},
		/* 204 Spnl <- (Sp (Newline Sp)?) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1073
			}
			{
				position1074 := position
				if !p.rules[ruleNewline]() {
					goto l1074
				}
				if !p.rules[ruleSp]() {
					goto l1074
				}
				goto l1075
			l1074:
				position = position1074
			}
		l1075:
			return true
		l1073:
			position = position0
			return false
		},
		/* 205 SpecialChar <- ('\'' / '"' / ((&[\\] '\\') | (&[#] '#') | (&[!] '!') | (&[<] '<') | (&[)] ')') | (&[(] '(') | (&[\]] ']') | (&[\[] '[') | (&[&] '&') | (&[`] '`') | (&[_] '_') | (&[*] '*') | (&[\"\'\-.^] ExtendedSpecialChar))) */
		func() bool {
			if !matchChar('\'') {
				goto l1078
			}
			goto l1077
		l1078:
			if !matchChar('"') {
				goto l1079
			}
			goto l1077
		l1079:
			{
				if position == len(p.Buffer) {
					goto l1076
				}
				switch p.Buffer[position] {
				case '\\':
					position++ // matchChar
					break
				case '#':
					position++ // matchChar
					break
				case '!':
					position++ // matchChar
					break
				case '<':
					position++ // matchChar
					break
				case ')':
					position++ // matchChar
					break
				case '(':
					position++ // matchChar
					break
				case ']':
					position++ // matchChar
					break
				case '[':
					position++ // matchChar
					break
				case '&':
					position++ // matchChar
					break
				case '`':
					position++ // matchChar
					break
				case '_':
					position++ // matchChar
					break
				case '*':
					position++ // matchChar
					break
				default:
					if !p.rules[ruleExtendedSpecialChar]() {
						goto l1076
					}
				}
			}
		l1077:
			return true
		l1076:
			return false
		},
		/* 206 NormalChar <- (!((&[\n\r] Newline) | (&[\t ] Spacechar) | (&[!-#&-*\-.<\[-`] SpecialChar)) .) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1082
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1082
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1082
					}
					break
				default:
					if !p.rules[ruleSpecialChar]() {
						goto l1082
					}
				}
			}
			goto l1081
		l1082:
			if !matchDot() {
				goto l1081
			}
			return true
		l1081:
			position = position0
			return false
		},
		/* 207 Alphanumeric <- ((&[\377] '\377') | (&[\376] '\376') | (&[\375] '\375') | (&[\374] '\374') | (&[\373] '\373') | (&[\372] '\372') | (&[\371] '\371') | (&[\370] '\370') | (&[\367] '\367') | (&[\366] '\366') | (&[\365] '\365') | (&[\364] '\364') | (&[\363] '\363') | (&[\362] '\362') | (&[\361] '\361') | (&[\360] '\360') | (&[\357] '\357') | (&[\356] '\356') | (&[\355] '\355') | (&[\354] '\354') | (&[\353] '\353') | (&[\352] '\352') | (&[\351] '\351') | (&[\350] '\350') | (&[\347] '\347') | (&[\346] '\346') | (&[\345] '\345') | (&[\344] '\344') | (&[\343] '\343') | (&[\342] '\342') | (&[\341] '\341') | (&[\340] '\340') | (&[\337] '\337') | (&[\336] '\336') | (&[\335] '\335') | (&[\334] '\334') | (&[\333] '\333') | (&[\332] '\332') | (&[\331] '\331') | (&[\330] '\330') | (&[\327] '\327') | (&[\326] '\326') | (&[\325] '\325') | (&[\324] '\324') | (&[\323] '\323') | (&[\322] '\322') | (&[\321] '\321') | (&[\320] '\320') | (&[\317] '\317') | (&[\316] '\316') | (&[\315] '\315') | (&[\314] '\314') | (&[\313] '\313') | (&[\312] '\312') | (&[\311] '\311') | (&[\310] '\310') | (&[\307] '\307') | (&[\306] '\306') | (&[\305] '\305') | (&[\304] '\304') | (&[\303] '\303') | (&[\302] '\302') | (&[\301] '\301') | (&[\300] '\300') | (&[\277] '\277') | (&[\276] '\276') | (&[\275] '\275') | (&[\274] '\274') | (&[\273] '\273') | (&[\272] '\272') | (&[\271] '\271') | (&[\270] '\270') | (&[\267] '\267') | (&[\266] '\266') | (&[\265] '\265') | (&[\264] '\264') | (&[\263] '\263') | (&[\262] '\262') | (&[\261] '\261') | (&[\260] '\260') | (&[\257] '\257') | (&[\256] '\256') | (&[\255] '\255') | (&[\254] '\254') | (&[\253] '\253') | (&[\252] '\252') | (&[\251] '\251') | (&[\250] '\250') | (&[\247] '\247') | (&[\246] '\246') | (&[\245] '\245') | (&[\244] '\244') | (&[\243] '\243') | (&[\242] '\242') | (&[\241] '\241') | (&[\240] '\240') | (&[\237] '\237') | (&[\236] '\236') | (&[\235] '\235') | (&[\234] '\234') | (&[\233] '\233') | (&[\232] '\232') | (&[\231] '\231') | (&[\230] '\230') | (&[\227] '\227') | (&[\226] '\226') | (&[\225] '\225') | (&[\224] '\224') | (&[\223] '\223') | (&[\222] '\222') | (&[\221] '\221') | (&[\220] '\220') | (&[\217] '\217') | (&[\216] '\216') | (&[\215] '\215') | (&[\214] '\214') | (&[\213] '\213') | (&[\212] '\212') | (&[\211] '\211') | (&[\210] '\210') | (&[\207] '\207') | (&[\206] '\206') | (&[\205] '\205') | (&[\204] '\204') | (&[\203] '\203') | (&[\202] '\202') | (&[\201] '\201') | (&[\200] '\200') | (&[0-9A-Za-z] [0-9A-Za-z])) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1084
				}
				switch p.Buffer[position] {
				case '\377':
					position++ // matchChar
					break
				case '\376':
					position++ // matchChar
					break
				case '\375':
					position++ // matchChar
					break
				case '\374':
					position++ // matchChar
					break
				case '\373':
					position++ // matchChar
					break
				case '\372':
					position++ // matchChar
					break
				case '\371':
					position++ // matchChar
					break
				case '\370':
					position++ // matchChar
					break
				case '\367':
					position++ // matchChar
					break
				case '\366':
					position++ // matchChar
					break
				case '\365':
					position++ // matchChar
					break
				case '\364':
					position++ // matchChar
					break
				case '\363':
					position++ // matchChar
					break
				case '\362':
					position++ // matchChar
					break
				case '\361':
					position++ // matchChar
					break
				case '\360':
					position++ // matchChar
					break
				case '\357':
					position++ // matchChar
					break
				case '\356':
					position++ // matchChar
					break
				case '\355':
					position++ // matchChar
					break
				case '\354':
					position++ // matchChar
					break
				case '\353':
					position++ // matchChar
					break
				case '\352':
					position++ // matchChar
					break
				case '\351':
					position++ // matchChar
					break
				case '\350':
					position++ // matchChar
					break
				case '\347':
					position++ // matchChar
					break
				case '\346':
					position++ // matchChar
					break
				case '\345':
					position++ // matchChar
					break
				case '\344':
					position++ // matchChar
					break
				case '\343':
					position++ // matchChar
					break
				case '\342':
					position++ // matchChar
					break
				case '\341':
					position++ // matchChar
					break
				case '\340':
					position++ // matchChar
					break
				case '\337':
					position++ // matchChar
					break
				case '\336':
					position++ // matchChar
					break
				case '\335':
					position++ // matchChar
					break
				case '\334':
					position++ // matchChar
					break
				case '\333':
					position++ // matchChar
					break
				case '\332':
					position++ // matchChar
					break
				case '\331':
					position++ // matchChar
					break
				case '\330':
					position++ // matchChar
					break
				case '\327':
					position++ // matchChar
					break
				case '\326':
					position++ // matchChar
					break
				case '\325':
					position++ // matchChar
					break
				case '\324':
					position++ // matchChar
					break
				case '\323':
					position++ // matchChar
					break
				case '\322':
					position++ // matchChar
					break
				case '\321':
					position++ // matchChar
					break
				case '\320':
					position++ // matchChar
					break
				case '\317':
					position++ // matchChar
					break
				case '\316':
					position++ // matchChar
					break
				case '\315':
					position++ // matchChar
					break
				case '\314':
					position++ // matchChar
					break
				case '\313':
					position++ // matchChar
					break
				case '\312':
					position++ // matchChar
					break
				case '\311':
					position++ // matchChar
					break
				case '\310':
					position++ // matchChar
					break
				case '\307':
					position++ // matchChar
					break
				case '\306':
					position++ // matchChar
					break
				case '\305':
					position++ // matchChar
					break
				case '\304':
					position++ // matchChar
					break
				case '\303':
					position++ // matchChar
					break
				case '\302':
					position++ // matchChar
					break
				case '\301':
					position++ // matchChar
					break
				case '\300':
					position++ // matchChar
					break
				case '\277':
					position++ // matchChar
					break
				case '\276':
					position++ // matchChar
					break
				case '\275':
					position++ // matchChar
					break
				case '\274':
					position++ // matchChar
					break
				case '\273':
					position++ // matchChar
					break
				case '\272':
					position++ // matchChar
					break
				case '\271':
					position++ // matchChar
					break
				case '\270':
					position++ // matchChar
					break
				case '\267':
					position++ // matchChar
					break
				case '\266':
					position++ // matchChar
					break
				case '\265':
					position++ // matchChar
					break
				case '\264':
					position++ // matchChar
					break
				case '\263':
					position++ // matchChar
					break
				case '\262':
					position++ // matchChar
					break
				case '\261':
					position++ // matchChar
					break
				case '\260':
					position++ // matchChar
					break
				case '\257':
					position++ // matchChar
					break
				case '\256':
					position++ // matchChar
					break
				case '\255':
					position++ // matchChar
					break
				case '\254':
					position++ // matchChar
					break
				case '\253':
					position++ // matchChar
					break
				case '\252':
					position++ // matchChar
					break
				case '\251':
					position++ // matchChar
					break
				case '\250':
					position++ // matchChar
					break
				case '\247':
					position++ // matchChar
					break
				case '\246':
					position++ // matchChar
					break
				case '\245':
					position++ // matchChar
					break
				case '\244':
					position++ // matchChar
					break
				case '\243':
					position++ // matchChar
					break
				case '\242':
					position++ // matchChar
					break
				case '\241':
					position++ // matchChar
					break
				case '\240':
					position++ // matchChar
					break
				case '\237':
					position++ // matchChar
					break
				case '\236':
					position++ // matchChar
					break
				case '\235':
					position++ // matchChar
					break
				case '\234':
					position++ // matchChar
					break
				case '\233':
					position++ // matchChar
					break
				case '\232':
					position++ // matchChar
					break
				case '\231':
					position++ // matchChar
					break
				case '\230':
					position++ // matchChar
					break
				case '\227':
					position++ // matchChar
					break
				case '\226':
					position++ // matchChar
					break
				case '\225':
					position++ // matchChar
					break
				case '\224':
					position++ // matchChar
					break
				case '\223':
					position++ // matchChar
					break
				case '\222':
					position++ // matchChar
					break
				case '\221':
					position++ // matchChar
					break
				case '\220':
					position++ // matchChar
					break
				case '\217':
					position++ // matchChar
					break
				case '\216':
					position++ // matchChar
					break
				case '\215':
					position++ // matchChar
					break
				case '\214':
					position++ // matchChar
					break
				case '\213':
					position++ // matchChar
					break
				case '\212':
					position++ // matchChar
					break
				case '\211':
					position++ // matchChar
					break
				case '\210':
					position++ // matchChar
					break
				case '\207':
					position++ // matchChar
					break
				case '\206':
					position++ // matchChar
					break
				case '\205':
					position++ // matchChar
					break
				case '\204':
					position++ // matchChar
					break
				case '\203':
					position++ // matchChar
					break
				case '\202':
					position++ // matchChar
					break
				case '\201':
					position++ // matchChar
					break
				case '\200':
					position++ // matchChar
					break
				default:
					if !matchClass(4) {
						goto l1084
					}
				}
			}
			return true
		l1084:
			return false
		},
		/* 208 AlphanumericAscii <- [A-Za-z0-9] */
		func() bool {
			if !matchClass(5) {
				goto l1086
			}
			return true
		l1086:
			return false
		},
		/* 209 Digit <- [0-9] */
		func() bool {
			if !matchClass(0) {
				goto l1087
			}
			return true
		l1087:
			return false
		},
		/* 210 HexEntity <- (< '&' '#' [Xx] [0-9a-fA-F]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1088
			}
			if !matchChar('#') {
				goto l1088
			}
			if !matchClass(6) {
				goto l1088
			}
			if !matchClass(7) {
				goto l1088
			}
		l1089:
			if !matchClass(7) {
				goto l1090
			}
			goto l1089
		l1090:
			if !matchChar(';') {
				goto l1088
			}
			end = position
			return true
		l1088:
			position = position0
			return false
		},
		/* 211 DecEntity <- (< '&' '#' [0-9]+ > ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1091
			}
			if !matchChar('#') {
				goto l1091
			}
			if !matchClass(0) {
				goto l1091
			}
		l1092:
			if !matchClass(0) {
				goto l1093
			}
			goto l1092
		l1093:
			end = position
			if !matchChar(';') {
				goto l1091
			}
			end = position
			return true
		l1091:
			position = position0
			return false
		},
		/* 212 CharEntity <- (< '&' [A-Za-z0-9]+ ';' >) */
		func() bool {
			position0 := position
			begin = position
			if !matchChar('&') {
				goto l1094
			}
			if !matchClass(5) {
				goto l1094
			}
		l1095:
			if !matchClass(5) {
				goto l1096
			}
			goto l1095
		l1096:
			if !matchChar(';') {
				goto l1094
			}
			end = position
			return true
		l1094:
			position = position0
			return false
		},
		/* 213 NonindentSpace <- ('   ' / '  ' / ' ' / '') */
		func() bool {
			if !matchString("   ") {
				goto l1099
			}
			goto l1098
		l1099:
			if !matchString("  ") {
				goto l1100
			}
			goto l1098
		l1100:
			if !matchChar(' ') {
				goto l1101
			}
			goto l1098
		l1101:
		l1098:
			return true
		},
		/* 214 Indent <- ((&[ ] '    ') | (&[\t] '\t')) */
		func() bool {
			{
				if position == len(p.Buffer) {
					goto l1102
				}
				switch p.Buffer[position] {
				case ' ':
					position++
					if !matchString("   ") {
						goto l1102
					}
					break
				case '\t':
					position++ // matchChar
					break
				default:
					goto l1102
				}
			}
			return true
		l1102:
			return false
		},
		/* 215 IndentedLine <- (Indent Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1104
			}
			if !p.rules[ruleLine]() {
				goto l1104
			}
			return true
		l1104:
			position = position0
			return false
		},
		/* 216 OptionallyIndentedLine <- (Indent? Line) */
		func() bool {
			position0 := position
			if !p.rules[ruleIndent]() {
				goto l1106
			}
		l1106:
			if !p.rules[ruleLine]() {
				goto l1105
			}
			return true
		l1105:
			position = position0
			return false
		},
		/* 217 StartList <- (&. { yy = nil }) */
		func() bool {
			if !(position < len(p.Buffer)) {
				goto l1108
			}
			do(86)
			return true
		l1108:
			return false
		},
		/* 218 Line <- (RawLine { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleRawLine]() {
				goto l1109
			}
			do(87)
			return true
		l1109:
			position = position0
			return false
		},
		/* 219 RawLine <- ((< (!'\r' !'\n' .)* Newline >) / (< .+ > !.)) */
		func() bool {
			position0 := position
			{
				position1111 := position
				begin = position
			l1113:
				if position == len(p.Buffer) {
					goto l1114
				}
				switch p.Buffer[position] {
				case '\r', '\n':
					goto l1114
				default:
					position++
				}
				goto l1113
			l1114:
				if !p.rules[ruleNewline]() {
					goto l1112
				}
				end = position
				goto l1111
			l1112:
				position = position1111
				begin = position
				if !matchDot() {
					goto l1110
				}
			l1115:
				if !matchDot() {
					goto l1116
				}
				goto l1115
			l1116:
				end = position
				if (position < len(p.Buffer)) {
					goto l1110
				}
			}
		l1111:
			return true
		l1110:
			position = position0
			return false
		},
		/* 220 SkipBlock <- (HtmlBlock / ((!'#' !SetextBottom1 !SetextBottom2 !BlankLine RawLine)+ BlankLine*) / BlankLine+ / RawLine) */
		func() bool {
			position0 := position
			{
				position1118 := position
				if !p.rules[ruleHtmlBlock]() {
					goto l1119
				}
				goto l1118
			l1119:
				if peekChar('#') {
					goto l1120
				}
				if !p.rules[ruleSetextBottom1]() {
					goto l1123
				}
				goto l1120
			l1123:
				if !p.rules[ruleSetextBottom2]() {
					goto l1124
				}
				goto l1120
			l1124:
				if !p.rules[ruleBlankLine]() {
					goto l1125
				}
				goto l1120
			l1125:
				if !p.rules[ruleRawLine]() {
					goto l1120
				}
			l1121:
				{
					position1122 := position
					if peekChar('#') {
						goto l1122
					}
					if !p.rules[ruleSetextBottom1]() {
						goto l1126
					}
					goto l1122
				l1126:
					if !p.rules[ruleSetextBottom2]() {
						goto l1127
					}
					goto l1122
				l1127:
					if !p.rules[ruleBlankLine]() {
						goto l1128
					}
					goto l1122
				l1128:
					if !p.rules[ruleRawLine]() {
						goto l1122
					}
					goto l1121
				l1122:
					position = position1122
				}
			l1129:
				if !p.rules[ruleBlankLine]() {
					goto l1130
				}
				goto l1129
			l1130:
				goto l1118
			l1120:
				position = position1118
				if !p.rules[ruleBlankLine]() {
					goto l1131
				}
			l1132:
				if !p.rules[ruleBlankLine]() {
					goto l1133
				}
				goto l1132
			l1133:
				goto l1118
			l1131:
				position = position1118
				if !p.rules[ruleRawLine]() {
					goto l1117
				}
			}
		l1118:
			return true
		l1117:
			position = position0
			return false
		},
		/* 221 ExtendedSpecialChar <- ((&[^] (&{p.extension.Notes} '^')) | (&[\"\'\-.] (&{p.extension.Smart} ((&[\"] '"') | (&[\'] '\'') | (&[\-] '-') | (&[.] '.'))))) */
		func() bool {
			position0 := position
			{
				if position == len(p.Buffer) {
					goto l1134
				}
				switch p.Buffer[position] {
				case '^':
					if !(p.extension.Notes) {
						goto l1134
					}
					if !matchChar('^') {
						goto l1134
					}
					break
				default:
					if !(p.extension.Smart) {
						goto l1134
					}
					{
						if position == len(p.Buffer) {
							goto l1134
						}
						switch p.Buffer[position] {
						case '"':
							position++ // matchChar
							break
						case '\'':
							position++ // matchChar
							break
						case '-':
							position++ // matchChar
							break
						case '.':
							position++ // matchChar
							break
						default:
							goto l1134
						}
					}
				}
			}
			return true
		l1134:
			position = position0
			return false
		},
		/* 222 Smart <- (&{p.extension.Smart} (SingleQuoted / ((&[\'] Apostrophe) | (&[\"] DoubleQuoted) | (&[\-] Dash) | (&[.] Ellipsis)))) */
		func() bool {
			if !(p.extension.Smart) {
				goto l1137
			}
			if !p.rules[ruleSingleQuoted]() {
				goto l1139
			}
			goto l1138
		l1139:
			{
				if position == len(p.Buffer) {
					goto l1137
				}
				switch p.Buffer[position] {
				case '\'':
					if !p.rules[ruleApostrophe]() {
						goto l1137
					}
					break
				case '"':
					if !p.rules[ruleDoubleQuoted]() {
						goto l1137
					}
					break
				case '-':
					if !p.rules[ruleDash]() {
						goto l1137
					}
					break
				case '.':
					if !p.rules[ruleEllipsis]() {
						goto l1137
					}
					break
				default:
					goto l1137
				}
			}
		l1138:
			return true
		l1137:
			return false
		},
		/* 223 Apostrophe <- ('\'' { yy = p.mkElem(APOSTROPHE) }) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1141
			}
			do(88)
			return true
		l1141:
			position = position0
			return false
		},
		/* 224 Ellipsis <- (('...' / '. . .') { yy = p.mkElem(ELLIPSIS) }) */
		func() bool {
			position0 := position
			if !matchString("...") {
				goto l1144
			}
			goto l1143
		l1144:
			if !matchString(". . .") {
				goto l1142
			}
		l1143:
			do(89)
			return true
		l1142:
			position = position0
			return false
		},
		/* 225 Dash <- (EmDash / EnDash) */
		func() bool {
			if !p.rules[ruleEmDash]() {
				goto l1147
			}
			goto l1146
		l1147:
			if !p.rules[ruleEnDash]() {
				goto l1145
			}
		l1146:
			return true
		l1145:
			return false
		},
		/* 226 EnDash <- ('-' &[0-9] { yy = p.mkElem(ENDASH) }) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l1148
			}
			if !peekClass(0) {
				goto l1148
			}
			do(90)
			return true
		l1148:
			position = position0
			return false
		},
		/* 227 EmDash <- (('---' / '--') { yy = p.mkElem(EMDASH) }) */
		func() bool {
			position0 := position
			if !matchString("---") {
				goto l1151
			}
			goto l1150
		l1151:
			if !matchString("--") {
				goto l1149
			}
		l1150:
			do(91)
			return true
		l1149:
			position = position0
			return false
		},
		/* 228 SingleQuoteStart <- ('\'' !((&[\n\r] Newline) | (&[\t ] Spacechar))) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1152
			}
			{
				if position == len(p.Buffer) {
					goto l1153
				}
				switch p.Buffer[position] {
				case '\n', '\r':
					if !p.rules[ruleNewline]() {
						goto l1153
					}
					break
				case '\t', ' ':
					if !p.rules[ruleSpacechar]() {
						goto l1153
					}
					break
				default:
					goto l1153
				}
			}
			goto l1152
		l1153:
			return true
		l1152:
			position = position0
			return false
		},
		/* 229 SingleQuoteEnd <- ('\'' !Alphanumeric) */
		func() bool {
			position0 := position
			if !matchChar('\'') {
				goto l1155
			}
			if !p.rules[ruleAlphanumeric]() {
				goto l1156
			}
			goto l1155
		l1156:
			return true
		l1155:
			position = position0
			return false
		},
		/* 230 SingleQuoted <- (SingleQuoteStart StartList (!SingleQuoteEnd Inline { a = cons(b, a) })+ SingleQuoteEnd { yy = p.mkList(SINGLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleSingleQuoteStart]() {
				goto l1157
			}
			if !p.rules[ruleStartList]() {
				goto l1157
			}
			doarg(yySet, -2)
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1160
			}
			goto l1157
		l1160:
			if !p.rules[ruleInline]() {
				goto l1157
			}
			doarg(yySet, -1)
			do(92)
		l1158:
			{
				position1159, thunkPosition1159 := position, thunkPosition
				if !p.rules[ruleSingleQuoteEnd]() {
					goto l1161
				}
				goto l1159
			l1161:
				if !p.rules[ruleInline]() {
					goto l1159
				}
				doarg(yySet, -1)
				do(92)
				goto l1158
			l1159:
				position, thunkPosition = position1159, thunkPosition1159
			}
			if !p.rules[ruleSingleQuoteEnd]() {
				goto l1157
			}
			do(93)
			doarg(yyPop, 2)
			return true
		l1157:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 231 DoubleQuoteStart <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1162
			}
			return true
		l1162:
			return false
		},
		/* 232 DoubleQuoteEnd <- '"' */
		func() bool {
			if !matchChar('"') {
				goto l1163
			}
			return true
		l1163:
			return false
		},
		/* 233 DoubleQuoted <- ('"' StartList (!'"' Inline { a = cons(b, a) })+ '"' { yy = p.mkList(DOUBLEQUOTED, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('"') {
				goto l1164
			}
			if !p.rules[ruleStartList]() {
				goto l1164
			}
			doarg(yySet, -2)
			if peekChar('"') {
				goto l1164
			}
			if !p.rules[ruleInline]() {
				goto l1164
			}
			doarg(yySet, -1)
			do(94)
		l1165:
			{
				position1166, thunkPosition1166 := position, thunkPosition
				if peekChar('"') {
					goto l1166
				}
				if !p.rules[ruleInline]() {
					goto l1166
				}
				doarg(yySet, -1)
				do(94)
				goto l1165
			l1166:
				position, thunkPosition = position1166, thunkPosition1166
			}
			if !matchChar('"') {
				goto l1164
			}
			do(95)
			doarg(yyPop, 2)
			return true
		l1164:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 234 NoteReference <- (&{p.extension.Notes} RawNoteReference {
                    if match, ok := p.find_note(ref.contents.str); ok {
                        yy = p.mkElem(NOTE)
                        yy.children = match.children
                        yy.contents.str = ""
                    } else {
                        yy = p.mkString("[^"+ref.contents.str+"]")
                    }
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1167
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1167
			}
			doarg(yySet, -1)
			do(96)
			doarg(yyPop, 1)
			return true
		l1167:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 235 RawNoteReference <- ('[^' < (!Newline !']' .)+ > ']' { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			if !matchString("[^") {
				goto l1168
			}
			begin = position
			if !p.rules[ruleNewline]() {
				goto l1171
			}
			goto l1168
		l1171:
			if peekChar(']') {
				goto l1168
			}
			if !matchDot() {
				goto l1168
			}
		l1169:
			{
				position1170 := position
				if !p.rules[ruleNewline]() {
					goto l1172
				}
				goto l1170
			l1172:
				if peekChar(']') {
					goto l1170
				}
				if !matchDot() {
					goto l1170
				}
				goto l1169
			l1170:
				position = position1170
			}
			end = position
			if !matchChar(']') {
				goto l1168
			}
			do(97)
			return true
		l1168:
			position = position0
			return false
		},
		/* 236 Note <- (&{p.extension.Notes} NonindentSpace RawNoteReference ':' Sp StartList (RawNoteBlock { a = cons(yy, a) }) (&Indent RawNoteBlock { a = cons(yy, a) })* {   yy = p.mkList(NOTE, a)
                    yy.contents.str = ref.contents.str
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !(p.extension.Notes) {
				goto l1173
			}
			if !p.rules[ruleNonindentSpace]() {
				goto l1173
			}
			if !p.rules[ruleRawNoteReference]() {
				goto l1173
			}
			doarg(yySet, -2)
			if !matchChar(':') {
				goto l1173
			}
			if !p.rules[ruleSp]() {
				goto l1173
			}
			if !p.rules[ruleStartList]() {
				goto l1173
			}
			doarg(yySet, -1)
			if !p.rules[ruleRawNoteBlock]() {
				goto l1173
			}
			do(98)
		l1174:
			{
				position1175, thunkPosition1175 := position, thunkPosition
				{
					position1176 := position
					if !p.rules[ruleIndent]() {
						goto l1175
					}
					position = position1176
				}
				if !p.rules[ruleRawNoteBlock]() {
					goto l1175
				}
				do(99)
				goto l1174
			l1175:
				position, thunkPosition = position1175, thunkPosition1175
			}
			do(100)
			doarg(yyPop, 2)
			return true
		l1173:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 237 InlineNote <- (&{p.extension.Notes} '^[' StartList (!']' Inline { a = cons(yy, a) })+ ']' { yy = p.mkList(NOTE, a)
                  yy.contents.str = "" }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Notes) {
				goto l1177
			}
			if !matchString("^[") {
				goto l1177
			}
			if !p.rules[ruleStartList]() {
				goto l1177
			}
			doarg(yySet, -1)
			if peekChar(']') {
				goto l1177
			}
			if !p.rules[ruleInline]() {
				goto l1177
			}
			do(101)
		l1178:
			{
				position1179 := position
				if peekChar(']') {
					goto l1179
				}
				if !p.rules[ruleInline]() {
					goto l1179
				}
				do(101)
				goto l1178
			l1179:
				position = position1179
			}
			if !matchChar(']') {
				goto l1177
			}
			do(102)
			doarg(yyPop, 1)
			return true
		l1177:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 238 Notes <- (StartList ((Note { a = cons(b, a) }) / SkipBlock)* { p.notes = reverse(a) } commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1180
			}
			doarg(yySet, -1)
		l1181:
			{
				position1182, thunkPosition1182 := position, thunkPosition
				{
					position1183, thunkPosition1183 := position, thunkPosition
					if !p.rules[ruleNote]() {
						goto l1184
					}
					doarg(yySet, -2)
					do(103)
					goto l1183
				l1184:
					position, thunkPosition = position1183, thunkPosition1183
					if !p.rules[ruleSkipBlock]() {
						goto l1182
					}
				}
			l1183:
				goto l1181
			l1182:
				position, thunkPosition = position1182, thunkPosition1182
			}
			do(104)
			if !(commit(thunkPosition0)) {
				goto l1180
			}
			doarg(yyPop, 2)
			return true
		l1180:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 239 RawNoteBlock <- (StartList (!BlankLine OptionallyIndentedLine { a = cons(yy, a) })+ (< BlankLine* > { a = cons(p.mkString(yytext), a) }) {   yy = p.mkStringFromList(a, true)
                    yy.key = RAW
                }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1185
			}
			doarg(yySet, -1)
			if !p.rules[ruleBlankLine]() {
				goto l1188
			}
			goto l1185
		l1188:
			if !p.rules[ruleOptionallyIndentedLine]() {
				goto l1185
			}
			do(105)
		l1186:
			{
				position1187 := position
				if !p.rules[ruleBlankLine]() {
					goto l1189
				}
				goto l1187
			l1189:
				if !p.rules[ruleOptionallyIndentedLine]() {
					goto l1187
				}
				do(105)
				goto l1186
			l1187:
				position = position1187
			}
			begin = position
		l1190:
			if !p.rules[ruleBlankLine]() {
				goto l1191
			}
			goto l1190
		l1191:
			end = position
			do(106)
			do(107)
			doarg(yyPop, 1)
			return true
		l1185:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 240 DefinitionList <- (&{p.extension.Dlists} StartList (Definition { a = cons(yy, a) })+ { yy = p.mkList(DEFINITIONLIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !(p.extension.Dlists) {
				goto l1192
			}
			if !p.rules[ruleStartList]() {
				goto l1192
			}
			doarg(yySet, -1)
			if !p.rules[ruleDefinition]() {
				goto l1192
			}
			do(108)
		l1193:
			{
				position1194, thunkPosition1194 := position, thunkPosition
				if !p.rules[ruleDefinition]() {
					goto l1194
				}
				do(108)
				goto l1193
			l1194:
				position, thunkPosition = position1194, thunkPosition1194
			}
			do(109)
			doarg(yyPop, 1)
			return true
		l1192:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 241 Definition <- (&(NonindentSpace !Defmark Nonspacechar RawLine BlankLine? Defmark) StartList (DListTitle { a = cons(yy, a) })+ (DefTight / DefLoose) {
				for e := yy.children; e != nil; e = e.next {
					e.key = DEFDATA
				}
				a = cons(yy, a)
			} { yy = p.mkList(LIST, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			{
				position1196 := position
				if !p.rules[ruleNonindentSpace]() {
					goto l1195
				}
				if !p.rules[ruleDefmark]() {
					goto l1197
				}
				goto l1195
			l1197:
				if !p.rules[ruleNonspacechar]() {
					goto l1195
				}
				if !p.rules[ruleRawLine]() {
					goto l1195
				}
				if !p.rules[ruleBlankLine]() {
					goto l1198
				}
			l1198:
				if !p.rules[ruleDefmark]() {
					goto l1195
				}
				position = position1196
			}
			if !p.rules[ruleStartList]() {
				goto l1195
			}
			doarg(yySet, -1)
			if !p.rules[ruleDListTitle]() {
				goto l1195
			}
			do(110)
		l1200:
			{
				position1201, thunkPosition1201 := position, thunkPosition
				if !p.rules[ruleDListTitle]() {
					goto l1201
				}
				do(110)
				goto l1200
			l1201:
				position, thunkPosition = position1201, thunkPosition1201
			}
			if !p.rules[ruleDefTight]() {
				goto l1203
			}
			goto l1202
		l1203:
			if !p.rules[ruleDefLoose]() {
				goto l1195
			}
		l1202:
			do(111)
			do(112)
			doarg(yyPop, 1)
			return true
		l1195:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 242 DListTitle <- (NonindentSpace !Defmark &Nonspacechar StartList (!Endline Inline { a = cons(yy, a) })+ Sp Newline {	yy = p.mkList(LIST, a)
				yy.key = DEFTITLE
			}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleNonindentSpace]() {
				goto l1204
			}
			if !p.rules[ruleDefmark]() {
				goto l1205
			}
			goto l1204
		l1205:
			{
				position1206 := position
				if !p.rules[ruleNonspacechar]() {
					goto l1204
				}
				position = position1206
			}
			if !p.rules[ruleStartList]() {
				goto l1204
			}
			doarg(yySet, -1)
			if !p.rules[ruleEndline]() {
				goto l1209
			}
			goto l1204
		l1209:
			if !p.rules[ruleInline]() {
				goto l1204
			}
			do(113)
		l1207:
			{
				position1208 := position
				if !p.rules[ruleEndline]() {
					goto l1210
				}
				goto l1208
			l1210:
				if !p.rules[ruleInline]() {
					goto l1208
				}
				do(113)
				goto l1207
			l1208:
				position = position1208
			}
			if !p.rules[ruleSp]() {
				goto l1204
			}
			if !p.rules[ruleNewline]() {
				goto l1204
			}
			do(114)
			doarg(yyPop, 1)
			return true
		l1204:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 243 DefTight <- (&Defmark ListTight) */
		func() bool {
			{
				position1212 := position
				if !p.rules[ruleDefmark]() {
					goto l1211
				}
				position = position1212
			}
			if !p.rules[ruleListTight]() {
				goto l1211
			}
			return true
		l1211:
			return false
		},
		/* 244 DefLoose <- (BlankLine &Defmark ListLoose) */
		func() bool {
			position0 := position
			if !p.rules[ruleBlankLine]() {
				goto l1213
			}
			{
				position1214 := position
				if !p.rules[ruleDefmark]() {
					goto l1213
				}
				position = position1214
			}
			if !p.rules[ruleListLoose]() {
				goto l1213
			}
			return true
		l1213:
			position = position0
			return false
		},
		/* 245 Defmark <- (NonindentSpace ((&[~] '~') | (&[:] ':')) Spacechar+) */
		func() bool {
			position0 := position
			if !p.rules[ruleNonindentSpace]() {
				goto l1215
			}
			{
				if position == len(p.Buffer) {
					goto l1215
				}
				switch p.Buffer[position] {
				case '~':
					position++ // matchChar
					break
				case ':':
					position++ // matchChar
					break
				default:
					goto l1215
				}
			}
			if !p.rules[ruleSpacechar]() {
				goto l1215
			}
		l1217:
			if !p.rules[ruleSpacechar]() {
				goto l1218
			}
			goto l1217
		l1218:
			return true
		l1215:
			position = position0
			return false
		},
		/* 246 DefMarker <- (&{p.extension.Dlists} Defmark) */
		func() bool {
			if !(p.extension.Dlists) {
				goto l1219
			}
			if !p.rules[ruleDefmark]() {
				goto l1219
			}
			return true
		l1219:
			return false
		},
		/* 247 Table <- (StartList StartList (TableCaption { b = cons(yy, b) })? TableBody { yy.key = TABLEHEAD; a = cons(yy, a) } (SeparatorLine { append_list(yy, a) }) (TableBody { a = cons(yy, a) }) (BlankLine !TableCaption TableBody { a = cons(yy, a) } &(TableCaption / BlankLine))* ((TableCaption { b = cons(yy, b) } &BlankLine) / &BlankLine) {
        if b != nil { append_list(b,a) }
        yy = p.mkList(TABLE, a)
    }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleStartList]() {
				goto l1220
			}
			doarg(yySet, -2)
			if !p.rules[ruleStartList]() {
				goto l1220
			}
			doarg(yySet, -1)
			{
				position1221, thunkPosition1221 := position, thunkPosition
				if !p.rules[ruleTableCaption]() {
					goto l1221
				}
				do(115)
				goto l1222
			l1221:
				position, thunkPosition = position1221, thunkPosition1221
			}
		l1222:
			if !p.rules[ruleTableBody]() {
				goto l1220
			}
			do(116)
			if !p.rules[ruleSeparatorLine]() {
				goto l1220
			}
			do(117)
			if !p.rules[ruleTableBody]() {
				goto l1220
			}
			do(118)
		l1223:
			{
				position1224, thunkPosition1224 := position, thunkPosition
				if !p.rules[ruleBlankLine]() {
					goto l1224
				}
				if !p.rules[ruleTableCaption]() {
					goto l1225
				}
				goto l1224
			l1225:
				if !p.rules[ruleTableBody]() {
					goto l1224
				}
				do(119)
				{
					position1226, thunkPosition1226 := position, thunkPosition
					if !p.rules[ruleTableCaption]() {
						goto l1228
					}
					goto l1227
				l1228:
					if !p.rules[ruleBlankLine]() {
						goto l1224
					}
				l1227:
					position, thunkPosition = position1226, thunkPosition1226
				}
				goto l1223
			l1224:
				position, thunkPosition = position1224, thunkPosition1224
			}
			{
				position1229, thunkPosition1229 := position, thunkPosition
				if !p.rules[ruleTableCaption]() {
					goto l1230
				}
				do(120)
				{
					position1231 := position
					if !p.rules[ruleBlankLine]() {
						goto l1230
					}
					position = position1231
				}
				goto l1229
			l1230:
				position, thunkPosition = position1229, thunkPosition1229
				{
					position1232 := position
					if !p.rules[ruleBlankLine]() {
						goto l1220
					}
					position = position1232
				}
			}
		l1229:
			do(121)
			doarg(yyPop, 2)
			return true
		l1220:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 248 TableBody <- (StartList (TableRow { a = cons(yy, a) })+ { yy = p.mkList(TABLEBODY, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1233
			}
			doarg(yySet, -1)
			if !p.rules[ruleTableRow]() {
				goto l1233
			}
			do(122)
		l1234:
			{
				position1235, thunkPosition1235 := position, thunkPosition
				if !p.rules[ruleTableRow]() {
					goto l1235
				}
				do(122)
				goto l1234
			l1235:
				position, thunkPosition = position1235, thunkPosition1235
			}
			do(123)
			doarg(yyPop, 1)
			return true
		l1233:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 249 TableRow <- (StartList (!SeparatorLine &TableLine '|'? (TableCell { a = cons(yy, a) })+) Sp Newline { yy = p.mkList(TABLEROW, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1236
			}
			doarg(yySet, -1)
			if !p.rules[ruleSeparatorLine]() {
				goto l1237
			}
			goto l1236
		l1237:
			{
				position1238 := position
				if !p.rules[ruleTableLine]() {
					goto l1236
				}
				position = position1238
			}
			matchChar('|')
			if !p.rules[ruleTableCell]() {
				goto l1236
			}
			do(124)
		l1239:
			{
				position1240 := position
				if !p.rules[ruleTableCell]() {
					goto l1240
				}
				do(124)
				goto l1239
			l1240:
				position = position1240
			}
			if !p.rules[ruleSp]() {
				goto l1236
			}
			if !p.rules[ruleNewline]() {
				goto l1236
			}
			do(125)
			doarg(yyPop, 1)
			return true
		l1236:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 250 TableLine <- ((!Newline !'|' .)* '|') */
		func() bool {
			position0 := position
		l1242:
			{
				position1243 := position
				if !p.rules[ruleNewline]() {
					goto l1244
				}
				goto l1243
			l1244:
				if peekChar('|') {
					goto l1243
				}
				if !matchDot() {
					goto l1243
				}
				goto l1242
			l1243:
				position = position1243
			}
			if !matchChar('|') {
				goto l1241
			}
			return true
		l1241:
			position = position0
			return false
		},
		/* 251 TableCell <- (ExtendedCell / EmptyCell / FullCell) */
		func() bool {
			if !p.rules[ruleExtendedCell]() {
				goto l1247
			}
			goto l1246
		l1247:
			if !p.rules[ruleEmptyCell]() {
				goto l1248
			}
			goto l1246
		l1248:
			if !p.rules[ruleFullCell]() {
				goto l1245
			}
		l1246:
			return true
		l1245:
			return false
		},
		/* 252 ExtendedCell <- ((EmptyCell / FullCell) < '|'+ > {
        span := p.mkString(yytext)
        span.key = CELLSPAN
        span.next = yy.children
        yy.children = span
    }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleEmptyCell]() {
				goto l1251
			}
			goto l1250
		l1251:
			if !p.rules[ruleFullCell]() {
				goto l1249
			}
		l1250:
			begin = position
			if !matchChar('|') {
				goto l1249
			}
		l1252:
			if !matchChar('|') {
				goto l1253
			}
			goto l1252
		l1253:
			end = position
			do(126)
			return true
		l1249:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 253 CellStr <- (< (!'|' NormalChar) ((!'|' NormalChar) / ('_'+ &Alphanumeric))* > { yy = p.mkString(yytext) }) */
		func() bool {
			position0 := position
			begin = position
			if peekChar('|') {
				goto l1254
			}
			if !p.rules[ruleNormalChar]() {
				goto l1254
			}
		l1255:
			{
				position1256 := position
				if peekChar('|') {
					goto l1258
				}
				if !p.rules[ruleNormalChar]() {
					goto l1258
				}
				goto l1257
			l1258:
				if !matchChar('_') {
					goto l1256
				}
			l1259:
				if !matchChar('_') {
					goto l1260
				}
				goto l1259
			l1260:
				{
					position1261 := position
					if !p.rules[ruleAlphanumeric]() {
						goto l1256
					}
					position = position1261
				}
			l1257:
				goto l1255
			l1256:
				position = position1256
			}
			end = position
			do(127)
			return true
		l1254:
			position = position0
			return false
		},
		/* 254 FullCell <- (Sp StartList (((!'|' CellStr) / (!Newline !Endline !'|' !Str !(Sp &'|') Inline)) { a = cons(yy, a) })+ Sp '|'? { yy = p.mkList(TABLECELL, a) }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleSp]() {
				goto l1262
			}
			if !p.rules[ruleStartList]() {
				goto l1262
			}
			doarg(yySet, -1)
			if peekChar('|') {
				goto l1266
			}
			if !p.rules[ruleCellStr]() {
				goto l1266
			}
			goto l1265
		l1266:
			if !p.rules[ruleNewline]() {
				goto l1267
			}
			goto l1262
		l1267:
			if !p.rules[ruleEndline]() {
				goto l1268
			}
			goto l1262
		l1268:
			if peekChar('|') {
				goto l1262
			}
			if !p.rules[ruleStr]() {
				goto l1269
			}
			goto l1262
		l1269:
			{
				position1270 := position
				if !p.rules[ruleSp]() {
					goto l1270
				}
				if !peekChar('|') {
					goto l1270
				}
				goto l1262
			l1270:
				position = position1270
			}
			if !p.rules[ruleInline]() {
				goto l1262
			}
		l1265:
			do(128)
		l1263:
			{
				position1264, thunkPosition1264 := position, thunkPosition
				if peekChar('|') {
					goto l1272
				}
				if !p.rules[ruleCellStr]() {
					goto l1272
				}
				goto l1271
			l1272:
				if !p.rules[ruleNewline]() {
					goto l1273
				}
				goto l1264
			l1273:
				if !p.rules[ruleEndline]() {
					goto l1274
				}
				goto l1264
			l1274:
				if peekChar('|') {
					goto l1264
				}
				if !p.rules[ruleStr]() {
					goto l1275
				}
				goto l1264
			l1275:
				{
					position1276 := position
					if !p.rules[ruleSp]() {
						goto l1276
					}
					if !peekChar('|') {
						goto l1276
					}
					goto l1264
				l1276:
					position = position1276
				}
				if !p.rules[ruleInline]() {
					goto l1264
				}
			l1271:
				do(128)
				goto l1263
			l1264:
				position, thunkPosition = position1264, thunkPosition1264
			}
			if !p.rules[ruleSp]() {
				goto l1262
			}
			matchChar('|')
			do(129)
			doarg(yyPop, 1)
			return true
		l1262:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 255 EmptyCell <- (Sp '|' { yy = p.mkElem(TABLECELL) }) */
		func() bool {
			position0 := position
			if !p.rules[ruleSp]() {
				goto l1277
			}
			if !matchChar('|') {
				goto l1277
			}
			do(130)
			return true
		l1277:
			position = position0
			return false
		},
		/* 256 SeparatorLine <- (StartList &TableLine '|'? (AlignmentCell { a = cons(yy, a) })+ Sp Newline {
        yy = p.mkStringFromList(a, false);
        yy.key = TABLESEPARATOR;
    }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleStartList]() {
				goto l1278
			}
			doarg(yySet, -1)
			{
				position1279 := position
				if !p.rules[ruleTableLine]() {
					goto l1278
				}
				position = position1279
			}
			matchChar('|')
			if !p.rules[ruleAlignmentCell]() {
				goto l1278
			}
			do(131)
		l1280:
			{
				position1281 := position
				if !p.rules[ruleAlignmentCell]() {
					goto l1281
				}
				do(131)
				goto l1280
			l1281:
				position = position1281
			}
			if !p.rules[ruleSp]() {
				goto l1278
			}
			if !p.rules[ruleNewline]() {
				goto l1278
			}
			do(132)
			doarg(yyPop, 1)
			return true
		l1278:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 257 AlignmentCell <- (Sp (!'|' (LeftAlignWrap / CenterAlignWrap / RightAlignWrap / LeftAlign / ((&[\-] RightAlign) | (&[:] CenterAlign)))) Sp '|'?) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[ruleSp]() {
				goto l1282
			}
			if peekChar('|') {
				goto l1282
			}
			if !p.rules[ruleLeftAlignWrap]() {
				goto l1284
			}
			goto l1283
		l1284:
			if !p.rules[ruleCenterAlignWrap]() {
				goto l1285
			}
			goto l1283
		l1285:
			if !p.rules[ruleRightAlignWrap]() {
				goto l1286
			}
			goto l1283
		l1286:
			if !p.rules[ruleLeftAlign]() {
				goto l1287
			}
			goto l1283
		l1287:
			{
				if position == len(p.Buffer) {
					goto l1282
				}
				switch p.Buffer[position] {
				case '-':
					if !p.rules[ruleRightAlign]() {
						goto l1282
					}
					break
				case ':':
					if !p.rules[ruleCenterAlign]() {
						goto l1282
					}
					break
				default:
					goto l1282
				}
			}
		l1283:
			if !p.rules[ruleSp]() {
				goto l1282
			}
			matchChar('|')
			return true
		l1282:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 258 LeftAlignWrap <- (':'? '-'+ '+' &(!'-' !':') { yy = p.mkString("L");}) */
		func() bool {
			position0 := position
			matchChar(':')
			if !matchChar('-') {
				goto l1289
			}
		l1290:
			if !matchChar('-') {
				goto l1291
			}
			goto l1290
		l1291:
			if !matchChar('+') {
				goto l1289
			}
			if peekChar('-') {
				goto l1289
			}
			if peekChar(':') {
				goto l1289
			}
			do(133)
			return true
		l1289:
			position = position0
			return false
		},
		/* 259 LeftAlign <- (':'? '-'+ &(!'-' !':') { yy = p.mkString("l");}) */
		func() bool {
			position0 := position
			matchChar(':')
			if !matchChar('-') {
				goto l1293
			}
		l1294:
			if !matchChar('-') {
				goto l1295
			}
			goto l1294
		l1295:
			if peekChar('-') {
				goto l1293
			}
			if peekChar(':') {
				goto l1293
			}
			do(134)
			return true
		l1293:
			position = position0
			return false
		},
		/* 260 CenterAlignWrap <- (':' '-'* '+' ':' &(!'-' !':') { yy = p.mkString("C");}) */
		func() bool {
			position0 := position
			if !matchChar(':') {
				goto l1297
			}
		l1298:
			if !matchChar('-') {
				goto l1299
			}
			goto l1298
		l1299:
			if !matchChar('+') {
				goto l1297
			}
			if !matchChar(':') {
				goto l1297
			}
			if peekChar('-') {
				goto l1297
			}
			if peekChar(':') {
				goto l1297
			}
			do(135)
			return true
		l1297:
			position = position0
			return false
		},
		/* 261 CenterAlign <- (':' '-'* ':' &(!'-' !':') { yy = p.mkString("c");}) */
		func() bool {
			position0 := position
			if !matchChar(':') {
				goto l1301
			}
		l1302:
			if !matchChar('-') {
				goto l1303
			}
			goto l1302
		l1303:
			if !matchChar(':') {
				goto l1301
			}
			if peekChar('-') {
				goto l1301
			}
			if peekChar(':') {
				goto l1301
			}
			do(136)
			return true
		l1301:
			position = position0
			return false
		},
		/* 262 RightAlignWrap <- ('-'+ ':' '+' &(!'-' !':') { yy = p.mkString("R");}) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l1305
			}
		l1306:
			if !matchChar('-') {
				goto l1307
			}
			goto l1306
		l1307:
			if !matchChar(':') {
				goto l1305
			}
			if !matchChar('+') {
				goto l1305
			}
			if peekChar('-') {
				goto l1305
			}
			if peekChar(':') {
				goto l1305
			}
			do(137)
			return true
		l1305:
			position = position0
			return false
		},
		/* 263 RightAlign <- ('-'+ ':' &(!'-' !':') { yy = p.mkString("r");}) */
		func() bool {
			position0 := position
			if !matchChar('-') {
				goto l1309
			}
		l1310:
			if !matchChar('-') {
				goto l1311
			}
			goto l1310
		l1311:
			if !matchChar(':') {
				goto l1309
			}
			if peekChar('-') {
				goto l1309
			}
			if peekChar(':') {
				goto l1309
			}
			do(138)
			return true
		l1309:
			position = position0
			return false
		},
		/* 264 CellDivider <- '|' */
		func() bool {
			if !matchChar('|') {
				goto l1313
			}
			return true
		l1313:
			return false
		},
		/* 265 TableCaption <- (StartList Label (Label { b = c; b.key = TABLELABEL;})? Sp Newline {
    yy = a
    yy.key = TABLECAPTION
    if b != nil && b.key == TABLELABEL {
        b.next = yy.children
        yy.children = b
    }
}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 3)
			if !p.rules[ruleStartList]() {
				goto l1314
			}
			doarg(yySet, -1)
			if !p.rules[ruleLabel]() {
				goto l1314
			}
			doarg(yySet, -2)
			{
				position1315, thunkPosition1315 := position, thunkPosition
				if !p.rules[ruleLabel]() {
					goto l1315
				}
				doarg(yySet, -3)
				do(139)
				goto l1316
			l1315:
				position, thunkPosition = position1315, thunkPosition1315
			}
		l1316:
			if !p.rules[ruleSp]() {
				goto l1314
			}
			if !p.rules[ruleNewline]() {
				goto l1314
			}
			do(140)
			doarg(yyPop, 3)
			return true
		l1314:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
	}
}


/*
 * List manipulation functions
 */

/* cons - cons an element onto a list, returning pointer to new head
 */
func cons(new, list *element) *element {
	new.next = list
	return new
}

/* reverse - reverse a list, returning pointer to new list
 */
func reverse(list *element) (new *element) {
	for list != nil {
		next := list.next
		new = cons(list, new)
		list = next
	}
	return
}

/* append_list - add element to end of list */
func append_list(new *element, list *element) {
  step := list

  for step.next != nil {
    step = step.next
  }

  new.next = nil
  step.next = new
}

/*
 *  Auxiliary functions for parsing actions.
 *  These make it easier to build up data structures (including lists)
 *  in the parsing actions.
 */

/* p.mkElem - generic constructor for element
 */
func (p *yyParser) mkElem(key int) *element {
	r := p.state.heap.row
	if len(r) == 0 {
		r = p.state.heap.nextRow()
	}
	e := &r[0]
	*e = element{}
	p.state.heap.row = r[1:]
	e.key = key
	return e
}

/* p.mkString - constructor for STR element
 */
func (p *yyParser) mkString(s string) (result *element) {
	result = p.mkElem(STR)
	result.contents.str = s
	return
}

/* p.mkStringFromList - makes STR element by concatenating a
 * reversed list of strings, adding optional extra newline
 */
func (p *yyParser) mkStringFromList(list *element, extra_newline bool) (result *element) {
	s := ""
	for list = reverse(list); list != nil; list = list.next {
		s += list.contents.str
	}

	if extra_newline {
		s += "\n"
	}
	result = p.mkElem(STR)
	result.contents.str = s
	return
}

/* p.mkList - makes new list with key 'key' and children the reverse of 'lst'.
 * This is designed to be used with cons to build lists in a parser action.
 * The reversing is necessary because cons adds to the head of a list.
 */
func (p *yyParser) mkList(key int, lst *element) (el *element) {
	el = p.mkElem(key)
	el.children = reverse(lst)
	return
}

/* p.mkLink - constructor for LINK element
 */
func (p *yyParser) mkLink(label *element, url, title string) (el *element) {
	el = p.mkElem(LINK)
	el.contents.link = &link{label: label, url: url, title: title}
	return
}

/* match_inlines - returns true if inline lists match (case-insensitive...)
 */
func match_inlines(l1, l2 *element) bool {
	for l1 != nil && l2 != nil {
		if l1.key != l2.key {
			return false
		}
		switch l1.key {
		case SPACE, LINEBREAK, ELLIPSIS, EMDASH, ENDASH, APOSTROPHE:
			break
		case CODE, STR, HTML:
			if strings.ToUpper(l1.contents.str) != strings.ToUpper(l2.contents.str) {
				return false
			}
		case EMPH, STRONG, LIST, SINGLEQUOTED, DOUBLEQUOTED:
			if !match_inlines(l1.children, l2.children) {
				return false
			}
		case LINK, IMAGE:
			return false /* No links or images within links */
		default:
			log.Fatalf("match_inlines encountered unknown key = %d\n", l1.key)
		}
		l1 = l1.next
		l2 = l2.next
	}
	return l1 == nil && l2 == nil /* return true if both lists exhausted */
}

/* find_reference - return true if link found in references matching label.
 * 'link' is modified with the matching url and title.
 */
func (p *yyParser) findReference(label *element) (*link, bool) {
	for cur := p.references; cur != nil; cur = cur.next {
		l := cur.contents.link
		if match_inlines(label, l.label) {
			return l, true
		}
	}
	return nil, false
}

/* find_note - return true if note found in notes matching label.
 * if found, 'result' is set to point to matched note.
 */
func (p *yyParser) find_note(label string) (*element, bool) {
	for el := p.notes; el != nil; el = el.next {
		if label == el.contents.str {
			return el, true
		}
	}
	return nil, false
}

/* print tree of elements, for debugging only.
 */
func print_tree(w io.Writer, elt *element, indent int) {
	var key string

	for elt != nil {
		for i := 0; i < indent; i++ {
			fmt.Fprint(w, "\t")
		}
		key = keynames[elt.key]
		if key == "" {
			key = "?"
		}
		if elt.key == STR {
			fmt.Fprintf(w, "%p:\t%s\t'%s'\n", elt, key, elt.contents.str)
		} else {
			fmt.Fprintf(w, "%p:\t%s %p\n", elt, key, elt.next)
		}
		if elt.children != nil {
			print_tree(w, elt.children, indent+1)
		}
		elt = elt.next
	}
}

var keynames = [numVAL]string{
	LIST:           "LIST",
	RAW:            "RAW",
	SPACE:          "SPACE",
	LINEBREAK:      "LINEBREAK",
	ELLIPSIS:       "ELLIPSIS",
	EMDASH:         "EMDASH",
	ENDASH:         "ENDASH",
	APOSTROPHE:     "APOSTROPHE",
	SINGLEQUOTED:   "SINGLEQUOTED",
	DOUBLEQUOTED:   "DOUBLEQUOTED",
	STR:            "STR",
	LINK:           "LINK",
	IMAGE:          "IMAGE",
	CODE:           "CODE",
	HTML:           "HTML",
	EMPH:           "EMPH",
	STRONG:         "STRONG",
	PLAIN:          "PLAIN",
	PARA:           "PARA",
	LISTITEM:       "LISTITEM",
	BULLETLIST:     "BULLETLIST",
	ORDEREDLIST:    "ORDEREDLIST",
	H1:             "H1",
	H2:             "H2",
	H3:             "H3",
	H4:             "H4",
	H5:             "H5",
	H6:             "H6",
	BLOCKQUOTE:     "BLOCKQUOTE",
	VERBATIM:       "VERBATIM",
	HTMLBLOCK:      "HTMLBLOCK",
	HRULE:          "HRULE",
	REFERENCE:      "REFERENCE",
	NOTE:           "NOTE",
	DEFINITIONLIST: "DEFINITIONLIST",
	DEFTITLE:       "DEFTITLE",
	DEFDATA:        "DEFDATA",
}
