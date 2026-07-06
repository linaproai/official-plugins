// This file verifies dynamic-demo scheduled-job domain capability usage without
// requiring a running Wasm host service.

package dynamicservice

import (
	"context"
	"slices"
	"testing"
	"time"

	jobv1 "lina-core/api/job/v1"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/jobcap"
)

// TestDynamicServiceInitializesJobsCapabilityClient verifies New wires the
// guest-side scheduled-job domain client into the dynamic sample service.
func TestDynamicServiceInitializesJobsCapabilityClient(t *testing.T) {
	created := New()
	service, ok := created.(*serviceImpl)
	if !ok {
		t.Fatalf("expected New to return *serviceImpl, got %T", created)
	}
	if service.jobsSvc == nil {
		t.Fatal("expected dynamic service to initialize Jobs capability client")
	}
}

// TestDynamicJobCapabilityUsageCoversFullContract verifies dynamic plugin
// service code can consume the complete runtime jobcap.Service surface.
func TestDynamicJobCapabilityUsageCoversFullContract(t *testing.T) {
	jobs := &recordingDynamicJobcapService{createID: "dynamic-job-42"}
	service := &serviceImpl{jobsSvc: jobs}

	if service.jobsSvc == nil {
		t.Fatal("expected test service to expose Jobs capability service")
	}
	exerciseDynamicJobCapabilityUsage(t, t.Context(), service.jobsSvc)
	jobs.assertFullUsage(t)
}

// recordingDynamicJobcapService records all jobcap.Service method calls.
type recordingDynamicJobcapService struct {
	createID        jobcap.JobID
	createdInput    jobcap.SaveInput
	getID           jobcap.JobID
	batchIDs        []jobcap.JobID
	listInput       jobcap.ListInput
	visibleIDs      []jobcap.JobID
	updatedInput    jobcap.UpdateInput
	runID           jobcap.JobID
	statusID        jobcap.JobID
	status          jobv1.Status
	deleteID        jobcap.JobID
	createCalled    bool
	getCalled       bool
	batchGetCalled  bool
	listCalled      bool
	ensureCalled    bool
	updateCalled    bool
	runCalled       bool
	setStatusCalled bool
	deleteCalled    bool
}

// Get returns one fake visible job projection.
func (s *recordingDynamicJobcapService) Get(_ context.Context, id jobcap.JobID) (*jobcap.JobInfo, error) {
	s.getCalled = true
	s.getID = id
	return s.info(id), nil
}

// BatchGet returns one fake visible job projection plus requested missing IDs.
func (s *recordingDynamicJobcapService) BatchGet(
	_ context.Context,
	ids []jobcap.JobID,
) (*capmodel.BatchResult[*jobcap.JobInfo, jobcap.JobID], error) {
	s.batchGetCalled = true
	s.batchIDs = append([]jobcap.JobID(nil), ids...)
	result := &capmodel.BatchResult[*jobcap.JobInfo, jobcap.JobID]{
		Items:      map[jobcap.JobID]*jobcap.JobInfo{},
		MissingIDs: []jobcap.JobID{},
	}
	for _, id := range ids {
		if id == s.createID {
			result.Items[id] = s.info(id)
			continue
		}
		result.MissingIDs = append(result.MissingIDs, id)
	}
	return result, nil
}

// List returns one fake bounded job page.
func (s *recordingDynamicJobcapService) List(
	_ context.Context,
	input jobcap.ListInput,
) (*capmodel.PageResult[*jobcap.JobInfo], error) {
	s.listCalled = true
	s.listInput = input
	return &capmodel.PageResult[*jobcap.JobInfo]{
		Items: []*jobcap.JobInfo{s.info(s.createID)},
		Total: 1,
	}, nil
}

// EnsureVisible records one visibility check request.
func (s *recordingDynamicJobcapService) EnsureVisible(_ context.Context, ids []jobcap.JobID) error {
	s.ensureCalled = true
	s.visibleIDs = append([]jobcap.JobID(nil), ids...)
	return nil
}

// Create records one governed job create request.
func (s *recordingDynamicJobcapService) Create(_ context.Context, input jobcap.SaveInput) (jobcap.JobID, error) {
	s.createCalled = true
	s.createdInput = input
	return s.createID, nil
}

// Update records one governed job update request.
func (s *recordingDynamicJobcapService) Update(_ context.Context, input jobcap.UpdateInput) error {
	s.updateCalled = true
	s.updatedInput = input
	return nil
}

// Delete records one governed job delete request.
func (s *recordingDynamicJobcapService) Delete(_ context.Context, id jobcap.JobID) error {
	s.deleteCalled = true
	s.deleteID = id
	return nil
}

// Run records one governed job execution request.
func (s *recordingDynamicJobcapService) Run(_ context.Context, id jobcap.JobID) error {
	s.runCalled = true
	s.runID = id
	return nil
}

// SetStatus records one governed job status transition request.
func (s *recordingDynamicJobcapService) SetStatus(_ context.Context, id jobcap.JobID, status jobv1.Status) error {
	s.setStatusCalled = true
	s.statusID = id
	s.status = status
	return nil
}

// info builds one fake job projection with a task-level log retention policy.
func (s *recordingDynamicJobcapService) info(id jobcap.JobID) *jobcap.JobInfo {
	return &jobcap.JobInfo{
		ID:     id,
		Name:   "Dynamic Plugin Jobcap Smoke",
		Group:  "dynamic-demo",
		Status: jobv1.StatusEnabled,
		LogRetentionOverride: &jobcap.LogRetentionOption{
			Mode:  jobv1.RetentionModeDays,
			Value: 60,
		},
	}
}

