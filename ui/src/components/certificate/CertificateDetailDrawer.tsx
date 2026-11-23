import { startTransition, useState } from "react";
import { IconX } from "@tabler/icons-react";
import { useControllableValue, useGetState } from "ahooks";
import { Button, Drawer, Flex } from "antd";

import Show from "@/components/Show";
import { type CertificateModel } from "@/domain/certificate";
import { useTriggerElement } from "@/hooks";
import CertificateDetail from "./CertificateDetail";

export interface CertificateDetailDrawerProps {
  afterClose?: () => void;
  data?: CertificateModel;
  loading?: boolean;
  open?: boolean;
  trigger?: React.ReactNode;
  onOpenChange?: (open: boolean) => void;
}

const CertificateDetailDrawer = ({ afterClose, data, loading, trigger, ...props }: CertificateDetailDrawerProps) => {
  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const triggerEl = useTriggerElement(trigger, { onClick: () => setOpen(true) });

  return (
    <>
      {triggerEl}

      <Drawer
        afterOpenChange={(open) => !open && afterClose?.()}
        autoFocus
        closeIcon={false}
        destroyOnHidden
        open={open}
        loading={loading}
        placement="right"
        size="large"
        title={
          <Flex align="center" justify="space-between" gap="small">
            <div className="flex-1 truncate">{data ? `Certificate #${data.id}` : "Certificate"}</div>
            <Button
              className="ant-drawer-close"
              style={{ marginInline: 0 }}
              icon={<IconX size="1.25em" />}
              size="small"
              type="text"
              onClick={() => setOpen(false)}
            />
          </Flex>
        }
        onClose={() => setOpen(false)}
      >
        <Show when={!!data}>
          <CertificateDetail data={data!} />
        </Show>
      </Drawer>
    </>
  );
};

const useDrawer = () => {
  type DataType = CertificateDetailDrawerProps["data"];
  const [data, setData, getData] = useGetState<DataType>();
  const [loading, setLoading] = useState<boolean>();
  const [open, setOpen] = useState(false);

  const onOpenChange = (open: boolean) => {
    setOpen(open);
  };

  return {
    drawerProps: {
      afterClose: () => {
        startTransition(() => {
          if (!open) {
            setData(void 0);
            setLoading(void 0);
          }
        });
      },
      data,
      loading,
      open,
      onOpenChange,
    },

    open: ({ data, loading }: { data: NonNullable<DataType>; loading?: boolean }) => {
      setData(data);
      setLoading(loading);
      setOpen(true);

      return {
        safeUpdate: ({ data, loading }: { data?: NonNullable<DataType>; loading?: boolean }) => {
          if (data != null) {
            if (data.id !== getData()?.id) return; // 确保数据不脏读

            setData(data);
          }

          if (loading != null) {
            setLoading(loading);
          }
        },
      };
    },
    close: () => {
      setOpen(false);
    },
  };
};

const _default = Object.assign(CertificateDetailDrawer, {
  useDrawer,
});

export default _default;
