package dynamicapi

import (
	"context"

	"lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
)

// IDynamicV1 defines the backend API contract for the dynamic sample plugin.
type IDynamicV1 interface {
	BackendSummary(ctx context.Context, req *v1.BackendSummaryReq) (res *v1.BackendSummaryRes, err error)
	DemoRecordList(ctx context.Context, req *v1.DemoRecordListReq) (res *v1.DemoRecordListRes, err error)
	DemoRecord(ctx context.Context, req *v1.DemoRecordReq) (res *v1.DemoRecordRes, err error)
	CreateDemoRecord(ctx context.Context, req *v1.CreateDemoRecordReq) (res *v1.CreateDemoRecordRes, err error)
	UpdateDemoRecord(ctx context.Context, req *v1.UpdateDemoRecordReq) (res *v1.UpdateDemoRecordRes, err error)
	DeleteDemoRecord(ctx context.Context, req *v1.DeleteDemoRecordReq) (res *v1.DeleteDemoRecordRes, err error)
	DownloadDemoRecordAttachment(ctx context.Context, req *v1.DownloadDemoRecordAttachmentReq) (res *v1.DownloadDemoRecordAttachmentRes, err error)
	HostCallDemo(ctx context.Context, req *v1.HostCallDemoReq) (res *v1.HostCallDemoRes, err error)
	ManifestDemo(ctx context.Context, req *v1.ManifestDemoReq) (res *v1.ManifestDemoRes, err error)
	BeforeInstall(ctx context.Context, req *v1.BeforeInstallReq) (res *v1.LifecycleDecisionRes, err error)
	AfterInstall(ctx context.Context, req *v1.AfterInstallReq) (res *v1.LifecycleDecisionRes, err error)
	BeforeUpgrade(ctx context.Context, req *v1.BeforeUpgradeReq) (res *v1.LifecycleDecisionRes, err error)
	Upgrade(ctx context.Context, req *v1.UpgradeReq) (res *v1.LifecycleDecisionRes, err error)
	AfterUpgrade(ctx context.Context, req *v1.AfterUpgradeReq) (res *v1.LifecycleDecisionRes, err error)
	BeforeDisable(ctx context.Context, req *v1.BeforeDisableReq) (res *v1.LifecycleDecisionRes, err error)
	AfterDisable(ctx context.Context, req *v1.AfterDisableReq) (res *v1.LifecycleDecisionRes, err error)
	BeforeUninstall(ctx context.Context, req *v1.BeforeUninstallReq) (res *v1.LifecycleDecisionRes, err error)
	Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.LifecycleDecisionRes, err error)
	AfterUninstall(ctx context.Context, req *v1.AfterUninstallReq) (res *v1.LifecycleDecisionRes, err error)
	BeforeTenantDisable(ctx context.Context, req *v1.BeforeTenantDisableReq) (res *v1.LifecycleDecisionRes, err error)
	AfterTenantDisable(ctx context.Context, req *v1.AfterTenantDisableReq) (res *v1.LifecycleDecisionRes, err error)
	BeforeTenantDelete(ctx context.Context, req *v1.BeforeTenantDeleteReq) (res *v1.LifecycleDecisionRes, err error)
	AfterTenantDelete(ctx context.Context, req *v1.AfterTenantDeleteReq) (res *v1.LifecycleDecisionRes, err error)
	BeforeInstallModeChange(ctx context.Context, req *v1.BeforeInstallModeChangeReq) (res *v1.LifecycleDecisionRes, err error)
	AfterInstallModeChange(ctx context.Context, req *v1.AfterInstallModeChangeReq) (res *v1.LifecycleDecisionRes, err error)
	RegisterJobs(ctx context.Context, req *v1.RegisterJobsReq) (res *v1.RegisterJobsRes, err error)
	JobHeartbeat(ctx context.Context, req *v1.JobHeartbeatReq) (res *v1.JobHeartbeatRes, err error)
}
