import nlsAccess from "./nls.access.json";
import nlsCertificate from "./nls.certificate.json";
import nlsCommon from "./nls.common.json";
import nlsDashboard from "./nls.dashboard.json";
import nlsLogin from "./nls.login.json";
import nlsPlugin from "./nls.plugin.json";
import nlsPreset from "./nls.preset.json";
import nlsProvider from "./nls.provider.json";
import nlsSettings from "./nls.settings.json";
import nlsWorkflow from "./nls.workflow.json";
import nlsWorkflowNodes from "./nls.workflow.nodes.json";
import nlsWorkflowRuns from "./nls.workflow.runs.json";
import { buildTranslations } from "../utils";

export default buildTranslations(
  nlsCommon,
  nlsLogin,
  nlsDashboard,
  nlsSettings,
  nlsProvider,
  nlsAccess,
  nlsPreset,
  nlsPlugin,
  nlsCertificate,
  nlsWorkflow,
  nlsWorkflowNodes,
  nlsWorkflowRuns
);
