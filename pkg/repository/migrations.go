package repository

import "context"

type Migrator interface {
	// Perform necessary migrations to allow a datastore to be used by
	// the app.
	Migrate(ctx context.Context) error
}

// DummyPopulator prepares any type (though it should be a repository)
// to populate its data storage instance with demonstration values for
// testing purposes.
type DummyPopulator interface {
	// Populate a concrete data store with dummy values for testing
	//
	// Make sure to run IsPrepopulated first to prevent multiple
	// populations
	PopulateDummyValues(ctx context.Context) error

	// Check if dummy values have been set.
	//
	// This should check for the presence of the values that will be
	// set in PopulateDummyValues, and return true if they are already
	// present in the db
	IsPrepopulated(ctx context.Context) bool

	// Remove the dummy values from the datastore.
	//
	// This should do the reverse of PopulateDummyValues; and is
	// only useful if you don't have a testing datastore for whatever
	// reason.
	CleanDummyValues(ctx context.Context) error
}
