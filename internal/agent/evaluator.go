package agent

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func EvaluateSimpleExpression(expression string, operations map[string]int, answMap map[string]float64, errChan chan error, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	re := regexp.MustCompile(`(\-??\d+(\.\d+)?)([+\-*/])(\-??\d+(\.\d+)?)`)
	res := re.FindStringSubmatch(expression)
	if len(res) < 5 {
		errChan <- errors.New("Invalid expression")
		return
	}
	parts := []string{res[1], res[3], res[4]}
	operand1, _ := strconv.ParseFloat(parts[0], 64)
	operand2, _ := strconv.ParseFloat(parts[2], 64)
	var result float64
	switch parts[1] {
	case "+":
		result = operand1 + operand2
		time.Sleep(time.Duration(operations["add"]) * time.Second)
	case "-":
		result = operand1 - operand2
		time.Sleep(time.Duration(operations["sub"]) * time.Second)
	case "*":
		result = operand1 * operand2
		time.Sleep(time.Duration(operations["mult"]) * time.Second)
	case "/":
		if operand2 == 0 {
			errChan <- fmt.Errorf("division by zero")
			return
		}
		result = operand1 / operand2
		time.Sleep(time.Duration(operations["div"]) * time.Second)
	default:
		errChan <- fmt.Errorf("invalid operator: %s", parts[1])
		return
	}
	errChan <- nil
	mu.Lock()
	defer mu.Unlock()
	answMap[expression] = result
}

func EvaluateComplexExpression(expression string, numGoroutines int, operations map[string]int) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	isValid := regexp.MustCompile(`^\(*\-??\d+(\.\d+)?([+\-/*]\(*\-??\d+(\.\d+)?\)*)+\)*$`).MatchString(expression) && ParenthesesCheck(expression)
	if !isValid {
		return 0, fmt.Errorf("not valid")
	}
	expression = ParenthesesClear(expression)
	simpleExprMap := ExtractSimpleExpressions(expression)
	for len(simpleExprMap) != 0 {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errChan := make(chan error, numGoroutines)
		var count int
		answMap := make(map[string]float64)
		for _, v := range simpleExprMap {
			if count < numGoroutines {
				count++
				wg.Add(1)
				go EvaluateSimpleExpression(v, operations, answMap, errChan, &mu, &wg)
			} else {
				count = 1
				wg.Wait()
				for i := 0; i < numGoroutines; i++ {
					select {
					case err := <-errChan:
						if err != nil {
							close(errChan)
							return 0, err
						}
					}
				}
				wg.Add(1)
				go EvaluateSimpleExpression(v, operations, answMap, errChan, &mu, &wg)
			}
		}
		wg.Wait()
		for i := 0; i < count; i++ {
			select {
			case err := <-errChan:
				if err != nil {
					close(errChan)
					return 0, err
				}
			}
		}
		close(errChan)
		for k, v := range simpleExprMap {
			result := answMap[v]
			old := len(expression)
			expression = expression[:k] + strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", result), "0"), ".") + expression[k+len(v):]
			expression = ParenthesesClear(expression)
			if old > len(expression) {
				diff := old - len(expression)
				for k1, v1 := range simpleExprMap {
					if k1 > k {
						delete(simpleExprMap, k1)
						simpleExprMap[k1-diff] = v1
					}
				}
			} else if old < len(expression) {
				diff := len(expression) - old
				for k1, v1 := range simpleExprMap {
					if k1 > k {
						delete(simpleExprMap, k1)
						simpleExprMap[k1+diff] = v1
					}
				}
			}
		}
		simpleExprMap = ExtractSimpleExpressions(expression)
	}
	return strconv.ParseFloat(expression, 64)
}
