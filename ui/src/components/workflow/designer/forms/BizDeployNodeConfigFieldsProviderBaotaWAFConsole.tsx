import type { getI18n } from "react-i18next";
import { useTranslation } from "react-i18next";
import { Form } from "antd";
import { z } from "zod";

import Tips from "@/components/Tips";

const BizDeployNodeConfigFieldsProviderBaotaWAFConsole = () => {
  const { t } = useTranslation();

  return (
    <>
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotawaf_console.guide") }}></span>} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {};
};

const getSchema = (_: { i18n?: ReturnType<typeof getI18n> }) => {
  return z.object({});
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBaotaWAFConsole, {
  getInitialValues,
  getSchema,
});

export default _default;
