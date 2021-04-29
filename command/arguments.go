package command

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	varargReg *regexp.Regexp
)

func init() {
	varargReg = regexp.MustCompile(`^\$\d+$`)
}

// Arguments is a template for a command, with $n replacement
// a, err := NewArguments("pim", "$1", "poum")
// v, err := a.Values("pam") // "pam" is $1, v is []string{"pim", "pam", "poum"}
type Arguments []Valueable

func (a Arguments) Values(args ...string) ([]string, error) {
	zargs := make([]string, len(a))
	for i := 0; i < len(a); i++ {
		v, err := a[i].Value(args...)
		if err != nil {
			return nil, err
		}
		zargs[i] = v
	}
	return zargs, nil
}

type Valueable interface {
	Value(args ...string) (string, error)
}

type vararg int

func (v vararg) Value(args ...string) (string, error) {
	if int(v) > len(args) {
		return "", fmt.Errorf("Bad $%d from %p", v, args)
	}
	return args[v-1], nil
}

type fixarg string

func (f fixarg) Value(args ...string) (string, error) {
	return string(f), nil
}

func NewArguments(args ...string) (Arguments, error) {
	zargs := make(Arguments, 0)
	for _, arg := range args {
		if varargReg.MatchString(arg) {
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "$"))
			if err != nil {
				return nil, err
			}
			zargs = append(zargs, vararg(n))
		} else {
			zargs = append(zargs, fixarg(arg))
		}
	}
	return zargs, nil
}
