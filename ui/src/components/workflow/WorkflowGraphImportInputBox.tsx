import { forwardRef, useImperativeHandle } from "react";
import { useTranslation } from "react-i18next";
import { Form, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { parse as parseYaml } from "yaml";
import { z } from "zod";

import CodeTextInput from "@/components/CodeTextInput";
import { WORKFLOW_NODE_TYPES, type WorkflowGraph, type WorkflowNode, type WorkflowNodeType } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";

export type WorkflowGraphImportInputBoxFormats = "json" | "yaml";

export interface WorkflowGraphImportInputBoxProps {
  className?: string;
  style?: React.CSSProperties;
}

export interface WorkflowGraphImportInputBoxInstance {
  validate: () => Promise<WorkflowGraph | undefined>;
}

const deserialize = (content: string | undefined, format: WorkflowGraphImportInputBoxFormats): WorkflowGraph | undefined => {
  if (!content?.trim()) return;

  let temp: any;
  switch (format) {
    case "json":
      temp = JSON.parse(content);
      break;

    case "yaml":
      temp = parseYaml(content);
      break;
  }

  const deepParse = (item: any): WorkflowNode => {
    item = item ?? {};

    const node: WorkflowNode = {
      id: item.id?.toString() ?? "",
      type: (item.type?.toString() ?? "") as WorkflowNodeType,
      data: {
        name: item.name?.toString() ?? "",
        disabled: item.disabled === true || item.disabled === "true",
        config: item.config,
      },
      blocks: Array.isArray(item.blocks) ? item.blocks.map((block: any) => deepParse(block)) : [],
    };

    if (item.data != null) {
      Object.entries(item.data).forEach(([k, v]) => {
        if (k === "id" || k === "type" || k === "meta" || k === "blocks") return;
        if (k === "name" || k === "disabled" || k === "config") return;
        node.data[k] = v;
      });
    }

    return node;
  };
  const nodes = Array.from(temp.nodes ?? []).map((item) => deepParse(item));

  return { nodes: nodes };
};

const WorkflowGraphImportInputBox = forwardRef<WorkflowGraphImportInputBoxInstance, WorkflowGraphImportInputBoxProps>(({ className, style }, ref) => {
  const { t } = useTranslation();

  const formSchema = z
    .object({
      format: z.enum(["json", "yaml"]),
      content: z.string().refine((v) => !!v?.trim(), t("workflow.detail.design.action.import.form.content.errmsg.invalid")),
    })
    .superRefine((values, ctx) => {
      let graph: WorkflowGraph | undefined;
      try {
        graph = deserialize(values.content, values.format);
      } catch {
        ctx.addIssue({
          code: "custom",
          message: t("workflow.detail.design.action.import.form.content.errmsg.invalid"),
          path: ["content"],
        });
        return;
      }

      if (graph) {
        const errmsgs: string[] = [];

        if (graph.nodes.at(0)?.type !== WORKFLOW_NODE_TYPES.START) {
          errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.first_node_start"));
        }

        if (graph.nodes.at(-1)?.type !== WORKFLOW_NODE_TYPES.END) {
          errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.last_node_end"));
        }

        let startNodeId: string | undefined;
        let startNodeDuplicated: boolean = false;
        const nodeIds = new Set<string>();
        const deepValidate = (node: WorkflowNode) => {
          // 验证字段：ID
          if (!/^(?![_-])[a-zA-Z0-9_-]{1,32}$/.test(node.id)) {
            errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.invalid_id", { nodeId: node.id }));
          }

          // 验证字段：配置项
          if (node.data.config != null) {
            if (typeof node.data.config !== "object" || Array.isArray(node.data.config)) {
              errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.invalid_config", { nodeId: node.id }));
            }
          }

          // 验证节点 ID 是否冲突
          if (nodeIds.has(node.id)) {
            errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.conflict_id", { nodeId: node.id }));
          } else {
            nodeIds.add(node.id);
          }

          // 验证开始节点是否重复
          if (node.type === WORKFLOW_NODE_TYPES.START) {
            if (startNodeId) {
              if (!startNodeDuplicated) {
                startNodeDuplicated = true;
                errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.duplicate_start"));
              }
            } else {
              startNodeId = node.id;
            }
          }

          // 验证 Condition 分支结构
          if (node.type === WORKFLOW_NODE_TYPES.CONDITION) {
            const blocks = node.blocks ?? [];
            const f1 = Array.isArray(blocks) && blocks.length > 0;
            const f2 = Array.from(blocks).every((block) => block.type === WORKFLOW_NODE_TYPES.BRANCHBLOCK);
            if (!f1 || !f2) {
              errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.abnormal_condition_branch", { nodeId: node.id }));
            }
          } else if (node.type === WORKFLOW_NODE_TYPES.BRANCHBLOCK) {
            const blocks = node.blocks ?? [];
            const f1 = Array.isArray(blocks);
            const f2 = Array.from(blocks).every((block) => block.type !== WORKFLOW_NODE_TYPES.BRANCHBLOCK);
            if (!f1 || !f2) {
              errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.abnormal_condition_branch", { nodeId: node.id }));
            }
          }

          // 验证 TryCatch 分支结构
          if (node.type === WORKFLOW_NODE_TYPES.TRYCATCH) {
            const blocks = node.blocks ?? [];
            const f1 = Array.isArray(blocks) && blocks.length >= 2;
            const f2 = Array.from(blocks).at(0)?.type === WORKFLOW_NODE_TYPES.TRYBLOCK;
            const f3 = Array.from(blocks).some((block) => block.type === WORKFLOW_NODE_TYPES.CATCHBLOCK);
            if (!f1 || !f2 || !f3) {
              errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.abnormal_try_catch_branch", { nodeId: node.id }));
            }
          } else if (node.type === WORKFLOW_NODE_TYPES.TRYBLOCK || node.type === WORKFLOW_NODE_TYPES.CATCHBLOCK) {
            const blocks = node.blocks ?? [];
            const f1 = Array.isArray(blocks);
            const f2 = Array.from(blocks).every((block) => block.type !== WORKFLOW_NODE_TYPES.TRYBLOCK);
            const f3 = Array.from(blocks).every((block) => block.type !== WORKFLOW_NODE_TYPES.CATCHBLOCK);
            if (!f1 || !f2 || !f3) {
              errmsgs.push(t("workflow.detail.design.action.import.form.content.errmsg.abnormal_try_catch_branch", { nodeId: node.id }));
            }
          }

          // 验证子节点
          if (Array.isArray(node.blocks)) {
            node.blocks.forEach((block) => deepValidate(block));
          }
        };
        graph.nodes.forEach((node) => deepValidate(node));

        const MAX_ISSUE_COUNT = 5;
        for (let i = 0; i < Math.min(MAX_ISSUE_COUNT, errmsgs.length); i++) {
          ctx.addIssue({
            code: "custom",
            message: errmsgs[i],
            path: ["content"],
          });
        }
      }
    });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    name: "workflowGraphExportInputBoxForm",
    initialValues: {
      format: "yaml",
      content: "",
    },
  });

  const fieldFormat = Form.useWatch<WorkflowGraphImportInputBoxFormats>("format", formInst);
  const fieldContent = Form.useWatch<string>("content", formInst);

  const handleFormatChange = (format: WorkflowGraphImportInputBoxFormats) => {
    formInst.setFieldValue("format", format);

    switch (format) {
      case "json":
        try {
          JSON.parse(fieldContent);
        } catch {
          formInst.setFieldValue("content", "");
        }
        break;

      case "yaml":
        try {
          parseYaml(fieldContent);
        } catch {
          formInst.setFieldValue("content", "");
        }
        break;
    }
  };

  useImperativeHandle(ref, () => {
    return {
      validate: async () => {
        const formValues = await formInst.validateFields();
        return deserialize(formValues.content, formValues.format);
      },
    };
  });

  return (
    <Form className={className} style={style} {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
      <Form.Item className="mb-4" name="format" label={t("workflow.detail.design.action.import.form.format.label")} rules={[formRule]}>
        <Radio.Group block onChange={(e) => handleFormatChange(e.target.value)}>
          <Radio.Button value="yaml">YAML</Radio.Button>
          <Radio.Button value="json">JSON</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Form.Item name="content" label={t("workflow.detail.design.action.import.form.content.label")} rules={[formRule]}>
        <CodeTextInput height="calc(min(60vh, 512px))" language={fieldFormat} lineWrapping={false} value={fieldContent} />
      </Form.Item>
    </Form>
  );
});

export default WorkflowGraphImportInputBox;
