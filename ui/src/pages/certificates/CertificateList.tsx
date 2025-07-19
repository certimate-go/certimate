import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { IconBrowserShare, IconReload, IconTrash } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { App, Button, Divider, Empty, Input, Menu, type MenuProps, Radio, Space, Table, type TableProps, Tooltip, Typography, theme } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import CertificateDetailDrawer from "@/components/certificate/CertificateDetailDrawer";
import { CERTIFICATE_SOURCES, type CertificateModel } from "@/domain/certificate";
import { list as listCertificates, type ListRequest as listCertificatesRequest, remove as removeCertificate } from "@/repository/certificate";
import { getErrMsg } from "@/utils/error";

const CertificateList = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const { t } = useTranslation();

  const { modal, notification } = App.useApp();
  const { token: themeToken } = theme.useToken();

  const tableColumns: TableProps<CertificateModel>["columns"] = [
    {
      key: "$index",
      align: "center",
      fixed: "left",
      width: 50,
      render: (_, __, index) => (page - 1) * pageSize + index + 1,
    },
    {
      key: "name",
      title: t("certificate.props.subject_alt_names"),
      render: (_, record) => <Typography.Text>{record.subjectAltNames}</Typography.Text>,
    },
    {
      key: "expiry",
      title: t("certificate.props.validity"),
      ellipsis: true,
      defaultFilteredValue: searchParams.has("state") ? [searchParams.get("state") as string] : undefined,
      filterDropdown: ({ setSelectedKeys, confirm, clearFilters }) => {
        const items: Required<MenuProps>["items"] = [
          ["expireSoon", "certificate.props.validity.filter.expire_soon"],
          ["expired", "certificate.props.validity.filter.expired"],
        ].map(([key, label]) => {
          return {
            key,
            label: <Radio checked={filters["state"] === key}>{t(label)}</Radio>,
            onClick: () => {
              if (filters["state"] !== key) {
                setPage(1);
                setFilters((prev) => ({ ...prev, state: key }));
                setSelectedKeys([key]);
              }

              confirm({ closeDropdown: true });
            },
          };
        });

        const handleResetClick = () => {
          setPage(1);
          setFilters((prev) => ({ ...prev, state: undefined }));
          setSelectedKeys([]);
          clearFilters?.();
          confirm();
        };

        const handleConfirmClick = () => {
          confirm();
        };

        return (
          <div style={{ padding: 0 }}>
            <Menu items={items} selectable={false} />
            <Divider className="my-0" />
            <Space className="w-full justify-end" style={{ padding: themeToken.paddingSM }}>
              <Button size="small" disabled={!filters.state} onClick={handleResetClick}>
                {t("common.button.reset")}
              </Button>
              <Button type="primary" size="small" onClick={handleConfirmClick}>
                {t("common.button.ok")}
              </Button>
            </Space>
          </div>
        );
      },
      render: (_, record) => {
        const total = dayjs(record.expireAt).diff(dayjs(record.created), "d") + 1;
        // 使用 isAfter 更精确地判断是否过期
        const isExpired = dayjs().isAfter(dayjs(record.expireAt));
        const leftDays = dayjs(record.expireAt).diff(dayjs(), "d");
        const leftHours = dayjs(record.expireAt).diff(dayjs(), "h");

        return (
          <Space className="max-w-full" direction="vertical" size={4}>
            {!isExpired ? (
              leftDays > 0 ? (
                <Typography.Text type="success">{t("certificate.props.validity.left_days", { left: leftDays, total })}</Typography.Text>
              ) : (
                <Typography.Text type="warning">{t("certificate.props.validity.less_than_day", { hours: leftHours > 0 ? leftHours : 1 })}</Typography.Text>
              )
            ) : (
              <Typography.Text type="danger">{t("certificate.props.validity.expired")}</Typography.Text>
            )}

            <Typography.Text type="secondary">
              {t("certificate.props.validity.expiration", { date: dayjs(record.expireAt).format("YYYY-MM-DD") })}
            </Typography.Text>
          </Space>
        );
      },
    },
    {
      key: "brand",
      title: t("certificate.props.brand"),
      render: (_, record) => (
        <Space className="max-w-full" direction="vertical" size={4}>
          <Typography.Text>{record.issuerOrg}</Typography.Text>
          <Typography.Text>{record.keyAlgorithm}</Typography.Text>
        </Space>
      ),
    },
    {
      key: "source",
      title: t("certificate.props.source"),
      ellipsis: true,
      render: (_, record) => {
        if (record.source === CERTIFICATE_SOURCES.WORKFLOW) {
          const workflowId = record.workflowId;
          return (
            <Space className="max-w-full" direction="vertical" size={4}>
              <Typography.Text>{t("certificate.props.source.workflow")}</Typography.Text>
              <Typography.Link
                type="secondary"
                ellipsis
                onClick={() => {
                  if (workflowId) {
                    navigate(`/workflows/${workflowId}`);
                  }
                }}
              >
                {record.expand?.workflowId?.name ?? <span className="font-mono">{t(`#${workflowId}`)}</span>}
              </Typography.Link>
            </Space>
          );
        } else if (record.source === CERTIFICATE_SOURCES.UPLOAD) {
          return <Typography.Text>{t("certificate.props.source.upload")}</Typography.Text>;
        }

        return <></>;
      },
    },
    {
      key: "createdAt",
      title: t("certificate.props.created_at"),
      ellipsis: true,
      render: (_, record) => {
        return dayjs(record.created!).format("YYYY-MM-DD HH:mm:ss");
      },
    },
    {
      key: "updatedAt",
      title: t("certificate.props.updated_at"),
      ellipsis: true,
      render: (_, record) => {
        return dayjs(record.updated!).format("YYYY-MM-DD HH:mm:ss");
      },
    },
    {
      key: "$action",
      align: "end",
      fixed: "right",
      width: 120,
      render: (_, record) => (
        <Space.Compact>
          <CertificateDetailDrawer
            data={record}
            trigger={
              <Tooltip title={t("certificate.action.view")}>
                <Button color="primary" icon={<IconBrowserShare size="1.25em" />} variant="text" />
              </Tooltip>
            }
          />

          <Tooltip title={t("certificate.action.delete")}>
            <Button color="danger" icon={<IconTrash size="1.25em" />} variant="text" onClick={() => handleDeleteClick(record)} />
          </Tooltip>
        </Space.Compact>
      ),
    },
  ];
  const [tableData, setTableData] = useState<CertificateModel[]>([]);
  const [tableTotal, setTableTotal] = useState<number>(0);

  const [filters, setFilters] = useState<Record<string, unknown>>(() => {
    return {
      keyword: searchParams.get("keyword"),
      state: searchParams.get("state"),
    };
  });

  const [page, setPage] = useState<number>(() => parseInt(+searchParams.get("page")! + "") || 1);
  const [pageSize, setPageSize] = useState<number>(() => parseInt(+searchParams.get("perPage")! + "") || 15);

  const {
    loading,
    error: loadedError,
    run: refreshData,
  } = useRequest(
    () => {
      return listCertificates({
        keyword: filters["keyword"] as string,
        state: filters["state"] as listCertificatesRequest["state"],
        page: page,
        perPage: pageSize,
      });
    },
    {
      refreshDeps: [filters, page, pageSize],
      onSuccess: (res) => {
        setTableData(res.items);
        setTableTotal(res.totalItems);
      },
      onError: (err) => {
        if (err instanceof ClientResponseError && err.isAbort) {
          return;
        }

        console.error(err);
        notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      },
    }
  );

  const handleSearch = (value: string) => {
    setFilters((prev) => ({ ...prev, keyword: value.trim() }));
    setPage(1);
  };

  const handleReloadClick = () => {
    if (loading) return;

    refreshData();
  };

  const handleDeleteClick = (certificate: CertificateModel) => {
    modal.confirm({
      title: <span className="text-error">{t("certificate.action.delete")}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("certificate.action.delete.confirm", { name: certificate.subjectAltNames }) }} />,
      icon: (
        <span className="anticon">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await removeCertificate(certificate);
          if (resp) {
            setTableData((prev) => prev.filter((item) => item.id !== certificate.id));
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  return (
    <div className="px-6 py-4">
      <div className="mx-auto max-w-320">
        <h1>{t("certificate.page.title")}</h1>
        <p className="text-base text-gray-500">{t("certificate.page.subtitle")}</p>

        <div className="flex items-center justify-between gap-x-2 gap-y-3 not-md:flex-col-reverse not-md:items-start not-md:justify-normal">
          <div className="flex w-full flex-1 items-center gap-x-2 md:max-w-200">
            <div className="flex-1">
              <Input.Search
                className="text-sm placeholder:text-sm"
                allowClear
                defaultValue={filters["keyword"] as string}
                placeholder={t("certificate.search.placeholder")}
                size="large"
                onSearch={handleSearch}
              />
            </div>
            <div>
              <Button icon={<IconReload size="1.25em" />} size="large" onClick={handleReloadClick} />
            </div>
          </div>
          <div></div>
        </div>

        <Table<CertificateModel>
          className="mt-4"
          columns={tableColumns}
          dataSource={tableData}
          loading={loading}
          locale={{
            emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={getErrMsg(loadedError ?? t("certificate.nodata"))} />,
          }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: tableTotal,
            showSizeChanger: true,
            onChange: (page: number, pageSize: number) => {
              setPage(page);
              setPageSize(pageSize);
            },
            onShowSizeChange: (page: number, pageSize: number) => {
              setPage(page);
              setPageSize(pageSize);
            },
          }}
          rowKey={(record) => record.id}
          scroll={{ x: "max(100%, 960px)" }}
        />
      </div>
    </div>
  );
};

export default CertificateList;
