package getopt

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestBase_Empty(t *testing.T) {
	args := []string{}
	assert.Empty(t, args)

	a, err := GetOptions(args)

	assert.NoError(t, err)
	assert.Empty(t, a)
}

func TestBase_NoDesc(t *testing.T) {
	args := []string{"notaflag", "also not a flag"}
	assert.NotEmpty(t, args)

	a, err := GetOptions(args)

	assert.NoError(t, err)
	assert.Equal(t, a, args)
}

func TestBase_OddDesc1(t *testing.T) {
	args := []string{"notaflag"}

	a, err := GetOptions(args, "x")

	// TODO: More specific?
	assert.ErrorContains(t, err, "odd number of arguments")
	assert.Equal(t, a, args)
}

func TestBase_OddDesc2(t *testing.T) {
	args := []string{"-x", "notaflag"}
	value := false

	a, err := GetOptions(args, "x", &value, "y")

	// TODO: More specific?
	assert.ErrorContains(t, err, "odd number of arguments")
	assert.False(t, value)
	assert.Equal(t, a, args)
}

func TestBase_ExplicitEnd(t *testing.T) {
	args := []string{"--flag", "--", "notaflag"}
	flag := false

	a, err := GetOptions(args, "flag", &flag)

	assert.NoError(t, err)
	assert.True(t, flag)
	assert.Equal(t, a, args[2:])
}

func TestSimple_Flag(t *testing.T) {
	args := []string{"--flag"}
	flag := false

	a, err := GetOptions(args, "flag", &flag)

	assert.NoError(t, err)
	assert.True(t, flag)
	assert.Empty(t, a)
}

func TestSimple_FlagExtra(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	flag := false

	a, err := GetOptions(args, "flag", &flag)

	assert.NoError(t, err)
	assert.True(t, flag)
	assert.Equal(t, a, args[1:])
}

func TestSimple_FlagTyped(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	flag := false

	a, err := GetOptions(args, "flag=b", &flag)

	assert.NoError(t, err)
	assert.True(t, flag)
	assert.Equal(t, a, args[1:])
}

// Type must match.
func TestSimple_FlagTypedMismatch(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	value := 10

	a, err := GetOptions(args, "flag=b", &value)

	assert.ErrorContains(t, err, "descriptor type mismatch")
	assert.Equal(t, value, 10)
	assert.Equal(t, a, args)
}

// Simple flags don't enable --noflag version.
func TestSimple_FlagNoFlag(t *testing.T) {
	args := []string{"--noflag", "not a flag"}
	flag := true

	a, err := GetOptions(args, "flag", &flag)

	assert.ErrorContains(t, err, "not recognized")
	assert.True(t, flag)
	assert.Equal(t, a, args)
}

// Simple flags don't allow optional.
func TestSimple_FlagNoOptional(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	flag := false

	a, err := GetOptions(args, "flag:b", &flag)

	assert.ErrorContains(t, err, "descriptor type mismatch")
	assert.False(t, flag)
	assert.Equal(t, a, args)
}

func TestNegatable_Base(t *testing.T) {
	args := []string{"--flag"}
	flag := false
	assert.NotEmpty(t, args)
	a, err := GetOptions(args, "flag!", &flag)
	assert.NoError(t, err)
	assert.True(t, flag)
	assert.Empty(t, a)
}

func TestNegatable_Extra(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	flag := false

	a, err := GetOptions(args, "flag!", &flag)

	assert.NoError(t, err)
	assert.True(t, flag)
	assert.Equal(t, a, args[1:])
}

func TestNegatable_ExtraNegative(t *testing.T) {
	args := []string{"--noflag", "not a flag"}
	flag := true

	a, err := GetOptions(args, "flag!", &flag)

	assert.NoError(t, err)
	assert.False(t, flag)
	assert.Equal(t, a, args[1:])
}

// Type must match.
func TestNegatable_TypeMismatch(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	value := 10

	a, err := GetOptions(args, "flag!", &value)

	assert.ErrorContains(t, err, "descriptor type mismatch")
	assert.Equal(t, value, 10)
	assert.Equal(t, a, args)
}

