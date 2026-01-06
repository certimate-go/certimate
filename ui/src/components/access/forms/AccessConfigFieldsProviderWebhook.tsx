import { getI18n, useTranslation } from "react-i18next";
import { IconChevronDown } from "@tabler/icons-react";
import { Button, Dropdown, Form, Input, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import CodeTextInput from "@/components/CodeTextInput";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { isJsonObject } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

export interface AccessConfigFormFieldsWebhookProps {
  usage?: "deployment" | "notification" | "none";
}

const AccessConfigFormFieldsProviderWebhook = ({ usage = "none" }: AccessConfigFormFieldsWebhookProps) => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues({ usage });

  const handleWebhookHeadersBlur = () => {
    let value = formInst.getFieldValue([parentNamePath, "headers"]);
    value = value.trim();
    value = value.replace(/(?<!\r)\n/g, "\r\n");
    formInst.setFieldValue([parentNamePath, "headers"], value);
  };

  const handleWebhookDataForDeploymentBlur = () => {
    const value = formInst.getFieldValue([parentNamePath, "data"]);
    try {
      const json = JSON.stringify(JSON.parse(value), null, 2);
      formInst.setFieldValue([parentNamePath, "data"], json);
    } catch {
      return;
    }
  };

  const handleWebhookDataForNotificationBlur = () => {
    const value = formInst.getFieldValue([parentNamePath, "data"]);
    try {
      const json = JSON.stringify(JSON.parse(value), null, 2);
      formInst.setFieldValue([parentNamePath, "data"], json);
    } catch {
      return;
    }
  };

  const handlePresetDataForDeploymentClick = () => {
    formInst.setFieldValue([parentNamePath, "method"], "POST");
    formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
    formInst.setFieldValue([parentNamePath, "data"], getInitialValues({ usage: "deployment" }).data);
  };

  const handlePresetDataForNotificationClick = (key: string) => {
    switch (key) {
      case "bark":
        formInst.setFieldValue([parentNamePath, "url"], "https://api.day.app/push");
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
        formInst.setFieldValue(
          [parentNamePath, "data"],
          JSON.stringify(
            {
              title: "${CERTIMATE_NOTIFIER_SUBJECT}",
              body: "${CERTIMATE_NOTIFIER_MESSAGE}",
              device_key: "<your-bark-device-key>",
            },
            null,
            2
          )
        );
        break;

      case "gotify":
        formInst.setFieldValue([parentNamePath, "url"], "https://<your-gotify-server>/");
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json\r\nAuthorization: Bearer <your-gotify-token>");
        formInst.setFieldValue(
          [parentNamePath, "data"],
          JSON.stringify(
            {
              title: "${CERTIMATE_NOTIFIER_SUBJECT}",
              message: "${CERTIMATE_NOTIFIER_MESSAGE}",
              priority: 1,
            },
            null,
            2
          )
        );
        break;

      case "ntfy":
        formInst.setFieldValue([parentNamePath, "url"], "https://<your-ntfy-server>/");
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
        formInst.setFieldValue(
          [parentNamePath, "data"],
          JSON.stringify(
            {
              topic: "<your-ntfy-topic>",
              title: "${CERTIMATE_NOTIFIER_SUBJECT}",
              message: "${CERTIMATE_NOTIFIER_MESSAGE}",
              priority: 1,
            },
            null,
            2
          )
        );
        break;

      case "pushover":
        formInst.setFieldValue([parentNamePath, "url"], "https://api.pushover.net/1/messages.json");
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
        formInst.setFieldValue(
          [parentNamePath, "data"],
          JSON.stringify(
            {
              token: "<your-pushover-token>",
              user: "<your-pushover-user>",
              title: "${CERTIMATE_NOTIFIER_SUBJECT}",
              message: "${CERTIMATE_NOTIFIER_MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      case "pushplus":
        formInst.setFieldValue([parentNamePath, "url"], "https://www.pushplus.plus/send");
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
        formInst.setFieldValue(
          [parentNamePath, "data"],
          JSON.stringify(
            {
              token: "<your-pushplus-token>",
              title: "${CERTIMATE_NOTIFIER_SUBJECT}",
              content: "${CERTIMATE_NOTIFIER_MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      case "serverchan3":
        formInst.setFieldValue([parentNamePath, "url"], "https://<your-serverchan-uid>.push.ft07.com/send/<your-serverchan-sendkey>.send");
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
        formInst.setFieldValue(
          [parentNamePath, "data"],
          JSON.stringify(
            {
              title: "${CERTIMATE_NOTIFIER_SUBJECT}",
              desp: "${CERTIMATE_NOTIFIER_MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      case "serverchanturbo":
        formInst.setFieldValue([parentNamePath, "url"], "https://sctapi.ftqq.com/<your-serverchan-key>.send");
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
        formInst.setFieldValue(
          [parentNamePath, "data"],
          JSON.stringify(
            {
              title: "${CERTIMATE_NOTIFIER_SUBJECT}",
              desp: "${CERTIMATE_NOTIFIER_MESSAGE}",
            },
            null,
            2
          )
        );
        break;

      default:
        formInst.setFieldValue([parentNamePath, "method"], "POST");
        formInst.setFieldValue([parentNamePath, "headers"], "Content-Type: application/json");
        formInst.setFieldValue([parentNamePath, "data"], getInitialValues({ usage: "notification" }).data);
        break;
    }
  };

  return (
    <>
      <Form.Item name={[parentNamePath, "url"]} initialValue={initialValues.url} label={t("access.form.webhook_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.webhook_url.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "method"]} initialValue={initialValues.method} label={t("access.form.webhook_method.label")} rules={[formRule]}>
        <Select
          options={["GET", "POST", "PUT", "PATCH", "DELETE"].map((s) => ({ label: s, value: s }))}
          placeholder={t("access.form.webhook_method.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "headers"]}
        initialValue={initialValues.headers}
        label={t("access.form.webhook_headers.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.webhook_headers.tooltip") }}></span>}
      >
        <CodeTextInput
          lineWrapping={false}
          height="auto"
          minHeight="64px"
          maxHeight="256px"
          placeholder={t("access.form.webhook_headers.placeholder")}
          onBlur={handleWebhookHeadersBlur}
        />
      </Form.Item>

      <Show when={usage === "deployment"}>
        <Form.Item className="relative" label={t("access.form.webhook_data.label")} extra={t("access.form.webhook_data.help")}>
          <div className="absolute -top-1.5 right-0 -translate-y-full">
            <Dropdown
              menu={{
                items: [
                  {
                    key: "common",
                    label: t("access.form.webhook_preset_data.common"),
                    onClick: handlePresetDataForDeploymentClick,
                  },
                ],
              }}
              trigger={["click"]}
            >
              <Button size="small" type="link">
                {t("access.form.webhook_preset_data")}
                <IconChevronDown size="1.25em" />
              </Button>
            </Dropdown>
          </div>
          <Form.Item name={[parentNamePath, "data"]} initialValue={initialValues.data} noStyle rules={[formRule]}>
            <CodeTextInput
              lineWrapping={false}
              height="auto"
              minHeight="64px"
              maxHeight="256px"
              language="json"
              placeholder={t("access.form.webhook_data.placeholder")}
              onBlur={handleWebhookDataForDeploymentBlur}
            />
          </Form.Item>
        </Form.Item>

        <Form.Item>
          <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.webhook_data.guide_for_deployment") }}></span>} />
        </Form.Item>
      </Show>

      <Show when={usage === "notification"}>
        <Form.Item className="relative" label={t("access.form.webhook_data.label")} extra={t("access.form.webhook_data.help")}>
          <div className="absolute -top-1.5 right-0 -translate-y-full">
            <Dropdown
              menu={{
                items: ["bark", "ntfy", "gotify", "pushover", "pushplus", "serverchan3", "serverchanturbo", "common"].map((key) => ({
                  key,
                  label: <span dangerouslySetInnerHTML={{ __html: t(`access.form.webhook_preset_data.${key}`) }}></span>,
                  onClick: () => handlePresetDataForNotificationClick(key),
                })),
              }}
              trigger={["click"]}
            >
              <Button size="small" type="link">
                {t("access.form.webhook_preset_data")}
                <IconChevronDown size="1.25em" />
              </Button>
            </Dropdown>
          </div>
          <Form.Item name={[parentNamePath, "data"]} initialValue={initialValues.data} noStyle rules={[formRule]}>
            <CodeTextInput
              lineWrapping={false}
              height="auto"
              minHeight="64px"
              maxHeight="256px"
              language="json"
              placeholder={t("access.form.webhook_data.placeholder")}
              onBlur={handleWebhookDataForNotificationBlur}
            />
          </Form.Item>
        </Form.Item>

        <Form.Item>
          <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.webhook_data.guide_for_notification") }}></span>} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "allowInsecureConnections"]}
        initialValue={initialValues.allowInsecureConnections}
        label={t("access.form.shared_allow_insecure_conns.label")}
        rules={[formRule]}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = ({ usage = "none" }: { usage?: "deployment" | "notification" | "none" }): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    url: "",
    method: "POST",
    headers: "Content-Type: application/json",
    allowInsecureConnections: false,
    data: JSON.stringify(
      usage === "deployment"
        ? {
            name: "${CERTIMATE_DEPLOYER_COMMONNAME}",
            cert: "${CERTIMATE_DEPLOYER_CERTIFICATE}",
            privkey: "${CERTIMATE_DEPLOYER_PRIVATEKEY}",
          }
        : usage === "notification"
          ? {
              subject: "${CERTIMATE_NOTIFIER_SUBJECT}",
              message: "${CERTIMATE_NOTIFIER_MESSAGE}",
            }
          : {},
      null,
      2
    ),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
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
    data: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return isJsonObject(v);
      }, t("access.form.webhook_data.errmsg.json_invalid")),
    allowInsecureConnections: z.boolean().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderWebhook, {
  getInitialValues,
  getSchema,
});

export default _default;
