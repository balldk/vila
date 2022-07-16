package parser

import (
	"strconv"
	"vila/ast"
	"vila/token"
)

const (
	_ int = iota
	LOWEST
	EQUAL   // ==
	COMPARE // > or <
	SUM     // +
	PRODUCT // *
	EXP     // ^
	PREFIX
	CALL
)

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			p.syntaxError("toán tử trung tố không tồn tại")
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseInt() ast.Expression {
	i := &ast.Int{Token: p.curToken}

	value, err := strconv.ParseInt(string(p.curToken.Literal), 0, 64)
	if err != nil {
		p.syntaxError("Không thể parse số nguyên này")
	}

	i.Value = value

	return i
}

func (p *Parser) parseReal() ast.Expression {
	re := &ast.Real{Token: p.curToken}

	value, err := strconv.ParseFloat(string(p.curToken.Literal), 64)
	if err != nil {
		p.syntaxError("Không thể parse số thực này")
	}

	re.Value = value

	return re
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: fn}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(token.RPAREN) {
			return args
		}
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}