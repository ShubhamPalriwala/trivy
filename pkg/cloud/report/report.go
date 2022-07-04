package report

import (
	"encoding/json"
	"io"

	"golang.org/x/xerrors"

	"github.com/aquasecurity/defsec/pkg/scan"
	dbTypes "github.com/aquasecurity/trivy-db/pkg/types"
	"github.com/aquasecurity/trivy/pkg/types"
)

const (
	allReport     = "all"
	summaryReport = "summary"

	tableFormat = "table"
	jsonFormat  = "json"

	schemaVersion = 0
)

type Option struct {
	Format     string
	Type       string
	Output     io.Writer
	Severities []dbTypes.Severity
	FromCache  bool
}

// Report represents a kubernetes scan report
type Report struct {
	SchemaVersion   int           `json:"schema_version"`
	AccountID       string        `json:"account_id"`
	Results         types.Results `json:"results"`
	ServicesInScope []string      `json:"services"`
}

func New(accountID string, defsecResults scan.Results, scopedServices []string) *Report {
	return &Report{
		SchemaVersion:   schemaVersion,
		AccountID:       accountID,
		Results:         convertResults(defsecResults),
		ServicesInScope: scopedServices,
	}
}

// Failed returns whether the aws report includes any "failed" results
func (r Report) Failed() bool {
	return r.Results.Failed()
}

// Write writes the results in the give format
func Write(report *Report, option Option) error {
	switch option.Format {
	case jsonFormat:
		return json.NewEncoder(option.Output).Encode(report)
	case tableFormat:
		return writeSummaryTable(option.Output, report, option.FromCache)
	default:
		return xerrors.Errorf(`unknown format %q. Use "json" or "table"`, option.Format)
	}
}

func (r *Report) ForService(service string) *Report {
	var results types.Results
	for _, result := range r.Results {
		var misconfigurations []types.DetectedMisconfiguration
		for _, misconfig := range result.Misconfigurations {
			if misconfig.CauseMetadata.Service == service {
				misconfigurations = append(misconfigurations, misconfig)
			}
		}
		if len(misconfigurations) > 0 {
			copied := result
			copied.Misconfigurations = misconfigurations
			results = append(results, copied)
		}
	}
	return &Report{
		SchemaVersion:   schemaVersion,
		AccountID:       r.AccountID,
		Results:         results,
		ServicesInScope: []string{service},
	}
}
