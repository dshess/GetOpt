package getopt

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

var descRe = func() func() (*regexp.Regexp, error) {
	var re *regexp.Regexp

	return func() (*regexp.Regexp, error) {
		if re != nil {
			return re, nil
		}

		// A descriptor is a flag name, an optional type, and
		// optional modifiers.  Not all modifiers apply to all
		// types.
		re, err := regexp.Compile("^([-_a-zA-Z0-9]+)([=:][bifs])?([!+@])?$")
		return re, err
	}
}()

func parseOption(oc *optionCollection, desc string, ptr any) error {
	// Parse the descriptor for requested flag attributes.
	negatable := false
	counting := false
	optional := false
	dArray := false
	dType := '?'

	re, err := descRe()
	if err != nil {
		return err
	}

	match := re.FindStringSubmatch(desc)
	if match == nil {
		// TODO: More-specific error reporting?
		return errors.New("descriptor not understood")
	}
	// desc, eType, modifier := match[1], match[2], match[3]
	// Can't get type directly, though, since it would be a string.

	if len(match[3]) > 0 {
		if match[3][0] == '!' {
			negatable = true
		} else if match[3][0] == '+' {
			counting = true
		} else if match[3][0] == '@' {
			dArray = true
		} else {
			return errors.New("descriptor not understood")
		}
	}

	if len(match[2]) > 0 {
		optional = (match[2][0] == ':')
		dType = rune(match[2][1])
	}
	name := match[1]

	// Probe the argument for type information.
	pType := '?'
	pArray := false

	switch ptr.(type) {
	case *bool:
		pType = 'b'
	case *int:
		pType = 'i'
	case *[]int:
		pType = 'i'
		pArray = true
	case *float64:
		pType = 'f'
	case *[]float64:
		pType = 'f'
		pArray = true
	case *string:
		pType = 's'
	case *[]string:
		pType = 's'
		pArray = true
	default:
		return errors.New("type not recognized")
	}

	// Make sure descriptor requests are valid for actual pointer type.
	if dType != '?' && dType != pType {
		// descriptor type doesn't match probed type.
		return errors.New("descriptor type mismatch")
	} else if counting && !(pType == 'i' && !pArray) {
		// counting requires direct integer only
		return errors.New("descriptor type mismatch")
	} else if negatable && pType != 'b' {
		// negatable requires boolean
		return errors.New("descriptor type mismatch")
	} else if dArray != pArray {
		// The array sense has to match.
		// TODO: Consider requiring the match only if explicit type info
		// is provided, so a name-only descriptor bound to an array
		// pointer can get array treatment.
		return errors.New("descriptor type mismatch")
	} else if optional && pArray {
		// Optional doesn't make sense with arrays.
		return errors.New("descriptor type mismatch")
	} else if optional && !(pType == 'i' || pType == 's' || pType == 'f') {
		// Optional only makes sense with int and string.
		return errors.New("descriptor type mismatch")
	}

	// Check for ambiguous options.  This could be tested in the adders, at
	// the expense of error handling boilerplate.
	if err := oc.checkNameConflict(name, negatable); err != nil {
		return err
	}

	// Setup processing object for the pointer.
	switch f := ptr.(type) {
	case *bool:
		if negatable {
			oc.addNegatableHandler(name, (*bool)(f))
		} else {
			oc.addSimpleHandler(name, (*bool)(f))
		}
	case *int:
		if counting {
			oc.addCountingHandler(name, (*int)(f))
		} else if optional {
			oc.addOptionalIntHandler(name, (*int)(f))
		} else {
			oc.addIntHandler(name, (*int)(f))
		}
	case *[]int:
		oc.addIntArrayHandler(name, (*[]int)(f))
	case *float64:
		if optional {
			oc.addOptionalFloatHandler(name, (*float64)(f))
		} else {
			oc.addFloatHandler(name, (*float64)(f))
		}
	case *[]float64:
		oc.addFloatArrayHandler(name, (*[]float64)(f))
	case *string:
		if optional {
			oc.addOptionalStringHandler(name, (*string)(f))
		} else {
			oc.addStringHandler(name, (*string)(f))
		}
	case *[]string:
		oc.addStringArrayHandler(name, (*[]string)(f))
	default:
		return errors.New("type not recognized")
	}

	return nil
}

