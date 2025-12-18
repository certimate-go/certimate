import { getI18n, useTranslation } from "react-i18next";
import { Form } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import FileTextInput from "@/components/FileTextInput";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderKubernetes = () => {
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
        name={[parentNamePath, "kubeConfig"]}
        initialValue={initialValues.kubeConfig}
        label={t("access.form.k8s_kubeconfig.label")}
        extra={t("access.form.k8s_kubeconfig.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.k8s_kubeconfig.tooltip") }}></span>}
      >
        <FileTextInput allowClear autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("access.form.k8s_kubeconfig.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {};
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    kubeConfig: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
      .nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderKubernetes, {
  getInitialValues,
  getSchema,
});

export default _default;
