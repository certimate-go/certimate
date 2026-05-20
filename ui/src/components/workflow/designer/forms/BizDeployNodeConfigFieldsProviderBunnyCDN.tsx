import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderBunnyCDN = () => {
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
        name={[parentNamePath, "pullZoneId"]}
        initialValue={initialValues.pullZoneId}
        label={t("workflow_node.deploy.form.bunny_cdn_pull_zone_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.bunny_cdn_pull_zone_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.bunny_cdn_pull_zone_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "hostname"]}
        initialValue={initialValues.hostname}
        label={t("workflow_node.deploy.form.bunny_cdn_hostname.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.bunny_cdn_hostname.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.bunny_cdn_hostname.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    pullZoneId: "",
    hostname: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    pullZoneId: z.union([z.string().nonempty(), z.int().positive()]),
    hostname: z.string().refine((v) => isDomain(v, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBunnyCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
