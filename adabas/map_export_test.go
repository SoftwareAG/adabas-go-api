package adabas

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const exportFileName = "/tmp/go-test-export.json"

func TestExportMap(t *testing.T) {
	os.Remove(exportFileName)
	url, _ := NewURL("23")
	dbURL := DatabaseURL{URL: *url, Fnr: 4}
	repository := NewMapRepositoryWithURL(dbURL)
	ada, _ := NewAdabas(url)
	err := repository.ExportMapRepository(ada, "", exportFileName)
	assert.NoError(t, err)
}