// assertFullUsage verifies every jobcap method was called with expected values.
func (s *recordingDynamicJobcapService) assertFullUsage(t *testing.T) {
	t.Helper()

	if !s.createCalled || !s.getCalled || !s.batchGetCalled || !s.listCalled ||
		!s.ensureCalled || !s.updateCalled || !s.runCalled || !s.setStatusCalled || !s.deleteCalled {
		t.Fatalf("expected every jobcap method to be called, got %#v", s)
	}
	assertDynamicRetention(t, s.createdInput.LogRetentionOverride, jobv1.RetentionModeDays, 60)
	assertDynamicRetention(t, s.updatedInput.LogRetentionOverride, jobv1.RetentionModeNone, 0)
	if s.getID != s.createID || s.runID != s.createID || s.statusID != s.createID || s.deleteID != s.createID {
		t.Fatalf("expected all single-job calls to use %q, got get=%q run=%q status=%q delete=%q",
			s.createID, s.getID, s.runID, s.statusID, s.deleteID)
	}
	if !slices.Equal(s.batchIDs, []jobcap.JobID{s.createID, "dynamic-missing-job"}) {
		t.Fatalf("unexpected batch IDs: %#v", s.batchIDs)
	}
	if !slices.Equal(s.visibleIDs, []jobcap.JobID{s.createID}) {
		t.Fatalf("unexpected visible IDs: %#v", s.visibleIDs)
	}
	if s.status != jobv1.StatusDisabled {
		t.Fatalf("expected status %q, got %q", jobv1.StatusDisabled, s.status)
	}
	if s.listInput.Keyword != "dynamic" ||
		s.listInput.Group != "dynamic-demo" ||
		s.listInput.Status != jobv1.StatusEnabled ||
		s.listInput.Page.PageNum != 1 ||
		s.listInput.Page.PageSize != 20 {
		t.Fatalf("unexpected list input: %#v", s.listInput)
	}
}

// exerciseDynamicJobCapabilityUsage calls every method on the Jobs capability.
func exerciseDynamicJobCapabilityUsage(t *testing.T, ctx context.Context, jobs jobcap.Service) {
	t.Helper()

	createdID, err := jobs.Create(ctx, jobcap.SaveInput{
		GroupID:        "dynamic-demo",
		Name:           "Dynamic Plugin Jobcap Smoke",
		Description:    "Covers dynamic-plugin scheduled-job capability usage.",
		Timeout:        30 * time.Second,
		ShellCmd:       "echo dynamic jobcap smoke",
		WorkDir:        "/tmp",
		Env:            map[string]string{"PLUGIN_ID": "linapro-demo-dynamic"},
		CronExpr:       "0 */30 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobv1.ScopeMasterOnly,
		Concurrency:    jobv1.ConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  3,
		Status:         jobv1.StatusEnabled,
		LogRetentionOverride: &jobcap.LogRetentionOption{
			Mode:  jobv1.RetentionModeDays,
			Value: 60,
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := jobs.Get(ctx, createdID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	assertDynamicRetention(t, got.LogRetentionOverride, jobv1.RetentionModeDays, 60)

	batch, err := jobs.BatchGet(ctx, []jobcap.JobID{createdID, "dynamic-missing-job"})
	if err != nil {
		t.Fatalf("BatchGet returned error: %v", err)
	}
	if batch.Items[createdID] == nil || len(batch.MissingIDs) != 1 {
		t.Fatalf("unexpected batch result: %#v", batch)
	}
	assertDynamicRetention(t, batch.Items[createdID].LogRetentionOverride, jobv1.RetentionModeDays, 60)

	page, err := jobs.List(ctx, jobcap.ListInput{
		Keyword: "dynamic",
		Group:   "dynamic-demo",
		Status:  jobv1.StatusEnabled,
		Page: capmodel.PageRequest{
			PageNum:  1,
			PageSize: 20,
		},
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("unexpected list result: %#v", page)
	}
	assertDynamicRetention(t, page.Items[0].LogRetentionOverride, jobv1.RetentionModeDays, 60)

	if err = jobs.EnsureVisible(ctx, []jobcap.JobID{createdID}); err != nil {
		t.Fatalf("EnsureVisible returned error: %v", err)
	}
	if err = jobs.Update(ctx, jobcap.UpdateInput{
		ID: createdID,
		SaveInput: jobcap.SaveInput{
			GroupID:        "dynamic-demo",
			Name:           "Dynamic Plugin Jobcap Smoke Updated",
			Description:    "Updates dynamic-plugin scheduled-job capability usage.",
			Timeout:        time.Minute,
			CronExpr:       "0 */45 * * * *",
			Timezone:       "Asia/Shanghai",
			Scope:          jobv1.ScopeMasterOnly,
			Concurrency:    jobv1.ConcurrencySingleton,
			MaxConcurrency: 1,
			MaxExecutions:  5,
			Status:         jobv1.StatusEnabled,
			LogRetentionOverride: &jobcap.LogRetentionOption{
				Mode:  jobv1.RetentionModeNone,
				Value: 0,
			},
		},
	}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if err = jobs.Run(ctx, createdID); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if err = jobs.SetStatus(ctx, createdID, jobv1.StatusDisabled); err != nil {
		t.Fatalf("SetStatus returned error: %v", err)
	}
	if err = jobs.Delete(ctx, createdID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

// assertDynamicRetention verifies one task-level log retention option.
func assertDynamicRetention(
	t *testing.T,
	option *jobcap.LogRetentionOption,
	mode jobv1.RetentionMode,
	value int64,
) {
	t.Helper()

	if option == nil || option.Mode != mode || option.Value != value {
		t.Fatalf("expected retention mode=%q value=%d, got %#v", mode, value, option)
	}
}
