import { useEffect, useMemo, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { IconDice6 } from "@tabler/icons-react";
import { type AnchorProps, Button, Form, type FormInstance, Input, Radio, Space } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import dayjs from "dayjs";
import { z } from "zod";

import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { WORKFLOW_TRIGGERS, type WorkflowNodeConfigForStart, defaultNodeConfigForStart } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { getNextCronExecutions, validateCronExpression } from "@/utils/cron";

import { NodeFormContextProvider } from "./_context";
import { NodeType } from "../nodes/typings";

export interface StartNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const StartNodeConfigForm = ({ node, ...props }: StartNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.Start) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.Start}`);
  }

  const { i18n, t } = useTranslation();

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForStart | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeStartConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldTrigger = Form.useWatch("trigger", formInst);
  const fieldTriggerCron = Form.useWatch("triggerCron", formInst);
  const [fieldTriggerCronExpectedExecutions, setFieldTriggerCronExpectedExecutions] = useState<Date[]>([]);
  useEffect(() => {
    setFieldTriggerCronExpectedExecutions(getNextCronExecutions(fieldTriggerCron!, 5));
  }, [fieldTriggerCron]);

  const handleTriggerChange = (value: string) => {
    if (value === WORKFLOW_TRIGGERS.SCHEDULED) {
      formInst.setFieldValue("triggerCron", initialValues?.triggerCron || "0 0 * * *");
    } else {
      formInst.setFieldValue("triggerCron", void 0);
    }
  };

  const handleRandomCronClick = () => {
    const m = Math.floor(Math.random() * 60);
    const h = Math.floor(Math.random() * 24);
    formInst.setFieldValue("triggerCron", `${m} ${h} * * *`);
  };

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item name="trigger" label={t("workflow_node.start.form.trigger.label")} rules={[formRule]}>
            <Radio.Group onChange={(e) => handleTriggerChange(e.target.value)}>
              <Radio value={WORKFLOW_TRIGGERS.MANUAL}>{t("workflow_node.start.form.trigger.option.manual.label")}</Radio>
              <Radio value={WORKFLOW_TRIGGERS.SCHEDULED}>{t("workflow_node.start.form.trigger.option.scheduled.label")}</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            hidden={fieldTrigger !== WORKFLOW_TRIGGERS.SCHEDULED}
            label={t("workflow_node.start.form.trigger_cron.label")}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.start.form.trigger_cron.tooltip") }}></span>}
            extra={
              <Show when={fieldTriggerCronExpectedExecutions.length > 0}>
                <div>
                  {t("workflow_node.start.form.trigger_cron.help")}
                  <br />
                  {fieldTriggerCronExpectedExecutions.map((date, index) => (
                    <span key={index}>
                      {dayjs(date).format("YYYY-MM-DD HH:mm:ss")}
                      <br />
                    </span>
                  ))}
                </div>
              </Show>
            }
          >
            <Space.Compact className="w-full">
              <Form.Item name="triggerCron" noStyle rules={[formRule]}>
                <Input placeholder={t("workflow_node.start.form.trigger_cron.placeholder")} />
              </Form.Item>
              <Button className="px-2" onClick={handleRandomCronClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Space.Compact>
          </Form.Item>

          <Show when={fieldTrigger === WORKFLOW_TRIGGERS.SCHEDULED}>
            <Form.Item>
              <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.start.form.trigger_cron.guide") }}></span>} />
            </Form.Item>
          </Show>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters"].map((key) => ({
    key: key,
    title: t(`workflow_node.start.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    trigger: WORKFLOW_TRIGGERS.MANUAL,
    ...(defaultNodeConfigForStart() as Nullish<z.infer<ReturnType<typeof getSchema>>>),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      trigger: z.string().nonempty(t("workflow_node.start.form.trigger.placeholder")),
      triggerCron: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.trigger === WORKFLOW_TRIGGERS.SCHEDULED) {
        if (!validateCronExpression(values.triggerCron!)) {
          ctx.addIssue({
            code: "custom",
            message: t("workflow_node.start.form.trigger_cron.errmsg.invalid"),
            path: ["triggerCron"],
          });
        }
      }
    });
};

const _default = Object.assign(StartNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