func TestCounting_Base(t *testing.T) {
	args := []string{"--count"}
	count := 3
	assert.NotEmpty(t, args)

	a, err := GetOptions(args, "count+", &count)

	assert.NoError(t, err)
	assert.Equal(t, count, 4)
	assert.Empty(t, a)
}

func TestCounting_BaseExtra(t *testing.T) {
	args := []string{"--count", "not a flag"}
	count := 3

	a, err := GetOptions(args, "count+", &count)

	assert.NoError(t, err)
	assert.Equal(t, count, 4)
	assert.Equal(t, a, args[1:])
}

// Counting flags don't enable --noflag version.
func TestCounting_NoNegate(t *testing.T) {
	args := []string{"--nocount", "not a flag"}
	count := 3

	a, err := GetOptions(args, "count+", &count)

	assert.ErrorContains(t, err, "not recognized")
	assert.Equal(t, count, 3)
	assert.Equal(t, a, args)
}

// Type must match.
func TestCounting_TypeMismatch(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	flag := false

	a, err := GetOptions(args, "count+", &flag)

	assert.ErrorContains(t, err, "descriptor type mismatch")
	assert.False(t, flag)
	assert.Equal(t, a, args)
}

func TestInteger_NoValue(t *testing.T) {
	args := []string{"--value"}
	value := 3

	a, err := GetOptions(args, "value", &value)

	assert.ErrorContains(t, err, "missing required argument")
	assert.Equal(t, value, 3)
	assert.Equal(t, a, args)
}

func TestInteger_InvalidArg(t *testing.T) {
	args := []string{"--value", "not a flag"}
	value := 3

	a, err := GetOptions(args, "value", &value)

	// TODO: Should this be "Argument not a number" or similar?
	assert.ErrorContains(t, err, "invalid syntax")
	assert.Equal(t, value, 3)
	assert.Equal(t, a, args)
}

func TestInteger_SeparateValue(t *testing.T) {
	args := []string{"--value", "5"}
	value := 3

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5)
	assert.Empty(t, a)
}

func TestInteger_SeparateValueExtra(t *testing.T) {
	args := []string{"--value", "5", "not a flag"}
	value := 3

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5)
	assert.Equal(t, a, args[2:])
}

func TestInteger_InlineValue(t *testing.T) {
	args := []string{"--value=5"}
	value := 3

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5)
	assert.Empty(t, a)
}

func TestInteger_InlineValueExtra(t *testing.T) {
	args := []string{"--value=5", "not a flag"}
	value := 3

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5)
	assert.Equal(t, a, args[1:])
}

func TestInteger_Typed(t *testing.T) {
	args := []string{"--value", "5", "not a flag"}
	value := 3

	a, err := GetOptions(args, "value=i", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5)
	assert.Equal(t, a, args[2:])
}

func TestInteger_Optional(t *testing.T) {
	args := []string{"--value"}
	value := 10

	a, err := GetOptions(args, "value:i", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 0)
	assert.Empty(t, a)
}

func TestInteger_OptionalWithValue(t *testing.T) {
	args := []string{"--value", "5"}
	value := 10

	a, err := GetOptions(args, "value:i", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5)
	assert.Empty(t, a)
}

func TestInteger_OptionalWithFlag(t *testing.T) {
	args := []string{"--value", "--flag", "not a flag"}
	value := 10
	flag := false

	a, err := GetOptions(args, "value:i", &value, "flag", &flag)

	assert.NoError(t, err)
	assert.Equal(t, value, 0)
	assert.True(t, flag)
	assert.Equal(t, a, args[2:])
}

