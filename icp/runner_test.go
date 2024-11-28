package icp

import "testing"

func TestRunner_SearchByUnitName(t *testing.T) {
	runner := NewRunner(&Options{
		ProxyURL: "http://127.0.0.1:8080",
	})
	entries, err := runner.Search("王继淋")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(entries)
}
