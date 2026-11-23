import { startTransition, useState } from "react";
import { useTranslation } from "react-i18next";
import { useControllableValue } from "ahooks";
import { Modal } from "antd";

import { type WorkflowGraph } from "@/domain/workflow";
import { useTriggerElement } from "@/hooks";

import WorkflowGraphExportBox from "./WorkflowGraphExportBox";

export interface WorkflowGraphExportModalProps {
  afterClose?: () => void;
  data: WorkflowGraph;
  loading?: boolean;
  open?: boolean;
  trigger?: React.ReactNode;
  onOpenChange?: (open: boolean) => void;
}

const WorkflowGraphExportModal = ({ afterClose, data, loading, trigger, ...props }: WorkflowGraphExportModalProps) => {
  const { t } = useTranslation();

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const triggerEl = useTriggerElement(trigger, { onClick: () => setOpen(true) });

  const handleCancelClick = () => {
    setOpen(false);
  };

  return (
    <>
      {triggerEl}

      <Modal
        afterClose={afterClose}
        closable
        destroyOnHidden
        footer={null}
        loading={loading}
        open={open}
        title={t("workflow.detail.design.action.export.modal.title")}
        width="768px"
        onCancel={handleCancelClick}
      >
        <div className="py-3 pb-0">
          <WorkflowGraphExportBox data={data} />
        </div>
      </Modal>
    </>
  );
};

const useModal = () => {
  type DataType = WorkflowGraphExportModalProps["data"];
  const [data, setData] = useState<DataType>();
  const [open, setOpen] = useState(false);

  const onOpenChange = (open: boolean) => {
    setOpen(open);
  };

  return {
    modalProps: {
      afterClose: () => {
        startTransition(() => {
          if (!open) {
            setData(void 0);
          }
        });
      },
      data: data!,
      open,
      onOpenChange,
    },

    open: ({ data }: { data: DataType }) => {
      setData(data);
      setOpen(true);
    },
    close: () => {
      setOpen(false);
    },
  };
};

const _default = Object.assign(WorkflowGraphExportModal, {
  useModal,
});

export default _default;
