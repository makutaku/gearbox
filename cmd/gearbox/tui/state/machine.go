package state

import (
	"fmt"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
)

// Stage represents a discrete unit of work in a state machine
type Stage struct {
	Name           string
	Description    string
	Action         func() tea.Cmd
	IsCompleteFunc func() bool
	Reset          func() error
	Timeout        time.Duration
	
	// Internal state
	IsComplete bool
	Error      error
	StartTime  time.Time
	EndTime    time.Time
}

// StateMachine manages a sequence of stages for complex workflows
type StateMachine struct {
	name        string
	stages      []Stage
	currentIdx  int
	isRunning   bool
	isPaused    bool
	completed   bool
	failed      bool
	failedStage int
	startTime   time.Time
	endTime     time.Time
}

// NewStateMachine creates a new state machine
func NewStateMachine(name string, stages []Stage) *StateMachine {
	return &StateMachine{
		name:        name,
		stages:      stages,
		currentIdx:  0,
		failedStage: -1,
	}
}

// Start begins execution of the state machine
func (sm *StateMachine) Start() tea.Cmd {
	if sm.isRunning {
		return nil
	}
	
	sm.isRunning = true
	sm.startTime = time.Now()
	
	return sm.executeCurrentStage()
}

// Pause pauses execution of the state machine
func (sm *StateMachine) Pause() {
	sm.isPaused = true
}

// Resume resumes execution of the state machine
func (sm *StateMachine) Resume() tea.Cmd {
	if !sm.isPaused {
		return nil
	}
	
	sm.isPaused = false
	return sm.executeCurrentStage()
}

// Reset resets the state machine to the beginning
func (sm *StateMachine) Reset() error {
	sm.currentIdx = 0
	sm.isRunning = false
	sm.isPaused = false
	sm.completed = false
	sm.failed = false
	sm.failedStage = -1
	sm.startTime = time.Time{}
	sm.endTime = time.Time{}
	
	// Reset all stages
	for i := range sm.stages {
		sm.stages[i].IsComplete = false
		sm.stages[i].Error = nil
		sm.stages[i].StartTime = time.Time{}
		sm.stages[i].EndTime = time.Time{}
		
		if sm.stages[i].Reset != nil {
			if err := sm.stages[i].Reset(); err != nil {
				return fmt.Errorf("failed to reset stage %s: %w", sm.stages[i].Name, err)
			}
		}
	}
	
	return nil
}

// Next advances to the next stage
func (sm *StateMachine) Next() tea.Cmd {
	if !sm.isRunning || sm.isPaused || sm.completed || sm.failed {
		return nil
	}
	
	// Mark current stage as complete
	if sm.currentIdx < len(sm.stages) {
		sm.stages[sm.currentIdx].IsComplete = true
		sm.stages[sm.currentIdx].EndTime = time.Now()
	}
	
	sm.currentIdx++
	
	// Check if we've completed all stages
	if sm.currentIdx >= len(sm.stages) {
		sm.completed = true
		sm.isRunning = false
		sm.endTime = time.Now()
		return sm.onComplete()
	}
	
	return sm.executeCurrentStage()
}

// Fail marks the current stage as failed
func (sm *StateMachine) Fail(err error) tea.Cmd {
	if sm.currentIdx < len(sm.stages) {
		sm.stages[sm.currentIdx].Error = err
		sm.stages[sm.currentIdx].EndTime = time.Now()
	}
	
	sm.failed = true
	sm.failedStage = sm.currentIdx
	sm.isRunning = false
	sm.endTime = time.Now()
	
	return sm.onFailure(err)
}

// executeCurrentStage executes the current stage
func (sm *StateMachine) executeCurrentStage() tea.Cmd {
	if sm.currentIdx >= len(sm.stages) {
		return nil
	}
	
	stage := &sm.stages[sm.currentIdx]
	
	// Check if stage is already complete
	if stage.IsCompleteFunc != nil && stage.IsCompleteFunc() {
		stage.IsComplete = true
		return sm.Next()
	}
	
	// Start the stage
	stage.StartTime = time.Now()
	
	if stage.Action != nil {
		return stage.Action()
	}
	
	// If no action, consider it complete
	return sm.Next()
}

