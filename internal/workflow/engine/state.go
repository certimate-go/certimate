package engine

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"
)

type VariableState struct {
	Scope     string // 零值时表示全局的，否则表示指定节点的
	Key       string
	Value     any
	ValueType string
}

func (s VariableState) ValueString() string {
	switch s.ValueType {
	case stateValTypeString:
		return fmt.Sprintf("%s", s.Value)
	case stateValTypeNumber:
		return fmt.Sprintf("%d", s.Value)
	case stateValTypeBoolean:
		return strconv.FormatBool(s.Value.(bool))
	case stateValTypeDateTime:
		valueAsTime := s.Value.(time.Time)
		if valueAsTime.IsZero() {
			return "-"
		}
		return valueAsTime.Format(time.RFC3339)
	default:
		return fmt.Sprintf("[%s]%v", s.ValueType, s.Value)
	}
}

type VariableManager interface {
	All() []VariableState
	Erase()

	Add(entry VariableState)
	Set(name string, value any, valueType string)
	SetScoped(scope string, name string, value any, valueType string)
	Get(name string) (*VariableState, bool)
	GetScoped(scope string, key string) (*VariableState, bool)
	Take(key string) (*VariableState, bool)
	TakeScoped(scope string, key string) (*VariableState, bool)
	Remove(key string) bool
	RemoveScoped(scope string, key string) bool
}

type variableManager struct {
	statesMtx sync.RWMutex
	states    []VariableState
}

var _ VariableManager = (*variableManager)(nil)

func (m *variableManager) All() []VariableState {
	m.statesMtx.RLock()
	defer m.statesMtx.RUnlock()

	if m.states == nil {
		return make([]VariableState, 0)
	}

	return slices.Clone(m.states)
}

func (m *variableManager) Erase() {
	m.statesMtx.Lock()
	defer m.statesMtx.Unlock()

	m.states = make([]VariableState, 0)
}

func (m *variableManager) Add(state VariableState) {
	m.statesMtx.Lock()
	defer m.statesMtx.Unlock()

	if m.states == nil {
		m.states = make([]VariableState, 0)
	}

	for i, item := range m.states {
		if item.Scope == state.Scope && item.Key == state.Key {
			m.states[i] = state
			return
		}
	}
	m.states = append(m.states, state)
}

func (m *variableManager) Set(key string, value any, valueType string) {
	m.SetScoped("", key, value, valueType)
}

func (m *variableManager) SetScoped(scope string, key string, value any, valueType string) {
	m.Add(VariableState{Scope: scope, Key: key, Value: value, ValueType: valueType})
}

func (m *variableManager) Get(key string) (*VariableState, bool) {
	return m.GetScoped("", key)
}

func (m *variableManager) GetScoped(scope string, key string) (*VariableState, bool) {
	m.statesMtx.RLock()
	defer m.statesMtx.RUnlock()

	if m.states == nil {
		return nil, false
	}

	for _, item := range m.states {
		if item.Scope == scope && item.Key == key {
			return &item, true
		}
	}
	return nil, false
}

func (m *variableManager) Take(key string) (*VariableState, bool) {
	return m.TakeScoped("", key)
}

func (m *variableManager) TakeScoped(scope string, key string) (*VariableState, bool) {
	m.statesMtx.Lock()
	defer m.statesMtx.Unlock()

	if m.states == nil {
		return nil, false
	}

	for i, item := range m.states {
		if item.Scope == scope && item.Key == key {
			m.states = slices.Delete(m.states, i, i+1)
			return &item, true
		}
	}
	return nil, false
}

func (m *variableManager) Remove(key string) bool {
	return m.RemoveScoped("", key)
}

func (m *variableManager) RemoveScoped(scope string, key string) bool {
	_, ok := m.TakeScoped(scope, key)
	return ok
}

func newVariableManager() VariableManager {
	return &variableManager{
		states: make([]VariableState, 0),
	}
}

type InOutState struct {
	NodeId     string
	Type       string
	Name       string
	Value      any
	ValueType  string
	Persistent bool
}

func (s InOutState) ValueString() string {
	switch s.ValueType {
	case stateValTypeString:
		return s.Value.(string)
	case stateValTypeNumber:
		return fmt.Sprintf("%d", s.Value)
	case stateValTypeBoolean:
		return strconv.FormatBool(s.Value.(bool))
	default:
		return fmt.Sprintf("%v", s.Value)
	}
}

