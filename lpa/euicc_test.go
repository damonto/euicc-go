package lpa

import "testing"

func TestOptionsNormalizeRejectsNil(t *testing.T) {
	var opts *Options
	if err := opts.Normalize(); err == nil {
		t.Error("Options.Normalize() error = nil for nil options")
	}
	if _, err := New(nil); err == nil {
		t.Error("New(nil) error = nil")
	}
}

func TestOptionsNormalizeRejectsEmptyPrefixedVersion(t *testing.T) {
	opts := &Options{AdminProtocolVersion: "v"}
	if err := opts.Normalize(); err == nil {
		t.Error("Options.Normalize() error = nil for version v")
	}
}
