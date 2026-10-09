package error

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

func TestGMNContractMatchesMigrationTokens(t *testing.T) {
	migrationPath := filepath.Join("..", "..", "..", "migrations", "002_service_functions.up.sql")
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}

	re := regexp.MustCompile(`GMN_[A-Z0-9_]+`)
	found := re.FindAllString(string(content), -1)
	if len(found) == 0 {
		t.Fatal("no GMN_* tokens found in migration")
	}

	migrationSet := make(map[string]struct{}, len(found))
	for _, token := range found {
		migrationSet[token] = struct{}{}
	}

	contractSet := make(map[string]struct{}, len(allGMNTokens))
	for _, token := range GMNTokens() {
		contractSet[token] = struct{}{}
	}

	var missingInContract []string
	for token := range migrationSet {
		if _, ok := contractSet[token]; !ok {
			missingInContract = append(missingInContract, token)
		}
	}

	var missingInMigration []string
	for token := range contractSet {
		if _, ok := migrationSet[token]; !ok {
			missingInMigration = append(missingInMigration, token)
		}
	}

	sort.Strings(missingInContract)
	sort.Strings(missingInMigration)

	if len(missingInContract) > 0 || len(missingInMigration) > 0 {
		t.Fatalf("GMN token contract drift detected; missingInContract=%v missingInMigration=%v", missingInContract, missingInMigration)
	}
}
