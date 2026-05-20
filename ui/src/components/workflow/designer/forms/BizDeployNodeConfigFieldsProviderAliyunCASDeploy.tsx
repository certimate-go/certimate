import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";

const BizDeployNodeConfigFieldsProviderAliyunCASDeploy = () => {
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
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_casdeploy.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.aliyun_casdeploy_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_casdeploy_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_casdeploy_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceIds"]}
        initialValue={initialValues.resourceIds}
        label={t("workflow_node.deploy.form.aliyun_casdeploy_resource_ids.label")}
        extra={t("workflow_node.deploy.form.aliyun_casdeploy_resource_ids.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_casdeploy_resource_ids.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.aliyun_casdeploy_resource_ids.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.aliyun_casdeploy_resource_ids.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.aliyun_casdeploy_resource_ids.multiple_input_modal.placeholder")}
          separator={MULTIPLE_INPUT_SEPARATOR}
          splitOptions={{ removeEmpty: true, trimSpace: true }}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "contactIds"]}
        initialValue={initialValues.contactIds}
        label={t("workflow_node.deploy.form.aliyun_casdeploy_contact_ids.label")}
        extra={t("workflow_node.deploy.form.aliyun_casdeploy_contact_ids.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_casdeploy_contact_ids.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.aliyun_casdeploy_contact_ids.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.aliyun_casdeploy_contact_ids.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.aliyun_casdeploy_contact_ids.multiple_input_modal.placeholder")}
          separator={MULTIPLE_INPUT_SEPARATOR}
          splitOptions={{ removeEmpty: true, trimSpace: true }}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    resourceIds: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    region: z.string().nonempty(),
    resourceIds: z
      .string()
      .nonempty()
      .refine((v) => {
        return v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => /^[1-9]\d*$/.test(e));
      }, t("workflow_node.deploy.form.aliyun_casdeploy_resource_ids.errmsg.invalid")),
    contactIds: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => /^[1-9]\d*$/.test(e));
      }, t("workflow_node.deploy.form.aliyun_casdeploy_contact_ids.errmsg.invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAliyunCASDeploy, {
  getInitialValues,
  getSchema,
});

export default _default;
