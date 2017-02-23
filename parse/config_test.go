// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"testing"
)

func assertNoErr (t *testing.T, name string, err error) {
	if err != nil {
		t.Errorf("Error in %s: %v", name, err)
	}
}
func assertErr (t *testing.T, name string, err error) {
	if err == nil {
		t.Errorf("No error in %s: %v", name, err)
	}
}
func sliceToSet (slice []string) map[string]int {
	result := make(map[string]int, len(slice))
	for _, elem := range slice {
		result[elem] = 1
	}
	return result
}
func setToSlice (set map[string]int) []string {
	result := make([]string, len(set))
	i := 0
	for k, _ := range set {
		result[i] = k
		i++
	}
	return result
}
// a - b
func setDifference (a, b map[string]int) map[string]int {
	result := make(map[string]int, 0)
	for k, _ := range a {
		_, contained := b[k]
		if !contained {
			result[k] = 1
		}
	}
	return result
}

func assertKeyList (t *testing.T, name string, expected []string, actual []string) {
	expSet := sliceToSet(expected)
	actSet := sliceToSet(actual)

	// Check for non-unique elements first
	if len(expected) != len(expSet) {
		t.Errorf("Error in %s: non-unique elements in expected set", name)
	} else if len(actual) != len(actSet) {
		t.Errorf("Error in %s: non-unique elements in actual set", name)
	} else {
		expOnly := setToSlice(setDifference(expSet, actSet))
		actOnly := setToSlice(setDifference(actSet, expSet))
		if 0 != len(expOnly) || 0 != len(actOnly) {
			t.Errorf("Error in %s: disparate elements. %v only in expected set. %v only in actual set.",
				name, expOnly, actOnly)
		}
	}
}

func TestConfigKeySet (t *testing.T) {
	configString := "abc: {def: {ghi: 1, jkl: 2}, mno: [1, 2], pqr: 3}"
	configTree, err := New("key-set-test").Parse(configString)
	assertNoErr(t, "creating config tree", err)
	config := configTree.GetConfig()

	keySet, err := config.GetKeySet("")
	assertNoErr(t, "extracting root keys", err)
	assertKeyList(t, "root keys", []string{"abc"}, keySet)
	
	keySet, err = config.GetKeySet("abc")
	assertNoErr(t, "extracting abc keys", err)
	assertKeyList(t, "abc keys", []string{"def", "mno", "pqr"}, keySet)

	keySet, err = config.GetKeySet("abc.def")
	assertNoErr(t, "extracting abc.def keys", err)
	assertKeyList(t, "abc.def keys", []string{"ghi", "jkl"}, keySet)

	keySet, err = config.GetKeySet("abc.mno")
	assertErr(t, "Extracting abc.mno keys", err)

	keySet, err = config.GetKeySet("abc.pqr")
	assertErr(t, "Extracting abc.pqr keys", err)

	keySet, err = config.GetKeySet("xyz")
	assertErr(t, "Extracting xyz keys", err)
}
