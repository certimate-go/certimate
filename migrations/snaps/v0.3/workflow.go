package snaps

// This is a definition backup of WorkflowNode for v0.3.
type WorkflowNode struct {
	Id       string          `json:"id"`
	Type     string          `json:"type"`
	Name     string          `json:"name"`
	Config   map[string]any  `json:"config,omitempty"`
	Next     *WorkflowNode   `json:"next,omitempty"`
	Branches []*WorkflowNode `json:"branches,omitempty"`
}