// GetCurrentStage returns the currently executing stage
func (sm *StateMachine) GetCurrentStage() *Stage {
	if sm.currentIdx >= len(sm.stages) {
		return nil
	}
	return &sm.stages[sm.currentIdx]
}

// GetStageByName returns a stage by name
func (sm *StateMachine) GetStageByName(name string) *Stage {
	for i := range sm.stages {
		if sm.stages[i].Name == name {
			return &sm.stages[i]
		}
	}
	return nil
}

// GetProgress returns the current progress as a percentage
func (sm *StateMachine) GetProgress() float64 {
	if len(sm.stages) == 0 {
		return 0
	}
	
	completedStages := 0
	for _, stage := range sm.stages {
		if stage.IsComplete {
			completedStages++
		}
	}
	
	return float64(completedStages) / float64(len(sm.stages)) * 100
}

// GetStatus returns the current status of the state machine
func (sm *StateMachine) GetStatus() MachineStatus {
	switch {
	case sm.failed:
		return StatusFailed
	case sm.completed:
		return StatusCompleted
	case sm.isPaused:
		return StatusPaused
	case sm.isRunning:
		return StatusRunning
	default:
		return StatusReady
	}
}

// GetSummary returns a summary of the state machine execution
func (sm *StateMachine) GetSummary() MachineSummary {
	summary := MachineSummary{
		Name:           sm.name,
		Status:         sm.GetStatus(),
		Progress:       sm.GetProgress(),
		TotalStages:    len(sm.stages),
		CompletedStages: 0,
		FailedStage:    sm.failedStage,
		StartTime:      sm.startTime,
		EndTime:        sm.endTime,
		Duration:       sm.getDuration(),
	}
	
	for _, stage := range sm.stages {
		if stage.IsComplete {
			summary.CompletedStages++
		}
	}
	
	return summary
}

// getDuration returns the total execution duration
func (sm *StateMachine) getDuration() time.Duration {
	if sm.startTime.IsZero() {
		return 0
	}
	
	endTime := sm.endTime
	if endTime.IsZero() && sm.isRunning {
		endTime = time.Now()
	}
	
	return endTime.Sub(sm.startTime)
}

// onComplete is called when the state machine completes successfully
func (sm *StateMachine) onComplete() tea.Cmd {
	return func() tea.Msg {
		return StateMachineCompleteMsg{
			Name:     sm.name,
			Duration: sm.getDuration(),
			Summary:  sm.GetSummary(),
		}
	}
}

// onFailure is called when the state machine fails
func (sm *StateMachine) onFailure(err error) tea.Cmd {
	return func() tea.Msg {
		return StateMachineFailedMsg{
			Name:        sm.name,
			Error:       err,
			FailedStage: sm.failedStage,
			Duration:    sm.getDuration(),
			Summary:     sm.GetSummary(),
		}
	}
}

// MachineStatus represents the status of a state machine
type MachineStatus int

const (
	StatusReady MachineStatus = iota
	StatusRunning
	StatusPaused
	StatusCompleted
	StatusFailed
)

// String returns a string representation of the status
func (s MachineStatus) String() string {
	switch s {
	case StatusReady:
		return "Ready"
	case StatusRunning:
		return "Running"
	case StatusPaused:
		return "Paused"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// MachineSummary provides a summary of state machine execution
type MachineSummary struct {
	Name            string
	Status          MachineStatus
	Progress        float64
	TotalStages     int
	CompletedStages int
	FailedStage     int
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
}

// Messages for state machine events

// StateMachineCompleteMsg is sent when a state machine completes
type StateMachineCompleteMsg struct {
	Name     string
	Duration time.Duration
	Summary  MachineSummary
}

// StateMachineFailedMsg is sent when a state machine fails
type StateMachineFailedMsg struct {
	Name        string
	Error       error
	FailedStage int
	Duration    time.Duration
	Summary     MachineSummary
}

// StageCompleteMsg is sent when a stage completes
type StageCompleteMsg struct {
	MachineName string
	StageName   string
	Duration    time.Duration
}

// StageFailedMsg is sent when a stage fails
type StageFailedMsg struct {
	MachineName string
	StageName   string
	Error       error
	Duration    time.Duration
}