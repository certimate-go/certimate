import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconDownload, IconPlus, IconRefresh } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { App, Button, Empty, Form, Input, Modal, Select, Space, Table, type TableProps, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import dayjs from "dayjs";
import { saveAs } from "file-saver";
import { z } from "zod";

import { exportAccount, importAccount, list as listAcmeAccounts, rotateAccount } from "@/api/acmeAccounts";
import Show from "@/components/Show";
import { type ACMEAccountView } from "@/domain/acmeAccount";
import { CA_PROVIDERS, caProvidersMap } from "@/domain/provider";
import { useAntdForm, useAppSettings } from "@/hooks";
import { unwrapErrMsg } from "@/utils/error";

const SettingsAcmeAccounts = () => {
  const { t } = useTranslation();
  const { message, modal, notification } = App.useApp();
  const { appSettings: globalAppSettings } = useAppSettings();

  const [caFilter, setCaFilter] = useState<string | undefined>();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(() => globalAppSettings.defaultPerPage ?? 15);

  const {
    data: listData,
    loading,
    refresh,
  } = useRequest(
    async () => {
      const res = await listAcmeAccounts({
        page,
        perPage: pageSize,
        ca: caFilter || undefined,
      });
      return {
        items: res.data?.items ?? [],
        totalItems: res.data?.totalItems ?? 0,
      };
    },
    {
      refreshDeps: [page, pageSize, caFilter],
      onError: (err) => {
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
      },
    }
  );

  const tableData = listData?.items ?? [];
  const tableTotal = listData?.totalItems ?? 0;

  const [importOpen, setImportOpen] = useState(false);
  const [exportPem, setExportPem] = useState<{ id: string; pem: string } | null>(null);

  const caOptions = useMemo(() => {
    return Array.from(caProvidersMap.values()).map((p) => ({
      value: p.type,
      label: t(p.name),
    }));
  }, [t]);

  const importSchema = z
    .object({
      privateKeyPem: z.string().min(1, t("common.errmsg.form_invalid")),
      ca: z.string().min(1, t("common.errmsg.form_invalid")),
      acmeDirUrl: z.string().optional(),
      email: z.union([z.string().email(), z.literal("")]).optional(),
    })
    .superRefine((val, ctx) => {
      if (val.ca === CA_PROVIDERS.ACMECA && !val.acmeDirUrl?.trim()) {
        ctx.addIssue({ code: "custom", message: t("common.errmsg.form_invalid"), path: ["acmeDirUrl"] });
      }
    });
  const importRule = createSchemaFieldRule(importSchema);
  const {
    form: importForm,
    formPending: importPending,
    formProps: importFormProps,
  } = useAntdForm<z.infer<typeof importSchema>>({
    onSubmit: async (values) => {
      try {
        await importAccount({
          privateKeyPem: values.privateKeyPem,
          ca: values.ca,
          acmeDirUrl: values.acmeDirUrl || undefined,
          email: values.email || undefined,
        });
        message.success(t("common.text.operation_succeeded"));
        setImportOpen(false);
        importForm.resetFields();
        setPage(1);
        refresh();
      } catch (err) {
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
      }
    },
  });

  const caWatch = Form.useWatch("ca", importForm);

  const handleExport = (record: ACMEAccountView) => {
    modal.confirm({
      title: t("settings.acmeaccounts.export.confirm.title"),
      content: t("settings.acmeaccounts.export.confirm.content"),
      okText: t("common.button.confirm"),
      cancelText: t("common.button.cancel"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const res = await exportAccount(record.id);
          setExportPem({ id: record.id, pem: res.data.privateKeyPem });
        } catch (err) {
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
        }
      },
    });
  };

  const handleRotate = (record: ACMEAccountView) => {
    modal.confirm({
      title: t("settings.acmeaccounts.rotate.confirm.title"),
      content: t("settings.acmeaccounts.rotate.confirm.content"),
      okText: t("common.button.confirm"),
      cancelText: t("common.button.cancel"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await rotateAccount(record.id);
          message.success(t("common.text.operation_succeeded"));
          refresh();
        } catch (err) {
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
        }
      },
    });
  };

  const handlePaginationChange = (nextPage: number, nextPageSize: number) => {
    setPage(nextPage);
    setPageSize(nextPageSize);
  };

  const handleCaFilterChange = (value: string | undefined) => {
    setCaFilter(value);
    setPage(1);
  };

  const providerLabel = (ca: string) => {
    const p = caProvidersMap.get(ca);
    return p ? t(p.name) : ca;
  };

  const columns: TableProps<ACMEAccountView>["columns"] = [
    {
      key: "ca",
      title: t("settings.acmeaccounts.table.ca"),
      dataIndex: "ca",
      width: 180,
      ellipsis: true,
      render: (v: string) => providerLabel(v),
    },
    {
      key: "email",
      title: t("settings.acmeaccounts.table.email"),
      dataIndex: "email",
      ellipsis: true,
    },
    {
      key: "acmeAcctUrl",
      title: t("settings.acmeaccounts.table.account_uri"),
      dataIndex: "acmeAcctUrl",
      ellipsis: true,
      render: (v: string) => (
        <Typography.Text copyable={{ text: v }} ellipsis className="max-w-full">
          {v}
        </Typography.Text>
      ),
    },
    {
      key: "acmeDirUrl",
      title: t("settings.acmeaccounts.table.directory"),
      dataIndex: "acmeDirUrl",
      ellipsis: true,
      responsive: ["lg"],
    },
    {
      key: "created",
      title: t("settings.acmeaccounts.table.created"),
      dataIndex: "created",
      width: 160,
      responsive: ["xl"],
      render: (v: string) => (v ? dayjs(v).format("YYYY-MM-DD HH:mm") : "-"),
    },
    {
      key: "$action",
      title: t("settings.acmeaccounts.table.actions"),
      width: 200,
      align: "end",
      fixed: "right",
      render: (_, record) => (
        <Space size="small">
          <Button type="link" size="small" icon={<IconDownload size="1em" />} onClick={() => handleExport(record)}>
            {t("settings.acmeaccounts.action.export")}
          </Button>
          <Button type="link" size="small" danger icon={<IconRefresh size="1em" />} onClick={() => handleRotate(record)}>
            {t("settings.acmeaccounts.action.rotate")}
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 className="mb-1!">{t("settings.acmeaccounts.title")}</h2>
          <Typography.Paragraph type="secondary" className="mb-0!">
            {t("settings.acmeaccounts.tips")}
          </Typography.Paragraph>
        </div>
        <Space wrap>
          <Select
            allowClear
            className="min-w-48"
            options={caOptions}
            placeholder={t("settings.acmeaccounts.filter.ca_placeholder")}
            showSearch
            optionFilterProp="label"
            value={caFilter}
            onChange={handleCaFilterChange}
          />
          <Button icon={<IconRefresh size="1.1em" />} onClick={() => refresh()} loading={loading}>
            {t("common.button.reload")}
          </Button>
          <Button type="primary" icon={<IconPlus size="1.1em" />} onClick={() => setImportOpen(true)}>
            {t("settings.acmeaccounts.action.import")}
          </Button>
        </Space>
      </div>

      <Table
        rowKey="id"
        size="small"
        loading={loading}
        columns={columns}
        dataSource={tableData}
        scroll={{ x: "max(100%, 960px)" }}
        locale={{
          emptyText: <Empty description={!loading ? t("settings.acmeaccounts.empty") : " "} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
        }}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: tableTotal,
          showSizeChanger: true,
          onChange: handlePaginationChange,
          onShowSizeChange: handlePaginationChange,
        }}
      />

      <Modal
        title={t("settings.acmeaccounts.import.modal.title")}
        open={importOpen}
        onCancel={() => {
          setImportOpen(false);
          importForm.resetFields();
        }}
        onOk={() => importForm.submit()}
        confirmLoading={importPending}
        destroyOnHidden
        width={640}
      >
        <Form {...importFormProps} form={importForm} layout="vertical" className="mt-4">
          <Form.Item name="ca" label={t("settings.acmeaccounts.import.form.ca")} rules={[importRule]}>
            <Select options={caOptions} showSearch optionFilterProp="label" placeholder={t("settings.acmeaccounts.import.form.ca_placeholder")} />
          </Form.Item>
          <Show when={caWatch === CA_PROVIDERS.ACMECA}>
            <Form.Item name="acmeDirUrl" label={t("settings.acmeaccounts.import.form.directory")} rules={[importRule]}>
              <Input placeholder={t("settings.acmeaccounts.import.form.directory_placeholder")} />
            </Form.Item>
          </Show>
          <Form.Item
            name="privateKeyPem"
            label={t("settings.acmeaccounts.import.form.private_key")}
            rules={[importRule]}
            extra={t("settings.acmeaccounts.import.form.private_key_extra")}
          >
            <Input.TextArea autoSize={{ minRows: 6, maxRows: 12 }} placeholder={t("settings.acmeaccounts.import.form.private_key_placeholder")} />
          </Form.Item>
          <Form.Item
            name="email"
            label={t("settings.acmeaccounts.import.form.email")}
            rules={[importRule]}
            extra={t("settings.acmeaccounts.import.form.email_extra")}
          >
            <Input placeholder={t("settings.acmeaccounts.import.form.email_placeholder")} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t("settings.acmeaccounts.export.modal.title")}
        open={!!exportPem}
        onCancel={() => setExportPem(null)}
        footer={[
          <Button
            key="copy"
            onClick={async () => {
              if (!exportPem) return;
              await navigator.clipboard.writeText(exportPem.pem);
              message.success(t("common.text.copied"));
            }}
          >
            {t("common.button.copy")}
          </Button>,
          <Button
            key="download"
            type="primary"
            onClick={() => {
              if (!exportPem) return;
              const blob = new Blob([exportPem.pem], { type: "application/x-pem-file" });
              saveAs(blob, `acme-account-${exportPem.id}.pem`);
            }}
          >
            {t("common.button.download")}
          </Button>,
        ]}
        width={640}
      >
        <Typography.Paragraph type="secondary">{t("settings.acmeaccounts.export.modal.tips")}</Typography.Paragraph>
        <Input.TextArea value={exportPem?.pem ?? ""} readOnly autoSize={{ minRows: 8, maxRows: 16 }} />
      </Modal>
    </div>
  );
};

export default SettingsAcmeAccounts;
