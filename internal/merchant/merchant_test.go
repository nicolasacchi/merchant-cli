package merchant

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type fakeIt struct {
	items []int
	i     int
	info  *iterator.PageInfo
}

func (f *fakeIt) Next() (int, error) {
	if f.i >= len(f.items) {
		return 0, iterator.Done
	}
	f.i++
	return f.items[f.i-1], nil
}
func (f *fakeIt) PageInfo() *iterator.PageInfo { return f.info }

func TestDrainStopsAtLimitAndReturnsToken(t *testing.T) {
	info := &iterator.PageInfo{Token: "next"}
	got, tok, err := drain[int](&fakeIt{items: []int{1, 2, 3}, info: info}, 2)
	if err != nil || len(got) != 2 || tok != "next" {
		t.Fatalf("got=%v tok=%q err=%v", got, tok, err)
	}
	got, tok, err = drain[int](&fakeIt{items: []int{1}, info: info}, 5)
	if err != nil || len(got) != 1 || tok != "" {
		t.Fatalf("got=%v tok=%q err=%v", got, tok, err)
	}
}

func TestCredType(t *testing.T) {
	dir := t.TempDir()
	user := filepath.Join(dir, "u.json")
	sa := filepath.Join(dir, "s.json")
	_ = os.WriteFile(user, []byte(`{"type":"authorized_user"}`), 0o600)
	_ = os.WriteFile(sa, []byte(`{"type":"service_account"}`), 0o600)
	if credType(user) != option.AuthorizedUser || credType(sa) != option.ServiceAccount {
		t.Fatal("credential type detection wrong")
	}
}
