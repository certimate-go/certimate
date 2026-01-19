import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderProxmoxVE = () => {
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
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.proxmoxve.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "nodeName"]}
        initialValue={initialValues.nodeName}
        label={t("workflow_node.deploy.form.proxmoxve_node_name.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.proxmoxve_node_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "autoRestart"]}
        initialValue={initialValues.autoRestart}
        label={t("workflow_node.deploy.form.proxmoxve_auto_restart.label")}
        rules={[formRule]}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    nodeName: "",
    autoRestart: true,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    nodeName: z.string().nonempty(t("workflow_node.deploy.form.proxmoxve_node_name.placeholder")),
    autoRestart: z.boolean().nullish(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderProxmoxVE, {
  getInitialValues,
  getSchema,
});

export default _default;
