import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizUploadNodeConfigFieldsProviderAutoSSL = () => {
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
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.upload.form.autossl_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.upload.form.autossl_domain.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.upload.form.autossl_domain.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certId"]}
        initialValue={initialValues.certId}
        label={t("workflow_node.upload.form.autossl_cert_id.label")}
        extra={t("workflow_node.upload.form.autossl_cert_id.help")}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.upload.form.autossl_cert_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.upload.form.autossl_cert_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    domain: "",
    certId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    domain: z.string().nonempty(),
    certId: z.string().nullish(),
  });
};

const _default = Object.assign(BizUploadNodeConfigFieldsProviderAutoSSL, {
  getInitialValues,
  getSchema,
});

export default _default;
