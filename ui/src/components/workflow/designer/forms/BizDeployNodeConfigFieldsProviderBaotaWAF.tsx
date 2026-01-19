import { getI18n, useTranslation } from "react-i18next";
import { Form, InputNumber } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Tips from "@/components/Tips";
import { isPortNumber } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";

const BizDeployNodeConfigFieldsProviderBaotaWAF = () => {
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
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotawaf.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "siteNames"]}
        initialValue={initialValues.siteNames}
        label={t("workflow_node.deploy.form.baotawaf_site_names.label")}
        extra={t("workflow_node.deploy.form.baotawaf_site_names.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotawaf_site_names.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.baotawaf_site_names.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.baotawaf_site_names.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.baotawaf_site_names.multiple_input_modal.placeholder")}
          separator={MULTIPLE_INPUT_SEPARATOR}
          splitOptions={{ removeEmpty: true, trimSpace: true }}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "sitePort"]}
        initialValue={initialValues.sitePort}
        label={t("workflow_node.deploy.form.baotawaf_site_port.label")}
        rules={[formRule]}
      >
        <InputNumber style={{ width: "100%" }} placeholder={t("access.form.ssh_port.placeholder")} min={1} max={65535} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    siteNames: "",
    sitePort: 443,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    siteNames: z
      .string()
      .nonempty(t("workflow_node.deploy.form.baotawaf_site_names.placeholder"))
      .refine(
        (v) => {
          if (!v) return false;
          return v.split(MULTIPLE_INPUT_SEPARATOR).every((s) => !!s.trim());
        },
        { error: t("workflow_node.deploy.form.baotawaf_site_names.placeholder") }
      ),
    sitePort: z.coerce.number().refine((v) => isPortNumber(v), t("common.errmsg.port_invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBaotaWAF, {
  getInitialValues,
  getSchema,
});

export default _default;
