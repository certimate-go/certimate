import { memo } from "react";
import { useTranslation } from "react-i18next";
import { Button, theme } from "antd";

import { type WorkflowNode } from "@/domain/workflow";
import { useZustandShallowSelector } from "@/hooks";
import { useWorkflowStore } from "@/stores/workflow";

import AddNode from "./AddNode";
import WorkflowElement from "../WorkflowElement";
import { type SharedNodeProps } from "./_SharedNode";

export type BrandNodeProps = SharedNodeProps;

const BranchNode = ({ node, disabled }: BrandNodeProps) => {
  const { t } = useTranslation();

  const { addBranch } = useWorkflowStore(useZustandShallowSelector(["addBranch"]));

  const { token: themeToken } = theme.useToken();

  const renderBranch = (node: WorkflowNode, branchNodeId?: string, branchIndex?: number) => {
    const elements: JSX.Element[] = [];

    let current = node as typeof node | undefined;
    while (current) {
      elements.push(<WorkflowElement key={current.id} node={current} disabled={disabled} branchId={branchNodeId} branchIndex={branchIndex} />);
      current = current.next;
    }

    return elements;
  };

  return (
    <>
      <div
        className="relative flex gap-x-16 before:absolute before:inset-x-[128px] before:top-0 before:h-[2px] before:bg-stone-200 before:content-[''] after:absolute after:inset-x-[128px] after:bottom-0 after:h-[2px] after:bg-stone-200 after:content-['']"
        style={{
          backgroundColor: themeToken.colorBgContainer,
        }}
      >
        <Button
          className="absolute left-1/2 z-1 -translate-x-1/2 -translate-y-1/2 text-xs"
          disabled={disabled}
          size="small"
          shape="round"
          variant="outlined"
          onClick={() => {
            addBranch(node.id);
          }}
        >
          {t("workflow_node.action.add_branch")}
        </Button>

        {node.branches?.map((branch, index) => (
          <div
            key={branch.id}
            className="relative flex flex-col items-center before:absolute before:left-1/2 before:top-0 before:h-full before:w-[2px] before:-translate-x-1/2 before:bg-stone-200 before:content-['']"
          >
            {index == 0 && (
              <>
                <div className="absolute -left-[1px] -top-1 h-2 w-1/2" style={{ backgroundColor: themeToken.colorBgContainer }}></div>
                <div className="absolute -bottom-1 -left-[1px] z-50 h-2 w-1/2" style={{ backgroundColor: themeToken.colorBgContainer }}></div>
              </>
            )}
            {index == node.branches?.length! - 1 && (
              <>
                <div className="absolute -right-[1px] -top-1 h-2 w-1/2" style={{ backgroundColor: themeToken.colorBgContainer }}></div>
                <div className="absolute -bottom-1 -right-[1px] z-50 h-2 w-1/2" style={{ backgroundColor: themeToken.colorBgContainer }}></div>
              </>
            )}
            <div className="relative flex flex-col items-center">{renderBranch(branch, node.id, index)}</div>
          </div>
        ))}
      </div>

      <AddNode node={node} disabled={disabled} />
    </>
  );
};

export default memo(BranchNode);
