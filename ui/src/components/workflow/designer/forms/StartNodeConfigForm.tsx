import { useEffect, useMemo, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import QuestionCircleOutlined from "@ant-design/icons/QuestionCircleOutlined";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { IconDice6 } from "@tabler/icons-react";
import { type AnchorProps, Button, Form, type FormInstance, Input, Radio, Space, Tooltip, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import dayjs from "dayjs";
import { z } from "zod";

import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { WORKFLOW_TRIGGERS, type WorkflowNodeConfigForStart, type WorkflowSimpleSchedule, defaultNodeConfigForStart } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { cronToSimpleSchedule, getNextCronExecutions, simpleScheduleToCron, validateCronExpression } from "@/utils/cron";

import { NodeFormContextProvider } from "./_context";
import SimplifiedScheduleInput from "./SimplifiedScheduleInput";
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
  const initialSchedule = useMemo(() => cronToSimpleSchedule(initialValues?.triggerCron || "0 0 * * *"), [initialValues?.triggerCron]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeStartConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldTrigger = Form.useWatch("trigger", formInst);
  const fieldTriggerCron = Form.useWatch("triggerCron", formInst);
  const [useSimpleEditor, setUseSimpleEditor] = useState(() => {
    return initialValues?.trigger === WORKFLOW_TRIGGERS.SCHEDULED ? initialSchedule != null : true;
  });
  const [simpleSchedule, setSimpleSchedule] = useState<WorkflowSimpleSchedule>(() => {
    return initialSchedule ?? defaultNodeConfigForStart().triggerSchedule!;
  });
  const fieldTriggerCronExpectedExecutions = useMemo(() => getNextCronExecutions(fieldTriggerCron!, 5), [fieldTriggerCron]);
  useEffect(() => {
    if (fieldTrigger !== WORKFLOW_TRIGGERS.SCHEDULED || !useSimpleEditor || fieldTriggerCron) return;

    try {
      formInst.setFieldValue("triggerCron", simpleScheduleToCron(simpleSchedule));
    } catch {
      formInst.setFieldValue("triggerCron", void 0);
    }
  }, [fieldTrigger, fieldTriggerCron, formInst, simpleSchedule, useSimpleEditor]);

  const handleTriggerChange = (value: string) => {
    if (value === WORKFLOW_TRIGGERS.SCHEDULED) {
      const parsedSchedule = cronToSimpleSchedule(initialValues?.triggerCron || "");
      const nextSchedule = parsedSchedule ?? simpleSchedule;

      setUseSimpleEditor(parsedSchedule != null || !initialValues?.triggerCron);
      setSimpleSchedule(nextSchedule);
      formInst.setFieldValue("triggerCron", initialValues?.triggerCron || simpleScheduleToCron(nextSchedule));
    } else {
      formInst.setFieldValue("triggerCron", void 0);
    }
  };

  const handleSimpleScheduleChange = (value: WorkflowSimpleSchedule) => {
    setSimpleSchedule(value);

    try {
      formInst.setFieldValue("triggerCron", simpleScheduleToCron(value));
    } catch {
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
            label={
              useSimpleEditor ? (
                <div className="inline-flex items-center gap-1.5">
                  <span>{t("workflow_node.start.form.schedule.label")}</span>
                  <Tooltip title={t("workflow_node.start.form.schedule.tooltip")}>
                    <span className="ant-form-item-tooltip" tabIndex={-1}>
                      <QuestionCircleOutlined />
                    </span>
                  </Tooltip>
                  <Typography.Link className="text-xs" onClick={() => setUseSimpleEditor(false)}>
                    {t("workflow_node.start.form.schedule.switch_to_cron")}
                  </Typography.Link>
                </div>
              ) : (
                t("workflow_node.start.form.trigger_cron.label")
              )
            }
            tooltip={useSimpleEditor ? undefined : <span dangerouslySetInnerHTML={{ __html: t("workflow_node.start.form.trigger_cron.tooltip") }}></span>}
            extra={
              <Show when={fieldTriggerCronExpectedExecutions.length > 0}>
                <div className="mt-2 text-xs/6 text-gray-400">
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
            <Show
              when={useSimpleEditor}
              fallback={
                <Space.Compact className="w-full">
                  <Form.Item name="triggerCron" noStyle rules={[formRule]}>
                    <Input placeholder={t("workflow_node.start.form.trigger_cron.placeholder")} />
                  </Form.Item>
                  <Tooltip title={t("common.text.random_roll")}>
                    <Button className="px-2" onClick={handleRandomCronClick}>
                      <IconDice6 size="1.25em" />
                    </Button>
                  </Tooltip>
                </Space.Compact>
              }
            >
              <div>
                <Form.Item name="triggerCron" hidden rules={[formRule]}>
                  <Input />
                </Form.Item>
                <SimplifiedScheduleInput value={simpleSchedule} onChange={handleSimpleScheduleChange} />
              </div>
            </Show>
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
      trigger: z.string().nonempty(),
      triggerCron: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.trigger === WORKFLOW_TRIGGERS.SCHEDULED) {
        const scTriggerCron = z.string().refine((v) => validateCronExpression(v), t("workflow_node.start.form.trigger_cron.errmsg.invalid"));
        const spTriggerCron = scTriggerCron.safeParse(values.triggerCron);
        if (!spTriggerCron.success) {
          ctx.addIssue({
            code: "custom",
            message: z.treeifyError(spTriggerCron.error).errors.join(),
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
