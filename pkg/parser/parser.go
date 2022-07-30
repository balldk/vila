package parser

import (
	"fmt"
	"vila/pkg/ast"
	"vila/pkg/errorhandler"
	"vila/pkg/lexer"
	"vila/pkg/token"
)

const IDENT_SIZE = 4

func New(l *lexer.Lexer, errors *errorhandler.ErrorList) *Parser {
	p := &Parser{
		l:             l,
		Errors:        errors,
		identLevel:    0,
		peekPeekToken: token.Token{Type: ""},
	}
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.Ident, p.parseIdentifier)
	p.registerPrefix(token.Int, p.parseInt)
	p.registerPrefix(token.Real, p.parseReal)
	p.registerPrefix(token.True, p.parseBoolean)
	p.registerPrefix(token.False, p.parseBoolean)
	p.registerPrefix(token.Bang, p.parsePrefixExpression)
	p.registerPrefix(token.Minus, p.parsePrefixExpression)
	p.registerPrefix(token.Plus, p.parsePrefixExpression)
	p.registerPrefix(token.LParen, p.parseGroupExpression)
	p.registerPrefix(token.LBracket, p.parseInterval)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.Plus, p.parseInfixExpression)
	p.registerInfix(token.Minus, p.parseInfixExpression)
	p.registerInfix(token.Slash, p.parseInfixExpression)
	p.registerInfix(token.Asterisk, p.parseInfixExpression)
	p.registerInfix(token.Dot, p.parseInfixExpression)
	p.registerInfix(token.Hat, p.parseInfixExpression)
	p.registerInfix(token.Equal, p.parseInfixExpression)
	p.registerInfix(token.NotEqual, p.parseInfixExpression)
	p.registerInfix(token.Less, p.parseInfixExpression)
	p.registerInfix(token.Greater, p.parseInfixExpression)
	p.registerInfix(token.LessEqual, p.parseInfixExpression)
	p.registerInfix(token.GreaterEqual, p.parseInfixExpression)
	p.registerInfix(token.LParen, p.parseCallExpression)
	p.registerInfix(token.If, p.parseIfExpression)

	p.advanceToken()
	p.advanceToken()

	return p
}

type Parser struct {
	l          *lexer.Lexer
	Errors     *errorhandler.ErrorList
	identLevel int

	curToken      token.Token
	peekToken     token.Token
	peekPeekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) advanceToken() {
	if p.peekPeekToken.Type != "" {
		p.curToken = p.peekToken
		p.peekToken = p.peekPeekToken
		p.peekPeekToken = token.Token{Type: ""}
	} else {
		p.curToken = p.peekToken
		p.peekToken = p.l.AdvanceToken()
	}
	// fmt.Println(p.curToken)
}

func (p *Parser) insertPeekToken(tok token.Token) {
	p.peekPeekToken = p.peekToken
	p.peekToken = tok
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
	}

	return program
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curIsStatementSeperator() bool {
	return p.curTokenIs(token.Semicolon) || p.curTokenIs(token.Endline) || p.curTokenIs(token.EOF)
}

func (p *Parser) peekIsStatementSeperator() bool {
	return p.peekTokenIs(token.Semicolon) || p.peekTokenIs(token.Endline) || p.peekTokenIs(token.EOF)
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.advanceToken()
		return true
	} else {
		p.advanceToken()
		p.expectError(t)
		return false
	}
}

func (p *Parser) expectCur(t token.TokenType) bool {
	if p.curTokenIs(t) {
		p.advanceToken()
		return true
	} else {
		p.expectError(t)
		p.advanceToken()
		return false
	}
}

func (p *Parser) skipEndline() {
	// skip semicolon
	for p.curTokenIs(token.Semicolon) {
		p.advanceToken()
	}
	// skip consecutive endline
	for p.curTokenIs(token.Endline) && p.peekTokenIs(token.Endline) {
		p.advanceToken()
	}
}

func (p *Parser) updateIdentLevel() {
	p.skipEndline()

	if p.curTokenIs(token.Endline) {
		length := len(p.curToken.Literal)
		if length%IDENT_SIZE != 0 {
			p.invalidIndent()
			return
		}

		level := length / 4
		if level > p.identLevel {
			p.invalidIndent()
			return
		}
		p.identLevel = level
		p.advanceToken()
	}
}

func (p *Parser) syntaxError(message string) {
	p.Errors.AddParserError(message, p.curToken)
}

func (p *Parser) invalidSyntax() {
	p.Errors.AddParserError("Cú pháp không hợp lệ", p.curToken)
}

func (p *Parser) invalidIndent() {
	p.syntaxError("Thụt dòng không hợp lệ")
}

func (p *Parser) expectError(tokType token.TokenType) {
	var msg string

	if p.curIsStatementSeperator() {
		msg = fmt.Sprintf("Thiếu `%s`", string(tokType))
	} else {
		msg = fmt.Sprintf("Cần `%s` thay vì `%s`", string(tokType), string(p.curToken.Literal))
	}
	p.syntaxError(msg)
}