package request

import (
    "testing"

    . "github.com/smartystreets/goconvey/convey"
)

func PrepFunc(t *testing.T, f func()) func() {
	return func() {
        f()
    }
}

func TestPolicyRequest(t *testing.T) {
    Convey("Sample test", t, PrepFunc(t, func() {
        t.SkipNow()
    }))
}
