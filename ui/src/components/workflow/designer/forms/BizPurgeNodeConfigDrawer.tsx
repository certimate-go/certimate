import { useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { Form } from "antd";

import { type WorkflowNodeConfigForBizPurge } from "@/domain/workflow";

import { NodeConfigDrawer } from "./_shared";
import BizPurgeNodeConfigForm from "./BizPurgeNodeConfigForm";
import { NodeType } from "../nodes/typings";

export interface BizPurgeNodeConfigDrawerProps {
  afterClose?: () => void;
  loading?: boolean;
  node: FlowNodeEntity;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

const BizPurgeNodeConfigDrawer = ({ node, ...props }: BizPurgeNodeConfigDrawerProps) => {
  if (node.flowNodeType !== NodeType.BizPurge) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizPurge}`);
  }

  const { i18n } = useTranslation();

  const [formInst] = Form.useForm();

  const fieldProvider = Form.useWatch<WorkflowNodeConfigForBizPurge["provider"]>("provider", { form: formInst, preserve: true });

  return (
    <NodeConfigDrawer
      anchor={fieldProvider ? { items: BizPurgeNodeConfigForm.getAnchorItems({ i18n }) } : false}
      footer={fieldProvider ? void 0 : false}
      form={formInst}
      node={node}
      {...props}
    >
      <BizPurgeNodeConfigForm form={formInst} node={node} />
    </NodeConfigDrawer>
  );
};

export default BizPurgeNodeConfigDrawer;
