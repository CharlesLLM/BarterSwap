package main

import "testing"

func TestMainWithoutDatabaseURL(testContext *testing.T) {
	testContext.Setenv("DATABASE_URL", "")
	main()
}

func TestMainWithInvalidDatabaseURL(testContext *testing.T) {
	testContext.Setenv("DATABASE_URL", "://")
	main()
}