func parseOptions(oc *optionCollection, a ...any) error {
	// Always two there are.  No more.  No less.  A Descriptor and a
	// Pointer.
	for len(a) >= 2 {
		var desc string

		switch f := a[0].(type) {
		case string:
			desc = string(f)
		default:
			return errors.New("descriptor must be string")
		}

		err := parseOption(oc, desc, a[1])
		if err != nil {
			return err
		}
		a = a[2:]
	}

	if len(a) == 1 {
		return errors.New("odd number of arguments")
	}

	return nil
}

func processArgs(oc *optionCollection, args []string) ([]string, error) {
	rest := args

	for len(rest) > 0 {
		name := rest[0]
		if !strings.HasPrefix(name, "--") {
			break
		}
		name = name[2:]
		rest = rest[1:]
		if len(name) == 0 {
			break
		}

		var zeroOrOne []string

		i := strings.IndexRune(name, '=')
		if i > -1 {
			zeroOrOne = []string{name[i+1:]}
			name = name[:i]
		}

		h, ok := oc.handlers[name]
		if !ok {
			return args, errors.New("Arg " + name + " not recognized")
		}

		if h.getType() == optionNoArg {
			// Nothing
		} else if len(zeroOrOne) > 0 {
			// Nothing, already have an arg
		} else if len(rest) < 1 {
			if h.getType() == optionRequiredArg {
				return args, errors.New("missing required argument")
			}
			// For optional, no more args is fine
		} else if h.getType() == optionOptionalArg && strings.HasPrefix(rest[0], "--") {
			// Nothing, next arg looks flag-like
		} else {
			zeroOrOne = rest[0:1]
			rest = rest[1:]
		}

		c, err := h.handle(zeroOrOne)
		if err != nil {
			return args, err
		}
		oc.committers = append(oc.committers, c)
	}
	oc.commit()
	return rest, nil
}

// Process the argument descriptors and pointers from the passed slice of
// arguments.  Returns the remaining arguments in case of success, or the
// original arguments in case of error.
func GetOptions(args []string, a ...any) ([]string, error) {
	oc := newOptionCollection()

	err := parseOptions(oc, a...)
	if err != nil {
		return args, err
	}

	return processArgs(oc, args)
}

// GetOSOptions wraps [GetOptions] to read options from [os.Args][1:],
// destructively updating [os.Args][1:] in case of success.
func GetOSOptions(a ...any) error {
	ret, err := GetOptions(os.Args[1:], a...)
	if err != nil {
		return err
	}

	replace := []string{os.Args[1]}
	os.Args = append(replace, ret...)
	return err
}

// Perl's GetOptions() takes descriptors:
// GetOptions("length=i" => \$length,
//            "file=s" => \$file,
//            "verbose!" => \$verbose)
//
// So something like:
// GetOptions("length", &int,
//            "file", &string,
//            "verbose", &bool)

// GetOpt::Long notes:
// "name" is a simple no-arg bool
// "name+" is a counting no-arg int
// "name!" is a negatable no-arg bool
// "name=t" is a typed arg.  Type can be s (string), i (integer), or f (float)
// "name:t" is an optional typed arg, zero if next arg looks flag-like
// "name=t@" is a typed arg which feeds an array.
// "name=t{n}" is n values of the type and populates an array.
// "name=t{n,m}" is n to m values
// "name=s%" takes key=value arguments and populates a hash.
// "name|alt=t" allows alternate names
//
// Uncertain:
// "name:10" integer option default 10
// "name:+i" counting integer

// Allow -v

// Allow an option-processing closure which takes things as input.
// Allow a flag closure which eats the flag (such as for --help).

// TODO: Probably don't implement {n} for now.
// TODO: Probably don't implement % for now.
// TODO: Implement alternate names.
// TODO: Allow short options as alternates.
// TODO: Allow short option batching.
