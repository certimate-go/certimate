import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderAWSIAM = () => {
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
        label={t("workflow_node.deploy.form.aws_iam_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_iam_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_iam_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificatePath"]}
        initialValue={initialValues.certificatePath}
        label={t("workflow_node.deploy.form.aws_iam_certificate_path.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_iam_certificate_path.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.aws_iam_certificate_path.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    certificatePath: "/",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    region: z.string().nonempty(),
    certificatePath: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return v.startsWith("/") && v.endsWith("/");
      }, t("workflow_node.deploy.form.aws_iam_certificate_path.errmsg.invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAWSIAM, {
  getInitialValues,
  getSchema,
});

export default _default;
