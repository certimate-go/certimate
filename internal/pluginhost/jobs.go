package pluginhost

import (
	"sync"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
)

type JobState string

const (
	JobQueued      JobState = "queued"
	JobDownloading JobState = "downloading"
	JobVerifying   JobState = "verifying"
	JobExtracting  JobState = "extracting"
	JobReloading   JobState = "reloading"
	JobInstalled   JobState = "installed"
	JobFailed      JobState = "failed"
)

func (s JobState) Terminal() bool {
	return s == JobInstalled || s == JobFailed
}

type JobStatus struct {
	ProviderType string        `json:"providerType"`
	State        JobState      `json:"state"`
	Stage        string        `json:"stage,omitempty"`
	Error        string        `json:"error,omitempty"`
	Downloaded   int64         `json:"downloaded,omitempty"`
	Total        int64         `json:"total,omitempty"`
	Result       *ReloadResult `json:"result,omitempty"`
}

type InstallJob struct {
	providerType string
	state        JobState
	stage        string
	err          string
	downloaded   int64
	total        int64
	result       *ReloadResult
	startedAt    time.Time
	mu           sync.RWMutex
}

func newInstallJob(providerType string) *InstallJob {
	return &InstallJob{providerType: providerType, state: JobQueued, startedAt: time.Now()}
}

func (j *InstallJob) setState(s JobState, stage string) {
	if j == nil {
		return
	}
	j.mu.Lock()
	j.state = s
	j.stage = stage
	j.mu.Unlock()
}

func (j *InstallJob) fail(msg string) {
	if j == nil {
		return
	}
	j.mu.Lock()
	j.state = JobFailed
	j.err = msg
	j.mu.Unlock()
}

func (j *InstallJob) succeed(result *ReloadResult) {
	if j == nil {
		return
	}
	j.mu.Lock()
	j.state = JobInstalled
	j.result = result
	j.mu.Unlock()
}

func (j *InstallJob) setProgress(downloaded, total int64) {
	if j == nil {
		return
	}
	j.mu.Lock()
	j.downloaded = downloaded
	j.total = max(j.total, total)
	j.mu.Unlock()
}

func (j *InstallJob) Status() JobStatus {
	if j == nil {
		return JobStatus{}
	}
	j.mu.RLock()
	defer j.mu.RUnlock()
	total := max(0, j.total)
	return JobStatus{
		ProviderType: j.providerType,
		State:        j.state,
		Stage:        j.stage,
		Error:        j.err,
		Downloaded:   j.downloaded,
		Total:        total,
		Result:       j.result,
	}
}

type jobStore struct {
	mu   sync.Mutex
	jobs map[string]*InstallJob
}

func newJobStore() *jobStore {
	return &jobStore{jobs: make(map[string]*InstallJob)}
}

func (s *jobStore) get(providerType string) *InstallJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jobs[providerType]
}

func (s *jobStore) set(j *InstallJob) {
	s.mu.Lock()
	s.jobs[j.providerType] = j
	s.mu.Unlock()
}

type opKind string

const (
	opNone    opKind = ""
	opInstall opKind = "install"
	opUpdate  opKind = "update"
	opDelete  opKind = "delete"
)

type pluginOps struct {
	mu     sync.Mutex
	active map[string]opKind
}

func newPluginOps() *pluginOps {
	return &pluginOps{active: make(map[string]opKind)}
}

var ErrOpInProgress = domain.NewError(409, "market: another operation is in progress for this plugin")

func (p *pluginOps) claim(providerType string, kind opKind) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if cur := p.active[providerType]; cur != opNone {
		return ErrOpInProgress
	}
	p.active[providerType] = kind
	return nil
}

func (p *pluginOps) release(providerType string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.active, providerType)
}
