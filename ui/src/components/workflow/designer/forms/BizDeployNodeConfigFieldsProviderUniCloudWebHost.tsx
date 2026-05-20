import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderUniCloudWebHost = () => {
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
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.unicloud_webhost.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "spaceProvider"]}
        initialValue={initialValues.spaceProvider}
        label={t("workflow_node.deploy.form.unicloud_webhost_space_provider.label")}
        rules={[formRule]}
      >
        <Select
          options={["aliyun", "alipay", "tencent"].map((s) => ({
            label: t(`workflow_node.deploy.form.unicloud_webhost_space_provider.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.unicloud_webhost_space_provider.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "spaceId"]}
        initialValue={initialValues.spaceId}
        label={t("workflow_node.deploy.form.unicloud_webhost_space_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.unicloud_webhost_space_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.unicloud_webhost_space_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.deploy.form.unicloud_webhost_domain.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.unicloud_webhost_domain.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    spaceProvider: "tencent",
    spaceId: "",
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    spaceProvider: z.string().nonempty(),
    spaceId: z.string().nonempty(),
    domain: z.string().refine((v) => isDomain(v), t("common.errmsg.domain_invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderUniCloudWebHost, {
  getInitialValues,
  getSchema,
});

export default _default;
