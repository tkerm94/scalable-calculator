package agent

import "testing"

func TestParenthesesCheck(t *testing.T) {
	cases := []struct {
		name       string
		expression string
		want       bool
	}{
		{
			name:       "not matched",
			expression: ")1+1(",
			want:       false,
		},
		{
			name:       "some extra",
			expression: "(1+1))))",
			want:       false,
		},
		{
			name:       "correct",
			expression: "((1+1))",
			want:       true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := ParenthesesCheck(tc.expression)
			if got != tc.want {
				t.Errorf("ParenthesesCheck(%v) = %v; want %v", tc.expression, got, tc.want)
			}
		})
	}
}

func TestParenthesesClear(t *testing.T) {
	cases := []struct {
		name       string
		expression string
		want       string
	}{
		{
			name:       "extra on one side",
			expression: "((1+1)))",
			want:       "(1+1))",
		},
		{
			name:       "no parentheses",
			expression: "1+1",
			want:       "1+1",
		},
		{
			name:       "matched",
			expression: "((1+1))",
			want:       "(1+1)",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := ParenthesesClear(tc.expression)
			if got != tc.want {
				t.Errorf("ParenthesesClear(%v) = %v; want %v", tc.expression, got, tc.want)
			}
		})
	}
}

func TestExtractSimpleExpressions(t *testing.T) {
	cases := []struct {
		name       string
		expression string
		want       map[int]string
	}{
		{
			name:       "one simple expression",
			expression: "1+1+1",
			want: map[int]string{
				0: "1+1",
			},
		},
		{
			name:       "two simple expression",
			expression: "1+1-1+1",
			want: map[int]string{
				0: "1+1",
				4: "1+1",
			},
		},
		{
			name:       "addition before multiplying",
			expression: "1+1*2",
			want: map[int]string{
				2: "1*2",
			},
		},
		{
			name:       "negative number at the start",
			expression: "-1+1",
			want: map[int]string{
				0: "-1+1",
			},
		},
		{
			name:       "more negative numbers",
			expression: "-1-1-1-1",
			want: map[int]string{
				0: "-1-1",
			},
		},
		{
			name:       "negative number in the middle",
			expression: "1+1-1-1+1",
			want: map[int]string{
				0: "1+1",
				5: "-1+1",
			},
		},
		{
			name:       "multiplying multiple numbers",
			expression: "1*1*1*1",
			want: map[int]string{
				0: "1*1",
				4: "1*1",
			},
		},
		{
			name:       "dividing multiple numbers",
			expression: "1/1/1/1",
			want: map[int]string{
				0: "1/1",
			},
		},
		{
			name:       "multiplying before dividing",
			expression: "1*1/1",
			want: map[int]string{
				0: "1*1",
			},
		},
		{
			name:       "dividing before multiplying",
			expression: "1/1*1",
			want: map[int]string{
				0: "1/1",
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractSimpleExpressions(tc.expression)
			for key, val := range got {
				if val != tc.want[key] {
					t.Errorf("ExtractSimpleExpressions(%v) = %v; want %v", tc.expression, got, tc.want)
					break
				}
			}
		})
	}
}
