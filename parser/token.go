package parser

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	tokenComment = "comment"
	tokenOption  = "option"
	tokenAccount = "account"
	tokenPad     = "pad"
	tokenHeading = "heading"
	tokenBody    = "body"
	tokenBalance = "balance"
)

var dateRegex = `\d{4}-\d{2}-\d{2}`

func lexComment(line string) (token, bool) {
	if strings.HasPrefix(strings.TrimSpace(line), ";") {
		return token{tokenComment, line, []string{line}}, true
	}
	return nilToken, false
}

func lexOption(line string) (token, bool) {
	var optionRegexp = regexp.MustCompile(`option (\S+) (\S+)`)
	if m := optionRegexp.FindStringSubmatch(line); m != nil {
		return token{tokenOption, m[0], m}, true
	}
	return nilToken, false
}

func lexAccount(line string) (token, bool) {
	accountRegexp := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})\s(open|note)\s(\S+)\s?(\S+)?$`)
	if m := accountRegexp.FindStringSubmatch(line); m != nil {
		return token{tokenAccount, m[0], m}, true
	}
	return nilToken, false
}

func lexPad(line string) (token, bool) {
	accountRegexp := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})\s(pad)\s(\S+)\s?(\S+)?`)
	if m := accountRegexp.FindStringSubmatch(line); m != nil {
		return token{tokenPad, m[0], m}, true
	}
	return nilToken, false
}

func lexHeading(line string) (token, bool) {
	headRegexp := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})\s+(\*|!)\s+(\S+)\s(\S+)\s?(#\S+)?`)
	if m := headRegexp.FindStringSubmatch(line); m != nil {
		return token{tokenHeading, m[0], m}, true
	}
	return nilToken, false
}

func lexBody(line string) (token, bool) {
	// (only?) body not started with date
	datePrefixRegexp := regexp.MustCompile(fmt.Sprintf("^%s", dateRegex))
	if datePrefixRegexp.MatchString(line) {
		return nilToken, false
	}

	// body can be only:
	// account
	// account number currency
	parts := strings.Fields(line)
	if len(parts) == 1 {
		return token{tokenBody, line, append([]string{parts[0]}, parts...)}, true
	}
	if len(parts) == 3 {
		joined := strings.Join(parts, " ")
		bodyRegexp := regexp.MustCompile(`(\S+)\s([-\+\.\d]+)\s(\S+)`)
		if m := bodyRegexp.FindStringSubmatch(joined); m != nil {
			return token{tokenBody, m[0], m}, true
		}
	}
	return nilToken, false
}

func lexBalance(line string) (token, bool) {
	balanceRegexp := regexp.MustCompile(`^[\s+]?(\d{4}-\d{2}-\d{2})\s+(balance)\s+(\S+)\s+([\+-\.\d]+)\s+(\S+)`)
	if m := balanceRegexp.FindStringSubmatch(line); m != nil {
		return token{tokenBalance, m[0], m}, true
	}
	return nilToken, false
}
