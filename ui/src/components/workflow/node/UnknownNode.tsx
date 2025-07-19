import { memo } from "react";
import { CloseCircleOutlined as CloseCircleOutlinedIcon } from "@ant-design/icons";
import { Alert, Button, Card } from "antd";

import { useZustandShallowSelector } from "@/hooks";
import { useWorkflowStore } from "@/stores/workflow";

import { type SharedNodeProps } from "./_SharedNode";
import AddNode from "./AddNode";

export type MonitorNodeProps = SharedNodeProps;

const UnknownNode = ({ node, disabled }: MonitorNodeProps) => {
  const { removeNode } = useWorkflowStore(useZustandShallowSelector(["removeNode"]));

  const handleClickRemove = () => {
    removeNode(node);
  };

  return (
    <>
      <div className="relative w-[256px] overflow-hidden">
        <Card className="shadow-md" styles={{ body: { padding: 0 } }} hoverable variant="borderless">
          <div className="cursor-pointer ">
            <Alert
              type="error"
              message={
                <div className="flex items-center justify-between gap-4 overflow-hidden">
                  <div className="flex-1 text-center text-xs">
                    INVALID NODE
                    <br />
                    PLEASE REMOVE
                  </div>
                  <Button color="primary" icon={<CloseCircleOutlinedIcon />} variant="text" onClick={handleClickRemove} />
                </div>
              }
            />
          </div>
        </Card>
      </div>

      <AddNode node={node} disabled={disabled} />
    </>
  );
};

export default memo(UnknownNode);
