package containerx

import (
	"testing"
)

func TestSet(t *testing.T) {
	s := NewSet("key1", "key2", "key3")

	ok := s.Has("key1")
	if !ok {
		t.Errorf("excepted has key1, but none")
	}

	ok = s.HasAll("some")
	if ok {
		t.Errorf("excepted does not have some")
	}

	s.Insert("key4", "key5")
	if s.Len() != 5 {
		t.Errorf("insert failed")
	}

	s.Delete("key1", "key2", "key3")
	if s.Len() != 2 {
		t.Errorf("delete failed, got set: %v", s)
	}

	slice := s.Slice()
	if s.Len() != len(slice) {
		t.Errorf("get slice error, slice: %v", slice)
	}
}