type InOutManager interface {
	All() []InOutState
	Erase()

	Add(state InOutState)
	Set(nodeId string, stype string, name string, value any, valueType string, persistent bool)
	Get(nodeId string, name string) (*InOutState, bool)
	Take(nodeId string, name string) (*InOutState, bool)
	Remove(nodeId string, name string) bool
}

type inoutManager struct {
	statesMtx sync.RWMutex
	states    []InOutState
}

var _ InOutManager = (*inoutManager)(nil)

func (m *inoutManager) All() []InOutState {
	m.statesMtx.RLock()
	defer m.statesMtx.RUnlock()

	if m.states == nil {
		return make([]InOutState, 0)
	}

	return slices.Clone(m.states)
}

func (m *inoutManager) Erase() {
	m.statesMtx.Lock()
	defer m.statesMtx.Unlock()

	m.states = make([]InOutState, 0)
}

func (m *inoutManager) Add(state InOutState) {
	m.statesMtx.Lock()
	defer m.statesMtx.Unlock()

	if m.states == nil {
		m.states = make([]InOutState, 0)
	}

	for i, item := range m.states {
		if item.NodeId == state.NodeId && item.Name == state.Name {
			m.states[i] = state
			return
		}
	}
	m.states = append(m.states, state)
}

func (m *inoutManager) Set(nodeId string, stype string, name string, value any, valueType string, persistent bool) {
	m.Add(InOutState{NodeId: nodeId, Type: stype, Name: name, Value: value, ValueType: valueType, Persistent: persistent})
}

func (m *inoutManager) Get(nodeId string, name string) (*InOutState, bool) {
	m.statesMtx.RLock()
	defer m.statesMtx.RUnlock()

	if m.states == nil {
		return nil, false
	}

	for _, item := range m.states {
		if item.NodeId == nodeId && item.Name == name {
			return &item, true
		}
	}
	return nil, false
}

func (m *inoutManager) Take(nodeId string, name string) (*InOutState, bool) {
	m.statesMtx.Lock()
	defer m.statesMtx.Unlock()

	if m.states == nil {
		return nil, false
	}

	for i, item := range m.states {
		if item.NodeId == nodeId && item.Name == name {
			m.states = slices.Delete(m.states, i, i+1)
			return &item, true
		}
	}
	return nil, false
}

func (m *inoutManager) Remove(nodeId string, name string) bool {
	_, ok := m.Take(nodeId, name)
	return ok
}

func newInOutManager() InOutManager {
	return &inoutManager{
		states: make([]InOutState, 0),
	}
}

const (
	stateValTypeBoolean  = "boolean"
	stateValTypeDateTime = "datetime"
	stateValTypeNumber   = "number"
	stateValTypeString   = "string"
)

const (
	stateIOTypeRef = "ref"
)

const (
	stateVarKeyWorkflowId                 = "workflow.id"                 // ValueType: "string"
	stateVarKeyWorkflowName               = "workflow.name"               // ValueType: "string"
	stateVarKeyRunId                      = "run.id"                      // ValueType: "string"
	stateVarKeyRunTrigger                 = "run.trigger"                 // ValueType: "string"
	stateVarKeyNodeId                     = "node.id"                     // ValueType: "string"
	stateVarKeyNodeName                   = "node.name"                   // ValueType: "string"
	stateVarKeyNodeSkipped                = "node.skipped"                // ValueType: "boolean"
	stateVarKeyErrorNodeId                = "error.nodeId"                // ValueType: "string"
	stateVarKeyErrorNodeName              = "error.nodeName"              // ValueType: "string"
	stateVarKeyErrorMessage               = "error.message"               // ValueType: "string"
	stateVarKeyCertificateDomain          = "certificate.domain"          // 已废弃，仅为兼容旧版而保留，请使用 [stateVarKeyCertificateCommonName]
	stateVarKeyCertificateDomains         = "certificate.domains"         // 已废弃，仅为兼容旧版而保留，请使用 [stateVarKeyCertificateSubjectAltNames]
	stateVarKeyCertificateCommonName      = "certificate.commonName"      // ValueType: "string"
	stateVarKeyCertificateSubjectAltNames = "certificate.subjectAltNames" // ValueType: "string"
	stateVarKeyCertificateNotBefore       = "certificate.notBefore"       // ValueType: "datetime"
	stateVarKeyCertificateNotAfter        = "certificate.notAfter"        // ValueType: "datetime"
	stateVarKeyCertificateHoursLeft       = "certificate.hoursLeft"       // ValueType: "number"
	stateVarKeyCertificateDaysLeft        = "certificate.daysLeft"        // ValueType: "number"
	stateVarKeyCertificateValidity        = "certificate.validity"        // ValueType: "boolean"
)
