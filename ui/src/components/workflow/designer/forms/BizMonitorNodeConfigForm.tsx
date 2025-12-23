import { useMemo } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { type AnchorProps, Form, type FormInstance, Input, InputNumber } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";
import { type WorkflowNodeConfigForBizMonitor, defaultNodeConfigForBizMonitor } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { isDomain, isHostname, isPortNumber } from "@/utils/validator";

import { NodeFormContextProvider } from "./_context";
import { NodeType } from "../nodes/typings";

export interface BizMonitorNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const BizMonitorNodeConfigForm = ({ node, ...props }: BizMonitorNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizMonitor) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizMonitor}`);
  }

  const { i18n, t } = useTranslation();

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForBizMonitor | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizMonitorConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item>
            <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.monitor.form.guide") }}></span>} />
          </Form.Item>

          <div className="flex space-x-2">
            <div className="w-2/3">
              <Form.Item name="host" label={t("workflow_node.monitor.form.host.label")} rules={[formRule]}>
                <Input placeholder={t("workflow_node.monitor.form.host.placeholder")} />
              </Form.Item>
            </div>

            <div className="w-1/3">
              <Form.Item name="port" label={t("workflow_node.monitor.form.port.label")} rules={[formRule]}>
                <InputNumber style={{ width: "100%" }} min={1} max={65535} placeholder={t("workflow_node.monitor.form.port.placeholder")} />
              </Form.Item>
            </div>
          </div>

          <Form.Item name="domain" label={t("workflow_node.monitor.form.domain.label")} extra={t("workflow_node.monitor.form.domain.help")} rules={[formRule]}>
            <Input placeholder={t("workflow_node.monitor.form.domain.placeholder")} />
          </Form.Item>

          <Form.Item name="requestPath" label={t("workflow_node.monitor.form.request_path.label")} rules={[formRule]}>
            <Input placeholder={t("workflow_node.monitor.form.request_path.placeholder")} />
          </Form.Item>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters"].map((key) => ({
    key: key,
    title: t(`workflow_node.monitor.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return defaultNodeConfigForBizMonitor();
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    host: z.string().refine((v) => isHostname(v), t("common.errmsg.host_invalid")),
    port: z.coerce.number().refine((v) => isPortNumber(v), t("common.errmsg.port_invalid")),
    domain: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return isDomain(v);
      }, t("common.errmsg.domain_invalid")),
    requestPath: z.string().nullish(),
  });
};

const _default = Object.assign(BizMonitorNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
