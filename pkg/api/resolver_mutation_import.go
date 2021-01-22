package api

import (
	"context"

	"github.com/stashapp/stashdb/pkg/manager/bulkimport"
	"github.com/stashapp/stashdb/pkg/models"
)

func (r *mutationResolver) AnalyzeData(ctx context.Context, input models.BulkImportInput) (*models.BulkAnalyzeResult, error) {
	return bulkimport.Analyze(input)
}

func (r *mutationResolver) ImportData(ctx context.Context, input models.BulkImportInput) (*models.BulkImportResult, error) {
	data, err := bulkimport.Analyze(input)
	if err != nil {
		return nil, err
	}

	return bulkimport.ApplyImport(data)
}
