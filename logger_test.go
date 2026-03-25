package main

import (
	"errors"
	"testing"
)

func TestNewLogger(t *testing.T) {
	l := NewLogger()
	if l == nil {
		t.Fatal("NewLogger() returned nil")
	}
	if l.fields == nil {
		t.Error("NewLogger() fields should not be nil")
	}
	if len(l.fields) != 0 {
		t.Errorf("NewLogger() fields length = %v, want 0", len(l.fields))
	}
	if l.err != nil {
		t.Error("NewLogger() err should be nil")
	}
}

func TestLogger_WithField(t *testing.T) {
	l := NewLogger()
	l2 := l.WithField("key", "value")
	
	if l2.fields["key"] != "value" {
		t.Errorf("WithField() key = %v, want 'value'", l2.fields["key"])
	}
	if len(l.fields) != 0 {
		t.Error("WithField() should not modify original logger")
	}
}

func TestLogger_WithFields(t *testing.T) {
	l := NewLogger()
	fields := Fields{
		"key1": "value1",
		"key2": 42,
	}
	l2 := l.WithFields(fields)
	
	if l2.fields["key1"] != "value1" {
		t.Errorf("WithFields() key1 = %v, want 'value1'", l2.fields["key1"])
	}
	if l2.fields["key2"] != 42 {
		t.Errorf("WithFields() key2 = %v, want 42", l2.fields["key2"])
	}
	if len(l.fields) != 0 {
		t.Error("WithFields() should not modify original logger")
	}
}

func TestLogger_WithFields_Merge(t *testing.T) {
	l := NewLogger().WithField("existing", "old")
	l2 := l.WithFields(Fields{
		"new": "value",
	})
	
	if l2.fields["existing"] != "old" {
		t.Error("WithFields() should preserve existing fields")
	}
	if l2.fields["new"] != "value" {
		t.Error("WithFields() should add new fields")
	}
}

func TestLogger_WithError(t *testing.T) {
	l := NewLogger()
	err := errors.New("test error")
	l2 := l.WithError(err)
	
	if l2.err != err {
		t.Errorf("WithError() err = %v, want %v", l2.err, err)
	}
	if l.err != nil {
		t.Error("WithError() should not modify original logger")
	}
}

func TestLogger_WithError_PreservesFields(t *testing.T) {
	l := NewLogger().WithField("key", "value")
	err := errors.New("test error")
	l2 := l.WithError(err)
	
	if l2.fields["key"] != "value" {
		t.Error("WithError() should preserve fields")
	}
	if l2.err != err {
		t.Errorf("WithError() err = %v, want %v", l2.err, err)
	}
}
