import { startTransition, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useControllableValue } from "ahooks";
import { Button, Flex, Modal } from "antd";

import { type WorkflowGraph } from "@/domain/workflow";
import { useTriggerElement } from "@/hooks";

import WorkflowImportExportForm, { type WorkflowGraphImportInputBoxInstance } from "./WorkflowGraphImportInputBox";

export interface WorkflowGraphImportModalProps {
  afterClose?: () => void;
  open?: boolean;
  trigger?: React.ReactNode;
  onCancel?: () => void;
  onOk?: (graph: WorkflowGraph) => void;
  onOpenChange?: (open: boolean) => void;
}

const WorkflowGraphImportModal = ({ afterClose, trigger, onCancel, onOk, ...props }: WorkflowGraphImportModalProps) => {
  const { t } = useTranslation();

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const triggerEl = useTriggerElement(trigger, { onClick: () => setOpen(true) });

  const graphInputBoxRef = useRef<WorkflowGraphImportInputBoxInstance>(null);

  const handleCancelClick = () => {
    setOpen(false);

    onCancel?.();
  };

  const handleOkClick = async () => {
    const graph = await graphInputBoxRef.current!.validate();

    setOpen(false);

    if (graph != null) {
      onOk?.(graph);
    }
  };

  return (
    <>
      {triggerEl}

      <Modal
        afterClose={afterClose}
        closable
        destroyOnHidden
        footer={
          <Flex className="px-2" justify="end" gap="small">
            <Button onClick={handleCancelClick}>{t("common.button.cancel")}</Button>
            <Button type="primary" onClick={handleOkClick}>
              {t("workflow.detail.design.action.import.modal.ok_button")}
            </Button>
          </Flex>
        }
        open={open}
        title={t("workflow.detail.design.action.import.modal.title")}
        width="768px"
        onCancel={handleCancelClick}
      >
        <div className="py-3">
          <WorkflowImportExportForm ref={graphInputBoxRef} />
        </div>
      </Modal>
    </>
  );
};

const useModal = () => {
  const [open, setOpen] = useState(false);
  const [onOkHandler, setOnOkHandler] = useState<{ handler: WorkflowGraphImportModalProps["onOk"] }>();

  const onOpenChange = (open: boolean) => {
    setOpen(open);
  };

  return {
    modalProps: {
      afterClose: () => {
        startTransition(() => {
          if (!open) {
            setOnOkHandler(void 0);
          }
        });
      },
      open,
      onOk: (graph: WorkflowGraph) => {
        onOkHandler?.handler?.(graph);
      },
      onOpenChange,
    },

    open: () => {
      setOpen(true);

      const { promise, resolve } = Promise.withResolvers<WorkflowGraph>();
      setOnOkHandler({ handler: (graph) => resolve(graph) });
      return promise;
    },
    close: () => {
      setOpen(false);
    },
  };
};

const _default = Object.assign(WorkflowGraphImportModal, {
  useModal,
});

export default _default;
