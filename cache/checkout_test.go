package cache

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"os"
	"testing"
)

func TestCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllTestCases() {
		for _, strat := range []strategy.CheckoutStrategy{strategy.CopyStrategy, strategy.LinkStrategy} {
			t.Run(fmt.Sprintf("%s %+v", strat, testCase), func(t *testing.T) {
				testCheckoutIntegration(strat, testCase, t)
			})
		}
	}
}

func testCheckoutIntegration(strat strategy.CheckoutStrategy, statusStart artifact.Status, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(statusStart)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache := LocalCache{Dir: dirs.CacheDir}

	checkoutErr := cache.Checkout(dirs.WorkDir, &art, strat)

	if !statusStart.HasChecksum {
		if checkoutErr != nil {
			return
		}
		t.Fatal("expected Checkout to raise invalid checksum error")
	}

	if !statusStart.ChecksumInCache {
		if checkoutErr != nil {
			return
		}
		t.Fatal("expected Checkout to raise missing checksum in cache error")
	}

	if statusStart.WorkspaceStatus != artifact.Absent {
		if os.IsExist(checkoutErr) {
			return
		}
		t.Fatalf("expected Checkout to raise Exist error, got %#v", checkoutErr)
	} else if checkoutErr != nil {
		t.Fatal(checkoutErr)
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}
	statusWant := artifact.Status{
		HasChecksum:     true,
		ChecksumInCache: true,
		ContentsMatch:   true,
	}
	switch strat {
	case strategy.CopyStrategy:
		statusWant.WorkspaceStatus = artifact.RegularFile
	case strategy.LinkStrategy:
		statusWant.WorkspaceStatus = artifact.Link
	}

	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}
