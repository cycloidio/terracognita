package validate

import "testing"

func TestElasticEmailAddress(t *testing.T) {
	testCases := []struct {
		Input    string
		Expected bool
	}{
		{
			Input:    "abc@xyz.com",
			Expected: true,
		},
		{
			Input:    "abc@dyg@jad.com",
			Expected: false,
		},
		{
			Input:    "abc@",
			Expected: false,
		},
		{
			Input:    "abc@xyz",
			Expected: false,
		},
		{
			Input:    "abcdyg@jad.com.cdhc",
			Expected: true,
		},
	}
	for _, v := range testCases {
		_, errors := ElasticEmailAddress(v.Input, "email_address")
		result := len(errors) == 0
		if result != v.Expected {
			t.Fatalf("Expected the result to be %t but got %t (and %d errors)", v.Expected, result, len(errors))
		}
	}
}
