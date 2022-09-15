package finder

import "testing"

func TestGetDirs(t *testing.T) {

	dirs, err := getDirs("/mnt/md0/columbus")
	if err != nil {
		t.Fatalf("Fail: %s\n", err)
	}
	t.Logf("%#v\n", dirs)
	t.Logf("len: %d\n", len(dirs))

}

func TestFind(t *testing.T) {

	r, err := Find("elmasy.com", "/mnt/md0/columbus")
	if err != nil {
		t.Fatalf("Fail: %s\n", err)
	}

	t.Logf("%#v\n", r)
}
