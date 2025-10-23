import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderBaotaPanelGoSite = () => {
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
        name={[parentNamePath, "siteName"]}
        initialValue={initialValues.siteName}
        label={t("workflow_node.deploy.form.baotapanelgo_site_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotapanelgo_site_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.baotapanelgo_site_name.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    siteName: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    siteName: z.string().nonempty(t("workflow_node.deploy.form.baotapanelgo_site_name.placeholder")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBaotaPanelGoSite, {
  getInitialValues,
  getSchema,
});

export default _default;
