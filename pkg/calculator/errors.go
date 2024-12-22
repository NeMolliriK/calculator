package calculator

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
type InvalidCharacter struct {
	InvalidChar rune
}

func (e *InvalidCharacterCombinationError) Error() string {
	return fmt.Sprintf("invalid character combination: %s", e.InvalidChars)
}
func (e *InvalidCharacterAtBeginningOrEndError) Error() string {
	return fmt.Sprintf(
		"an expression cannot begin or end with this symbol: %s",
		string(e.InvalidChar),
	)
}
func (e *InvalidUseOfParentheses) Error() string {
	return fmt.Sprintf("invalid use of parentheses in the expression: %s", e.Note)
}
func (e *InvalidCharacter) Error() string {
	return fmt.Sprintf(
		"an invalid character is present in the expression: %s",
		string(e.InvalidChar),
	)
}
