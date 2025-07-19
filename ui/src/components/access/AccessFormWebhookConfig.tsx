import { useTranslation } from "react-i18next";
import { IconChevronDown } from "@tabler/icons-react";
import { Alert, Button, Dropdown, Form, type FormInstance, Input, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import CodeInput from "@/components/CodeInput";
import Show from "@/components/Show";
import { type AccessConfigForWebhook } from "@/domain/access";

type AccessFormWebhookConfigFieldValues = Nullish<AccessConfigForWebhook>;

export type AccessFormWebhookConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormWebhookConfigFieldValues;
  usage?: "deployment" | "notification" | "none";
  onValuesChange?: (values: AccessFormWebhookConfigFieldValues) => void;
};

const initFormModel = (): AccessFormWebhookConfigFieldValues => {
  return {
    url: "",
    method: "POST",
    headers: "Content-Type: application/json",
    allowInsecureConnections: false,
    defaultDataForDeployment: JSON.stringify(
      {
        name: "${DOMAINS}",
        cert: "${CERTIFICATE}",
        privkey: "${PRIVATE_KEY}",
      },
      null,
      2
    ),
    defaultDataForNotification: JSON.stringify(
      {
        subject: "${SUBJECT}",
        message: "${MESSAGE}",
      },
      null,
      2
    ),
  };
};

