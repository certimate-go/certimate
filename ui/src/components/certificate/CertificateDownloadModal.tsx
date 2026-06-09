import { useState } from "react";
import { useTranslation } from "react-i18next";
import { IconChevronDown, IconChevronUp } from "@tabler/icons-react";
import { useControllableValue } from "ahooks";
import { App, Button, Divider, Form, Input, Modal, Select, Table, type TableProps } from "antd";
import { saveAs } from "file-saver";

import { download as downloadCertificate } from "@/api/certificates";
import { CERTIFICATE_FORMATS, type CertificateFormatType, type CertificateModel } from "@/domain/certificate";
import { useAntdForm, useTriggerElement } from "@/hooks";

export interface CertificateDownloadModalProps {
  className?: string;
  style?: React.CSSProperties;
  afterClose?: () => void;
  data: Pick<CertificateModel, "id" | "subjectAltNames">;
  open?: boolean;
  trigger?: React.ReactNode;
  onOpenChange?: (open: boolean) => void;
}

const CertificateDownloadModal = ({ afterClose, data, trigger, ...props }: CertificateDownloadModalProps) => {
  const { t } = useTranslation();

  const { message } = App.useApp();

  const [open, _setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });
  const setOpen = (open: boolean) => {
    _setOpen(open);

    if (!open) {
      setTableExpandedKeys([]);
      formInst.resetFields();
    }
  };

  const triggerEl = useTriggerElement(trigger, { onClick: () => setOpen(true) });

  const { form: formInst, formProps } = useAntdForm({
    name: "certficateDownloadAdvancedForm",
  });

  type TableRecord = {
    format: CertificateFormatType;
    advanced?: boolean;
    advancedForm?: React.ReactNode;
  };
  const tableColumns: TableProps<TableRecord>["columns"] = [
    {
      key: "format",
      title: t("certificate.action.download.modal.form.format.label"),
      render: (_, record) => (
        <div className="space-y-2">
          <div>{t(`certificate.action.download.modal.form.format.option.${record.format.toLowerCase()}.label`)}</div>
          <div>{t(`certificate.action.download.modal.form.format.option.${record.format.toLowerCase()}.description`)}</div>
        </div>
      ),
    },
    {
      key: "$action",
      align: "end",
      width: 32,
      render: (_, record) => (
        <div className="flex items-center justify-end">
          <Button color="primary" size="small" variant="link" onClick={() => handleDownloadClick(record.format)}>
            {t("common.button.download")}
          </Button>
          <Divider orientation="vertical" />
          <Button
            color="primary"
            disabled={!record.advanced}
            icon={tableExpandedKeys.includes(record.format) ? <IconChevronUp size="1.25em" /> : <IconChevronDown size="1.25em" />}
            iconPlacement="end"
            size="small"
            variant="link"
            onClick={() => {
              setTableExpandedKeys((prev) => (prev.includes(record.format) ? [] : [record.format]));
              formInst.resetFields();
            }}
          >
            {t("certificate.action.download.modal.action.advanced.button")}
          </Button>
        </div>
      ),
    },
  ];
  const tableData: TableRecord[] = [
    {
      format: CERTIFICATE_FORMATS.PEM,
    },
    {
      format: CERTIFICATE_FORMATS.PFX,
      advanced: true,
      advancedForm: (
        <Form form={formInst} layout="vertical" {...formProps}>
          <div className="flex gap-x-4 not-md:flex-wrap">
            <div className="w-1/2 not-md:w-full">
              <Form.Item name="pfxPassword" label={t("certificate.action.download.modal.form.pfx_password.label")}>
                <Input placeholder={t("certificate.action.download.modal.form.pfx_password.placeholder")} />
              </Form.Item>
            </div>
            <div className="w-1/2 not-md:w-full">
              <Form.Item name="pfxEncoder" label={t("certificate.action.download.modal.form.pfx_encoder.label")}>
                <Select
                  options={["LegacyRC2", "LegacyDES", "Modern2023", "Modern2026"].map((s) => ({
                    label: t(`certificate.action.download.modal.form.pfx_encoder.option.${s.toLowerCase()}.label`),
                    value: s,
                  }))}
                  placeholder={t("certificate.action.download.modal.form.pfx_encoder.placeholder")}
                />
              </Form.Item>
            </div>
          </div>
        </Form>
      ),
    },
    {
      format: CERTIFICATE_FORMATS.JKS,
      advanced: true,
      advancedForm: (
        <Form form={formInst} layout="vertical" {...formProps}>
          <div className="flex gap-x-4 not-md:flex-wrap">
            <div className="w-1/2 not-md:w-full">
              <Form.Item name="jksAlias" label={t("certificate.action.download.modal.form.jks_alias.label")}>
                <Input placeholder={t("certificate.action.download.modal.form.jks_alias.placeholder")} />
              </Form.Item>
            </div>
            <div className="w-1/2 not-md:w-full">
              <Form.Item name="jksKeypass" label={t("certificate.action.download.modal.form.jks_keypass.label")}>
                <Input placeholder={t("certificate.action.download.modal.form.jks_keypass.placeholder")} />
              </Form.Item>
            </div>
          </div>
          <div className="flex gap-x-4 not-md:flex-wrap">
            <div className="w-1/2 not-md:w-full">
              <Form.Item name="jksStorepass" label={t("certificate.action.download.modal.form.jks_storepass.label")}>
                <Input placeholder={t("certificate.action.download.modal.form.jks_storepass.placeholder")} />
              </Form.Item>
            </div>
          </div>
        </Form>
      ),
    },
  ];
  const [tableExpandedKeys, setTableExpandedKeys] = useState<string[]>([]);

  const handleCancelClick = () => {
    setOpen(false);
  };

  const handleDownloadClick = async (format: CertificateFormatType) => {
    await formInst.validateFields();

    try {
      const res = await downloadCertificate(data.id, format, formInst.getFieldsValue());
      const bstr = atob(res.data.zipBytes);
      const u8arr = Uint8Array.from(bstr, (ch) => ch.charCodeAt(0));
      const blob = new Blob([u8arr], { type: "application/zip" });
      saveAs(blob, `${data.subjectAltNames}_${data.id}_${format}.zip`.toLowerCase());

      setOpen(false);
    } catch (err) {
      console.error(err);
      message.warning(t("common.text.operation_failed"));
    }
  };

  return (
    <>
      {triggerEl}

      <Modal
        afterClose={afterClose}
        closable
        destroyOnHidden
        footer={null}
        open={open}
        title={t("certificate.action.download.modal.title")}
        width="768px"
        onCancel={handleCancelClick}
      >
        <div className="py-3 pb-0">
          <Table<TableRecord>
            columns={tableColumns}
            dataSource={tableData}
            expandable={{
              expandedRowKeys: tableExpandedKeys,
              expandedRowRender: (record) => record.advancedForm,
              rowExpandable: (record) => !!record.advanced,
              showExpandColumn: false,
            }}
            rowHoverable={false}
            pagination={false}
            rowKey={(record) => record.format}
            size="medium"
          />
        </div>
      </Modal>
    </>
  );
};

export default CertificateDownloadModal;
