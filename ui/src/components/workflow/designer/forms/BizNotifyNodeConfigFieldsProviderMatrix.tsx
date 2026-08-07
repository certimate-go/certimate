import type { getI18n } from "react-i18next";
import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizNotifyNodeConfigFieldsProviderMatrix = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item
        name={[parentNamePath, "roomId"]}
        label={t("workflow_node.notify.form.matrix_room_id.label")}
        extra={t("workflow_node.notify.form.matrix_room_id.help")}
        rules={[formRule]}
        initialValue={initialValues?.roomId}
      >
        <Input allowClear placeholder={t("workflow_node.notify.form.matrix_room_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<Record<string, unknown>> => {
  return {};
};

const getSchema = (_opts?: { i18n?: ReturnType<typeof getI18n> }) => {
  return z.object({
    roomId: z.string().nullish(),
  });
};

const _default = Object.assign(BizNotifyNodeConfigFieldsProviderMatrix, {
  getInitialValues,
  getSchema,
});

export default _default;
