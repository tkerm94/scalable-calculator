package agent

import (
	"sync"
	"testing"
)

func TestEvaluateSimpleExpression(t *testing.T) {
	cases := []struct {
		name       string
		expression string
		want_err   bool
		want       float64
	}{
		{
			name:       "letter as first operand",
			expression: "a+1",
			want_err:   true,
		},
		{
			name:       "letter as second operand",
			expression: "1-a",
			want_err:   true,
		},
		{
			name:       "division by zero",
			expression: "1/0",
			want_err:   true,
		},
		{
			name:       "add",
			expression: "1+1",
			want_err:   false,
			want:       2,
		},
		{
			name:       "sub",
			expression: "1-1",
			want_err:   false,
			want:       0,
		},
		{
			name:       "mult",
			expression: "2*2",
			want_err:   false,
			want:       4,
		},
		{
			name:       "div",
			expression: "4/2",
			want_err:   false,
			want:       2,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			answMap := make(map[string]float64)
			errChan := make(chan error, 1)
			operations := map[string]int{
				"add":  0,
				"sub":  0,
				"mult": 0,
				"div":  0,
			}
			var wg sync.WaitGroup
			var mu sync.Mutex
			wg.Add(1)
			EvaluateSimpleExpression(tc.expression, operations, answMap, errChan, &mu, &wg)
			close(errChan)
			if tc.want_err {
				got_err := false
				for err := range errChan {
					if err != nil {
						got_err = true
					}
				}
				if !got_err {
					t.Errorf("EvaluateSimpleExpression(%v, ...); want error, but got none", tc.expression)
				}
			} else {
				got := answMap[tc.expression]
				if got != tc.want {
					t.Errorf("EvaluateSimpleExpression(%v, ...) = %v; want %v", tc.expression, got, tc.want)
				}
			}
		})
	}
}

func TestEvaluateComplexExpression(t *testing.T) {
	cases := []struct {
		name       string
		expression string
		want_err   bool
		want       float64
	}{
		{
			name:       "some letters",
			expression: "abc",
			want_err:   true,
		},
		{
			name:       "parentheses chaos",
			expression: "(1+1))))",
			want_err:   true,
		},
		{
			name:       "division by zero",
			expression: "1/0",
			want_err:   true,
		},
		{
			name:       "valid expression",
			expression: "1+1-1",
			want_err:   false,
			want:       1,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			operations := map[string]int{
				"add":  0,
				"sub":  0,
				"mult": 0,
				"div":  0,
			}
			got, err := EvaluateComplexExpression(tc.expression, 5, operations)
			if tc.want_err {
				if err == nil {
					t.Errorf("EvaluateComplexExpression(%v, ...); want error, but got none", tc.expression)
				}
			} else {
				if got != tc.want {
					t.Errorf("EvaluateComplexExpression(%v, ...) = %v; want %v", tc.expression, got, tc.want)
				}
			}
		})
	}
}
