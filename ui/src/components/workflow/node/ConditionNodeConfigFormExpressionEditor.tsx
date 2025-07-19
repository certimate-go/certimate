import { forwardRef, useEffect, useImperativeHandle, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { CloseOutlined as CloseOutlinedIcon, PlusOutlined } from "@ant-design/icons";
import { useControllableValue } from "ahooks";
import { Button, Form, Input, Radio, Select, theme } from "antd";

import Show from "@/components/Show";
import type { Expr, ExprComparisonOperator, ExprLogicalOperator, ExprValue, ExprValueSelector, ExprValueType } from "@/domain/workflow";
import { ExprType } from "@/domain/workflow";
import { useAntdFormName, useZustandShallowSelector } from "@/hooks";
import { useWorkflowStore } from "@/stores/workflow";

export type ConditionNodeConfigFormExpressionEditorProps = {
  className?: string;
  style?: React.CSSProperties;
  defaultValue?: Expr;
  disabled?: boolean;
  nodeId: string;
  value?: Expr;
  onChange?: (value: Expr) => void;
};

export type ConditionNodeConfigFormExpressionEditorInstance = {
  validate: () => Promise<void>;
};

// 表单内部使用的扁平结构
type ConditionItem = {
  // 选择器，格式为 "${nodeId}#${outputName}#${valueType}"
  // 将 [ExprValueSelector] 转为字符串形式，以便于结构化存储。
  leftSelector?: string;
  // 比较运算符。
  operator?: ExprComparisonOperator;
  // 值。
  // 将 [ExprValue] 转为字符串形式，以便于结构化存储。
  rightValue?: string;
};

type ConditionFormValues = {
  conditions: ConditionItem[];
  logicalOperator: ExprLogicalOperator;
};

const initFormModel = (): ConditionFormValues => {
  return {
    conditions: [{}],
    logicalOperator: "and",
  };
};

const exprToFormValues = (expr?: Expr): ConditionFormValues => {
  if (!expr) return initFormModel();

  const conditions: ConditionItem[] = [];
  let logicalOp: ExprLogicalOperator = "and";

  const extractExpr = (expr: Expr): void => {
    if (expr.type === ExprType.Comparison) {
      if (expr.left.type == ExprType.Variant && expr.right.type == ExprType.Constant) {
        conditions.push({
          leftSelector: expr.left.selector?.id != null ? `${expr.left.selector.id}#${expr.left.selector.name}#${expr.left.selector.type}` : undefined,
          operator: expr.operator != null ? expr.operator : undefined,
          rightValue: expr.right?.value != null ? String(expr.right.value) : undefined,
        });
      } else {
        console.warn("[certimate] invalid comparison expression: left must be a variant and right must be a constant", expr);
      }
    } else if (expr.type === ExprType.Logical) {
      logicalOp = expr.operator || "and";
      extractExpr(expr.left);
      extractExpr(expr.right);
    }
  };

  extractExpr(expr);

  return {
    conditions: conditions,
    logicalOperator: logicalOp,
  };
};

const formValuesToExpr = (values: ConditionFormValues): Expr | undefined => {
  const wrapExpr = (condition: ConditionItem): Expr => {
    const [id, name, type] = (condition.leftSelector?.split("#") ?? ["", "", ""]) as [string, string, ExprValueType];
    const valid = !!id && !!name && !!type;

    const left: Expr = {
      type: ExprType.Variant,
      selector: valid
        ? {
            id: id,
            name: name,
            type: type,
          }
        : ({} as ExprValueSelector),
    };

    const right: Expr = {
      type: ExprType.Constant,
      value: condition.rightValue!,
      valueType: type,
    };

    return {
      type: ExprType.Comparison,
      operator: condition.operator!,
      left,
      right,
    };
  };

  if (values.conditions.length === 0) {
    return undefined;
  }

  // 只有一个条件时，直接返回比较表达式
  if (values.conditions.length === 1) {
    const { leftSelector, operator, rightValue } = values.conditions[0];
    if (!leftSelector || !operator || !rightValue) {
      return undefined;
    }
    return wrapExpr(values.conditions[0]);
  }

  // 多个条件时，通过逻辑运算符连接
  let expr: Expr = wrapExpr(values.conditions[0]);
  for (let i = 1; i < values.conditions.length; i++) {
    expr = {
      type: ExprType.Logical,
      operator: values.logicalOperator,
      left: expr,
      right: wrapExpr(values.conditions[i]),
    };
  }
  return expr;
};

const ConditionNodeConfigFormExpressionEditor = forwardRef<ConditionNodeConfigFormExpressionEditorInstance, ConditionNodeConfigFormExpressionEditorProps>(
  ({ className, style, disabled, nodeId, ...props }, ref) => {
    const { t } = useTranslation();

    const { token: themeToken } = theme.useToken();

    const { getWorkflowOuptutBeforeId } = useWorkflowStore(useZustandShallowSelector(["updateNode", "getWorkflowOuptutBeforeId"]));

    const [value, setValue] = useControllableValue<Expr | undefined>(props, {
      valuePropName: "value",
      defaultValuePropName: "defaultValue",
      trigger: "onChange",
    });

    const [formInst] = Form.useForm<ConditionFormValues>();
    const formName = useAntdFormName({ form: formInst, name: "workflowNodeConditionConfigFormExpressionEditorForm" });
    const [formModel, setFormModel] = useState<ConditionFormValues>(initFormModel());

    useEffect(() => {
      if (value) {
        const formValues = exprToFormValues(value);
        formInst.setFieldsValue(formValues);
        setFormModel(formValues);
      } else {
        formInst.resetFields();
        setFormModel(initFormModel());
      }
    }, [value]);

    const ciSelectorCandidates = useMemo(() => {
      const previousNodes = getWorkflowOuptutBeforeId(nodeId);
      return previousNodes
        .map((node) => {
          const group = {
            label: node.name,
            options: Array<{ label: string; value: string }>(),
          };

          for (const output of node.outputs ?? []) {
            switch (output.type) {
              case "certificate":
                group.options.push({
                  label: `${output.label} - ${t("workflow.variables.selector.validity.label")}`,
                  value: `${node.id}#${output.name}.validity#boolean`,
                });
                group.options.push({
                  label: `${output.label} - ${t("workflow.variables.selector.days_left.label")}`,
                  value: `${node.id}#${output.name}.daysLeft#number`,
                });
                break;

              default:
                group.options.push({
                  label: `${output.label}`,
                  value: `${node.id}#${output.name}#${output.type}`,
                });
                console.warn("[certimate] invalid workflow output type in condition expressions", output);
                break;
            }
          }

          return group;
        })
        .filter((item) => item.options.length > 0);
    }, [nodeId]);

    const getValueTypeBySelector = (selector: string): ExprValueType | undefined => {
      if (!selector) return;

      const parts = selector.split("#");
      if (parts.length >= 3) {
        return parts[2].toLowerCase() as ExprValueType;
      }
    };

    const getOperatorsBySelector = (selector: string): { value: ExprComparisonOperator; label: string }[] => {
      const valueType = getValueTypeBySelector(selector);
      return getOperatorsByValueType(valueType!);
    };

    const getOperatorsByValueType = (valueType: ExprValue): { value: ExprComparisonOperator; label: string }[] => {
      switch (valueType) {
        case "number":
          return [
            { value: "eq", label: t("workflow_node.condition.form.expression.operator.option.eq.label") },
            { value: "neq", label: t("workflow_node.condition.form.expression.operator.option.neq.label") },
            { value: "gt", label: t("workflow_node.condition.form.expression.operator.option.gt.label") },
            { value: "gte", label: t("workflow_node.condition.form.expression.operator.option.gte.label") },
            { value: "lt", label: t("workflow_node.condition.form.expression.operator.option.lt.label") },
            { value: "lte", label: t("workflow_node.condition.form.expression.operator.option.lte.label") },
          ];

        case "string":
          return [
            { value: "eq", label: t("workflow_node.condition.form.expression.operator.option.eq.label") },
            { value: "neq", label: t("workflow_node.condition.form.expression.operator.option.neq.label") },
          ];

        case "boolean":
          return [
            { value: "eq", label: t("workflow_node.condition.form.expression.operator.option.eq.alias_is_label") },
            { value: "neq", label: t("workflow_node.condition.form.expression.operator.option.neq.alias_not_label") },
          ];

        default:
          return [];
      }
    };

    const handleFormChange = (_: undefined, values: ConditionFormValues) => {
      setValue(formValuesToExpr(values));
    };

    useImperativeHandle(ref, () => {
      return {
        validate: async () => {
          await formInst.validateFields();
        },
      } as ConditionNodeConfigFormExpressionEditorInstance;
    });

    return (
      <Form
        className={className}
        style={style}
        form={formInst}
        disabled={disabled}
        initialValues={formModel}
        layout="vertical"
        name={formName}
        onValuesChange={handleFormChange}
      >
        <Show when={formModel.conditions?.length > 1}>
          <Form.Item name="logicalOperator" rules={[{ required: true, message: t("workflow_node.condition.form.expression.logical_operator.errmsg") }]}>
            <Radio.Group block>
              <Radio.Button value="and">{t("workflow_node.condition.form.expression.logical_operator.option.and.label")}</Radio.Button>
              <Radio.Button value="or">{t("workflow_node.condition.form.expression.logical_operator.option.or.label")}</Radio.Button>
            </Radio.Group>
          </Form.Item>
        </Show>

        <Form.List name="conditions">
          {(fields, { add, remove }) => (
            <div className="flex flex-col gap-2">
              {fields.map(({ key, name: index, ...rest }) => (
                <div key={key} className="flex gap-2">
                  {/* 左：变量选择器 */}
                  <Form.Item
                    className="mb-0 flex-1"
                    name={[index, "leftSelector"]}
                    rules={[{ required: true, message: t("workflow_node.condition.form.expression.variable.errmsg") }]}
                    {...rest}
                  >
                    <Select
                      labelRender={({ label, value }) => {
                        if (value != null) {
                          const group = ciSelectorCandidates.find((group) => group.options.some((option) => option.value === value));
                          return `${group?.label} - ${label}`;
                        }

                        return (
                          <span style={{ color: themeToken.colorTextPlaceholder }}>{t("workflow_node.condition.form.expression.variable.placeholder")}</span>
                        );
                      }}
                      options={ciSelectorCandidates}
                      placeholder={t("workflow_node.condition.form.expression.variable.placeholder")}
                    />
                  </Form.Item>

                  {/* 中：运算符选择器，根据变量类型决定选项 */}
                  <Form.Item
                    noStyle
                    shouldUpdate={(prevValues, currentValues) => {
                      return prevValues.conditions?.[index]?.leftSelector !== currentValues.conditions?.[index]?.leftSelector;
                    }}
                  >
                    {({ getFieldValue }) => {
                      const leftSelector = getFieldValue(["conditions", index, "leftSelector"]);
                      const operators = getOperatorsBySelector(leftSelector);

                      return (
                        <Form.Item
                          className="mb-0 w-36"
                          name={[index, "operator"]}
                          rules={[{ required: true, message: t("workflow_node.condition.form.expression.operator.errmsg") }]}
                          {...rest}
                        >
                          <Select
                            open={operators.length === 0 ? false : undefined}
                            options={operators}
                            placeholder={t("workflow_node.condition.form.expression.operator.placeholder")}
                          />
                        </Form.Item>
                      );
                    }}
                  </Form.Item>

                  {/* 右：输入控件，根据变量类型决定组件 */}
                  <Form.Item
                    noStyle
                    shouldUpdate={(prevValues, currentValues) => {
                      return prevValues.conditions?.[index]?.leftSelector !== currentValues.conditions?.[index]?.leftSelector;
                    }}
                  >
                    {({ getFieldValue }) => {
                      const leftSelector = getFieldValue(["conditions", index, "leftSelector"]);
                      const valueType = getValueTypeBySelector(leftSelector);

                      return (
                        <Form.Item
                          className="mb-0 w-36"
                          name={[index, "rightValue"]}
                          rules={[{ required: true, message: t("workflow_node.condition.form.expression.value.errmsg") }]}
                          {...rest}
                        >
                          {valueType === "string" ? (
                            <Input placeholder={t("workflow_node.condition.form.expression.value.placeholder")} />
                          ) : valueType === "number" ? (
                            <Input type="number" placeholder={t("workflow_node.condition.form.expression.value.placeholder")} />
                          ) : valueType === "boolean" ? (
                            <Select placeholder={t("workflow_node.condition.form.expression.value.placeholder")}>
                              <Select.Option value="true">{t("workflow_node.condition.form.expression.value.option.true.label")}</Select.Option>
                              <Select.Option value="false">{t("workflow_node.condition.form.expression.value.option.false.label")}</Select.Option>
                            </Select>
                          ) : (
                            <Input readOnly placeholder={t("workflow_node.condition.form.expression.value.placeholder")} />
                          )}
                        </Form.Item>
                      );
                    }}
                  </Form.Item>

                  <Button
                    className="my-1"
                    color="default"
                    disabled={disabled}
                    icon={<CloseOutlinedIcon />}
                    size="small"
                    type="text"
                    onClick={() => remove(index)}
                  />
                </div>
              ))}

              <Form.Item>
                <Button type="dashed" block icon={<PlusOutlined />} onClick={() => add({})}>
                  {t("workflow_node.condition.form.expression.add_condition.button")}
                </Button>
              </Form.Item>
            </div>
          )}
        </Form.List>
      </Form>
    );
  }
);

export default ConditionNodeConfigFormExpressionEditor;
