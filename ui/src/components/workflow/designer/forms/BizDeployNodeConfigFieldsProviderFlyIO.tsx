import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderFlyIO = () => {
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
        name={[parentNamePath, "appName"]}
        initialValue={initialValues.appName}
        label={t("workflow_node.deploy.form.flyio_app_name.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.flyio_app_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "hostname"]}
        initialValue={initialValues.hostname}
        label={t("workflow_node.deploy.form.flyio_hostname.label")}
        extra={t("workflow_node.deploy.form.flyio_hostname.help")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.flyio_hostname.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    appName: "",
    hostname: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    appName: z.string().nonempty(t("workflow_node.deploy.form.flyio_app_name.placeholder")),
    hostname: z.string().nonempty(t("workflow_node.deploy.form.flyio_hostname.placeholder")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderFlyIO, {
  getInitialValues,
  getSchema,
});

export default _default;
