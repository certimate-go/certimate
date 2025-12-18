import { getI18n, useTranslation } from "react-i18next";
import { IconCircleArrowDown, IconCircleArrowUp, IconCircleMinus, IconCirclePlus } from "@tabler/icons-react";
import { Button, Collapse, Form, Input, InputNumber, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import FileTextInput from "@/components/FileTextInput";
import { mergeCls } from "@/utils/css";
import { validDomainName, validIPv4Address, validIPv6Address, validPortNumber } from "@/utils/validators";

import { useFormNestedFieldsContext } from "./_context";

const AUTH_METHOD_NONE = "none" as const;
const AUTH_METHOD_PASSWORD = "password" as const;
const AUTH_METHOD_KEY = "key" as const;

const AccessConfigFormFieldsProviderSSH = ({ disabled }: { disabled?: boolean }) => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldAuthMethod = Form.useWatch([parentNamePath, "authMethod"], formInst);
  const fieldJumpServers = Form.useWatch([parentNamePath, "jumpServers"], formInst);

  return (
    <>
      <div className="flex space-x-2">
        <div className="w-2/3">
          <Form.Item name={[parentNamePath, "host"]} initialValue={initialValues.host} label={t("access.form.ssh_host.label")} rules={[formRule]}>
            <Input placeholder={t("access.form.ssh_host.placeholder")} />
          </Form.Item>
        </div>

        <div className="w-1/3">
          <Form.Item name={[parentNamePath, "port"]} initialValue={initialValues.port} label={t("access.form.ssh_port.label")} rules={[formRule]}>
            <InputNumber style={{ width: "100%" }} min={1} max={65535} placeholder={t("access.form.ssh_port.placeholder")} />
          </Form.Item>
        </div>
      </div>

      <Form.Item
        name={[parentNamePath, "authMethod"]}
        initialValue={initialValues.authMethod}
        label={t("access.form.ssh_auth_method.label")}
        rules={[formRule]}
      >
        <Radio.Group block>
          <Radio.Button value={AUTH_METHOD_NONE}>{t("access.form.ssh_auth_method.option.none.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_PASSWORD}>{t("access.form.ssh_auth_method.option.password.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_KEY}>{t("access.form.ssh_auth_method.option.key.label")}</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Form.Item name={[parentNamePath, "username"]} initialValue={initialValues.username} label={t("access.form.ssh_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.ssh_username.placeholder")} />
      </Form.Item>

      <Show when={fieldAuthMethod === AUTH_METHOD_PASSWORD}>
        <Form.Item name={[parentNamePath, "password"]} initialValue={initialValues.password} label={t("access.form.ssh_password.label")} rules={[formRule]}>
          <Input.Password autoComplete="new-password" placeholder={t("access.form.ssh_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldAuthMethod === AUTH_METHOD_KEY}>
        <Form.Item name={[parentNamePath, "key"]} initialValue={initialValues.key} label={t("access.form.ssh_key.label")} rules={[formRule]}>
          <FileTextInput autoSize={{ minRows: 1, maxRows: 5 }} placeholder={t("access.form.ssh_key.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "keyPassphrase"]}
          initialValue={initialValues.keyPassphrase}
          label={t("access.form.ssh_key_passphrase.label")}
          rules={[formRule]}
        >
          <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.ssh_key_passphrase.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item label={t("access.form.ssh_jump_servers.label")}>
        <Form.List name={[parentNamePath, "jumpServers"]}>
          {(fields, { add, remove, move }) => (
            <div className="flex flex-col gap-2">
              <Collapse
                className={mergeCls({ hidden: !fields.length })}
                items={fields?.map(({ key, name: index }) => {
                  const subfieldHost = fieldJumpServers?.[index]?.host;
                  const subfieldPort = fieldJumpServers?.[index]?.post;
                  const subfieldAuthMethod = fieldJumpServers?.[index]?.authMethod;

                  const subfieldHostAndPort =
                    !!subfieldHost && !!subfieldPort
                      ? `${subfieldHost}:${subfieldPort}`
                      : subfieldHost
                        ? subfieldHost
                        : subfieldPort
                          ? `:${subfieldPort}`
                          : "unknown";

                  return {
                    key: key,
                    forceRender: true,
                    label: (
                      <span className="select-none">
                        [{t("access.form.ssh_jump_servers.item.label")} {index + 1}] {subfieldHostAndPort}
                      </span>
                    ),
                    extra: !disabled && (
                      <div className="flex items-center justify-end">
                        <Button
                          icon={<IconCircleArrowUp size="1.25em" />}
                          color="default"
                          disabled={index === 0}
                          size="small"
                          type="text"
                          onClick={(e) => {
                            move(index, index - 1);
                            e.stopPropagation();
                          }}
                        />
                        <Button
                          icon={<IconCircleArrowDown size="1.25em" />}
                          color="default"
                          disabled={index === fields.length - 1}
                          size="small"
                          type="text"
                          onClick={(e) => {
                            move(index, index + 1);
                            e.stopPropagation();
                          }}
                        />
                        <Button
                          icon={<IconCircleMinus size="1.25em" />}
                          color="default"
                          size="small"
                          type="text"
                          onClick={(e) => {
                            remove(index);
                            e.stopPropagation();
                          }}
                        />
                      </div>
                    ),
                    children: (
                      <>
                        <div className="flex space-x-2">
                          <div className="w-2/3">
                            <Form.Item name={[index, "host"]} label={t("access.form.ssh_host.label")} shouldUpdate rules={[formRule]}>
                              <Input placeholder={t("access.form.ssh_host.placeholder")} />
                            </Form.Item>
                          </div>
                          <div className="w-1/3">
                            <Form.Item name={[index, "port"]} label={t("access.form.ssh_port.label")} shouldUpdate rules={[formRule]}>
                              <InputNumber style={{ width: "100%" }} placeholder={t("access.form.ssh_port.placeholder")} min={1} max={65535} />
                            </Form.Item>
                          </div>
                        </div>

                        <Form.Item name={[index, "authMethod"]} label={t("access.form.ssh_auth_method.label")} shouldUpdate rules={[formRule]}>
                          <Radio.Group
                            options={[AUTH_METHOD_NONE, AUTH_METHOD_PASSWORD, AUTH_METHOD_KEY].map((s) => ({
                              key: s,
                              label: t(`access.form.ssh_auth_method.option.${s}.label`),
                              value: s,
                            }))}
                          />
                        </Form.Item>

                        <Form.Item name={[index, "username"]} label={t("access.form.ssh_username.label")} shouldUpdate rules={[formRule]}>
                          <Input autoComplete="new-password" placeholder={t("access.form.ssh_username.placeholder")} />
                        </Form.Item>

                        <Show when={subfieldAuthMethod === AUTH_METHOD_PASSWORD}>
                          <Form.Item name={[index, "password"]} label={t("access.form.ssh_password.label")} shouldUpdate rules={[formRule]}>
                            <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.ssh_password.placeholder")} />
                          </Form.Item>
                        </Show>

                        <Show when={subfieldAuthMethod === AUTH_METHOD_KEY}>
                          <Form.Item name={[index, "key"]} label={t("access.form.ssh_key.label")} shouldUpdate rules={[formRule]}>
                            <FileTextInput allowClear autoSize={{ minRows: 1, maxRows: 5 }} placeholder={t("access.form.ssh_key.placeholder")} />
                          </Form.Item>

                          <Form.Item name={[index, "keyPassphrase"]} label={t("access.form.ssh_key_passphrase.label")} shouldUpdate rules={[formRule]}>
                            <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.ssh_key_passphrase.placeholder")} />
                          </Form.Item>
                        </Show>
                      </>
                    ),
                  };
                })}
              />
              <Button
                className="w-full"
                type="dashed"
                icon={<IconCirclePlus size="1.25em" />}
                onClick={() => {
                  add();
                  setTimeout(() => {
                    const jumpServer = getInitialValues();
                    delete jumpServer.jumpServers;
                    formInst.setFieldValue([parentNamePath, "jumpServers", (fieldJumpServers?.length ?? 0) + 1 - 1], jumpServer);
                  }, 0);
                }}
              >
                {t("access.form.ssh_jump_servers.add")}
              </Button>
            </div>
          )}
        </Form.List>
        <Form.Item name={[parentNamePath, "jumpServers"]} noStyle rules={[formRule]} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    host: "127.0.0.1",
    port: 22,
    authMethod: AUTH_METHOD_PASSWORD,
    username: "root",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  const baseSchema = z
    .object({
      host: z.string().refine((v) => validDomainName(v) || validIPv4Address(v) || validIPv6Address(v), t("common.errmsg.host_invalid")),
      port: z.coerce
        .number()
        .int(t("access.form.ssh_port.placeholder"))
        .refine((v) => validPortNumber(v), t("common.errmsg.port_invalid")),
      authMethod: z.literal([AUTH_METHOD_NONE, AUTH_METHOD_PASSWORD, AUTH_METHOD_KEY], t("access.form.ssh_auth_method.placeholder")),
      username: z.string().nonempty(t("access.form.ssh_username.placeholder")),
      password: z.string().nullish(),
      key: z
        .string()
        .max(20480, t("common.errmsg.string_max", { max: 20480 }))
        .nullish(),
      keyPassphrase: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.authMethod) {
        case AUTH_METHOD_PASSWORD:
          {
            if (!values.password?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("access.form.ssh_password.placeholder"),
                path: ["password"],
              });
            }
          }
          break;

        case AUTH_METHOD_KEY:
          {
            if (!values.key?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("access.form.ssh_key.placeholder"),
                path: ["key"],
              });
            }
          }
          break;
      }
    });

  return baseSchema.safeExtend({
    jumpServers: z
      .array(baseSchema, t("access.form.ssh_jump_servers.errmsg.invalid"))
      .nullish()
      .refine((v) => {
        if (v == null) return true;
        return v.every((item) => baseSchema.safeParse(item).success);
      }, t("access.form.ssh_jump_servers.errmsg.invalid")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderSSH, {
  getInitialValues,
  getSchema,
});

export default _default;
