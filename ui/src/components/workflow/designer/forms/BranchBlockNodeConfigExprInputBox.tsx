import { forwardRef, useEffect, useImperativeHandle, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconCircleMinus, IconCirclePlus } from "@tabler/icons-react";
import { useControllableValue } from "ahooks";
import { Button, Form, Input, Radio, Select, theme } from "antd";

import Show from "@/components/Show";
import {
  type Expr,
  type ExprComparisonOperator,
  type ExprLogicalOperator,
  ExprType,
  type ExprValue,
  type ExprValueSelector,
  type ExprValueType,
} from "@/domain/workflow";
import { useAntdFormName } from "@/hooks";

import { useNodeFormContext } from "./_context";
import { getAllPreviousNodes } from "../_util";
import { NodeType } from "../nodes/typings";

export interface BranchBlockNodeConfigExprInputBoxProps {
  className?: string;
  style?: React.CSSProperties;
  defaultValue?: Expr;
  value?: Expr;
  onChange?: (value: Expr) => void;
}

export interface BranchBlockNodeConfigExprInputBoxInstance {
  validate: () => Promise<Expr | undefined>;
}

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

const exprToFormValues = (expr: Expr | undefined): ConditionFormValues => {
  if (!expr) return getInitialValues();

  const conditions: ConditionItem[] = [];
  let logicalOp: ExprLogicalOperator = "and";

  const extractExpr = (expr: Expr): void => {
    if (expr.type === ExprType.Comparison) {
      if (expr.left.type == ExprType.Variant && expr.right.type == ExprType.Constant) {
        conditions.push({
          leftSelector: expr.left.selector?.id != null ? `${expr.left.selector.id}#${expr.left.selector.name}#${expr.left.selector.type}` : void 0,
          operator: expr.operator != null ? expr.operator : void 0,
          rightValue: expr.right?.value != null ? String(expr.right.value) : void 0,
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
    return;
  }

  // 只有一个条件时，直接返回比较表达式
  if (values.conditions.length === 1) {
    const { leftSelector, operator, rightValue } = values.conditions[0];
    if (!leftSelector || !operator || !rightValue) {
      return;
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

const BranchBlockNodeConfigExprInputBox = forwardRef<BranchBlockNodeConfigExprInputBoxInstance, BranchBlockNodeConfigExprInputBoxProps>(
  ({ className, style, ...props }, ref) => {
    const { t } = useTranslation();

    const { token: themeToken } = theme.useToken();

    const [value, setValue] = useControllableValue<Expr | undefined>(props, {
      valuePropName: "value",
      defaultValuePropName: "defaultValue",
      trigger: "onChange",
    });

    const { node } = useNodeFormContext();

    const [formInst] = Form.useForm<ConditionFormValues>();
    const formName = useAntdFormName({ form: formInst, name: "workflowNodeBranchBlockConfigExprInputBoxForm" });
    const [formModel, setFormModel] = useState<ConditionFormValues>(getInitialValues());

    useEffect(() => {
      if (value) {
        const formValues = exprToFormValues(value);
        formInst.setFieldsValue(formValues);
        setFormModel(formValues);
      } else {
        formInst.resetFields();
        setFormModel(getInitialValues());
      }
    }, [value]);

    const ciSelectorOptions = useMemo(() => {
      return getAllPreviousNodes(node)
        .filter((node) => node.flowNodeType === NodeType.BizApply || node.flowNodeType === NodeType.BizUpload || node.flowNodeType === NodeType.BizMonitor)
        .map((node) => {
          const form = node.form;
          const group = {
            data: {
              name: form?.getValueIn("name"),
              ...form?.values,
            },
            label: (
              <div className="flex items-center justify-between gap-4 overflow-hidden">
                <div className="flex-1 truncate">{form?.getValueIn("name")}</div>
                <div className="origin-right scale-90 font-mono text-xs" style={{ color: themeToken.colorTextSecondary }}>
                  (NodeID: {node.id})
                </div>
              </div>
            ),
            options: Array<{ label: string; value: string }>(),
          };

          group.options.push({
            label: `${t("workflow.variables.type.certificate.label")} - ${t("workflow.variables.selector.hours_left.label")}`,
            value: `${node.id}#certificate.hoursLeft#number`,
          });
          group.options.push({
            label: `${t("workflow.variables.type.certificate.label")} - ${t("workflow.variables.selector.days_left.label")}`,
            value: `${node.id}#certificate.daysLeft#number`,
          });
          group.options.push({
            label: `${t("workflow.variables.type.certificate.label")} - ${t("workflow.variables.selector.validity.label")}`,
            value: `${node.id}#certificate.validity#boolean`,
          });

          return group;
        })
        .filter((item) => item.options.length > 0);
    }, [node]);

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
            { value: "eq", label: t("workflow_node.branch_block.form.expression.operator.option.eq.label") },
            { value: "neq", label: t("workflow_node.branch_block.form.expression.operator.option.neq.label") },
            { value: "gt", label: t("workflow_node.branch_block.form.expression.operator.option.gt.label") },
            { value: "gte", label: t("workflow_node.branch_block.form.expression.operator.option.gte.label") },
            { value: "lt", label: t("workflow_node.branch_block.form.expression.operator.option.lt.label") },
            { value: "lte", label: t("workflow_node.branch_block.form.expression.operator.option.lte.label") },
          ];

        case "string":
          return [
            { value: "eq", label: t("workflow_node.branch_block.form.expression.operator.option.eq.label") },
            { value: "neq", label: t("workflow_node.branch_block.form.expression.operator.option.neq.label") },
          ];

        case "boolean":
          return [
            { value: "eq", label: t("workflow_node.branch_block.form.expression.operator.option.eq.alias_is_label") },
            { value: "neq", label: t("workflow_node.branch_block.form.expression.operator.option.neq.alias_not_label") },
          ];

        default:
          return [];
      }
    };

    const handleFormChange = (_: unknown, values: ConditionFormValues) => {
      // TODO: 这里直接用参数 `values` 会丢失部分字段，引发 Issue #1096。
      // 暂时先用 `getFieldsValue()` 代替，待排查原因，疑似与 antd v6 升级有关。
      values = formInst.getFieldsValue();
      setValue(formValuesToExpr(values));
    };

    useImperativeHandle(ref, () => {
      return {
        validate: async () => {
          const formValues = await formInst.validateFields();
          return formValuesToExpr(formValues);
        },
      };
    });

    return (
      <Form className={className} style={style} form={formInst} initialValues={formModel} layout="vertical" name={formName} onValuesChange={handleFormChange}>
        <Show when={formModel.conditions?.length > 1}>
          <Form.Item name="logicalOperator" rules={[{ required: true, message: t("workflow_node.branch_block.form.expression.logical_operator.errmsg") }]}>
            <Radio.Group block>
              <Radio.Button value="and">{t("workflow_node.branch_block.form.expression.logical_operator.option.and.label")}</Radio.Button>
              <Radio.Button value="or">{t("workflow_node.branch_block.form.expression.logical_operator.option.or.label")}</Radio.Button>
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
                    rules={[{ required: true, message: t("workflow_node.branch_block.form.expression.variable.errmsg") }]}
                    {...rest}
                  >
                    <Select
                      labelRender={({ label, value }) => {
                        if (value != null) {
                          const group = ciSelectorOptions.find((group) => group.options.some((option) => option.value === value));
                          return `${group?.data?.name} - ${label}`;
                        }

                        return (
                          <span style={{ color: themeToken.colorTextPlaceholder }}>{t("workflow_node.branch_block.form.expression.variable.placeholder")}</span>
                        );
                      }}
                      options={ciSelectorOptions}
                      placeholder={t("workflow_node.branch_block.form.expression.variable.placeholder")}
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
                          rules={[{ required: true, message: t("workflow_node.branch_block.form.expression.operator.errmsg") }]}
                          {...rest}
                        >
                          <Select
                            open={operators.length === 0 ? false : void 0}
                            options={operators}
                            placeholder={t("workflow_node.branch_block.form.expression.operator.placeholder")}
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
                          rules={[{ required: true, message: t("workflow_node.branch_block.form.expression.value.errmsg") }]}
                          {...rest}
                        >
                          {valueType === "string" ? (
                            <Input placeholder={t("workflow_node.branch_block.form.expression.value.placeholder")} />
                          ) : valueType === "number" ? (
                            <Input type="number" placeholder={t("workflow_node.branch_block.form.expression.value.placeholder")} />
                          ) : valueType === "boolean" ? (
                            <Select placeholder={t("workflow_node.branch_block.form.expression.value.placeholder")}>
                              <Select.Option value="true">{t("workflow_node.branch_block.form.expression.value.option.true.label")}</Select.Option>
                              <Select.Option value="false">{t("workflow_node.branch_block.form.expression.value.option.false.label")}</Select.Option>
                            </Select>
                          ) : (
                            <Input readOnly placeholder={t("workflow_node.branch_block.form.expression.value.placeholder")} />
                          )}
                        </Form.Item>
                      );
                    }}
                  </Form.Item>

                  <Button color="default" icon={<IconCircleMinus size="1.25em" />} type="text" onClick={() => remove(index)} />
                </div>
              ))}

              <Form.Item>
                <Button type="dashed" block icon={<IconCirclePlus size="1.25em" />} onClick={() => add({})}>
                  {t("workflow_node.branch_block.form.expression.add_condition.button")}
                </Button>
              </Form.Item>
            </div>
          )}
        </Form.List>
      </Form>
    );
  }
);

const getInitialValues = (): ConditionFormValues => {
  return {
    conditions: [{}],
    logicalOperator: "and",
  };
};

export default BranchBlockNodeConfigExprInputBox;
