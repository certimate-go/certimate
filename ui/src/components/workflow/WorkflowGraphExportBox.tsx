import { useEffect, useState } from "react";
import { CopyToClipboard } from "react-copy-to-clipboard";
import { useTranslation } from "react-i18next";
import { FlowNodeBaseType, FlowNodeSplitType } from "@flowgram.ai/fixed-layout-editor";
import { IconClipboard } from "@tabler/icons-react";
import { App, Button, Form, Radio, Tooltip } from "antd";
import { stringify as stringifyYaml } from "yaml";

import CodeTextInput from "@/components/CodeTextInput";
import { type WorkflowGraph, type WorkflowNode } from "@/domain/workflow";

import { getAllNodeRegistries } from "./designer/nodes";

export type WorkflowGraphExportBoxFormats = "json" | "yaml";

export interface WorkflowGraphExportBoxProps {
  className?: string;
  style?: React.CSSProperties;
  data: WorkflowGraph;
}

const serialize = (graph: WorkflowGraph | undefined, format: WorkflowGraphExportBoxFormats): string | undefined => {
  if (!graph) return;

  const nodeRegistries = getAllNodeRegistries();

  const deepConvert = (node: WorkflowNode): Map<string, unknown> => {
    // 利用 Map 来保证字段序列化的有序性
    const map = new Map<string, unknown>([
      ["id", node.id],
      ["type", node.type],
      ["name", node.data.name],
    ]);

    if (node.data.disabled != null) {
      if (node.data.disabled) {
        map.set("disabled", node.data.disabled);
      }
    }

    if (node.data.config != null) {
      map.set("config", node.data.config);
    }

    if (node.blocks != null) {
      const branchLikeNodeTypes = [
        FlowNodeBaseType.BLOCK,
        FlowNodeSplitType.SIMPLE_SPLIT,
        FlowNodeSplitType.DYNAMIC_SPLIT,
        FlowNodeSplitType.STATIC_SPLIT,
      ] as string[];
      const hasChildren = node.blocks.length > 0 || branchLikeNodeTypes.includes(nodeRegistries.find((r) => r.type === node.type)?.extend ?? "");
      if (hasChildren) {
        const childern = node.blocks.map((block) => deepConvert(block));
        map.set("blocks", childern);
      }
    }

    Object.entries(node.data).forEach(([k, v]) => {
      if (k === "name" || k === "disabled" || k === "config") return;
      map.set(k, v);
    });

    return map;
  };
  const nodes = graph.nodes.map((node) => deepConvert(node));

  let content: string = "";
  switch (format) {
    case "json":
      content = JSON.stringify(
        { ...graph, nodes },
        (_, value) => {
          if (value instanceof Map) {
            return Object.fromEntries(value.entries());
          } else {
            return value;
          }
        },
        2
      );
      break;

    case "yaml":
      content = stringifyYaml(
        { ...graph, nodes },
        {
          indent: 2,
          defaultKeyType: "PLAIN",
          defaultStringType: "QUOTE_DOUBLE",
        }
      );
      break;
  }

  return content;
};

const WorkflowGraphExportBox = ({ className, style, data }: WorkflowGraphExportBoxProps) => {
  const { t } = useTranslation();

  const { message } = App.useApp();

  const [format, setFormat] = useState<WorkflowGraphExportBoxFormats>("yaml");
  const [content, setContent] = useState<string>();

  useEffect(() => {
    setContent(serialize(data, format));
  }, [data, format]);

  return (
    <Form className={className} style={style} layout="vertical">
      <Form.Item className="mb-4" label={t("workflow.detail.design.action.export.form.format.label")}>
        <Radio.Group block value={format} onChange={(e) => setFormat(e.target.value)}>
          <Radio.Button value="yaml">YAML</Radio.Button>
          <Radio.Button value="json">JSON</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Form.Item label={t("workflow.detail.design.action.export.form.content.label")}>
        <div className="absolute -top-1.5 right-0 -translate-y-full">
          <Tooltip title={t("common.button.copy")}>
            <CopyToClipboard
              text={content!}
              onCopy={() => {
                message.success(t("common.text.copied"));
              }}
            >
              <Button icon={<IconClipboard size="1.25em" />} disabled={!content} size="small" type="text" />
            </CopyToClipboard>
          </Tooltip>
        </div>
        <CodeTextInput height="calc(min(60vh, 512px))" language={format} lineWrapping={false} value={content} readOnly />
      </Form.Item>
    </Form>
  );
};

export default WorkflowGraphExportBox;
