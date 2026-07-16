package postgres

import "testing"

type sqlStateErrorStub string

func (err sqlStateErrorStub) Error() string { return string(err) }
func (sqlStateErrorStub) SQLState() string  { return uniqueViolationSQLState }

func TestEmbeddedSchemaAndSQLState(testContext *testing.T) {
	if schema == "" {
		testContext.Fatal("le schéma SQL embarqué est vide")
	}
	if !hasSQLState(sqlStateErrorStub("doublon"), uniqueViolationSQLState) {
		testContext.Fatal("le code SQL d'unicité devrait être reconnu")
	}
	if hasSQLState(sqlStateErrorStub("doublon"), "00000") {
		testContext.Fatal("un autre code SQL ne devrait pas être reconnu")
	}
}
