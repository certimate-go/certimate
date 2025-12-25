import { useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { Form } from "antd";

import { type WorkflowNodeConfigForBizApply } from "@/domain/workflow";

import { NodeConfigDrawer } from "./_shared";
import BizApplyNodeConfigForm from "./BizApplyNodeConfigForm";
import { NodeType } from "../nodes/typings";

export interface BizApplyNodeConfigDrawerProps {
  afterClose?: () => void;
  loading?: boolean;
  node: FlowNodeEntity;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

const BizApplyNodeConfigDrawer = ({ node, ...props }: BizApplyNodeConfigDrawerProps) => {
  if (node.flowNodeType !== NodeType.BizApply) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizApply}`);
  }

  const { i18n } = useTranslation();

  const [formInst] = Form.useForm();

  const fieldIdentifier = Form.useWatch<WorkflowNodeConfigForBizApply["identifier"]>("identifier", { form: formInst, preserve: true });

  return (
    <NodeConfigDrawer
      anchor={fieldIdentifier ? { items: BizApplyNodeConfigForm.getAnchorItems({ i18n }) } : false}
      footer={fieldIdentifier ? void 0 : false}
      form={formInst}
      node={node}
      {...props}
    >
      <BizApplyNodeConfigForm form={formInst} node={node} />
    </NodeConfigDrawer>
  );
};

export default BizApplyNodeConfigDrawer;
