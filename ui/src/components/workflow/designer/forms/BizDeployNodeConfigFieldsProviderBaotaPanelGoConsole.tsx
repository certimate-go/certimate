import { getI18n } from "react-i18next";
// import { getI18n, useTranslation } from "react-i18next";
// import { Form, Switch } from "antd";
// import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

// import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderBaotaPanelConsoleGo = () => {
  // const { i18n, t } = useTranslation();

  // const { parentNamePath } = useFormNestedFieldsContext();
  // const formSchema = z.object({
  //   [parentNamePath]: getSchema({ i18n }),
  // });
  // const formRule = createSchemaFieldRule(formSchema);
  // const initialValues = getInitialValues();

  return <></>;
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {};
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({});
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBaotaPanelConsoleGo, {
  getInitialValues,
  getSchema,
});

export default _default;
