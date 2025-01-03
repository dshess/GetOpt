package getopt

import (
	"errors"
	"strconv"
)

// TODO: Right now, this is structured as a map of handler objects, which
// generate an array of commit objects.  This works just as well as a map of
// closures which generate an array of closures, and it has less boilderplate.
// BUT, each object also carries a context structure around 4k in size.

// optionCommitters allow deferring changes until the end of options processing.
type optionCommitter interface {
	commit()
}

// Store a value at a pointer on commit.
type optionSimpleCommitter[T any] struct {
	value  T
	option *T
}

func (o optionSimpleCommitter[_]) commit() {
	*o.option = o.value
}

// Increment a pointed-to value on commit.
type optionCountingCommitter struct {
	option *int
}

func (o optionCountingCommitter) commit() {
	*o.option++
}

// Append value to an array on commit.
type optionArrayCommitter[T any] struct {
	value  T
	option *[]T
}

func (o optionArrayCommitter[_]) commit() {
	*o.option = append(*o.option, o.value)
}

// optionHandler provides a hint as to how many arguments, and a handler to call
// with those arguments.  The handler generates an optionCommitter to be called
// later.
type optionType int

const (
	optionNoArg optionType = iota
	optionOptionalArg
	optionRequiredArg
)

type optionHandler interface {
	getType() optionType
	handle(args []string) (optionCommitter, error)
}

type optionSimpleHandler struct {
	t      optionType
	value  bool
	option *bool
}

func (oh optionSimpleHandler) getType() optionType {
	return oh.t
}
func (oh optionSimpleHandler) handle(args []string) (optionCommitter, error) {
	c := optionSimpleCommitter[bool]{oh.value, oh.option}
	return c, nil
}

type optionCountingHandler struct {
	t      optionType
	option *int
}

func (oh optionCountingHandler) getType() optionType {
	return oh.t
}
func (oh optionCountingHandler) handle(args []string) (optionCommitter, error) {
	c := optionCountingCommitter{oh.option}
	return c, nil
}

type optionIntHandler struct {
	t      optionType
	option *int
}

func (oh optionIntHandler) getType() optionType {
	return oh.t
}
func (oh optionIntHandler) handle(args []string) (optionCommitter, error) {
	if len(args) < 1 {
		args = []string{"0"}
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}
	c := optionSimpleCommitter[int]{i, oh.option}
	return c, nil
}

type optionIntArrayHandler struct {
	t      optionType
	option *[]int
}

func (oh optionIntArrayHandler) getType() optionType {
	return oh.t
}

func (oh optionIntArrayHandler) handle(args []string) (optionCommitter, error) {
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, err
	}
	c := optionArrayCommitter[int]{i, oh.option}
	return c, nil
}

type optionFloatHandler struct {
	t      optionType
	option *float64
}

func (oh optionFloatHandler) getType() optionType {
	return oh.t
}
func (oh optionFloatHandler) handle(args []string) (optionCommitter, error) {
	if len(args) < 1 {
		args = []string{"0"}
	}
	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return nil, err
	}
	c := optionSimpleCommitter[float64]{f, oh.option}
	return c, nil
}

type optionFloatArrayHandler struct {
	t      optionType
	option *[]float64
}

func (oh optionFloatArrayHandler) getType() optionType {
	return oh.t
}

func (oh optionFloatArrayHandler) handle(args []string) (optionCommitter, error) {
	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return nil, err
	}
	c := optionArrayCommitter[float64]{f, oh.option}
	return c, nil
}

type optionStringHandler struct {
	t      optionType
	option *string
}

func (oh optionStringHandler) getType() optionType {
	return oh.t
}
func (oh optionStringHandler) handle(args []string) (optionCommitter, error) {
	if len(args) < 1 {
		args = []string{""}
	}
	c := optionSimpleCommitter[string]{args[0], oh.option}
	return c, nil
}

type optionStringArrayHandler struct {
	t      optionType
	option *[]string
}

func (oh optionStringArrayHandler) getType() optionType {
	return oh.t
}
func (oh optionStringArrayHandler) handle(args []string) (optionCommitter, error) {
	c := optionArrayCommitter[string]{args[0], oh.option}
	return c, nil
}

type optionCollection struct {
	// <flag name> => <handler for that flag>
	handlers map[string]optionHandler

	// Defer updates until after all options are processed.
	committers []optionCommitter
}

func newOptionCollection() *optionCollection {
	return &optionCollection{
		make(map[string]optionHandler),
		make([]optionCommitter, 0, 10),
	}
}

func (oc *optionCollection) addSimpleHandler(name string, option *bool) {
	oc.handlers[name] = optionSimpleHandler{
		optionNoArg,
		true,
		option,
	}
}

func negatedName(name string) string {
	return "no" + name
}

func (oc *optionCollection) checkNameConflict(name string, negatable bool) error {
	if _, ok := oc.handlers[name]; ok {
		return errors.New("option already exists")
	}
	if negatable {
		if _, ok := oc.handlers[negatedName(name)]; ok {
			return errors.New("option already exists")
		}
	}
	return nil
}

func (oc *optionCollection) addNegatableHandler(name string, option *bool) {
	oc.handlers[name] = optionSimpleHandler{
		optionNoArg,
		true,
		option,
	}
	oc.handlers[negatedName(name)] = optionSimpleHandler{
		optionNoArg,
		false,
		option,
	}
}

func (oc *optionCollection) addCountingHandler(name string, option *int) {
	oc.handlers[name] = optionCountingHandler{
		optionNoArg,
		option,
	}
}

func (oc *optionCollection) addIntHandler(name string, option *int) {
	oc.handlers[name] = optionIntHandler{
		optionRequiredArg,
		option,
	}
}

func (oc *optionCollection) addOptionalIntHandler(name string, option *int) {
	oc.handlers[name] = optionIntHandler{
		optionOptionalArg,
		option,
	}
}

func (oc *optionCollection) addIntArrayHandler(name string, option *[]int) {
	oc.handlers[name] = optionIntArrayHandler{
		optionRequiredArg,
		option,
	}
}

func (oc *optionCollection) addFloatHandler(name string, option *float64) {
	oc.handlers[name] = optionFloatHandler{
		optionRequiredArg,
		option,
	}
}

func (oc *optionCollection) addOptionalFloatHandler(name string, option *float64) {
	oc.handlers[name] = optionFloatHandler{
		optionOptionalArg,
		option,
	}
}

func (oc *optionCollection) addFloatArrayHandler(name string, option *[]float64) {
	oc.handlers[name] = optionFloatArrayHandler{
		optionRequiredArg,
		option,
	}
}

func (oc *optionCollection) addStringHandler(name string, option *string) {
	oc.handlers[name] = optionStringHandler{
		optionRequiredArg,
		option,
	}
}

func (oc *optionCollection) addOptionalStringHandler(name string, option *string) {
	oc.handlers[name] = optionStringHandler{
		optionOptionalArg,
		option,
	}
}

func (oc *optionCollection) addStringArrayHandler(name string, option *[]string) {
	oc.handlers[name] = optionStringArrayHandler{
		optionRequiredArg,
		option,
	}
}

func (oc *optionCollection) commit() {
	for _, e := range oc.committers {
		e.commit()
	}
}
