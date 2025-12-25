import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AccessProviderSelect from "@/components/provider/AccessProviderSelect";
import { type AccessModel } from "@/domain/access";
import { ACCESS_PROVIDERS, ACCESS_USAGES } from "@/domain/provider";
import { useAntdForm } from "@/hooks";

import { FormNestedFieldsContextProvider } from "./forms/_context";
import { useProviderFilterByUsage } from "./forms/_hooks";
import AccessConfigFieldsProvider from "./forms/AccessConfigFieldsProvider";

export type AccessFormModes = "create" | "modify";
export type AccessFormUsages = "dns" | "hosting" | "dns-hosting" | "ca" | "notification";

export interface AccessFormProps {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: Nullish<MaybeModelRecord<AccessModel>>;
  form: FormInstance;
  mode: AccessFormModes;
  usage?: AccessFormUsages;
  onFormValuesChange?: (changedValues: Nullish<MaybeModelRecord<AccessModel>>, values: Nullish<MaybeModelRecord<AccessModel>>) => void;
}

const AccessForm = ({ className, style, disabled, initialValues, mode, usage, onFormValuesChange, ...props }: AccessFormProps) => {
  const { t } = useTranslation();

  const providerFilter = useProviderFilterByUsage(usage);

  const formSchema = z.object({
    name: z
      .string(t("access.form.name.placeholder"))
      .min(1, t("access.form.name.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    provider: z.enum(ACCESS_PROVIDERS, t("access.form.provider.placeholder")),
    config: z.any(),
    reserve: z.string().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "accessForm",
    initialValues: initialValues,
  });

  const fieldProvider = Form.useWatch("provider", { form: formInst, preserve: true });

  const renderNestedFieldProviderComponent = AccessConfigFieldsProvider.useComponent(fieldProvider, {
    initProps: (provider) => {
      let props: object = { disabled: disabled };

      switch (provider) {
        case ACCESS_PROVIDERS.WEBHOOK:
          {
            props = {
              ...props,
              usage: usage === "notification" ? "notification" : usage === "hosting" || usage === "dns-hosting" ? "deployment" : "none",
            };
          }
          break;
      }

      return props;
    },
    deps: [disabled, usage],
  });

  return (
    <Form
      className={className}
      style={style}
      {...formProps}
      clearOnDestroy={true}
      disabled={disabled}
      form={formInst}
      layout="vertical"
      preserve={false}
      scrollToFirstError
      onValuesChange={onFormValuesChange}
    >
      <Form.Item name="name" label={t("access.form.name.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.name.placeholder")} />
      </Form.Item>

      <Form.Item
        name="provider"
        label={t("access.form.provider.label")}
        extra={usage === "dns-hosting" ? <span dangerouslySetInnerHTML={{ __html: t("access.form.provider.help") }}></span> : null}
        rules={[formRule]}
      >
        <AccessProviderSelect
          disabled={mode !== "create"}
          placeholder={t("access.form.provider.placeholder")}
          showOptionTags={
            usage == null || (usage === "dns-hosting" ? { ["builtin"]: true, [ACCESS_USAGES.DNS]: true, [ACCESS_USAGES.HOSTING]: true } : { ["builtin"]: true })
          }
          showSearch={!disabled}
          onFilter={providerFilter}
        />
      </Form.Item>

      <FormNestedFieldsContextProvider value={{ parentNamePath: "config" }}>
        {renderNestedFieldProviderComponent && <>{renderNestedFieldProviderComponent}</>}
      </FormNestedFieldsContextProvider>
    </Form>
  );
};

const _default = Object.assign(AccessForm, {
  useProviderFilterByUsage,
});

export default _default;
