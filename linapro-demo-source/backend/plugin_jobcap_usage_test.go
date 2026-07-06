// This file verifies that the source demo plugin can consume the complete
// scheduled-job domain capability exposed through pluginhost Jobs registrars.

package backend

import (
	"context"
	"slices"
	"testing"
	"time"

	jobv1 "lina-core/api/job/v1"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/pluginhost"
)

// TestSourcePluginJobCapabilityUsageCoversFullContract verifies source-plugin
// Jobs registrar inputs expose the full runtime jobcap.Service surface.
func TestSourcePluginJobCapabilityUsageCoversFullContract(t *testing.T) {
	jobs := &recordingSourceJobcapService{createID: "source-job-42"}
	registrar := &recordingSourceJobsRegistrar{
		services: sourceJobcapCapabilityServices{jobs: jobs},
	}

	if err := registerBuiltinJobs(t.Context(), registrar); err != nil {
		t.Fatalf("expected built-in job registration to succeed, got error: %v", err)
	}
	if registrar.registeredName != sourcePluginEchoInspectionName {
		t.Fatalf("expected built-in job name %q, got %q", sourcePluginEchoInspectionName, registrar.registeredName)
	}
	if registrar.registeredHandler == nil {
		t.Fatal("expected built-in job handler to be registered")
	}
	if err := registrar.registeredHandler(t.Context()); err != nil {
		t.Fatalf("expected built-in job handler to succeed, got error: %v", err)
	}

	services := registrar.Services()
	if services == nil || services.Jobs() == nil {
		t.Fatal("expected source plugin registrar to expose Jobs capability service")
	}
	exerciseSourceJobCapabilityUsage(t, t.Context(), services.Jobs())
	jobs.assertFullUsage(t)
}

// sourceJobcapCapabilityServices exposes only the Jobs capability needed by
// this usage test while embedding the root interface for unused capabilities.
type sourceJobcapCapabilityServices struct {
	capability.Services
	jobs jobcap.Service
}

// Jobs returns the fake scheduled-job service under test.
func (s sourceJobcapCapabilityServices) Jobs() jobcap.Service {
	return s.jobs
}

// recordingSourceJobsRegistrar captures source-plugin Jobs registrations and
// exposes the configured fake capability directory.
type recordingSourceJobsRegistrar struct {
	services          capability.Services
	registeredPattern string
	registeredName    string
	registeredDisplay string
	registeredDesc    string
	registeredHandler pluginhost.JobHandler
}

// Add records one source-plugin job registration.
func (r *recordingSourceJobsRegistrar) Add(
	ctx context.Context,
	pattern string,
	name string,
	handler pluginhost.JobHandler,
) error {
	return r.AddWithMetadata(ctx, pattern, name, "", "", handler)
}

// AddWithMetadata records one source-plugin job registration with display metadata.
func (r *recordingSourceJobsRegistrar) AddWithMetadata(
	_ context.Context,
	pattern string,
	name string,
	displayName string,
	description string,
	handler pluginhost.JobHandler,
) error {
	r.registeredPattern = pattern
	r.registeredName = name
	r.registeredDisplay = displayName
	r.registeredDesc = description
	r.registeredHandler = handler
	return nil
}

// IsPrimaryNode reports a deterministic primary-node state for the test registrar.
func (r *recordingSourceJobsRegistrar) IsPrimaryNode() bool {
	return true
}

// Services returns the fake source-plugin capability directory.
func (r *recordingSourceJobsRegistrar) Services() capability.Services {
	return r.services
}

