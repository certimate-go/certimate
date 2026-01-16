import { getI18n, useTranslation } from "react-i18next";
import { IconBulb } from "@tabler/icons-react";
import { Button, Form, Input, Popover } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import CodeTextInput from "@/components/CodeTextInput";
import { isJsonObject } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const BizNotifyNodeConfigFieldsProviderWebhook = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const handleWebhookDataBlur = () => {
    const value = formInst.getFieldValue([parentNamePath, "webhookData"]);
    try {
      const json = JSON.stringify(JSON.parse(value), null, 2);
      formInst.setFieldValue([parentNamePath, "webhookData"], json);
    } catch {
      return;
    }
  };

  return (
    <>
      <Form.Item label={t("workflow_node.notify.form.webhook_data.label")} extra={t("workflow_node.notify.form.webhook_data.help")}>
        <div className="absolute -top-1.5 right-0 -translate-y-full">
          <Popover content={<div dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.webhook_data.vartips") }} />} mouseEnterDelay={1}>
            <Button color="default" size="small" variant="link">
              <IconBulb size="1.25em" />
            </Button>
          </Popover>
        </div>
        <Form.Item name={[parentNamePath, "webhookData"]} initialValue={initialValues.webhookData} noStyle rules={[formRule]}>
          <CodeTextInput
            lineWrapping={false}
            height="auto"
            minHeight="64px"
            maxHeight="256px"
            language="json"
            placeholder={t("workflow_node.notify.form.webhook_data.placeholder")}
            onBlur={handleWebhookDataBlur}
          />
        </Form.Item>
      </Form.Item>

      <Form.Item name={[parentNamePath, "timeout"]} label={t("workflow_node.notify.form.webhook_timeout.label")} rules={[formRule]}>
        <Input
          type="number"
          allowClear
          min={0}
          max={3600}
          placeholder={t("workflow_node.notify.form.webhook_timeout.placeholder")}
          suffix={t("workflow_node.notify.form.webhook_timeout.unit")}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {};
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    webhookData: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return isJsonObject(v);
      }, t("common.errmsg.json_invalid")),
    timeout: z.preprocess(
      (v) => (v == null || v === "" ? void 0 : Number(v)),
      z.number().int().gte(1, t("workflow_node.notify.form.webhook_timeout.placeholder")).nullish()
    ),
  });
};

const _default = Object.assign(BizNotifyNodeConfigFieldsProviderWebhook, {
  getInitialValues,
  getSchema,
});

export default _default;
