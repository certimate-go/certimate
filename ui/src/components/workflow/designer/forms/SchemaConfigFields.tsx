import { type ReactNode, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Form, Input, Radio, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import type { ProviderSchemaEnvelope, SchemaColumn } from "@/api/providerschema";
import CodeTextInput from "@/components/CodeTextInput";

import { useFormNestedFieldsContext } from "./_context";
import { deriveZodSchema } from "./schemaValidation";
import { isColumnVisible, unitFieldProps } from "./schemaValueTypes";

export interface SchemaConfigFieldsProps {
  envelope: ProviderSchemaEnvelope;
}

const SchemaConfigFields = ({ envelope }: SchemaConfigFieldsProps) => {
  const { t } = useTranslation();
  const { parentNamePath } = useFormNestedFieldsContext();
  const formInst = Form.useFormInstance();

  const providerConfig = Form.useWatch<Record<string, unknown> | undefined>(parentNamePath, { form: formInst, preserve: true });
  const getValue = (field: string) => providerConfig?.[field];

  const formSchema = useMemo(() => z.object({ [parentNamePath]: deriveZodSchema(envelope) }), [envelope, parentNamePath]);
  const formRule = createSchemaFieldRule(formSchema);

  const columns = envelope.schema.columns.filter((col) => isColumnVisible(col, getValue));

  return (
    <>
      {columns.map((col) => (
        <Form.Item
          key={col.name}
          name={[parentNamePath, col.name]}
          initialValue={col.default}
          label={col.labelKey ? t(col.labelKey) : col.name}
          tooltip={renderTooltip(col, t)}
          extra={renderExtra(col, t)}
          rules={[formRule as never]}
        >
          {renderControl(col, t)}
        </Form.Item>
      ))}
    </>
  );
};

const renderTooltip = (col: SchemaColumn, t: (key: string) => string): ReactNode => {
  if (!col.tooltipKey) return undefined;
  return col.tooltipHtml ? <span dangerouslySetInnerHTML={{ __html: t(col.tooltipKey) }} /> : t(col.tooltipKey);
};

const renderExtra = (col: SchemaColumn, t: (key: string) => string): ReactNode => {
  if (!col.extraKey) return undefined;
  return col.extraHtml ? <span dangerouslySetInnerHTML={{ __html: t(col.extraKey) }} /> : t(col.extraKey);
};

const optionList = (col: SchemaColumn, t: (key: string) => string) =>
  (col.options ?? []).map((o) => ({ label: o.labelKey ? t(o.labelKey) : o.value, value: o.value }));

const renderControl = (col: SchemaColumn, t: (key: string) => string): ReactNode => {
  const placeholder = col.placeholderKey ? t(col.placeholderKey) : undefined;
  switch (col.valueType) {
    case "switch":
      return <Switch />;
    case "number":
      return (
        <Input type="number" min={col.min} max={col.max} placeholder={placeholder} style={{ width: "100%" }} {...unitFieldProps("text", col.unitKey, t)} />
      );
    case "select":
      return <Select placeholder={placeholder} options={optionList(col, t)} style={{ width: "100%" }} />;
    case "radio":
      return <Radio.Group options={optionList(col, t)} />;
    case "secret":
      return <Input.Password placeholder={placeholder} autoComplete="new-password" />;
    case "code":
      return (
        <CodeTextInput language={col.language ?? "json"} lineWrapping={false} height="auto" minHeight="64px" maxHeight="256px" placeholder={placeholder} />
      );
    case "textarea":
      return <Input.TextArea placeholder={placeholder} autoSize={{ minRows: 2, maxRows: 6 }} />;
    case "text":
    case "autocomplete":
    default:
      return <Input placeholder={placeholder} {...unitFieldProps("text", col.unitKey, t)} />;
  }
};

export default SchemaConfigFields;
