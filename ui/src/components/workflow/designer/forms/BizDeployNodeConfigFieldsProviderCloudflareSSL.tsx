import { getI18n, useTranslation } from "react-i18next";
import { AutoComplete, Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderCloudflareSSL = () => {
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
        name={[parentNamePath, "environment"]}
        initialValue={initialValues.environment}
        label={t("workflow_node.deploy.form.cloudflare_ssl_environment.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.cloudflare_ssl_environment.tooltip") }}></span>}
      >
        <AutoComplete
          allowClear
          options={["production", "staging"].map((s) => ({ value: s }))}
          placeholder={t("workflow_node.deploy.form.cloudflare_ssl_environment.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "zoneId"]}
        initialValue={initialValues.zoneId}
        label={t("workflow_node.deploy.form.cloudflare_ssl_zone_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.cloudflare_ssl_zone_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.cloudflare_ssl_zone_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateId"]}
        initialValue={initialValues.certificateId}
        label={t("workflow_node.deploy.form.cloudflare_ssl_certificate_id.label")}
        extra={t("workflow_node.deploy.form.cloudflare_ssl_certificate_id.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.cloudflare_ssl_certificate_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.cloudflare_ssl_certificate_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    environment: "production",
    zoneId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    environment: z.string().nullish(),
    zoneId: z.string().nonempty(),
    certificateId: z.string().nullish(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderCloudflareSSL, {
  getInitialValues,
  getSchema,
});

export default _default;
