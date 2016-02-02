package eol

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"bitbucket.org/heindl/utils"
)

func TestPages(t *testing.T) {

	t.Parallel()

	Convey("should successfully return a list of encyclopedia of life search results", t, func() {

		response, err := Page(PageQuery{
			ID: 1045608,
			Synonyms: true,
			Images: 1,
			Text: 1,
			Details: true,
		})
		Println(utils.JsonOrSpew(response))
		So(err, ShouldBeNil)
		So(response.ScientificName, ShouldEqual, "Apis mellifera Linnaeus 1758")
		So(response.RichnessScore, ShouldEqual, 86.9941)
		So(len(response.Synonyms), ShouldEqual, 7)
		So(len(response.TaxonConcepts), ShouldEqual, 9)

	})
}
