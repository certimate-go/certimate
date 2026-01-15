package snaps

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pocketbase/pocketbase/core"
)

type WorkflowGraphWalker struct {
	visitors []WorkflowNodeVisitor
}

type WorkflowNodeVisitor func(node *WorkflowNode) (_changed bool, _err error)

func (w *WorkflowGraphWalker) Define(visitor WorkflowNodeVisitor) {
	if w.visitors == nil {
		w.visitors = make([]WorkflowNodeVisitor, 0)
	}
	w.visitors = append(w.visitors, visitor)
}

func (w *WorkflowGraphWalker) Visit(nodes []*WorkflowNode) (_changed bool, _err error) {
	changed := false

	if w.visitors == nil {
		return changed, nil
	}

	for _, node := range nodes {
		for _, visitor := range w.visitors {
			nodeChanged, err := visitor(node)
			if err != nil {
				return changed, err
			}
			if nodeChanged {
				changed = true
			}

			if len(node.Blocks) > 0 {
				blocksChanged, err := w.Visit(node.Blocks)
				if err != nil {
					return changed, err
				}
				if blocksChanged {
					changed = true
				}
			}
		}
	}

	return changed, nil
}

func (w *WorkflowGraphWalker) Migrate(record *core.Record, field string) (_changed bool, _err error) {
	f := record.Collection().Fields.GetByName(field)
	if f == nil {
		return false, fmt.Errorf("field '%s' not found", field)
	}

	if record.GetRaw(field) != nil {
		graph := make(map[string]any)
		if err := record.UnmarshalJSONField(field, &graph); err != nil {
			return false, err
		}

		if _, ok := graph["nodes"]; ok {
			nodes := make([]*WorkflowNode, 0)
			if err := mapstructure.Decode(graph["nodes"], &nodes); err != nil {
				return false, err
			}

			nodesChanged, err := w.Visit(nodes)
			if err != nil {
				return false, err
			} else if nodesChanged {
				graph["nodes"] = nodes
				record.Set(field, graph)
				return true, nil
			}
		}
	}

	return false, nil
}

// This is a definition copy of WorkflowNode.
// see: /internal/domain/workflow.go
type WorkflowNode struct {
	Id     string             `json:"id"`
	Type   string             `json:"type"`
	Data   WorkflowNodeData   `json:"data"`
	Blocks WorkflowNodeBlocks `json:"blocks,omitempty,omitzero"`
}

// This is a definition copy of []*WorkflowNode.
// see: /internal/domain/workflow.go
type WorkflowNodeBlocks []*WorkflowNode

// This is a definition copy of WorkflowNodeData.
// see: /internal/domain/workflow.go
type WorkflowNodeData struct {
	Name     string             `json:"name"`
	Disabled bool               `json:"disabled,omitempty,omitzero"`
	Config   WorkflowNodeConfig `json:"config,omitempty,omitzero"`
}

// This is a definition copy of WorkflowNodeConfig.
// see: /internal/domain/workflow.go
type WorkflowNodeConfig map[string]any

func (g WorkflowNodeBlocks) GetNodeById(nodeId string) (*WorkflowNode, bool) {
	return g.getNodeInBlocksById(g, nodeId)
}

func (g WorkflowNodeBlocks) getNodeInBlocksById(blocks WorkflowNodeBlocks, nodeId string) (*WorkflowNode, bool) {
	for _, node := range blocks {
		if node.Id == nodeId {
			return node, true
		}

		if len(node.Blocks) > 0 {
			if found, ok := g.getNodeInBlocksById(node.Blocks, nodeId); ok {
				return found, true
			}
		}
	}

	return nil, false
}
