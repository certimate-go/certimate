package engine

import (
	"github.com/certimate-go/certimate/internal/domain"
)

type Node = domain.WorkflowNode

type NodeType = domain.WorkflowNodeType

const (
	NodeTypeStart       = domain.WorkflowNodeTypeStart
	NodeTypeEnd         = domain.WorkflowNodeTypeEnd
	NodeTypeCondition   = domain.WorkflowNodeTypeCondition
	NodeTypeBranchBlock = domain.WorkflowNodeTypeBranchBlock
	NodeTypeTryCatch    = domain.WorkflowNodeTypeTryCatch
	NodeTypeTryBlock    = domain.WorkflowNodeTypeTryBlock
	NodeTypeCatchBlock  = domain.WorkflowNodeTypeCatchBlock
	NodeTypeDelay       = domain.WorkflowNodeTypeDelay
	NodeTypeBizApply    = domain.WorkflowNodeTypeBizApply
	NodeTypeBizUpload   = domain.WorkflowNodeTypeBizUpload
	NodeTypeBizMonitor  = domain.WorkflowNodeTypeBizMonitor
	NodeTypeBizDeploy   = domain.WorkflowNodeTypeBizDeploy
	NodeTypeBizPurge    = domain.WorkflowNodeTypeBizPurge
	NodeTypeBizNotify   = domain.WorkflowNodeTypeBizNotify
)

type Graph = domain.WorkflowGraph
