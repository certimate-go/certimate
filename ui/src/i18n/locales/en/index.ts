import nlsAccess from "./nls.access.json";
import nlsCertificate from "./nls.certificate.json";
import nlsCommon from "./nls.common.json";
import nlsDashboard from "./nls.dashboard.json";
import nlsLogin from "./nls.login.json";
import nlsPreset from "./nls.preset.json";
import nlsProvider from "./nls.provider.json";
import nlsSettings from "./nls.settings.json";
import nlsWorkflow from "./nls.workflow.json";
import nlsWorkflowNodes from "./nls.workflow.nodes.json";
import nlsWorkflowRuns from "./nls.workflow.runs.json";
import nlsWorkflowVars from "./nls.workflow.vars.json";

export default Object.freeze({
  ...nlsCommon,
  ...nlsLogin,
  ...nlsDashboard,
  ...nlsSettings,
  ...nlsProvider,
  ...nlsAccess,
  ...nlsPreset,
  ...nlsCertificate,
  ...nlsWorkflow,
  ...nlsWorkflowNodes,
  ...nlsWorkflowRuns,
  ...nlsWorkflowVars,
});
