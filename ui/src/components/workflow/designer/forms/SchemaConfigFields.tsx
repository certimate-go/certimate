import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { BetaSchemaForm, type ProFormColumnsType } from "@ant-design/pro-components";
import { Form } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import type { ProviderSchemaEnvelope, SchemaColumn } from "@/api/providerschema";

import { useFormNestedFieldsContext } from "./_context";
import { deriveZodSchema } from "./schemaValidation";
import { buildValueEnum, isColumnVisible, mapValueType } from "./schemaValueTypes";

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

  const columns = useMemo<ProFormColumnsType[]>(
    () => envelope.schema.columns.filter((col) => isColumnVisible(col, getValue)).map((col) => toColumn(col, parentNamePath, t, formRule)),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [envelope, parentNamePath, t, providerConfig]
  );

  return <BetaSchemaForm layoutType="Embed" submitter={false} columns={columns} />;
};

const toColumn = (col: SchemaColumn, parentNamePath: string, t: (key: string) => string, formRule: unknown): ProFormColumnsType => {
  const tooltip = col.tooltipKey
    ? col.tooltipHtml
      ? { title: <span dangerouslySetInnerHTML={{ __html: t(col.tooltipKey) }} /> }
      : t(col.tooltipKey)
    : undefined;

  const extra = col.extraKey ? col.extraHtml ? <span dangerouslySetInnerHTML={{ __html: t(col.extraKey) }} /> : t(col.extraKey) : undefined;

  return {
    title: col.labelKey ? t(col.labelKey) : col.name,
    name: [parentNamePath, col.name],
    valueType: mapValueType(col.valueType),
    valueEnum: buildValueEnum(col, t),
    initialValue: col.default,
    tooltip,
    fieldProps: {
      placeholder: col.placeholderKey ? t(col.placeholderKey) : undefined,
      min: col.min,
      max: col.max,
    },
    formItemProps: {
      extra,
      rules: [formRule as never],
    },
  } as ProFormColumnsType;
};

export default SchemaConfigFields;
