package conformance

import (
	"testing"

	"github.com/apigee/registry/rpc"
)

func InitReport(t *testing.T) *rpc.ConformanceReport {
	t.Helper()
	return &rpc.ConformanceReport{
		GuidelineReportGroups: []*rpc.GuidelineReportGroup{
			{
				Status:           rpc.Guideline_STATUS_UNSPECIFIED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			{
				Status:           rpc.Guideline_PROPOSED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			{
				Status:           rpc.Guideline_ACTIVE,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			{
				Status:           rpc.Guideline_DEPRECATED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			{
				Status:           rpc.Guideline_DISABLED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
		},
	}
}

func InitRuleReportGroups(t *testing.T) []*rpc.RuleReportGroup {
	t.Helper()
	return []*rpc.RuleReportGroup{
		{
			Severity:    rpc.Rule_SEVERITY_UNSPECIFIED,
			RuleReports: []*rpc.RuleReport{},
		},
		{
			Severity:    rpc.Rule_ERROR,
			RuleReports: []*rpc.RuleReport{},
		},
		{
			Severity:    rpc.Rule_WARNING,
			RuleReports: []*rpc.RuleReport{},
		},
		{
			Severity:    rpc.Rule_INFO,
			RuleReports: []*rpc.RuleReport{},
		},
		{
			Severity:    rpc.Rule_HINT,
			RuleReports: []*rpc.RuleReport{},
		},
	}
}
