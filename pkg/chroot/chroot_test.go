package chroot_test

import (
	"os"
	"testing"

	. "github.com/nobodyawesomer/redkit/pkg/chroot"
)

func TestNew(t *testing.T) {
	New("chroot")
	if !testDirExists("chroot") {
		t.FailNow()
	}
}

func testDirExists(testDir string) bool {
	testfile, err := os.Open(testDir)
	if err != nil {
		return false
	}
	testfileStat := unwrap(testfile.Stat())

	return testfileStat.IsDir()
}

func unwrap[R any](returnable R, err error) R {
	if err != nil {
		panic(err)
	}
	return returnable
}

// hmm.... catch funcs? generic on error and tests instanceof specific error
// func calm[R any](returnable R, recoverFuncs ...func(...any)) R {
// 	if r := recover(); r != nil {
// 		for _, fun := range recoverFuncs {
// 			fun(r)
// 		}
// 	}
// 	return returnable
// }

// func logwrap[R any](returnable R, err error) R {
// 	return calm(unwrap(returnable, err), log.Panic)
// }
