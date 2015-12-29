package db_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	. "github.com/starkandwayne/shield/db"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/starkandwayne/shield/timestamp"
	"time"
)

var _ = Describe("Archive Management", func() {
	TARGET_UUID := uuid.NewRandom()
	STORE_UUID := uuid.NewRandom()
	ARCHIVE_UUID := uuid.NewRandom()

	var db *DB

	shouldHaveArchiveStatus := func(id uuid.UUID, status string) {
		a, err := db.GetAnnotatedArchive(id)
		Ω(err).ShouldNot(HaveOccurred(), "Retrieving the archive should not have failed")
		Expect(a).ShouldNot(BeNil(), "An archive should be returned")
		Expect(a.Status).Should(Equal(status), "the archive should have correct status")
	}

	shouldHavePurgeReason := func(id uuid.UUID, reason string) {
		a, err := db.GetAnnotatedArchive(id)
		Ω(err).ShouldNot(HaveOccurred(), "Retrieving the archive should not have failed")
		Expect(a).ShouldNot(BeNil(), "An archive should be returned")
		Expect(a.PurgeReason).Should(Equal(reason), "the archive should have correct purge_reason")
	}

	BeforeEach(func() {
		var err error
		db, err = Database(
			// need a target
			`INSERT INTO targets (uuid, plugin, endpoint, agent) VALUES ("`+TARGET_UUID.String()+`", "target_plugin", "target_endpoint", "127.0.0.1:5444")`,
			// need a store
			`INSERT INTO stores (uuid, plugin, endpoint) VALUES ("`+STORE_UUID.String()+`", "store_plugin", "store_endpoint")`,
			// need an ARCHIVE
			`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status, notes, purge_reason)
				VALUES ("`+ARCHIVE_UUID.String()+`", "`+TARGET_UUID.String()+`",
				        "`+STORE_UUID.String()+`", "key", 0, 0, "valid", "my_notes", "")`,
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(db).ShouldNot(BeNil())

		shouldHaveArchiveStatus(ARCHIVE_UUID, "valid")
		shouldHavePurgeReason(ARCHIVE_UUID, "")
	})

	It("Archives can be invalidated", func() {
		err := db.InvalidateArchive(ARCHIVE_UUID)
		Expect(err).ShouldNot(HaveOccurred())

		shouldHaveArchiveStatus(ARCHIVE_UUID, "invalid")
	})

	It("Archives can be expired", func() {
		err := db.ExpireArchive(ARCHIVE_UUID)
		Expect(err).ShouldNot(HaveOccurred())

		shouldHaveArchiveStatus(ARCHIVE_UUID, "expired")
	})

	Describe("Purging archives", func() {
		It("with an archive whose status is 'valid'", func() {
			err := db.PurgeArchive(ARCHIVE_UUID)
			Expect(err).Should(HaveOccurred(), "should generate an error")

			shouldHaveArchiveStatus(ARCHIVE_UUID, "valid")
		})

		It("with an archive whose status is 'invalid'", func() {
			err := db.InvalidateArchive(ARCHIVE_UUID)
			Expect(err).ShouldNot(HaveOccurred(), "Invalidating archive should not have generated an error")

			err = db.PurgeArchive(ARCHIVE_UUID)
			Expect(err).ShouldNot(HaveOccurred(), "Purging archive should not have generated an error")

			shouldHaveArchiveStatus(ARCHIVE_UUID, "purged")
			shouldHavePurgeReason(ARCHIVE_UUID, "invalid")
		})

		It("If they are 'expired'", func() {
			err := db.ExpireArchive(ARCHIVE_UUID)
			Expect(err).ShouldNot(HaveOccurred(), "should not generate an error")

			err = db.PurgeArchive(ARCHIVE_UUID)
			Expect(err).ShouldNot(HaveOccurred(), "Purging archive should not have generated an error")

			shouldHaveArchiveStatus(ARCHIVE_UUID, "purged")
			shouldHavePurgeReason(ARCHIVE_UUID, "expired")
		})
	})

	Describe("Archive Retrieval", func() {
		TARGET2_UUID := uuid.NewRandom()
		STORE2_UUID := uuid.NewRandom()
		ARCHIVE_PURGED := uuid.NewRandom()
		ARCHIVE_INVALID := uuid.NewRandom()
		ARCHIVE_EXPIRED := uuid.NewRandom()
		ARCHIVE_TARGET2 := uuid.NewRandom()
		ARCHIVE_STORE2 := uuid.NewRandom()
		BeforeEach(func() {
			var err error
			db.Exec(`INSERT INTO targets (uuid, plugin, endpoint, agent) VALUES("` + TARGET2_UUID.String() + `","target_plugin2", "target_endpoint2", "127.0.0.1:5444")`)
			err = db.Exec(`INSERT INTO stores (uuid, plugin, endpoint ) VALUES("` + STORE2_UUID.String() + `","store_plugin2", "store_endpoint2")`)
			Expect(err).ShouldNot(HaveOccurred())
			err = db.Exec(`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status) VALUES("` +
				ARCHIVE_PURGED.String() + `","` + TARGET_UUID.String() + `", "` + STORE_UUID.String() +
				`", "key", 10, 10, "purged")`)
			Expect(err).ShouldNot(HaveOccurred())
			err = db.Exec(`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status) VALUES("` +
				ARCHIVE_INVALID.String() + `","` + TARGET_UUID.String() + `", "` + STORE_UUID.String() +
				`", "key", 10, 10, "invalid")`)
			Expect(err).ShouldNot(HaveOccurred())
			err = db.Exec(`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status) VALUES("` +
				ARCHIVE_EXPIRED.String() + `","` + TARGET_UUID.String() + `", "` + STORE_UUID.String() +
				`", "key", 20, 20, "expired")`)
			Expect(err).ShouldNot(HaveOccurred())
			err = db.Exec(`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status) VALUES("` +
				ARCHIVE_TARGET2.String() + `","` + TARGET2_UUID.String() + `", "` + STORE_UUID.String() +
				`", "key", 20, 20, "valid")`)
			Expect(err).ShouldNot(HaveOccurred())
			err = db.Exec(`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status) VALUES("` +
				ARCHIVE_STORE2.String() + `","` + TARGET_UUID.String() + `", "` + STORE2_UUID.String() +
				`", "key", 20, 20, "invalid")`)
			Expect(err).ShouldNot(HaveOccurred())
		})
		Describe("Of Individual archives", func() {
			It("Should return the requested archive", func() {
				a, err := db.GetAnnotatedArchive(ARCHIVE_UUID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(a).ShouldNot(BeNil())
				Expect(a).Should(BeEquivalentTo(&AnnotatedArchive{
					UUID:           ARCHIVE_UUID.String(),
					StoreKey:       "key",
					TakenAt:        NewTimestamp(time.Unix(0, 0).UTC()),
					ExpiresAt:      NewTimestamp(time.Unix(0, 0).UTC()),
					Notes:          "my_notes",
					Status:         "valid",
					PurgeReason:    "",
					TargetUUID:     TARGET_UUID.String(),
					TargetPlugin:   "target_plugin",
					TargetEndpoint: "target_endpoint",
					StoreUUID:      STORE_UUID.String(),
					StoreEndpoint:  "store_endpoint",
					StorePlugin:    "store_plugin",
				}))
			})
			It("Should return error nil/nil if no records exist", func() {
				a, err := db.GetAnnotatedArchive(uuid.NewRandom())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(a).Should(BeNil())
			})
		})

		Describe("Of multiple archives", func() {
			It("When filtering by Status", func() {
				filter := ArchiveFilter{
					WithStatus: []string{"purged"},
				}
				archives, err := db.GetAllAnnotatedArchives(&filter)
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				Expect(len(archives)).Should(Equal(1), "returns the correct number of archives")
				Expect(archives[0].UUID).Should(Equal(ARCHIVE_PURGED.String()), "returns the correct archive")
			})
			It("When filtering by Target", func() {
				filter := ArchiveFilter{
					ForTarget: TARGET2_UUID.String(),
				}
				archives, err := db.GetAllAnnotatedArchives(&filter)
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				Expect(len(archives)).Should(Equal(1), "returns the correct number of archives")
				Expect(archives[0].UUID).Should(Equal(ARCHIVE_TARGET2.String()), "returns the correct archive")
			})
			It("When filtering by Store", func() {
				filter := ArchiveFilter{
					ForStore: STORE2_UUID.String(),
				}
				archives, err := db.GetAllAnnotatedArchives(&filter)
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				Expect(len(archives)).Should(Equal(1), "returns the correct number of archives")
				Expect(archives[0].UUID).Should(Equal(ARCHIVE_STORE2.String()), "returns the correct archive")
			})
			It("When filtering with After", func() {
				t := time.Unix(15, 0).UTC()
				filter := ArchiveFilter{
					After: &t,
				}
				archives, err := db.GetAllAnnotatedArchives(&filter)
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				Expect(len(archives)).Should(Equal(3), "returns the correct number of archives")

				var uuids []string
				for _, e := range archives {
					uuids = append(uuids, e.UUID)
				}
				Expect(uuids).Should(ConsistOf([]string{ARCHIVE_EXPIRED.String(), ARCHIVE_TARGET2.String(), ARCHIVE_STORE2.String()}),
					"returns the correct archives")
			})
			It("When filtering with Before", func() {
				t := time.Unix(5, 0).UTC()
				filter := ArchiveFilter{
					Before: &t,
				}
				archives, err := db.GetAllAnnotatedArchives(&filter)
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				Expect(len(archives)).Should(Equal(1), "returns the correct number of archives")
				Expect(archives[0].UUID).Should(Equal(ARCHIVE_UUID.String()), "returns the correct archive in the first result")
			})
			It("When filtering via a combination of values", func() {
				t := time.Unix(15, 0).UTC()
				filter := ArchiveFilter{
					WithStatus: []string{"invalid"},
					After:      &t,
				}
				archives, err := db.GetAllAnnotatedArchives(&filter)
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				Expect(len(archives)).Should(Equal(1), "returns the correct number of archives")
				Expect(archives[0].UUID).Should(Equal(ARCHIVE_STORE2.String()), "returns the correct archive")

			})
			It("When filtering by WithoutStatus", func() {
				filter := ArchiveFilter{
					WithOutStatus: []string{"valid"},
				}
				archives, err := db.GetAllAnnotatedArchives(&filter)
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				Expect(len(archives)).Should(Equal(4), "returns the correct number of archives")

				var uuids []string
				for _, e := range archives {
					uuids = append(uuids, e.UUID)
				}
				Expect(uuids).Should(ConsistOf([]string{ARCHIVE_EXPIRED.String(), ARCHIVE_PURGED.String(), ARCHIVE_INVALID.String(), ARCHIVE_STORE2.String()}),
					"returns the correct archives")
			})
		})

		Describe("GetArchivesNeedingPurge", func() {
			var expectedArchiveCount int

			BeforeEach(func() {
				all, err := db.GetAllAnnotatedArchives(&ArchiveFilter{})
				Expect(err).ShouldNot(HaveOccurred())
				valid, err := db.GetAllAnnotatedArchives(&ArchiveFilter{WithStatus: []string{"valid"}})
				Expect(err).ShouldNot(HaveOccurred())
				purged, err := db.GetAllAnnotatedArchives(&ArchiveFilter{WithStatus: []string{"purged"}})
				Expect(err).ShouldNot(HaveOccurred())
				expectedArchiveCount = len(all) - len(valid) - len(purged)
			})

			It("returns all jobs whose status is not 'purged' or 'valid'", func() {
				archives, err := db.GetArchivesNeedingPurge()
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				for _, archive := range archives {
					Expect(archive.Status).ShouldNot(Equal("purged"), "does not return 'purged' archives")
					Expect(archive.Status).ShouldNot(Equal("valid"), "does not return 'valid' archives")
				}
				Expect(len(archives)).Should(Equal(expectedArchiveCount), "returns the correct number of archives")
			})
		})

		Describe("GetExpiredArchives()", func() {
			UNEXPIRED_ARCHIVE := uuid.NewRandom()
			UNEXPIRED_ARCHIVE2 := uuid.NewRandom()
			var expectedArchiveCount int
			BeforeEach(func() {
				// insert archive expiring in a day
				err := db.Exec(`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status) VALUES("`+
					UNEXPIRED_ARCHIVE.String()+`","`+TARGET_UUID.String()+`", "`+STORE2_UUID.String()+
					`", "key", 20, $1, "valid")`, time.Now().Unix())

				Expect(err).ShouldNot(HaveOccurred())
				err = db.Exec(`INSERT INTO archives (uuid, target_uuid, store_uuid, store_key, taken_at, expires_at, status) VALUES("` +
					UNEXPIRED_ARCHIVE2.String() + `","` + TARGET_UUID.String() + `", "` + STORE2_UUID.String() +
					`", "key", 20, 20, "invalid")`)
				Expect(err).ShouldNot(HaveOccurred())
				// get expeted count of expired archives
				all, err := db.GetAllAnnotatedArchives(&ArchiveFilter{})
				Expect(err).ShouldNot(HaveOccurred())

				expectedArchiveCount = len(all) - 6 // There are 2 directly above, and at last checked 4 non-valid but expirable archives in the BeforeEach calls
			})
			It("returns all jobs who have expired", func() {
				archives, err := db.GetExpiredArchives()
				Expect(err).ShouldNot(HaveOccurred(), "does not error")
				for _, archive := range archives {
					Expect(archive.UUID).ShouldNot(Equal(UNEXPIRED_ARCHIVE), "does not return the unexpired archive")
					Expect(archive.UUID).ShouldNot(Equal(UNEXPIRED_ARCHIVE2), "does not return the expired but not 'valid' archive")
					Expect(archive.ExpiresAt.Time()).Should(BeTemporally("<", time.Now()), "does not return archives that have not expired yet")
					Expect(archive.Status).Should(Equal("valid"), "does not return archives that aren't valid")
				}
				Expect(len(archives)).Should(Equal(expectedArchiveCount), "returns the correct number of archives")
			})

		})
	})
})