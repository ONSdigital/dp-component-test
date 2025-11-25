package validator

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	invalidString = "abcdefgh"
)

func TestValidateTimestamp(t *testing.T) {
	Convey("Given a valid timestamp string", t, func() {
		ts := "2025-10-30T14:43:39Z"

		Convey("When it is validated", func() {
			valid := ValidateTimestamp(ts)

			Convey("Then it is valid", func() {
				So(valid, ShouldBeTrue)
			})
		})
	})

	Convey("Given an invalid timestamp string", t, func() {
		ts := invalidString

		Convey("When it is validated", func() {
			valid := ValidateTimestamp(ts)

			Convey("Then it is invalid", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})
}

func TestValidateRecentTimestamp(t *testing.T) {
	Convey("Given a valid recent timestamp string", t, func() {
		ts := time.Now().Format(time.RFC3339)

		Convey("When it is validated", func() {
			valid := ValidateRecentTimestamp(ts)

			Convey("Then it is valid", func() {
				So(valid, ShouldBeTrue)
			})
		})
	})

	Convey("Given an invalid timestamp string", t, func() {
		ts := invalidString

		Convey("When it is validated", func() {
			valid := ValidateRecentTimestamp(ts)

			Convey("Then it is invalid", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})

	Convey("Given a non-recent timestamp string", t, func() {
		ts := time.Now().Add(-11 * time.Second).Format(time.RFC3339)

		Convey("When it is validated", func() {
			valid := ValidateRecentTimestamp(ts)

			Convey("Then it is invalid", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})
}

func TestValidateUUID(t *testing.T) {
	Convey("Given a valid uuid", t, func() {
		id := "2bcbce7d-d0a6-427e-9334-dda37f62a81b"

		Convey("When it is validated", func() {
			valid := ValidateUUID(id)

			Convey("Then it is valid", func() {
				So(valid, ShouldBeTrue)
			})
		})
	})

	Convey("Given an invalid uuid", t, func() {
		id := invalidString

		Convey("When it is validated", func() {
			valid := ValidateUUID(id)

			Convey("Then it is valid", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})
}

func TestValidateURL(t *testing.T) {
	Convey("Given a valid url", t, func() {
		url := "http://localhost:8080/economy"

		Convey("When it is validated", func() {
			valid := ValidateURL(url)

			Convey("Then it is valid", func() {
				So(valid, ShouldBeTrue)
			})
		})
	})

	Convey("Given an invalid uuid", t, func() {
		url := invalidString

		Convey("When it is validated", func() {
			valid := ValidateURL(url)

			Convey("Then it is invalid", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})
}

func TestValidateURIPath(t *testing.T) {
	Convey("Given a valid URI path", t, func() {
		uriPath := "/economy/data"

		Convey("When it is validated", func() {
			valid := ValidateURIPath(uriPath)

			Convey("Then it is valid", func() {
				So(valid, ShouldBeTrue)
			})
		})
	})

	Convey("Given an invalid URI path", t, func() {
		uriPath := ""

		Convey("When it is validated", func() {
			valid := ValidateURIPath(uriPath)

			Convey("Then it is invalid", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})
}
