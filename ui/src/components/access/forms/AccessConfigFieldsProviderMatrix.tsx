import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { core, z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderMatrix = () => {
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
        name={[parentNamePath, "serverUrl"]}
        initialValue={initialValues.serverUrl}
        label={t("access.form.matrix_server_url.label")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.matrix_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "userId"]}
        initialValue={initialValues.userId}
        label={t("access.form.matrix_user_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.matrix_user_id.tooltip") }} />}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.matrix_user_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessToken"]}
        initialValue={initialValues.accessToken}
        label={t("access.form.matrix_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.matrix_access_token.tooltip") }} />}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.matrix_access_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "roomId"]}
        initialValue={initialValues.roomId}
        label={t("access.form.matrix_room_id.label")}
        extra={t("access.form.matrix_room_id.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.matrix_room_id.tooltip") }} />}
      >
        <Input allowClear placeholder={t("access.form.matrix_room_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    serverUrl: "https://matrix-client.matrix.org",
    userId: "",
    accessToken: "",
    roomId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    serverUrl: z.url({ protocol: core.regexes.httpProtocol }),
    userId: z.string().nonempty().startsWith("@"),
    accessToken: z.string().nonempty(),
    roomId: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderMatrix, {
  getInitialValues,
  getSchema,
});

export default _default;
