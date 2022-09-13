package fetcher

import (
	"context"
	"testing"
	"time"
)

func TestGenTempPath(t *testing.T) {

	t.Logf("%s\n", genTempPath("/mnt/md0/columbus/argon2022/list"))

}

func TestUnique(t *testing.T) {

	start := time.Now()

	err := unique(context.TODO(), "/mnt/md0/columbus/trustasia2023/list.test")
	if err != nil {
		t.Fatalf("Failed: %s\n", err)
	}
	t.Logf("unique success in %s!\n", time.Since(start))
}
