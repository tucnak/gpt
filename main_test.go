package main

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	var (
		all = func(msg ...message) []message { return msg }
		msg = func(role, s string) message {
			return message{Role: role, Content: s}
		}
	)
	var tests = map[string][]message{
		"a\n>>>>>>\nb":      all(msg(system, "a"), msg(user, "b")),
		">>>>>>\nb":         all(msg(user, "b")),
		"\n>>>>>\nb":        all(msg(user, "b")),
		"\t\n>>>\nb":        all(msg(user, "b")),
		"\t>>\nb":           all(msg(user, ">>\nb")),
		"\n\t>>>>>>\nb":     all(msg(user, "b")),
		"    >>>>>>\nb":     all(msg(user, "b")),
		"    >\nb":          all(msg(user, ">\nb")),
		"b":                 all(msg(user, "b")),
		"b\n":               all(msg(user, "b")),
		"\t>>>>\na\n\t<\nb": all(msg(user, "a\n\t<\nb")),
	}

	for s, want := range tests {
		if got := parse(s); !reflect.DeepEqual(got, want) {
			t.Errorf("parse(%q) = %v, want %v", s, got, want)
		}
	}
}
