import { forwardRef, memo, useEffect, useImperativeHandle, useState } from "react";
import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import dayjs from "dayjs";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { WORKFLOW_TRIGGERS, type WorkflowNodeConfigForStart, type WorkflowTriggerType, defaultNodeConfigForStart } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { getNextCronExecutions, validCronExpression } from "@/utils/cron";

type StartNodeConfigFormFieldValues = Partial<WorkflowNodeConfigForStart>;

export type StartNodeConfigFormProps = {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: StartNodeConfigFormFieldValues;
  onValuesChange?: (values: StartNodeConfigFormFieldValues) => void;
};

export type StartNodeConfigFormInstance = {
  getFieldsValue: () => ReturnType<FormInstance<StartNodeConfigFormFieldValues>["getFieldsValue"]>;
  resetFields: FormInstance<StartNodeConfigFormFieldValues>["resetFields"];
  validateFields: FormInstance<StartNodeConfigFormFieldValues>["validateFields"];
};

const initFormModel = (): StartNodeConfigFormFieldValues => {
  return defaultNodeConfigForStart();
};

const StartNodeConfigForm = forwardRef<StartNodeConfigFormInstance, StartNodeConfigFormProps>(
  ({ className, style, disabled, initialValues, onValuesChange }, ref) => {
    const { t } = useTranslation();

    const formSchema = z.object({
      trigger: z.string(t("workflow_node.start.form.trigger.placeholder")).min(1, t("workflow_node.start.form.trigger.placeholder")),
      triggerCron: z
        .string()
        .nullish()
        .refine((v) => {
          if (fieldTrigger !== WORKFLOW_TRIGGERS.AUTO) return true;
          return validCronExpression(v!);
        }, t("workflow_node.start.form.trigger_cron.errmsg.invalid")),
    });
    const formRule = createSchemaFieldRule(formSchema);
    const { form: formInst, formProps } = useAntdForm({
      name: "workflowNodeStartConfigForm",
      initialValues: initialValues ?? initFormModel(),
    });

    const fieldTrigger = Form.useWatch<WorkflowTriggerType>("trigger", formInst);
    const fieldTriggerCron = Form.useWatch<string>("triggerCron", formInst);
    const [fieldTriggerCronExpectedExecutions, setFieldTriggerCronExpectedExecutions] = useState<Date[]>([]);
    useEffect(() => {
      setFieldTriggerCronExpectedExecutions(getNextCronExecutions(fieldTriggerCron, 5));
    }, [fieldTriggerCron]);

    const handleTriggerChange = (value: string) => {
      if (value === WORKFLOW_TRIGGERS.AUTO) {
        formInst.setFieldValue("triggerCron", formProps.initialValues?.triggerCron || initFormModel().triggerCron);
      } else {
        formInst.setFieldValue("triggerCron", undefined);
      }

      onValuesChange?.(formInst.getFieldsValue(true));
    };

    const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
      onValuesChange?.(values as StartNodeConfigFormFieldValues);
    };

    useImperativeHandle(ref, () => {
      return {
        getFieldsValue: () => {
          return formInst.getFieldsValue(true);
        },
        resetFields: (fields) => {
          return formInst.resetFields(fields);
        },
        validateFields: (nameList, config) => {
          return formInst.validateFields(nameList, config);
        },
      } as StartNodeConfigFormInstance;
    });

    return (
      <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
        <Form.Item name="trigger" label={t("workflow_node.start.form.trigger.label")} rules={[formRule]}>
          <Radio.Group onChange={(e) => handleTriggerChange(e.target.value)}>
            <Radio value={WORKFLOW_TRIGGERS.AUTO}>{t("workflow_node.start.form.trigger.option.auto.label")}</Radio>
            <Radio value={WORKFLOW_TRIGGERS.MANUAL}>{t("workflow_node.start.form.trigger.option.manual.label")}</Radio>
          </Radio.Group>
        </Form.Item>

        <Form.Item
          name="triggerCron"
          label={t("workflow_node.start.form.trigger_cron.label")}
          hidden={fieldTrigger !== WORKFLOW_TRIGGERS.AUTO}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.start.form.trigger_cron.tooltip") }}></span>}
          extra={
            <Show when={fieldTriggerCronExpectedExecutions.length > 0}>
              <div>
                {t("workflow_node.start.form.trigger_cron.extra")}
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
          <Input placeholder={t("workflow_node.start.form.trigger_cron.placeholder")} />
        </Form.Item>

        <Show when={fieldTrigger === WORKFLOW_TRIGGERS.AUTO}>
          <Form.Item>
            <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.start.form.trigger_cron.guide") }}></span>} />
          </Form.Item>
        </Show>
      </Form>
    );
  }
);

export default memo(StartNodeConfigForm);
