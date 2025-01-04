GetOpt implements command-line flag parsing.  This is abandonware, as I
replaced it with [github.com/dshess/Opts], which uses descriptor functions
instead of parsing descriptor strings, because it is cleaner.

# Usage

Options-handling modeled on Perl's
[Getopt::Long](https://perldoc.perl.org/Getopt::Long).
To handle a string, an int, and a flag:

	data := "file.dat"
	length := 24
	var verbose bool
    err := GetOSOptions(
		"length=i", &length, // numeric
		"files=s", &data, // string
		"verbose", &verbose, // flag
	)
	if err != nil {
		log.Fatal("Error in command-line arguments:", err)
	}

If the command is passed "--files=hello.world --length 10 --verbose rest",
then after GetOSOptions, data will be "hello.world", length will be 10,
verbose will be true, and [os.Args](https://pkg.go.dev/os#Args)[1:] will be
[]string{"rest"}.

# Command line flag syntax

This code only handles --option style of options.  "--" ends option
processing.  Boolean options can only be negatable or simple, with no
parameters (so --option or --nooption).  Int, Float, or String options can
be provided as --option=value or --option value.  Optional options deliver
the provided value if the option is seen with no further arguments, or if
the next argument itself looks like an option.

# Option descriptors

The option list must have pairs of values, a string descriptor and a
pointer to someplace to store values.  It is an error if the pointed-to
type is not compatible with the descriptor.  It is also an error if
multiple pointers to the same variable are provided.

  - "flag", &boolValue - --flag sets boolValue to true
  - "flag!", &boolValue - --flag sets true, --noflag sets false
  - "count+", &intValue - each --count increments intValue
  - "value=i", &intValue - --value 15 or --value=15 sets intValue to 15
  - "value:i", &intValue - flat with optional value, if no value is
    provided (no more args, or next looks like a flag), stores 0 to
    intValue
  - "value=i@", &intArray - each occurance appends value to the array
  - "value", &intValue - infers that the value should be parsed as an integer
  - "value:", &intValue - optional with inferred integer type
  - "value@", &intArray - array with inferred integer type
  - "value=f", "value=f@", "value:f" with float-typed pointer for float
    version, or drop =f to infer the type.
  - "value=s", "value=s@", "value:s" with string-typed pointer for string
    version, or drop =s to infer the type.

Descriptors in the style of "value=s" are more in the style of
Getopt::Long, because Perl's typing is different than Go's.  Perl can infer
array versus scalar, but not int versus string.  Go can infer int vs
string, so it may make sense to not use typing in the descriptor.  OTOH,
the descriptor makes the type clear in context.
