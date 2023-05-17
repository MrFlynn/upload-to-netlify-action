package actions

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_NewLogger(t *testing.T) {
	diff := cmp.Diff(
		&Logger{Output: os.Stdout}, NewLogger(), cmpopts.IgnoreUnexported(os.File{}),
	)

	if diff != "" {
		t.Errorf("Logger mismatch (-want +got):\n%s", diff)
	}
}

const (
	content = "lorem ipsum"
	number  = 10
	format  = "%s %d"
)

func TestLogger_Debug(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Debug(content)

	diff := cmp.Diff("::debug::"+content+"\n", logger.Output.(*strings.Builder).String())
	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_Debugf(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Debugf(format, content, number)

	diff := cmp.Diff(
		"::debug::"+content+" "+strconv.Itoa(number)+"\n",
		logger.Output.(*strings.Builder).String(),
	)

	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_Info(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Info(content)

	diff := cmp.Diff(content+"\n", logger.Output.(*strings.Builder).String())
	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_Infof(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Infof(format, content, number)

	diff := cmp.Diff(
		content+" "+strconv.Itoa(number)+"\n",
		logger.Output.(*strings.Builder).String(),
	)

	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_Warn(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Warn(content)

	diff := cmp.Diff("::warning::"+content+"\n", logger.Output.(*strings.Builder).String())
	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_Warnf(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Warnf(format, content, number)

	diff := cmp.Diff(
		"::warning::"+content+" "+strconv.Itoa(number)+"\n",
		logger.Output.(*strings.Builder).String(),
	)

	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_Error(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Error(content)

	diff := cmp.Diff("::error::"+content+"\n", logger.Output.(*strings.Builder).String())
	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_Errorf(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.Errorf(format, content, number)

	diff := cmp.Diff(
		"::error::"+content+" "+strconv.Itoa(number)+"\n",
		logger.Output.(*strings.Builder).String(),
	)

	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

func TestLogger_SetSecret(t *testing.T) {
	logger := &Logger{Output: &strings.Builder{}}
	logger.SetSecret(content)

	diff := cmp.Diff("::add-mask::"+content+"\n", logger.Output.(*strings.Builder).String())
	if diff != "" {
		t.Errorf("Output mismatch (-want +got):\n%s", diff)
	}
}

type testGetInput struct {
	name          string
	value         string
	options       GetInputOptions
	expectedError error
}

var (
	environmentKey = "key"
	compareErrors  = cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}

		return x.Error() == y.Error()
	})
)

func Test_GetInput(t *testing.T) {
	testCases := []struct {
		result string
		testGetInput
	}{
		{
			result: "lorem ipsum",
			testGetInput: testGetInput{
				name:  "simple",
				value: "lorem ipsum",
				options: GetInputOptions{
					Required:       false,
					TrimWhitespace: false,
				},
				expectedError: nil,
			},
		},
		{
			result: "  lorem ipsum   ",
			testGetInput: testGetInput{
				name:  "with_whitespace",
				value: "  lorem ipsum   ",
				options: GetInputOptions{
					Required:       false,
					TrimWhitespace: false,
				},
				expectedError: nil,
			},
		},
		{
			testGetInput: testGetInput{
				name:  "required_with_missing_value",
				value: "",
				options: GetInputOptions{
					Required:       true,
					TrimWhitespace: false,
				},
				expectedError: fmt.Errorf(
					"input %s is required but was not given", environmentKey,
				),
			},
		},
		{
			result: "lorem ipsum",
			testGetInput: testGetInput{
				name:  "trim_whitespace",
				value: "  lorem ipsum   ",
				options: GetInputOptions{
					Required:       false,
					TrimWhitespace: true,
				},
				expectedError: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("INPUT_KEY", tc.value)

			result, err := GetInput(environmentKey, tc.options)
			if diff := cmp.Diff(tc.expectedError, err, compareErrors); diff != "" {
				t.Errorf("Error mismatch (-want +got):\n%s", diff)
				return
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(tc.result, result); diff != "" {
				t.Errorf("Value mismatch (-want +got):\n%s", diff)
				return
			}
		})
	}
}

func Test_GetMultilineInput(t *testing.T) {
	testCases := []struct {
		result []string
		testGetInput
	}{
		{
			result: []string{"lorem", "ipsum"},
			testGetInput: testGetInput{
				name:  "simple",
				value: "lorem\nipsum",
				options: GetInputOptions{
					Required:       false,
					TrimWhitespace: false,
				},
				expectedError: nil,
			},
		},
		{
			result: []string{},
			testGetInput: testGetInput{
				name:  "required_with_missing_value",
				value: "",
				options: GetInputOptions{
					Required:       true,
					TrimWhitespace: false,
				},
				expectedError: fmt.Errorf(
					"input %s is required but was not given", environmentKey,
				),
			},
		},
		{
			result: []string{"lorem", "ipsum"},
			testGetInput: testGetInput{
				name:  "trim_whitespace",
				value: "  lorem  \nipsum  ",
				options: GetInputOptions{
					Required:       false,
					TrimWhitespace: true,
				},
				expectedError: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("INPUT_KEY", tc.value)

			result, err := GetMultilineInput(environmentKey, tc.options)
			if diff := cmp.Diff(tc.expectedError, err, compareErrors); diff != "" {
				t.Errorf("Error mismatch (-want +got):\n%s", diff)
				return
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(tc.result, result); diff != "" {
				t.Errorf("Value mismatch (-want +got):\n%s", diff)
				return
			}
		})
	}
}