// recordingSourceJobcapService records all jobcap.Service method calls.
type recordingSourceJobcapService struct {
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
func (s *recordingSourceJobcapService) Get(_ context.Context, id jobcap.JobID) (*jobcap.JobInfo, error) {
	s.getCalled = true
	s.getID = id
	return s.info(id), nil
}

// BatchGet returns one fake visible job projection plus requested missing IDs.
func (s *recordingSourceJobcapService) BatchGet(
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
func (s *recordingSourceJobcapService) List(
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
func (s *recordingSourceJobcapService) EnsureVisible(_ context.Context, ids []jobcap.JobID) error {
	s.ensureCalled = true
	s.visibleIDs = append([]jobcap.JobID(nil), ids...)
	return nil
}

// Create records one governed job create request.
func (s *recordingSourceJobcapService) Create(_ context.Context, input jobcap.SaveInput) (jobcap.JobID, error) {
	s.createCalled = true
	s.createdInput = input
	return s.createID, nil
}

// Update records one governed job update request.
func (s *recordingSourceJobcapService) Update(_ context.Context, input jobcap.UpdateInput) error {
	s.updateCalled = true
	s.updatedInput = input
	return nil
}

// Delete records one governed job delete request.
func (s *recordingSourceJobcapService) Delete(_ context.Context, id jobcap.JobID) error {
	s.deleteCalled = true
	s.deleteID = id
	return nil
}

// Run records one governed job execution request.
func (s *recordingSourceJobcapService) Run(_ context.Context, id jobcap.JobID) error {
	s.runCalled = true
	s.runID = id
	return nil
}

// SetStatus records one governed job status transition request.
func (s *recordingSourceJobcapService) SetStatus(_ context.Context, id jobcap.JobID, status jobv1.Status) error {
	s.setStatusCalled = true
	s.statusID = id
	s.status = status
	return nil
}

// info builds one fake job projection with a task-level log retention policy.
func (s *recordingSourceJobcapService) info(id jobcap.JobID) *jobcap.JobInfo {
	return &jobcap.JobInfo{
		ID:     id,
		Name:   "Source Plugin Jobcap Smoke",
		Group:  "source-demo",
		Status: jobv1.StatusEnabled,
		LogRetentionOverride: &jobcap.LogRetentionOption{
			Mode:  jobv1.RetentionModeDays,
			Value: 60,
		},
	}
}

// assertFullUsage verifies every jobcap method was called with expected values.
func (s *recordingSourceJobcapService) assertFullUsage(t *testing.T) {
	t.Helper()

	if !s.createCalled || !s.getCalled || !s.batchGetCalled || !s.listCalled ||
		!s.ensureCalled || !s.updateCalled || !s.runCalled || !s.setStatusCalled || !s.deleteCalled {
		t.Fatalf("expected every jobcap method to be called, got %#v", s)
	}
	assertSourceRetention(t, s.createdInput.LogRetentionOverride, jobv1.RetentionModeDays, 60)
	assertSourceRetention(t, s.updatedInput.LogRetentionOverride, jobv1.RetentionModeNone, 0)
	if s.getID != s.createID || s.runID != s.createID || s.statusID != s.createID || s.deleteID != s.createID {
		t.Fatalf("expected all single-job calls to use %q, got get=%q run=%q status=%q delete=%q",
			s.createID, s.getID, s.runID, s.statusID, s.deleteID)
	}
	if !slices.Equal(s.batchIDs, []jobcap.JobID{s.createID, "source-missing-job"}) {
		t.Fatalf("unexpected batch IDs: %#v", s.batchIDs)
	}
	if !slices.Equal(s.visibleIDs, []jobcap.JobID{s.createID}) {
		t.Fatalf("unexpected visible IDs: %#v", s.visibleIDs)
	}
	if s.status != jobv1.StatusDisabled {
		t.Fatalf("expected status %q, got %q", jobv1.StatusDisabled, s.status)
	}
	if s.listInput.Keyword != "source" ||
		s.listInput.Group != "source-demo" ||
		s.listInput.Status != jobv1.StatusEnabled ||
		s.listInput.Page.PageNum != 1 ||
		s.listInput.Page.PageSize != 20 {
		t.Fatalf("unexpected list input: %#v", s.listInput)
	}
}

// exerciseSourceJobCapabilityUsage calls every method on the Jobs capability.
func exerciseSourceJobCapabilityUsage(t *testing.T, ctx context.Context, jobs jobcap.Service) {
	t.Helper()

	createdID, err := jobs.Create(ctx, jobcap.SaveInput{
		GroupID:        "source-demo",
		Name:           "Source Plugin Jobcap Smoke",
		Description:    "Covers source-plugin scheduled-job capability usage.",
		Timeout:        30 * time.Second,
		ShellCmd:       "echo source jobcap smoke",
		WorkDir:        "/tmp",
		Env:            map[string]string{"PLUGIN_ID": pluginID},
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
	assertSourceRetention(t, got.LogRetentionOverride, jobv1.RetentionModeDays, 60)

	batch, err := jobs.BatchGet(ctx, []jobcap.JobID{createdID, "source-missing-job"})
	if err != nil {
		t.Fatalf("BatchGet returned error: %v", err)
	}
	if batch.Items[createdID] == nil || len(batch.MissingIDs) != 1 {
		t.Fatalf("unexpected batch result: %#v", batch)
	}
	assertSourceRetention(t, batch.Items[createdID].LogRetentionOverride, jobv1.RetentionModeDays, 60)

	page, err := jobs.List(ctx, jobcap.ListInput{
		Keyword: "source",
		Group:   "source-demo",
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
	assertSourceRetention(t, page.Items[0].LogRetentionOverride, jobv1.RetentionModeDays, 60)

	if err = jobs.EnsureVisible(ctx, []jobcap.JobID{createdID}); err != nil {
		t.Fatalf("EnsureVisible returned error: %v", err)
	}
	if err = jobs.Update(ctx, jobcap.UpdateInput{
		ID: createdID,
		SaveInput: jobcap.SaveInput{
			GroupID:        "source-demo",
			Name:           "Source Plugin Jobcap Smoke Updated",
			Description:    "Updates source-plugin scheduled-job capability usage.",
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

// assertSourceRetention verifies one task-level log retention option.
func assertSourceRetention(
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
