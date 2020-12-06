// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
    "flag"
    "fmt"
    "testing"
)

var debug = flag.Bool("debug", false, "show the errors produced by the main tests")

type parseTest struct {
    name   string
    input  string
    ok     bool
    result string // what the user would see in an error message.
}

const (
    noError  = true
    hasError = false
)

var parseTests = []parseTest{
    {"empty", "", noError,
        ``},
    {"comment", "/*\n\n\n*/", noError,
        ``},
    {"equals", `akka.on = true`, noError,
        `akka = (on = (true))`},
    {"equals unquoted", `akka.duration = 1 second`, noError,
        `akka = (duration = (1 second))`},
    {"equals number", `akka.count = 10`, noError,
        `akka = (count = (10))`},
    {"equals twice",
        `akka.on = true,
         akka.count = 10,`,
        noError,
        `akka = (count = (10)on = (true))`},
    {"object",
        `akka {
            count = 10,
            on = true
        }`,
        noError,
        `akka = (count = (10)on = (true))`},
    {"array",
        `[10, true]`,
        noError,
        `10true`},
    {"imbed object",
        `akka {
            count = 10,
            embeded {
                on = true
            },
        },
        akka.duration = 1 second
        akka.count = 7,
        `,
        noError,
        `akka = (count = (7)duration = (1 second)embeded = (on = (true)))`},
    {"object array",
        `akka {
            count = 10,
            arr = [true, false]
        }`,
        noError,
        `akka = (arr = (truefalse)count = (10))`},
}

func testParse(doCopy bool, t *testing.T) {
    textFormat = "%q"
    defer func() { textFormat = "%s" }()
    for _, test := range parseTests {
        tmpl, err := New(test.name).Parse(test.input)
        switch {
            case err == nil && !test.ok:
            t.Errorf("%q: expected error; got none", test.name)
            continue
            case err != nil && test.ok:
            t.Errorf("%q: unexpected error: %v", test.name, err)
            continue
            case err != nil && !test.ok:
            // expected error, got one
            if *debug {
                fmt.Printf("%s: %s\n\t%s\n", test.name, test.input, err)
            }
            continue
        }
        var result string
        if doCopy {
            result = tmpl.Root.Copy().String()
        } else {
            result = tmpl.Root.String()
        }
        if result != test.result {
            t.Errorf("%s=(%q): got\n\t%v\nexpected\n\t%v", test.name, test.input, result, test.result)
        }
    }
}

func TestParse(t *testing.T) {
    testParse(false, t)
    testParse(true, t)
}
