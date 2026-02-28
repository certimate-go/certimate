import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isDomain } from "@/utils/validator";

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
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.deploy.form.flyio_domain.label")}
        extra={t("workflow_node.deploy.form.flyio_domain.help")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.flyio_domain.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    appName: "",
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    appName: z.string().nonempty(t("workflow_node.deploy.form.flyio_app_name.placeholder")),
    domain: z.string().refine((v) => isDomain(v), t("workflow_node.deploy.form.flyio_domain.placeholder")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderFlyIO, {
  getInitialValues,
  getSchema,
});

export default _default;
