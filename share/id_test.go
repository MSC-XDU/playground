package share

import "testing"

var idTestCases = [...]struct {
	num uint64
	id  string
}{
	{1844674407370955, "8rOwHkBOb"},
	{0, ""},
	{1111111111, "1dc6KX"},
	{1312312321312, "n6rKLh6"},
	{10101, "2CV"},
}

func TestItoID(t *testing.T) {
	for _, v := range idTestCases {
		if id, _ := ItoID(v.num); id != v.id {
			t.Fatalf("bad result %s for %d expected %s", id, v.num, v.id)
		}
	}
}

func TestIDtoI(t *testing.T) {
	for _, v := range idTestCases {
		if num, _ := IDtoI(v.id); num != v.num {
			t.Fatalf("bad result %d for %s expected %d", num, v.id, v.num)
		}
	}
}

func TestNewID(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatal(err)
	}

	e, err := idExist(id)
	if err != nil {
		t.Fatal(err)
	}

	if !e {
		t.Fatal("bloom filter false negative")
	}
}
