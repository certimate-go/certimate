package engine

import (
	"fmt"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
)

// resolveLastCertificateForNode finds the certificate that should drive skip/renew decisions.
//
// Lookup order:
//  1. certificate#id ref stored in the last workflow_output (works after skip-only runs,
//     where the cert's workflowRunRef still points at the original apply run);
//  2. certificate by lastOutput.RunId + nodeId (legacy apply-run path);
//  3. latest certificate by workflowId + nodeId (survives workflow run history cleanup
//     that cascade-deletes workflow_output).
func resolveLastCertificateForNode(execCtx *NodeExecutionContext, certRepo certificateRepository, lastOutput *domain.WorkflowOutput) (*domain.Certificate, error) {
	if lastOutput != nil {
		for _, entry := range lastOutput.Outputs {
			if entry == nil || entry.Name != "certificate" || entry.Type != stateIOTypeRef {
				continue
			}
			parts := strings.Split(entry.Value, "#")
			if len(parts) != 2 || parts[0] != domain.CollectionNameCertificate || parts[1] == "" {
				continue
			}
			certificate, err := certRepo.GetById(execCtx.Context(), parts[1])
			if err != nil && !domain.IsRecordNotFoundError(err) {
				return nil, fmt.Errorf("failed to get last certificate record of node #%s: %w", execCtx.Node.Id, err)
			}
			if certificate != nil {
				return certificate, nil
			}
		}

		certificate, err := certRepo.GetByWorkflowRunIdAndNodeId(execCtx.Context(), lastOutput.RunId, lastOutput.NodeId)
		if err != nil && !domain.IsRecordNotFoundError(err) {
			return nil, fmt.Errorf("failed to get last certificate record of node #%s: %w", execCtx.Node.Id, err)
		}
		if certificate != nil {
			return certificate, nil
		}
	}

	certificate, err := certRepo.GetByWorkflowIdAndNodeId(execCtx.Context(), execCtx.WorkflowId, execCtx.Node.Id)
	if err != nil && !domain.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("failed to get last certificate record of node #%s: %w", execCtx.Node.Id, err)
	}
	return certificate, nil
}
