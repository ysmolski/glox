// Code generated by "stringer -type token -linecomment tokens.go"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LeftParen-1]
	_ = x[RightParen-2]
	_ = x[LeftBrace-3]
	_ = x[RightBrace-4]
	_ = x[Comma-5]
	_ = x[Dot-6]
	_ = x[Minus-7]
	_ = x[Plus-8]
	_ = x[Semicolon-9]
	_ = x[Colon-10]
	_ = x[Question-11]
	_ = x[Slash-12]
	_ = x[Star-13]
	_ = x[Bang-14]
	_ = x[BangEqual-15]
	_ = x[Equal-16]
	_ = x[EqualEqual-17]
	_ = x[Greater-18]
	_ = x[GreaterEqual-19]
	_ = x[Less-20]
	_ = x[LessEqual-21]
	_ = x[Identifier-22]
	_ = x[String-23]
	_ = x[Number-24]
	_ = x[And-25]
	_ = x[Class-26]
	_ = x[Else-27]
	_ = x[False-28]
	_ = x[Fun-29]
	_ = x[For-30]
	_ = x[If-31]
	_ = x[Nil-32]
	_ = x[Or-33]
	_ = x[Print-34]
	_ = x[Return-35]
	_ = x[Super-36]
	_ = x[This-37]
	_ = x[True-38]
	_ = x[Var-39]
	_ = x[While-40]
	_ = x[EOF-41]
}

const _token_name = "(){},.-+;:?/*!!====>>=<<=identstringnumberandclasselsefalsefunforifnilorprintreturnsuperthistruevarwhileeof"

var _token_index = [...]uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 16, 17, 19, 20, 22, 23, 25, 30, 36, 42, 45, 50, 54, 59, 62, 65, 67, 70, 72, 77, 83, 88, 92, 96, 99, 104, 107}

func (i token) String() string {
	i -= 1
	if i >= token(len(_token_index)-1) {
		return "token(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _token_name[_token_index[i]:_token_index[i+1]]
}
