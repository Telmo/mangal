package database

import (
	"database/sql"
	"testing"

	"github.com/metafates/mangal/db"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDB(t *testing.T) {
	Convey("Given a test database setup", t, func() {
		db, err := db.GetDB()
		So(err, ShouldBeNil)
		defer db.Close()

		Convey("When pinging the database", func() {
			err := db.Ping()

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When getting another database connection", func() {
			db2, err := db.GetDB()
			defer db2.Close()

			Convey("Then it should return a valid connection", func() {
				So(err, ShouldBeNil)
				So(db2, ShouldNotBeNil)
			})
		})
	})
}

func TestMigrate(t *testing.T) {
	Convey("Given a test database setup", t, func() {
		database, err := db.GetDB()
		So(err, ShouldBeNil)
		defer database.Close()

		Convey("When running migrations", func() {
			err := db.Migrate(database, true)

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)

				Convey("And the series table should exist", func() {
					var tableExists bool
					err := database.QueryRow(`
						SELECT EXISTS (
							SELECT FROM information_schema.tables 
							WHERE table_schema = 'public' 
							AND table_name = 'series'
						)
					`).Scan(&tableExists)

					So(err, ShouldBeNil)
					So(tableExists, ShouldBeTrue)
				})
			})
		})
	})
}
