import { type ReactNode, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Form, Input, InputNumber, Radio, Select, Switch } from "antd";

import type { ProviderSchemaEnvelope, SchemaColumn } from "@/api/providerschema";
import CodeTextInput from "@/components/CodeTextInput";
import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";
import { deriveZodSchema, matchCondition } from "./schemaValidation";
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

  const formSchema = useMemo(() => deriveZodSchema(envelope), [envelope]);

  const columns = envelope.schema.columns.filter((col) => isColumnVisible(col, getValue));

  return (
    <>
      {envelope.schema.noticeKey && (
        <Form.Item>
          <Tips
            message={envelope.schema.noticeHtml ? <span dangerouslySetInnerHTML={{ __html: t(envelope.schema.noticeKey) }} /> : t(envelope.schema.noticeKey)}
          />
        </Form.Item>
      )}
      {columns.map((col) => (
        <Form.Item
          key={col.name}
          name={[parentNamePath, col.name]}
          initialValue={col.default}
          label={col.labelKey ? t(col.labelKey) : col.name}
          tooltip={renderTooltip(col, t)}
          extra={(col.extraWhen ?? []).every((condition) => matchCondition(condition, getValue(condition.field))) ? renderExtra(col, t) : undefined}
          rules={[
            {
              validator: () => {
                const result = formSchema.safeParse(formInst.getFieldValue(parentNamePath) ?? {});
                if (result.success) return Promise.resolve();
                const issue = result.error.issues.find((entry) => entry.path.length === 0 || entry.path[0] === col.name);
                return issue ? Promise.reject(new Error(issue.message)) : Promise.resolve();
              },
            },
          ]}
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
      return <InputNumber changeOnBlur={false} placeholder={placeholder} style={{ width: "100%" }} {...unitFieldProps("digit", col.unitKey, t)} />;
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
    case "multipleSplit":
      return (
        <MultipleSplitValueInput
          separator={col.separator ?? ";"}
          modalTitle={col.modalTitleKey ? t(col.modalTitleKey) : undefined}
          placeholderInModal={col.modalPlaceholderKey ? t(col.modalPlaceholderKey) : undefined}
          placeholder={placeholder}
          splitOptions={{ trimSpace: true, removeEmpty: true }}
        />
      );
    case "textarea":
      return <Input.TextArea placeholder={placeholder} autoSize={{ minRows: 2, maxRows: 6 }} />;
    case "text":
    case "autocomplete":
    default:
      return <Input type={col.inputType} allowClear={col.allowClear} placeholder={placeholder} {...unitFieldProps("text", col.unitKey, t)} />;
  }
};

export default SchemaConfigFields;
