package main

import "fmt"

type InvalidCharacterCombinationError struct {
	InvalidChars string
}
type InvalidCharacterAtBeginningOrEndError struct {
	InvalidChar rune
}
type InvalidUseOfParentheses struct {
	Note string
}

func (e *InvalidCharacterCombinationError) Error() string {
	return fmt.Sprintf("Invalid character combination: %s",
		e.InvalidChars)
}
func (e *InvalidCharacterAtBeginningOrEndError) Error() string {
	return fmt.Sprintf("An expression cannot begin or end with this "+
		"symbol: %s", string(e.InvalidChar))
}
func (e *InvalidUseOfParentheses) Error() string {
	return fmt.Sprintf("Invalid use of parentheses in the expression:"+
		" %s", e.Note)
}
