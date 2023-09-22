package algo

import (
	"testing"
)

// - single test
func TestRRW_Next_Single(t *testing.T) {
	w := &RRW{}
	w.Add("server1", 5)

	results := make(map[string]int)

	for i := 0; i < 100; i++ {
		s := w.Next().(string)
		results[s]++
	}

	if results["server1"] != 100 {
		t.Error("the algorithm is wrong", results)
	}
}

func TestRRW_Next(t *testing.T) {
	w := &RRW{}

	//- test v1
	w.Add("server1", 5)
	w.Add("server2", 2)
	w.Add("server3", 3)

	results := make(map[string]int)

	for i := 0; i < 100; i++ {
		s := w.Next().(string)
		t.Logf("s: %+v", s)
		results[s]++
	}

	t.Logf("w.Next() results: %+v", results)

	if results["server1"] != 50 || results["server2"] != 20 || results["server3"] != 30 {
		t.Error("the algorithm is wrong", results)
	}
	w.Reset()

	//- test v1.1
	results = make(map[string]int)

	for i := 0; i < 100; i++ {
		s := w.Next().(string)
		results[s]++
	}

	if results["server1"] != 50 || results["server2"] != 20 || results["server3"] != 30 {
		t.Error("the algorithm is wrong", results)
	}

	w.RemoveAll()

	//- test v2
	w.Add("server1", 7)
	w.Add("server2", 9)
	w.Add("server3", 13)

	results = make(map[string]int)

	for i := 0; i < 29000; i++ {
		s := w.Next().(string)
		results[s]++
	}

	if results["server1"] != 7000 || results["server2"] != 9000 || results["server3"] != 13000 {
		t.Error("the algorithm is wrong", results)
	}
}
