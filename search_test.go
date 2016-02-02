package eol

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSearch(t *testing.T) {

	t.Parallel()

	Convey("should successfully return a list of encyclopedia of life search results", t, func() {

		response, err := Search(SearchQuery{Query: "Ursus"})
		So(err, ShouldBeNil)
		So(len(response), ShouldEqual, 157)

		So(response[0].ID, ShouldEqual, 14349)
		So(response[0].Title, ShouldEqual, "Ursus")
		So(response[0].Link, ShouldEqual, "http://eol.org/14349?action=overview&controller=taxa")
		So(response[0].Content, ShouldEqual, "Ursus Linnaeus, 1758; Ursus; Ursus Arctos Bruinosus; Ursus Arctos Ssp.")

	})
}
