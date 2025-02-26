package db

import "testing"

var db *postgres

func connect() error {

}

func disconnect() {
	db.db.Close()
}

func TestXxx(t *testing.T) {

}