const AccessFormWebhookConfig = ({ form: formInst, formName, disabled, initialValues, usage, onValuesChange }: AccessFormWebhookConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    url: z.url(t("common.errmsg.url_invalid")),
    method: z.literal(["GET", "POST", "PUT", "PATCH", "DELETE"], t("access.form.webhook_method.placeholder")),
    headers: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;

        const lines = v.split(/\r?\n/);
        for (const line of lines) {
          if (line.split(":").length < 2) {
            return false;
          }
        }
        return true;
      }, t("access.form.webhook_headers.errmsg.invalid")),
    defaultDataForDeployment: z
      .string()
      .nullish()
      .refine((v) => {
        if (usage && usage !== "deployment") return true;
        if (!v) return true;

        try {
          const obj = JSON.parse(v);
          return typeof obj === "object" && !Array.isArray(obj);
        } catch {
          return false;
        }
      }, t("access.form.webhook_default_data.errmsg.json_invalid")),
    defaultDataForNotification: z
      .string()
      .nullish()
      .refine((v) => {
        if (usage && usage !== "notification") return true;
        if (!v) return true;

        try {
          const obj = JSON.parse(v);
          return typeof obj === "object" && !Array.isArray(obj);
        } catch {
          return false;
        }
      }, t("access.form.webhook_default_data.errmsg.json_invalid")),
    allowInsecureConnections: z.boolean().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const handleWebhookHeadersBlur = () => {
    let value = formInst.getFieldValue("headers");
    value = value.trim();
    value = value.replace(/(?<!\r)\n/g, "\r\n");
    formInst.setFieldValue("headers", value);
  };

  const handleWebhookDataForDeploymentBlur = () => {
    const value = formInst.getFieldValue("defaultDataForDeployment");
    try {
      const json = JSON.stringify(JSON.parse(value), null, 2);
      formInst.setFieldValue("defaultDataForDeployment", json);
    } catch {
      return;
    }
  };

  const handleWebhookDataForNotificationBlur = () => {
    const value = formInst.getFieldValue("defaultDataForNotification");
    try {
      const json = JSON.stringify(JSON.parse(value), null, 2);
      formInst.setFieldValue("defaultDataForNotification", json);
    } catch {
      return;
    }
  };

  const handlePresetDataForDeploymentClick = () => {
    formInst.setFieldValue("defaultDataForDeployment", initFormModel().defaultDataForDeployment);
  };

  const handlePresetDataForNotificationClick = (key: string) => {
    switch (key) {
      case "bark":
        formInst.setFieldValue("url", "https://api.day.app/push");
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json\r\nAuthorization: Bearer <your-gotify-token>");
        formInst.setFieldValue(
          "defaultDataForNotification",
          JSON.stringify(
            {
              title: "${SUBJECT}",
              body: "${MESSAGE}",
              device_key: "<your-bark-device-key>",
            },
            null,
            2
          )
        );
        break;

      case "gotify":
        formInst.setFieldValue("url", "https://<your-gotify-server>/");
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json\r\nAuthorization: Bearer <your-gotify-token>");
        formInst.setFieldValue(
          "defaultDataForNotification",
          JSON.stringify(
            {
              title: "${SUBJECT}",
              message: "${MESSAGE}",
              priority: 1,
            },
            null,
            2
          )
        );
        break;

      case "ntfy":
        formInst.setFieldValue("url", "https://<your-ntfy-server>/");
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json");
        formInst.setFieldValue(
          "defaultDataForNotification",
          JSON.stringify(
            {
              topic: "<your-ntfy-topic>",
              title: "${SUBJECT}",
              message: "${MESSAGE}",
              priority: 1,
            },
            null,
            2
          )
        );
        break;

      case "pushover":
        formInst.setFieldValue("url", "https://api.pushover.net/1/messages.json");
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json");
        formInst.setFieldValue(
          "defaultDataForNotification",
          JSON.stringify(
            {
              token: "<your-pushover-token>",
              user: "<your-pushover-user>",
              title: "${SUBJECT}",
              message: "${MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      case "pushplus":
        formInst.setFieldValue("url", "https://www.pushplus.plus/send");
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json");
        formInst.setFieldValue(
          "defaultDataForNotification",
          JSON.stringify(
            {
              token: "<your-pushplus-token>",
              title: "${SUBJECT}",
              content: "${MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      case "serverchan3":
        formInst.setFieldValue("url", "https://<your-serverchan-uid>.push.ft07.com/send/<your-serverchan-sendkey>.send");
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json");
        formInst.setFieldValue(
          "defaultDataForNotification",
          JSON.stringify(
            {
              title: "${SUBJECT}",
              desp: "${MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      case "serverchanturbo":
        formInst.setFieldValue("url", "https://sctapi.ftqq.com/<your-serverchan-key>.send");
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json");
        formInst.setFieldValue(
          "defaultDataForNotification",
          JSON.stringify(
            {
              text: "${SUBJECT}",
              desp: "${MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      default:
        formInst.setFieldValue("method", "POST");
        formInst.setFieldValue("headers", "Content-Type: application/json");
        formInst.setFieldValue("defaultDataForNotification", initFormModel().defaultDataForNotification);
        break;
    }
  };

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
      <Form.Item name="url" label={t("access.form.webhook_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.webhook_url.placeholder")} />
      </Form.Item>

      <Form.Item name="method" label={t("access.form.webhook_method.label")} rules={[formRule]}>
        <Select
          options={["GET", "POST", "PUT", "PATCH", "DELETE"].map((s) => ({ label: s, value: s }))}
          placeholder={t("access.form.webhook_method.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name="headers"
        label={t("access.form.webhook_headers.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.webhook_headers.tooltip") }}></span>}
      >
        <CodeInput
          height="auto"
          minHeight="64px"
          maxHeight="256px"
          placeholder={t("access.form.webhook_headers.placeholder")}
          onBlur={handleWebhookHeadersBlur}
        />
      </Form.Item>

      <Show when={!usage || usage === "deployment"}>
        <Form.Item noStyle>
          <label className="mb-1 block">
            <div className="flex w-full items-center justify-between gap-4">
              <div className="max-w-full grow truncate">
                <span>{t("access.form.webhook_default_data_for_deployment.label")}</span>
              </div>
              <div className="text-right">
                <Button size="small" type="link" onClick={handlePresetDataForDeploymentClick}>
                  {t("access.form.webhook_preset_data.button")}
                </Button>
              </div>
            </div>
          </label>
          <Form.Item name="defaultDataForDeployment" rules={[formRule]}>
            <CodeInput
              height="auto"
              minHeight="64px"
              maxHeight="256px"
              language="json"
              placeholder={t("access.form.webhook_default_data_for_deployment.placeholder")}
              onBlur={handleWebhookDataForDeploymentBlur}
            />
          </Form.Item>
        </Form.Item>

        <Form.Item>
          <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("access.form.webhook_default_data_for_deployment.guide") }}></span>} />
        </Form.Item>
      </Show>

      <Show when={!usage || usage === "notification"}>
        <Form.Item noStyle>
          <label className="mb-1 block">
            <div className="flex w-full items-center justify-between gap-4">
              <div className="max-w-full grow truncate">
                <span>{t("access.form.webhook_default_data_for_notification.label")}</span>
              </div>
              <div className="text-right">
                <Dropdown
                  menu={{
                    items: ["bark", "ntfy", "gotify", "pushover", "pushplus", "serverchan3", "serverchanturbo", "common"].map((key) => ({
                      key,
                      label: <span dangerouslySetInnerHTML={{ __html: t(`access.form.webhook_preset_data.option.${key}.label`) }}></span>,
                      onClick: () => handlePresetDataForNotificationClick(key),
                    })),
                  }}
                  trigger={["click"]}
                >
                  <Button size="small" type="link">
                    {t("access.form.webhook_preset_data.button")}
                    <IconChevronDown size="1.25em" />
                  </Button>
                </Dropdown>
              </div>
            </div>
          </label>
          <Form.Item name="defaultDataForNotification" rules={[formRule]}>
            <CodeInput
              height="auto"
              minHeight="64px"
              maxHeight="256px"
              language="json"
              placeholder={t("access.form.webhook_default_data_for_notification.placeholder")}
              onBlur={handleWebhookDataForNotificationBlur}
            />
          </Form.Item>
        </Form.Item>

        <Form.Item>
          <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("access.form.webhook_default_data_for_notification.guide") }}></span>} />
        </Form.Item>
      </Show>

      <Form.Item name="allowInsecureConnections" label={t("access.form.common_allow_insecure_conns.label")} rules={[formRule]}>
        <Switch
          checkedChildren={t("access.form.common_allow_insecure_conns.switch.on")}
          unCheckedChildren={t("access.form.common_allow_insecure_conns.switch.off")}
        />
      </Form.Item>
    </Form>
  );
};

export default AccessFormWebhookConfig;
