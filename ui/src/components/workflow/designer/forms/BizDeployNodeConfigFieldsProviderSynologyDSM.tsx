import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";

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
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.synologydsm.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateIdOrDesc"]}
        initialValue={initialValues.certificateIdOrDesc}
        label={t("workflow_node.deploy.form.synologydsm_certificate_id_or_desc.label")}
        extra={t("workflow_node.deploy.form.synologydsm_certificate_id_or_desc.help")}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.synologydsm_certificate_id_or_desc.tooltip") }}></span>}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.synologydsm_certificate_id_or_desc.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "isDefault"]}
        initialValue={initialValues.isDefault}
        label={t("workflow_node.deploy.form.synologydsm_is_default.label")}
        rules={[formRule]}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    isDefault: true,
  };
};

const getSchema = ({ i18n: _i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  return z.object({
    certificateIdOrDesc: z.string().nullish(),
    isDefault: z.boolean().nullish(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderSynologyDSM, {
  getInitialValues,
  getSchema,
});

export default _default;
