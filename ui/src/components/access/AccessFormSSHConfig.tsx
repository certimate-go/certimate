import { useTranslation } from "react-i18next";
import { ArrowDownOutlined, ArrowUpOutlined, CloseOutlined, PlusOutlined } from "@ant-design/icons";
import { Button, Collapse, Form, type FormInstance, Input, InputNumber, Select, Space } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";
import TextFileInput from "@/components/TextFileInput";
import { type AccessConfigForSSH } from "@/domain/access";
import { validDomainName, validIPv4Address, validIPv6Address, validPortNumber } from "@/utils/validators";

type AccessFormSSHConfigFieldValues = Nullish<AccessConfigForSSH>;

export type AccessFormSSHConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormSSHConfigFieldValues;
  onValuesChange?: (values: AccessFormSSHConfigFieldValues) => void;
};

const AUTH_METHOD_NONE = "none" as const;
const AUTH_METHOD_PASSWORD = "password" as const;
const AUTH_METHOD_KEY = "key" as const;

const initFormModel = (): AccessFormSSHConfigFieldValues => {
  return {
    host: "127.0.0.1",
    port: 22,
    authMethod: AUTH_METHOD_PASSWORD,
    username: "root",
  };
};

const AccessFormSSHConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormSSHConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    host: z.string().refine((v) => validDomainName(v) || validIPv4Address(v) || validIPv6Address(v), t("common.errmsg.host_invalid")),
    port: z.preprocess(
      (v) => Number(v),
      z
        .number()
        .int(t("access.form.ssh_port.placeholder"))
        .refine((v) => validPortNumber(v), t("common.errmsg.port_invalid"))
    ),
    authMethod: z.literal([AUTH_METHOD_NONE, AUTH_METHOD_PASSWORD, AUTH_METHOD_KEY], t("access.form.ssh_auth_method.placeholder")),
    username: z
      .string()
      .min(1, t("access.form.ssh_username.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    password: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine((v) => {
        if (fieldAuthMethod !== AUTH_METHOD_PASSWORD) return true;
        return !!v?.trim();
      }, t("access.form.ssh_password.placeholder")),
    key: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
      .nullish()
      .refine((v) => {
        if (fieldAuthMethod !== AUTH_METHOD_KEY) return true;
        return !!v?.trim();
      }, t("access.form.ssh_key.placeholder")),
    keyPassphrase: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
      .nullish()
      .refine((v) => !v || formInst.getFieldValue("key"), t("access.form.ssh_key.placeholder")),
    jumpServers: z
      .array(
        z.object({
          host: z.string().refine((v) => validDomainName(v) || validIPv4Address(v) || validIPv6Address(v), t("common.errmsg.host_invalid")),
          port: z.preprocess(
            (v) => Number(v),
            z
              .number()
              .int(t("access.form.ssh_port.placeholder"))
              .refine((v) => validPortNumber(v), t("common.errmsg.port_invalid"))
          ),
          authMethod: z.literal([AUTH_METHOD_NONE, AUTH_METHOD_PASSWORD, AUTH_METHOD_KEY], t("access.form.ssh_auth_method.placeholder")),
          username: z
            .string()
            .min(1, t("access.form.ssh_username.placeholder"))
            .max(64, t("common.errmsg.string_max", { max: 64 })),
          password: z
            .string()
            .max(64, t("common.errmsg.string_max", { max: 64 }))
            .nullish(),
          key: z
            .string()
            .max(20480, t("common.errmsg.string_max", { max: 20480 }))
            .nullish(),
          keyPassphrase: z
            .string()
            .max(20480, t("common.errmsg.string_max", { max: 20480 }))
            .nullish(),
        }),
        { error: t("access.form.ssh_jump_servers.errmsg.invalid") }
      )
      .nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const fieldAuthMethod = Form.useWatch("authMethod", formInst);

  const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
    onValuesChange?.(values);
  };

  return (
    <Form
      form={formInst}
      disabled={disabled}
      initialValues={initialValues ?? initFormModel()}
      layout="vertical"
      name={formName}
      onValuesChange={handleFormChange}
    >
      <div className="flex space-x-2">
        <div className="w-2/3">
          <Form.Item name="host" label={t("access.form.ssh_host.label")} rules={[formRule]}>
            <Input placeholder={t("access.form.ssh_host.placeholder")} />
          </Form.Item>
        </div>

        <div className="w-1/3">
          <Form.Item name="port" label={t("access.form.ssh_port.label")} rules={[formRule]}>
            <InputNumber className="w-full" min={1} max={65535} placeholder={t("access.form.ssh_port.placeholder")} />
          </Form.Item>
        </div>
      </div>

      <Form.Item name="authMethod" label={t("access.form.ssh_auth_method.label")} rules={[formRule]}>
        <Select placeholder={t("access.form.ssh_auth_method.placeholder")}>
          <Select.Option key={AUTH_METHOD_NONE} value={AUTH_METHOD_NONE}>
            {t("access.form.ssh_auth_method.option.none.label")}
          </Select.Option>
          <Select.Option key={AUTH_METHOD_PASSWORD} value={AUTH_METHOD_PASSWORD}>
            {t("access.form.ssh_auth_method.option.password.label")}
          </Select.Option>
          <Select.Option key={AUTH_METHOD_KEY} value={AUTH_METHOD_KEY}>
            {t("access.form.ssh_auth_method.option.key.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item name="username" label={t("access.form.ssh_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.ssh_username.placeholder")} />
      </Form.Item>

      <Show when={fieldAuthMethod === AUTH_METHOD_PASSWORD}>
        <Form.Item name="password" label={t("access.form.ssh_password.label")} rules={[formRule]}>
          <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.ssh_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldAuthMethod === AUTH_METHOD_KEY}>
        <Form.Item name="key" label={t("access.form.ssh_key.label")} rules={[formRule]}>
          <TextFileInput allowClear autoSize={{ minRows: 1, maxRows: 5 }} placeholder={t("access.form.ssh_key.placeholder")} />
        </Form.Item>

        <Form.Item name="keyPassphrase" label={t("access.form.ssh_key_passphrase.label")} rules={[formRule]}>
          <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.ssh_key_passphrase.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item name="jumpServers" label={t("access.form.ssh_jump_servers.label")} rules={[formRule]}>
        <Form.List name="jumpServers">
          {(fields, { add, remove, move }) => (
            <Space className="w-full" direction="vertical" size="small">
              {fields?.length > 0 ? (
                <Collapse
                  items={fields.map((field, index) => {
                    const Label = () => {
                      const host = Form.useWatch(["jumpServers", field.name, "host"], formInst);
                      const port = Form.useWatch(["jumpServers", field.name, "port"], formInst);
                      const addr = !!host && !!port ? `${host}:${port}` : host ? host : port ? `:${port}` : "unknown";
                      return (
                        <span className="select-none">
                          [{t("access.form.ssh_jump_servers.item.label")} {field.name + 1}] {addr}
                        </span>
                      );
                    };

                    const Fields = () => {
                      const authMethod = Form.useWatch(["jumpServers", field.name, "authMethod"], formInst);
                      return (
                        <>
                          <div className="flex space-x-2">
                            <div className="w-2/3">
                              <Form.Item name={[field.name, "host"]} label={t("access.form.ssh_host.label")} rules={[formRule]}>
                                <Input placeholder={t("access.form.ssh_host.placeholder")} />
                              </Form.Item>
                            </div>
                            <div className="w-1/3">
                              <Form.Item name={[field.name, "port"]} label={t("access.form.ssh_port.label")} rules={[formRule]}>
                                <InputNumber className="w-full" placeholder={t("access.form.ssh_port.placeholder")} min={1} max={65535} />
                              </Form.Item>
                            </div>
                          </div>

                          <Form.Item name={[field.name, "authMethod"]} label={t("access.form.ssh_auth_method.label")} rules={[formRule]}>
                            <Select placeholder={t("access.form.ssh_auth_method.placeholder")}>
                              <Select.Option key={AUTH_METHOD_NONE} value={AUTH_METHOD_NONE}>
                                {t("access.form.ssh_auth_method.option.none.label")}
                              </Select.Option>
                              <Select.Option key={AUTH_METHOD_PASSWORD} value={AUTH_METHOD_PASSWORD}>
                                {t("access.form.ssh_auth_method.option.password.label")}
                              </Select.Option>
                              <Select.Option key={AUTH_METHOD_KEY} value={AUTH_METHOD_KEY}>
                                {t("access.form.ssh_auth_method.option.key.label")}
                              </Select.Option>
                            </Select>
                          </Form.Item>

                          <Form.Item name={[field.name, "username"]} label={t("access.form.ssh_username.label")} rules={[formRule]}>
                            <Input autoComplete="new-password" placeholder={t("access.form.ssh_username.placeholder")} />
                          </Form.Item>

                          <Show when={authMethod === AUTH_METHOD_PASSWORD}>
                            <Form.Item name={[field.name, "password"]} label={t("access.form.ssh_password.label")} rules={[formRule]}>
                              <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.ssh_password.placeholder")} />
                            </Form.Item>
                          </Show>

                          <Show when={authMethod === AUTH_METHOD_KEY}>
                            <Form.Item name={[field.name, "key"]} label={t("access.form.ssh_key.label")} rules={[formRule]}>
                              <TextFileInput allowClear autoSize={{ minRows: 1, maxRows: 5 }} placeholder={t("access.form.ssh_key.placeholder")} />
                            </Form.Item>

                            <Form.Item name={[field.name, "keyPassphrase"]} label={t("access.form.ssh_key_passphrase.label")} rules={[formRule]}>
                              <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.ssh_key_passphrase.placeholder")} />
                            </Form.Item>
                          </Show>
                        </>
                      );
                    };

                    return {
                      key: field.key,
                      label: <Label />,
                      extra: (
                        <Space.Compact>
                          <Button
                            icon={<ArrowUpOutlined />}
                            color="default"
                            disabled={disabled || index === 0}
                            size="small"
                            type="text"
                            onClick={(e) => {
                              move(index, index - 1);
                              e.stopPropagation();
                            }}
                          />
                          <Button
                            icon={<ArrowDownOutlined />}
                            color="default"
                            disabled={disabled || index === fields.length - 1}
                            size="small"
                            type="text"
                            onClick={(e) => {
                              move(index, index + 1);
                              e.stopPropagation();
                            }}
                          />
                          <Button
                            icon={<CloseOutlined />}
                            color="default"
                            disabled={disabled}
                            size="small"
                            type="text"
                            onClick={(e) => {
                              remove(field.name);
                              e.stopPropagation();
                            }}
                          />
                        </Space.Compact>
                      ),
                      children: <Fields />,
                    };
                  })}
                />
              ) : null}
              <Button className="w-full" type="dashed" icon={<PlusOutlined />} onClick={() => add(initFormModel())}>
                {t("access.form.ssh_jump_servers.add")}
              </Button>
            </Space>
          )}
        </Form.List>
      </Form.Item>
    </Form>
  );
};

export default AccessFormSSHConfig;
