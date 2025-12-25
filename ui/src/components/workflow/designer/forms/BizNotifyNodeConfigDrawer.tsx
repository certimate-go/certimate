import { useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { Form } from "antd";

import { type WorkflowNodeConfigForBizNotify } from "@/domain/workflow";

import { NodeConfigDrawer } from "./_shared";
import BizNotifyNodeConfigForm from "./BizNotifyNodeConfigForm";
import { NodeType } from "../nodes/typings";

export interface BizNotifyNodeConfigDrawerProps {
  afterClose?: () => void;
  loading?: boolean;
  node: FlowNodeEntity;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

const BizNotifyNodeConfigDrawer = ({ node, ...props }: BizNotifyNodeConfigDrawerProps) => {
  if (node.flowNodeType !== NodeType.BizNotify) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizNotify}`);
  }

  const { i18n } = useTranslation();

  const [formInst] = Form.useForm();

  const fieldProvider = Form.useWatch<WorkflowNodeConfigForBizNotify["provider"]>("provider", { form: formInst, preserve: true });

  return (
    <NodeConfigDrawer
      anchor={fieldProvider ? { items: BizNotifyNodeConfigForm.getAnchorItems({ i18n }) } : false}
      footer={fieldProvider ? void 0 : false}
      form={formInst}
      node={node}
      {...props}
    >
      <BizNotifyNodeConfigForm form={formInst} node={node} />
    </NodeConfigDrawer>
  );
};

export default BizNotifyNodeConfigDrawer;