func TestInteger_Array1(t *testing.T) {
	args := []string{"--value", "5", "not a flag"}
	values := []int{}

	a, err := GetOptions(args, "value@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, values[0], 5)
	assert.Equal(t, a, args[2:])
}

func TestInteger_Array2(t *testing.T) {
	args := []string{"--value", "5", "--value=3", "not a flag"}
	values := []int{}

	a, err := GetOptions(args, "value@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, values[0], 5)
	assert.Equal(t, values[1], 3)
	assert.Equal(t, a, args[3:])
}

func TestInteger_TypedArray1(t *testing.T) {
	args := []string{"--value", "5", "not a flag"}
	values := []int{}

	a, err := GetOptions(args, "value=i@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, values[0], 5)
	assert.Equal(t, a, args[2:])
}

// Type must match.
func TestInteger_TypeMismatch(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	value := "hello"

	a, err := GetOptions(args, "value=i", &value)

	assert.ErrorContains(t, err, "descriptor type mismatch")
	assert.Equal(t, value, "hello")
	assert.Equal(t, a, args)
}

func TestFloat_NoValue(t *testing.T) {
	args := []string{"--value"}
	value := 3.14

	a, err := GetOptions(args, "value", &value)

	assert.ErrorContains(t, err, "missing required argument")
	assert.Equal(t, value, 3.14)
	assert.Equal(t, a, args)
}

func TestFloat_InvalidArg(t *testing.T) {
	args := []string{"--value", "not a flag"}
	value := 3.14

	a, err := GetOptions(args, "value", &value)

	// TODO: Should this be "Argument not a number" or similar?
	assert.ErrorContains(t, err, "invalid syntax")
	assert.Equal(t, value, 3.14)
	assert.Equal(t, a, args)
}

func TestFloat_SeparateValue(t *testing.T) {
	args := []string{"--value", "5.5"}
	value := 3.14

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5.5)
	assert.Empty(t, a)
}

func TestFloat_SeparateValueExtra(t *testing.T) {
	args := []string{"--value", "5.5", "not a flag"}
	value := 3.14

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5.5)
	assert.Equal(t, a, args[2:])
}

func TestFloat_InlineValue(t *testing.T) {
	args := []string{"--value=5.5"}
	value := 3.14

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5.5)
	assert.Empty(t, a)
}

func TestFloat_InlineValueExtra(t *testing.T) {
	args := []string{"--value=5.5", "not a flag"}
	value := 3.14

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5.5)
	assert.Equal(t, a, args[1:])
}

func TestFloat_Typed(t *testing.T) {
	args := []string{"--value", "5.5", "not a flag"}
	value := 3.14

	a, err := GetOptions(args, "value=f", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5.5)
	assert.Equal(t, a, args[2:])
}

func TestFloat_Optional(t *testing.T) {
	args := []string{"--value"}
	value := 3.14

	a, err := GetOptions(args, "value:f", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 0.0)
	assert.Empty(t, a)
}

func TestFloat_OptionalWithValue(t *testing.T) {
	args := []string{"--value", "5.5"}
	value := 3.14

	a, err := GetOptions(args, "value:f", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, 5.5)
	assert.Empty(t, a)
}

func TestFloat_OptionalWithFlag(t *testing.T) {
	args := []string{"--value", "--flag", "not a flag"}
	value := 3.14
	flag := false

	a, err := GetOptions(args, "value:f", &value, "flag", &flag)

	assert.NoError(t, err)
	assert.Equal(t, value, 0.0)
	assert.True(t, flag)
	assert.Equal(t, a, args[2:])
}

