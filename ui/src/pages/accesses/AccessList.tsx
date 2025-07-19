import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { IconCopy, IconEdit, IconPlus, IconReload, IconTrash } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { App, Avatar, Button, Empty, Input, Space, Table, type TableProps, Tabs, Tooltip, Typography } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import AccessEditDrawer, { type AccessEditDrawerProps } from "@/components/access/AccessEditDrawer";
import { type AccessModel } from "@/domain/access";
import { ACCESS_USAGES, accessProvidersMap } from "@/domain/provider";
import { useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { getErrMsg } from "@/utils/error";

type AccessUsageProp = AccessEditDrawerProps["usage"];

const AccessList = () => {
  const [searchParams] = useSearchParams();

  const { t } = useTranslation();

  const { modal, notification } = App.useApp();

  const { accesses, loadedAtOnce, fetchAccesses, deleteAccess } = useAccessesStore(
    useZustandShallowSelector(["accesses", "loadedAtOnce", "fetchAccesses", "deleteAccess"])
  );

  const tableColumns: TableProps<AccessModel>["columns"] = [
    {
      key: "$index",
      align: "center",
      fixed: "left",
      width: 50,
      render: (_, __, index) => (page - 1) * pageSize + index + 1,
    },
    {
      key: "name",
      title: t("access.props.name"),
      ellipsis: true,
      render: (_, record) => <>{record.name}</>,
    },
    {
      key: "provider",
      title: t("access.props.provider"),
      ellipsis: true,
      render: (_, record) => {
        return (
          <div className="flex max-w-full items-center gap-2 truncate overflow-hidden">
            <Avatar shape="square" src={accessProvidersMap.get(record.provider)?.icon} size="small" />
            <Typography.Text ellipsis>{t(accessProvidersMap.get(record.provider)?.name ?? "")}</Typography.Text>
          </div>
        );
      },
    },
    {
      key: "createdAt",
      title: t("access.props.created_at"),
      ellipsis: true,
      render: (_, record) => {
        return dayjs(record.created!).format("YYYY-MM-DD HH:mm:ss");
      },
    },
    {
      key: "updatedAt",
      title: t("access.props.updated_at"),
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
          <AccessEditDrawer
            data={record}
            usage={filters["usage"] as AccessUsageProp}
            scene="edit"
            trigger={
              <Tooltip title={t("access.action.edit")}>
                <Button color="primary" icon={<IconEdit size="1.25em" />} variant="text" />
              </Tooltip>
            }
          />

          <AccessEditDrawer
            data={{ ...record, id: undefined, name: `${record.name}-copy` }}
            usage={filters["usage"] as AccessUsageProp}
            scene="add"
            trigger={
              <Tooltip title={t("access.action.duplicate")}>
                <Button color="primary" icon={<IconCopy size="1.25em" />} variant="text" />
              </Tooltip>
            }
          />

          <Tooltip title={t("access.action.delete")}>
            <Button
              color="danger"
              icon={<IconTrash size="1.25em" />}
              variant="text"
              onClick={() => {
                handleDeleteClick(record);
              }}
            />
          </Tooltip>
        </Space.Compact>
      ),
    },
  ];
  const [tableData, setTableData] = useState<AccessModel[]>([]);
  const [tableTotal, setTableTotal] = useState<number>(0);

  const [filters, setFilters] = useState<Record<string, unknown>>(() => {
    return {
      usage: "both-dns-hosting" satisfies AccessUsageProp,
      keyword: searchParams.get("keyword"),
    };
  });

  const [page, setPage] = useState<number>(() => parseInt(+searchParams.get("page")! + "") || 1);
  const [pageSize, setPageSize] = useState<number>(() => parseInt(+searchParams.get("perPage")! + "") || 15);

  useEffect(() => {
    fetchAccesses().catch((err) => {
      if (err instanceof ClientResponseError && err.isAbort) {
        return;
      }

      console.error(err);
      notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
    });
  }, []);

  const { loading, run: refreshData } = useRequest(
    () => {
      const startIndex = (page - 1) * pageSize;
      const endIndex = startIndex + pageSize;
      const list = accesses
        .filter((e) => {
          const keyword = (filters["keyword"] as string | undefined)?.trim();
          if (keyword) {
            return e.name.includes(keyword);
          }

          return true;
        })
        .filter((e) => {
          const provider = accessProvidersMap.get(e.provider);
          switch (filters["usage"] as AccessUsageProp) {
            case "both-dns-hosting":
              return !e.reserve && (provider?.usages?.includes(ACCESS_USAGES.DNS) || provider?.usages?.includes(ACCESS_USAGES.HOSTING));
            case "ca-only":
              return e.reserve === "ca" && provider?.usages?.includes(ACCESS_USAGES.CA);
            case "notification-only":
              return e.reserve === "notification" && provider?.usages?.includes(ACCESS_USAGES.NOTIFICATION);
          }
        });
      return Promise.resolve({
        items: list.slice(startIndex, endIndex),
        totalItems: list.length,
      });
    },
    {
      refreshDeps: [accesses, filters, page, pageSize],
      onSuccess: (res) => {
        setTableData(res.items);
        setTableTotal(res.totalItems);
      },
    }
  );

  const handleTabChange = (key: string) => {
    setFilters((prev) => ({ ...prev, usage: key }));
    setPage(1);
  };

  const handleSearch = (value: string) => {
    setFilters((prev) => ({ ...prev, keyword: value }));
    setPage(1);
  };

  const handleReloadClick = () => {
    if (loading) return;

    fetchAccesses();
  };

  const handleDeleteClick = async (access: AccessModel) => {
    modal.confirm({
      title: <span className="text-error">{t("access.action.delete")}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("access.action.delete.confirm", { name: access.name }) }} />,
      icon: (
        <span className="anticon">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        // TODO: 有关联数据的不允许被删除
        try {
          await deleteAccess(access);
          refreshData();
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
        <h1>{t("access.page.title")}</h1>
        <p className="text-base text-gray-500">{t("access.page.subtitle")}</p>

        <div className="flex items-center justify-between gap-x-2 gap-y-3 not-md:flex-col-reverse not-md:items-start not-md:justify-normal">
          <div className="flex w-full flex-1 items-center gap-x-2 md:max-w-200">
            <div className="flex-1">
              <Input.Search
                className="text-sm placeholder:text-sm"
                allowClear
                defaultValue={filters["keyword"] as string}
                placeholder={t("access.search.placeholder")}
                size="large"
                onSearch={handleSearch}
              />
            </div>
            <div>
              <Button icon={<IconReload size="1.25em" />} size="large" onClick={handleReloadClick} />
            </div>
          </div>
          <div>
            <AccessEditDrawer
              usage={filters["usage"] as AccessUsageProp}
              scene="add"
              trigger={
                <Button className="text-sm" icon={<IconPlus size="1.25em" />} size="large" type="primary">
                  {t("access.action.create")}
                </Button>
              }
            />
          </div>
        </div>

        <Tabs
          className="mt-2 -mb-2"
          activeKey={filters["usage"] as string}
          items={[
            {
              key: "both-dns-hosting",
              label: t("access.props.usage.both_dns_hosting"),
            },
            {
              key: "ca-only",
              label: t("access.props.usage.ca_only"),
            },
            {
              key: "notification-only",
              label: t("access.props.usage.notification_only"),
            },
          ]}
          size="large"
          onChange={(key) => handleTabChange(key)}
        />
        <Table<AccessModel>
          columns={tableColumns}
          dataSource={tableData}
          loading={!loadedAtOnce || loading}
          locale={{
            emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t("access.nodata")} />,
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

export default AccessList;
