package agent

import (
	"regexp"
	"strings"
)

func ParenthesesCheck(expression string) bool {
	var count int
	for _, char := range expression {
		if string(char) == "(" {
			count++
		} else if string(char) == ")" {
			count--
		}
		if count < 0 {
			return false
		}
	}
	if count != 0 {
		return false
	}
	return true
}

func ParenthesesClear(expression string) string {
	patternOneNum := `\(\-??\d+(\.\d+)?\)`
	patternDups := `\({2,}.*?\){2,}`
	clearMap := make(map[int]string)
	processing := true
	for processing {
		processing = false
		re := regexp.MustCompile(patternOneNum)
		matches := re.FindAllStringSubmatchIndex(expression, -1)
		for _, match := range matches {
			startInd, endInd := match[0], match[1]
			simpleExpr := expression[startInd:endInd]
			clearMap[startInd] = simpleExpr
			processing = true
		}
		for k, v := range clearMap {
			expression = expression[:k] + expression[k+1:k+len(v)-1] + expression[k+len(v):]
			delete(clearMap, k)
		}
		re = regexp.MustCompile(patternDups)
		matches = re.FindAllStringSubmatchIndex(expression, -1)
		for _, match := range matches {
			startInd, endInd := match[0], match[1]
			simpleExpr := expression[startInd:endInd]
			clearMap[startInd] = simpleExpr
			processing = true
		}
		for k, v := range clearMap {
			expression = expression[:k] + expression[k+1:k+len(v)-1] + expression[k+len(v):]
			delete(clearMap, k)
		}
	}
	return expression
}

func ExtractSimpleExpressions(expression string) map[int]string {
	patternAddSub := `\d+(\.\d+)?[+\-]\-??\d+(\.\d+)?`
	patternMultDiv := `\d+(\.\d+)?[*/]\-??\d+(\.\d+)?`
	simpleExprMap := make(map[int]string)
	re := regexp.MustCompile(patternAddSub)
	matches := re.FindAllStringSubmatchIndex(expression, -1)
	for _, match := range matches {
		startInd, endInd := match[0], match[1]
		if startInd == 0 && endInd == len(expression) {
			simpleExpr := expression[startInd:endInd]
			simpleExprMap[startInd] = simpleExpr
		} else if startInd == 0 {
			if !strings.ContainsAny(string(expression[endInd]), "*/") {
				simpleExpr := expression[startInd:endInd]
				simpleExprMap[startInd] = simpleExpr
			}
		} else if endInd == len(expression) {
			if startInd == 1 {
				if string(expression[startInd-1]) == "-" {
					simpleExpr := expression[startInd-1 : endInd]
					simpleExprMap[startInd-1] = simpleExpr
					continue
				}
			} else if startInd >= 2 {
				if !regexp.MustCompile(`\d`).MatchString(string(expression[startInd-2])) && string(expression[startInd-1]) == "-" {
					simpleExpr := expression[startInd-1 : endInd]
					simpleExprMap[startInd-1] = simpleExpr
					continue
				}
			}
			if strings.ContainsAny(string(expression[startInd-1]), "+(") {
				simpleExpr := expression[startInd:endInd]
				simpleExprMap[startInd] = simpleExpr
			}
		} else {
			if startInd == 1 {
				if string(expression[startInd-1]) == "-" && !strings.ContainsAny(string(expression[endInd]), "*/") {
					simpleExpr := expression[startInd-1 : endInd]
					simpleExprMap[startInd-1] = simpleExpr
					continue
				}
			} else if startInd >= 2 {
				if !regexp.MustCompile(`\d`).MatchString(string(expression[startInd-2])) && string(expression[startInd-1]) == "-" && !strings.ContainsAny(string(expression[endInd]), "*/") {
					simpleExpr := expression[startInd-1 : endInd]
					simpleExprMap[startInd-1] = simpleExpr
					continue
				}
			}
			if strings.ContainsAny(string(expression[startInd-1]), "+(") && !strings.ContainsAny(string(expression[endInd]), "*/") {
				simpleExpr := expression[startInd:endInd]
				simpleExprMap[startInd] = simpleExpr
			}
		}
	}
	re = regexp.MustCompile(patternMultDiv)
	matches = re.FindAllStringSubmatchIndex(expression, -1)
	for _, match := range matches {
		startInd, endInd := match[0], match[1]
		if startInd == 0 {
			simpleExpr := expression[startInd:endInd]
			simpleExprMap[startInd] = simpleExpr
		} else {
			if startInd == 1 {
				if string(expression[startInd-1]) == "-" {
					simpleExpr := expression[startInd-1 : endInd]
					simpleExprMap[startInd-1] = simpleExpr
					continue
				}
			} else if startInd >= 2 {
				if !regexp.MustCompile(`\d`).MatchString(string(expression[startInd-2])) && string(expression[startInd-1]) == "-" {
					simpleExpr := expression[startInd-1 : endInd]
					simpleExprMap[startInd-1] = simpleExpr
					continue
				}
			}
			if !strings.ContainsAny(string(expression[startInd-1]), "/") {
				simpleExpr := expression[startInd:endInd]
				simpleExprMap[startInd] = simpleExpr
			}
		}
	}
	return simpleExprMap
}
