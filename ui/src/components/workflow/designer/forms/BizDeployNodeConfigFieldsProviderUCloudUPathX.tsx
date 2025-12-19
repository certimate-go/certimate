import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { validPortNumber } from "@/utils/validators";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderUCloudUPathX = () => {
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
        name={[parentNamePath, "acceleratorId"]}
        initialValue={initialValues.acceleratorId}
        label={t("workflow_node.deploy.form.ucloud_upathx_accelerator_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_upathx_accelerator_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_upathx_accelerator_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "listenerPort"]}
        initialValue={initialValues.listenerPort}
        label={t("workflow_node.deploy.form.ucloud_upathx_listener_port.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_upathx_listener_port.tooltip") }}></span>}
      >
        <Input type="number" min={1} max={65535} placeholder={t("workflow_node.deploy.form.ucloud_upathx_listener_port.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    acceleratorId: "",
    listenerPort: 443,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    acceleratorId: z.string().nonempty(t("workflow_node.deploy.form.ucloud_upathx_accelerator_id.placeholder")),
    listenerPort: z.coerce.number().refine((v) => validPortNumber(v), t("common.errmsg.port_invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderUCloudUPathX, {
  getInitialValues,
  getSchema,
});

export default _default;
