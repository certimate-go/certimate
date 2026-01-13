import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderSynologyDSM = () => {
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
        name={[parentNamePath, "certificateName"]}
        initialValue={initialValues.certificateName}
        label={t("workflow_node.deploy.form.synologydsm_certificate_name.label")}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.synologydsm_certificate_name.tooltip") }}></span>}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.synologydsm_certificate_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateId"]}
        initialValue={initialValues.certificateId}
        label={t("workflow_node.deploy.form.synologydsm_certificate_id.label")}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.synologydsm_certificate_id.tooltip") }}></span>}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.synologydsm_certificate_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "isDefault"]}
        initialValue={initialValues.isDefault}
        label={t("workflow_node.deploy.form.synologydsm_is_default.label")}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.synologydsm_is_default.tooltip") }}></span>}
        rules={[formRule]}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    certificateName: "",
    certificateId: "",
    isDefault: false,
  };
};

const getSchema = ({ i18n: _i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  return z.object({
    certificateName: z.string().nullish(),
    certificateId: z.string().nullish(),
    isDefault: z.boolean().nullish(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderSynologyDSM, {
  getInitialValues,
  getSchema,
});

export default _default;
