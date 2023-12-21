package utils

import (
	"flag"
	"testing"

	//nolint:revive //dot imports acceptable in test file
	. "github.com/smartystreets/goconvey/convey"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func TestRandomDatabase(t *testing.T) {
	if *componentFlag {
		t.Skip()
	} else {
		Convey("When the RandomDatabase function is called", t, func() {
			s := RandomDatabase()
			Convey("Then it returns a valid random string", func() {
				So(s, ShouldNotBeNil)
				So(len(s), ShouldEqual, 15)
				for _, c := range s {
					So(c, ShouldBeGreaterThanOrEqualTo, 'a')
					So(c, ShouldBeLessThanOrEqualTo, 'z')
				}
			})
		})
		Convey("When the RandomDatabase function is called multiple times", t, func() {
			seen := map[string]bool{}
			Convey("Then it returns different values", func() {
				for i := 0; i < 1000; i++ {
					s := RandomDatabase()
					So(seen[s], ShouldBeFalse)
					seen[s] = true
				}
			})
		})
	}
}
