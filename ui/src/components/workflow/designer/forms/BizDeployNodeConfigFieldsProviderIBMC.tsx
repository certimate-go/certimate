import { getI18n, useTranslation } from "react-i18next";
import { Form, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderIBMC = () => {
  const { i18n, t } = useTranslation();
  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({ [parentNamePath]: getSchema({ i18n }) });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <Form.Item
      name={[parentNamePath, "restartAfterImport"]}
      initialValue={initialValues.restartAfterImport}
      label={t("workflow_node.deploy.form.ibmc_restart_after_import.label")}
      rules={[formRule]}
    >
      <Switch />
    </Form.Item>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => ({ restartAfterImport: true });
const getSchema = ({ i18n: _i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => z.object({ restartAfterImport: z.boolean().nullish() });

export default Object.assign(BizDeployNodeConfigFieldsProviderIBMC, { getInitialValues, getSchema });
