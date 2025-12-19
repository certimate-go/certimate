import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { validDomainName } from "@/utils/validators";

import { useFormNestedFieldsContext } from "./_context";

const ACCESS_MODE_CNAME = "cname" as const;

const BizDeployNodeConfigFieldsProviderVolcEngineWAF = () => {
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
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.volcengine_waf_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_waf_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_waf_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessMode"]}
        initialValue={initialValues.accessMode}
        label={t("workflow_node.deploy.form.volcengine_waf_access_mode.label")}
        rules={[formRule]}
      >
        <Select
          options={[ACCESS_MODE_CNAME].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.volcengine_waf_access_mode.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.volcengine_waf_access_mode.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.deploy.form.volcengine_waf_domain.label")}
        rules={[formRule]}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.volcengine_waf_domain.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    accessMode: ACCESS_MODE_CNAME,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    region: z.string().nonempty(t("workflow_node.deploy.form.volcengine_waf_region.placeholder")),
    accessMode: z.literal([ACCESS_MODE_CNAME], t("workflow_node.deploy.form.volcengine_waf_access_mode.placeholder")),
    domain: z.string().refine((v) => validDomainName(v, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderVolcEngineWAF, {
  getInitialValues,
  getSchema,
});

export default _default;
