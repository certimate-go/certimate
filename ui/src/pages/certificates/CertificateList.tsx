import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { IconBrowserShare, IconCertificate, IconDots, IconExternalLink, IconReload, IconShieldCancel, IconTrash } from "@tabler/icons-react";
import { useMount, useRequest } from "ahooks";
import { App, Button, Dropdown, Input, Segmented, Skeleton, Table, type TableProps, Typography, theme } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import { revoke as revokeCertificate } from "@/api/certificates";
import CertificateDetailDrawer from "@/components/certificate/CertificateDetailDrawer";
import Empty from "@/components/Empty";
import Show from "@/components/Show";
import { CERTIFICATE_SOURCES, type CertificateModel } from "@/domain/certificate";
import { useAppSettings, useZustandShallowSelector } from "@/hooks";
import { get as getCertificate, list as listCertificates, remove as removeCertificate } from "@/repository/certificate";
import { usePersistenceSettingsStore } from "@/stores/settings";
import { getErrMsg } from "@/utils/error";

const CertificateList = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { message, modal, notification } = App.useApp();

  const { appSettings: globalAppSettings } = useAppSettings();

  const { settings: persistenceSettings, loadSettings: loadPersistenceSettings } = usePersistenceSettingsStore(
    useZustandShallowSelector(["settings", "loadSettings"])
  );
  useMount(() => loadPersistenceSettings(false));

  const [expiryThreshold, setExpiryThreshold] = useState(() => persistenceSettings.certificatesWarningDaysBeforeExpire || 0);
  useEffect(() => {
    setExpiryThreshold(persistenceSettings.certificatesWarningDaysBeforeExpire || 0);
  }, [persistenceSettings.certificatesWarningDaysBeforeExpire]);

  const [filters, setFilters] = useState<Record<string, unknown>>(() => {
    return {
      keyword: searchParams.get("keyword"),
      state: searchParams.get("state"),
    };
  });
  const [sorter, setSorter] = useState<ArrayElement<Parameters<NonNullable<TableProps<CertificateModel>["onChange"]>>[2]>>(() => {
    return {};
  });
  const [page, setPage] = useState<number>(() => parseInt(+searchParams.get("page")! + "") || 1);
  const [pageSize, setPageSize] = useState<number>(() => parseInt(+searchParams.get("perPage")! + "") || globalAppSettings.defaultPerPage!);

  const [tableData, setTableData] = useState<CertificateModel[]>([]);
  const [tableTotal, setTableTotal] = useState<number>(0);
  const [tableSelectedRowKeys, setTableSelectedRowKeys] = useState<string[]>([]);
  const tableColumns: TableProps<CertificateModel>["columns"] = [
    {
      key: "name",
      title: t("certificate.props.subject_alt_names"),
      render: (_, record) => <Typography.Text delete={record.isRevoked}>{record.subjectAltNames}</Typography.Text>,
    },
    {
      key: "validity",
      title: t("certificate.props.validity"),
      sorter: true,
      sortOrder: sorter.columnKey === "validity" ? sorter.order : void 0,
      render: (_, record) => {
        const total = dayjs(record.validityNotAfter).diff(dayjs(record.created), "d") + 1;
        const isRevoked = record.isRevoked;
        const isExpired = dayjs().isAfter(dayjs(record.validityNotAfter));
        const leftHours = dayjs(record.validityNotAfter).diff(dayjs(), "h");
        const leftDays = Math.round(leftHours / 24);

        return (
          <div className="flex max-w-full flex-col gap-1 truncate">
            {!isRevoked && !isExpired ? (
              leftDays >= expiryThreshold ? (
                <Typography.Text ellipsis type="success">
                  <span className="mr-2 inline-block size-2 rounded-full bg-success leading-2">&nbsp;</span>
                  {t("certificate.props.validity.left_days", { left: leftDays, total })}
                </Typography.Text>
              ) : (
                <Typography.Text ellipsis type="warning">
                  <span className="mr-2 inline-block size-2 rounded-full bg-warning leading-2">&nbsp;</span>
                  {leftDays >= 1
                    ? t("certificate.props.validity.left_days", { left: leftDays, total })
                    : t("certificate.props.validity.less_than_a_day", { total })}
                </Typography.Text>
              )
            ) : (
              <Typography.Text ellipsis type="danger">
                <span className="mr-2 inline-block size-2 rounded-full bg-error leading-2">&nbsp;</span>
                {isRevoked ? t("certificate.props.revoked") : t("certificate.props.validity.expired")}
              </Typography.Text>
            )}

            <Typography.Text ellipsis type="secondary">
              {t("certificate.props.validity.expiration", { date: dayjs(record.validityNotAfter).format("YYYY-MM-DD") })}
            </Typography.Text>
          </div>
        );
      },
    },
    {
      key: "brand",
      title: t("certificate.props.brand"),
      render: (_, record) => (
        <div className="flex max-w-full flex-col gap-1 truncate">
          <Typography.Text ellipsis>{record.issuerOrg || "\u00A0"}</Typography.Text>
          <Typography.Text ellipsis>{record.keyAlgorithm || "\u00A0"}</Typography.Text>
        </div>
      ),
    },
    {
      key: "source",
      title: t("certificate.props.source"),
      render: (_, record) => {
        const workflowId = record.workflowRef;
        return (
          <div className="flex max-w-full flex-col gap-1 truncate">
            <Typography.Text ellipsis>{t(`certificate.props.source.${record.source}`)}</Typography.Text>
            <Typography.Link
              ellipsis
              type="secondary"
              onClick={(e) => {
                e.stopPropagation();
                if (workflowId) {
                  navigate(`/workflows/${workflowId}`);
                }
              }}
            >
              {record.expand?.workflowRef?.name ?? <span className="font-mono">{t(`#${workflowId}`)}</span>}
            </Typography.Link>
          </div>
        );
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
      key: "$action",
      align: "end",
      fixed: "right",
      width: 64,
      render: (_, record) => (
        <Dropdown
          menu={{
            items: [
              {
                key: "view",
                label: t("certificate.action.view.menu"),
                icon: (
                  <span className="anticon scale-125">
                    <IconBrowserShare size="1em" />
                  </span>
                ),
                onClick: () => {
                  handleRecordDetailClick(record);
                },
              },
              {
                key: "revoke",
                label: t("certificate.action.revoke.menu"),
                danger: true,
                disabled: record.source !== CERTIFICATE_SOURCES.REQUEST || record.isRevoked,
                icon: (
                  <span className="anticon scale-125">
                    <IconShieldCancel size="1em" />
                  </span>
                ),
                onClick: () => {
                  handleRecordRevokeClick(record);
                },
              },
              {
                type: "divider",
              },
              {
                key: "delete",
                label: t("certificate.action.delete.menu"),
                danger: true,
                icon: (
                  <span className="anticon scale-125">
                    <IconTrash size="1em" />
                  </span>
                ),
                onClick: () => {
                  handleRecordDeleteClick(record);
                },
              },
            ],
          }}
          trigger={["click"]}
        >
          <Button icon={<IconDots size="1.25em" />} type="text" />
        </Dropdown>
      ),
      onCell: () => {
        return {
          onClick: (e) => {
            e.stopPropagation();
          },
        };
      },
    },
  ];
  const tableRowSelection: TableProps<CertificateModel>["rowSelection"] = {
    fixed: true,
    selectedRowKeys: tableSelectedRowKeys,
    renderCell(checked, _, index, node) {
      if (!checked) {
        return (
          <div className="group/selection">
            <div className="group-hover/selection:hidden">{(page - 1) * pageSize + index + 1}</div>
            <div className="hidden group-hover/selection:block">{node}</div>
          </div>
        );
      }
      return node;
    },
    onCell: () => {
      return {
        onClick: (e) => {
          e.stopPropagation();
        },
      };
    },
    onChange: (keys) => {
      setTableSelectedRowKeys(keys as string[]);
    },
  };

  const {
    loading,
    error: loadError,
    run: refreshData,
  } = useRequest(
    () => {
      const { columnKey: sorterKey, order: sorterOrder } = sorter;
      let sort: string | undefined;
      sort = sorterKey === "validity" ? "validityNotAfter" : void 0;
      sort = sort && (sorterOrder === "ascend" ? `${sort}` : sorterOrder === "descend" ? `-${sort}` : void 0);

      return listCertificates({
        keyword: filters["keyword"] as string,
        state: filters["state"] as Parameters<typeof listCertificates>[0]["state"],
        stateThreshold: expiryThreshold,
        sort: sort,
        page: page,
        perPage: pageSize,
      });
    },
    {
      refreshDeps: [expiryThreshold, filters, sorter, page, pageSize],
      onBefore: async () => {
        setSearchParams((prev) => {
          if (filters["keyword"]) {
            prev.set("keyword", filters["keyword"] as string);
          } else {
            prev.delete("keyword");
          }

          if (filters["state"]) {
            prev.set("state", filters["state"] as string);
          } else {
            prev.delete("state");
          }

          prev.set("page", page.toString());
          prev.set("perPage", pageSize.toString());

          return prev;
        });
      },
      onSuccess: (res) => {
        setTableData(res.items);
        setTableTotal(res.totalItems);
        setTableSelectedRowKeys([]);
      },
      onError: (err) => {
        if (err instanceof ClientResponseError && err.isAbort) {
          return;
        }

        console.error(err);
        notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });

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

  const handlePaginationChange = (page: number, pageSize: number) => {
    setPage(page);
    setPageSize(pageSize);
  };

  const { drawerProps: detailDrawerProps, ...detailDrawer } = CertificateDetailDrawer.useDrawer();

  const handleRecordDetailClick = (certificate: CertificateModel) => {
    const drawer = detailDrawer.open({ data: certificate, loading: true });
    getCertificate(certificate.id).then((data) => {
      drawer.safeUpdate({ data, loading: false });
    });
  };

  const handleRecordRevokeClick = (certificate: CertificateModel) => {
    modal.confirm({
      title: <span className="text-error">{t("certificate.action.revoke.modal.title", { name: certificate.subjectAltNames })}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("certificate.action.revoke.modal.content") }} />,
      icon: (
        <span className="anticon" role="img">
          <IconShieldCancel className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await revokeCertificate(certificate.id);
          if (resp) {
            message.success(t("common.text.operation_succeeded"));
            setTableData((prev) => prev.map((item) => (item.id === certificate.id ? { ...item, isRevoked: true } : item)));
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handleRecordDeleteClick = (certificate: CertificateModel) => {
    modal.confirm({
      title: <span className="text-error">{t("certificate.action.delete.modal.title", { name: certificate.subjectAltNames })}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("certificate.action.delete.modal.content") }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await removeCertificate(certificate);
          if (resp) {
            setTableSelectedRowKeys((prev) => prev.filter((key) => key !== certificate.id));
            setTableData((prev) => prev.filter((item) => item.id !== certificate.id));
            setTableTotal((prev) => prev - 1);
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handleBatchDeleteClick = () => {
    const records = tableData.filter((item) => tableSelectedRowKeys.includes(item.id));
    if (records.length === 0) {
      return;
    }

    modal.confirm({
      title: <span className="text-error">{t("certificate.action.batch_delete.modal.title")}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("certificate.action.batch_delete.modal.content", { count: records.length }) }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await removeCertificate(records);
          if (resp) {
            setTableData((prev) => prev.filter((item) => !records.some((record) => record.id === item.id)));
            setTableTotal((prev) => prev - records.length);
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("certificate.page.title")}</h1>
        <p className="text-base text-gray-500">{t("certificate.page.subtitle")}</p>
      </div>

      <div className="container">
        <div className="flex items-center justify-between gap-x-2 gap-y-3 not-md:flex-col-reverse not-md:items-start not-md:justify-normal">
          <div className="flex w-full flex-1 items-center gap-x-2 md:max-w-200">
            <div>
              <Segmented
                options={[
                  { label: <span className="text-sm">{t("certificate.props.validity.filter.all")}</span>, value: "" },
                  { label: <span className="text-sm">{t("certificate.props.validity.filter.expiring_soon")}</span>, value: "expiringSoon" },
                  { label: <span className="text-sm">{t("certificate.props.validity.filter.expired")}</span>, value: "expired" },
                ]}
                size="large"
                value={(filters["state"] as string) || ""}
                onChange={(value) => {
                  setPage(1);
                  setFilters((prev) => ({ ...prev, state: value }));
                }}
              />
            </div>
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

        <div className="relative mt-4">
          <Table<CertificateModel>
            columns={tableColumns}
            dataSource={tableData}
            loading={loading}
            locale={{
              emptyText: loading ? (
                <Skeleton />
              ) : (
                <Empty
                  className="py-24"
                  title={t("certificate.nodata.title")}
                  description={loadError ? getErrMsg(loadError) : t("certificate.nodata.description")}
                  icon={<IconCertificate size={24} />}
                  extra={
                    loadError ? (
                      <Button ghost icon={<IconReload size="1.25em" />} type="primary" onClick={handleReloadClick}>
                        {t("common.button.reload")}
                      </Button>
                    ) : (
                      <Button icon={<IconExternalLink size="1.25em" />} type="primary" onClick={() => navigate("/workflows")}>
                        {t("certificate.nodata.button")}
                      </Button>
                    )
                  }
                />
              ),
            }}
            pagination={{
              current: page,
              pageSize: pageSize,
              total: tableTotal,
              showSizeChanger: true,
              onChange: handlePaginationChange,
              onShowSizeChange: handlePaginationChange,
            }}
            rowClassName="cursor-pointer"
            rowKey={(record) => record.id}
            rowSelection={tableRowSelection}
            scroll={{ x: "max(100%, 960px)" }}
            onChange={(_, __, sorter) => {
              setSorter(Array.isArray(sorter) ? sorter[0] : sorter);
            }}
            onRow={(record) => ({
              onClick: () => {
                handleRecordDetailClick(record);
              },
            })}
          />

          <Show when={tableSelectedRowKeys.length > 0}>
            <div
              className="absolute top-0 right-0 left-[32px] z-10 h-[54px]"
              style={{
                left: "32px", // Match the width of the table row selection checkbox
                height: "54px", // Match the height of the table header
                background: themeToken.Table?.headerBg ?? themeToken.colorBgElevated,
              }}
            >
              <div className="flex size-full items-center justify-end gap-x-2 overflow-hidden px-4 py-2">
                <Button danger ghost onClick={handleBatchDeleteClick}>
                  {t("common.button.delete")}
                </Button>
              </div>
            </div>
          </Show>
        </div>

        <CertificateDetailDrawer {...detailDrawerProps} />
      </div>
    </div>
  );
};

export default CertificateList;
