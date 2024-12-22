package cmd

import (
	"testing"

	"github.com/metafates/mangal/db"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDBInit(t *testing.T) {
	Convey("Given a database connection", t, func() {
		database, err := db.GetDB()
		So(err, ShouldBeNil)
		defer database.Close()

		Convey("When initializing the database", func() {
			err := db.Migrate(database, true)

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDBSearch(t *testing.T) {
	Convey("Given a database with test data", t, func() {
		database, err := db.GetDB()
		So(err, ShouldBeNil)
		defer database.Close()

		err = db.Migrate(database, true)
		So(err, ShouldBeNil)

		Convey("When searching for a manga", func() {
			manga, err := db.SearchMangaByName(database, "Test Manga")

			Convey("Then it should not error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