func TestFloat_Array1(t *testing.T) {
	args := []string{"--value", "5.5", "not a flag"}
	values := []float64{}

	a, err := GetOptions(args, "value@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, values[0], 5.5)
	assert.Equal(t, a, args[2:])
}

func TestFloat_Array2(t *testing.T) {
	args := []string{"--value", "5.5", "--value=3.14", "not a flag"}
	values := []float64{}

	a, err := GetOptions(args, "value@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, values[0], 5.5)
	assert.Equal(t, values[1], 3.14)
	assert.Equal(t, a, args[3:])
}

func TestFloat_TypedArray1(t *testing.T) {
	args := []string{"--value", "5.5", "not a flag"}
	values := []float64{}

	a, err := GetOptions(args, "value=f@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, values[0], 5.5)
	assert.Equal(t, a, args[2:])
}

// Type must match.
func TestFloat_TypeMismatch(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	value := "hello"

	a, err := GetOptions(args, "value=f", &value)

	assert.ErrorContains(t, err, "descriptor type mismatch")
	assert.Equal(t, value, "hello")
	assert.Equal(t, a, args)
}

func TestString_NoValue(t *testing.T) {
	args := []string{"--value"}
	value := "hello"

	a, err := GetOptions(args, "value", &value)

	assert.ErrorContains(t, err, "missing required argument")
	assert.Equal(t, value, "hello")
	assert.Equal(t, a, args)
}

func TestString_SeparateValue(t *testing.T) {
	args := []string{"--value", "world"}
	value := "hello"

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, "world")
	assert.Empty(t, a)
}

func TestString_SeparateValueExtra(t *testing.T) {
	args := []string{"--value", "world", "not a flag"}
	value := "hello"

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, "world")
	assert.Equal(t, a, args[2:])
}

func TestString_InlineValue(t *testing.T) {
	args := []string{"--value=world"}
	value := "hello"

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, "world")
	assert.Empty(t, a)
}

func TestString_InlineValueExtra(t *testing.T) {
	args := []string{"--value=world", "not a flag"}
	value := "hello"

	a, err := GetOptions(args, "value", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, "world")
	assert.Equal(t, a, args[1:])
}

func TestString_Typed(t *testing.T) {
	args := []string{"--value", "world", "not a flag"}
	value := "hello"

	a, err := GetOptions(args, "value=s", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, "world")
	assert.Equal(t, a, args[2:])
}

func TestString_Optional(t *testing.T) {
	args := []string{"--value"}
	value := "something"

	a, err := GetOptions(args, "value:s", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, "")
	assert.Empty(t, a)
}

func TestString_OptionalWithValue(t *testing.T) {
	args := []string{"--value", "hello"}
	value := "something"

	a, err := GetOptions(args, "value:s", &value)

	assert.NoError(t, err)
	assert.Equal(t, value, "hello")
	assert.Empty(t, a)
}

func TestString_OptionalWithFlag(t *testing.T) {
	args := []string{"--value", "--flag", "not a flag"}
	value := "something"
	flag := false

	a, err := GetOptions(args, "value:s", &value, "flag", &flag)

	assert.NoError(t, err)
	assert.Equal(t, value, "")
	assert.True(t, flag)
	assert.Equal(t, a, args[2:])
}

func TestString_Array1(t *testing.T) {
	args := []string{"--value", "world", "not a flag"}
	values := []string{}

	a, err := GetOptions(args, "value@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, values[0], "world")
	assert.Equal(t, a, args[2:])
}

func TestString_Array2(t *testing.T) {
	args := []string{"--value", "world", "--value=earth", "not a flag"}
	values := []string{}

	a, err := GetOptions(args, "value@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, values[0], "world")
	assert.Equal(t, values[1], "earth")
	assert.Equal(t, a, args[3:])
}

func TestString_TypedArray1(t *testing.T) {
	args := []string{"--value", "world", "not a flag"}
	values := []string{}

	a, err := GetOptions(args, "value=s@", &values)

	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, values[0], "world")
	assert.Equal(t, a, args[2:])
}

// Type must match.
func TestString_TypeMismatch(t *testing.T) {
	args := []string{"--flag", "not a flag"}
	value := 10

	a, err := GetOptions(args, "value=s", &value)

	assert.ErrorContains(t, err, "descriptor type mismatch")
	assert.Equal(t, value, 10)
	assert.Equal(t, a, args)
}

func ExampleGetOptions() {
	args := []string{
		"--files=hello.world", "--length", "10", "--verbose", "rest",
	}

	data := "file.dat"
	length := 24
	var verbose bool
	rest, err := GetOptions(args,
		"length=i", &length, // numeric
		"files=s", &data, // string
		"verbose", &verbose, // flag
	)
	if err != nil {
		log.Fatal("Error in command-line arguments:", err)
	}
	fmt.Printf("length:%d, data:%s, verbose:%t, rest:%s\n",
		length, data, verbose, rest)
	// Output:
	// length:10, data:hello.world, verbose:true, rest:[rest]
}
