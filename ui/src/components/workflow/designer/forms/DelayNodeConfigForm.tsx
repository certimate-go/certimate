import { useMemo } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { type AnchorProps, Form, type FormInstance, InputNumber } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { type WorkflowNodeConfigForDelay, defaultNodeConfigForDelay } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";

import { NodeFormContextProvider } from "./_context";
import { NodeType } from "../nodes/typings";

export interface DelayNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const DelayNodeConfigForm = ({ node, ...props }: DelayNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.Delay) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.Delay}`);
  }

  const { i18n, t } = useTranslation();

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForDelay | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeDelayConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item name="wait" label={t("workflow_node.delay.form.wait.label")} rules={[formRule]}>
            <InputNumber
              style={{ width: "100%" }}
              min={0}
              max={3600}
              placeholder={t("workflow_node.delay.form.wait.placeholder")}
              suffix={t("workflow_node.delay.form.wait.unit")}
            />
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
    title: t(`workflow_node.delay.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    ...(defaultNodeConfigForDelay() as Nullish<z.infer<ReturnType<typeof getSchema>>>),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    wait: z.coerce.number().int().gt(0),
  });
};

const _default = Object.assign(DelayNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
