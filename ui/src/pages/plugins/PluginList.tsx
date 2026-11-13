import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconDownload, IconPuzzle, IconRefresh, IconTrash } from "@tabler/icons-react";
import { Alert, App, Button, Card, Empty, Input, Pagination, Progress, Segmented, Skeleton, Tag, Tooltip } from "antd";

import {
  type InstallJobStatus,
  type MarketCapabilityFilter,
  type MarketEntry,
  type MarketFilter,
  type MarketPage,
  deletePlugin,
  fetchMarketListing,
  getInstallStatus,
  installPlugin,
  isTerminalState,
  updatePlugin,
} from "@/api/pluginmarket";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

const POLL_MS = 1000;
const MAX_POLL_DELAY_MS = 30000;

const statusLabel: Record<string, string> = {
  not_installed: "plugin.market.status.not_installed",
  installed: "plugin.market.status.installed",
  update_available: "plugin.market.status.update_available",
  installed_manual: "plugin.market.status.installed_manual",
  unsupported_platform: "plugin.market.status.unsupported_platform",
};

const statusColor: Record<string, string> = {
  not_installed: "bg-gray-400",
  installed: "bg-green-500",
  update_available: "bg-blue-500",
  installed_manual: "bg-gray-400",
  unsupported_platform: "bg-gray-400",
};

const jobStageLabelKey = (state: InstallJobStatus["state"]): string => {
  switch (state) {
    case "queued":
      return "plugin.market.job.queued";
    case "downloading":
      return "plugin.market.job.downloading";
    case "verifying":
      return "plugin.market.job.verifying";
    case "extracting":
      return "plugin.market.job.extracting";
    case "reloading":
      return "plugin.market.job.reloading";
    default:
      return "plugin.market.job.installing";
  }
};

const formatBytes = (n: number): string => {
  if (!n || n <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let value = n;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024;
    i++;
  }
  const digits = value >= 100 || i === 0 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[i]}`;
};

const PluginList = () => {
  const { t } = useTranslation();

  return (
    <div className="px-6 py-4">
      <div className="mx-auto max-w-[1920px]">
        <h1>{t("plugin.page.heading")}</h1>
      </div>

      <div className="mx-auto max-w-[1920px]">
        <PluginMarket />
      </div>
    </div>
  );
};

const PluginMarket = () => {
  const { t, i18n } = useTranslation();
  const { message, modal } = App.useApp();
  const catalogReload = usePluginCatalogStore((s) => s.reload);

  const [listing, setListing] = useState<MarketPage>();
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(30);
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [operating, setOperating] = useState<Record<string, boolean>>({});
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] = useState<MarketFilter>("all");
  const [capabilityFilter, setCapabilityFilter] = useState<MarketCapabilityFilter>("all");
  const [jobs, setJobs] = useState<Record<string, InstallJobStatus>>({});
  const [speeds, setSpeeds] = useState<Record<string, number>>({});
  const [pollingErrors, setPollingErrors] = useState<Record<string, string>>({});
  const pollingRef = useRef<Set<string>>(new Set());
  const pollingRequestsRef = useRef<Set<string>>(new Set());
  const pollingRetryRef = useRef<Record<string, { attempts: number; nextAt: number }>>({});
  const lastSampleRef = useRef<Record<string, { bytes: number; t: number }>>({});

  const listingRequestRef = useRef({ generation: 0 });

  const loadListing = useCallback(
    async (refresh = false, quiet = false) => {
      const request = ++listingRequestRef.current.generation;
      if (!quiet) setLoading(true);
      setError(null);
      try {
        const data = await fetchMarketListing({ page, perPage, keyword, status: statusFilter, capability: capabilityFilter, locale: i18n.language, refresh });
        if (request === listingRequestRef.current.generation) setListing(data);
      } catch {
        if (request === listingRequestRef.current.generation) setError(t("plugin.market.error.load_failed"));
      } finally {
        if (request === listingRequestRef.current.generation) setLoading(false);
      }
    },
    [t, page, perPage, keyword, statusFilter, capabilityFilter, i18n.language]
  );

  const setOp = (pt: string, v: boolean) => {
    setOperating((prev) => ({ ...prev, [pt]: v }));
  };

  const clearPollingError = useCallback((pt: string) => {
    setPollingErrors((prev) => {
      if (!prev[pt]) return prev;
      const next = { ...prev };
      delete next[pt];
      return next;
    });
    delete pollingRetryRef.current[pt];
  }, []);

  const startPolling = useCallback(
    (pt: string) => {
      clearPollingError(pt);
      pollingRef.current.add(pt);
    },
    [clearPollingError]
  );

  const removeJob = useCallback(
    (pt: string) => {
      clearPollingError(pt);
      setJobs((prev) => {
        if (!prev[pt]) return prev;
        const next = { ...prev };
        delete next[pt];
        return next;
      });
      setSpeeds((prev) => {
        if (!prev[pt]) return prev;
        const next = { ...prev };
        delete next[pt];
        return next;
      });
      delete lastSampleRef.current[pt];
    },
    [clearPollingError]
  );

  const recordSample = useCallback((pt: string, status: InstallJobStatus) => {
    const prev = lastSampleRef.current[pt];
    const now = Date.now();
    if (status.state === "downloading" && prev && now > prev.t) {
      const dt = (now - prev.t) / 1000;
      const dd = (status.downloaded ?? 0) - prev.bytes;
      if (dt > 0 && dd >= 0) {
        setSpeeds((prev) => ({ ...prev, [pt]: dd / dt }));
      }
    }
    lastSampleRef.current[pt] = { bytes: status.downloaded ?? 0, t: now };
  }, []);

  useEffect(() => {
    const listingRequest = listingRequestRef.current;
    let cancelled = false;
    queueMicrotask(() => {
      if (!cancelled) void loadListing();
    });
    return () => {
      cancelled = true;
      listingRequest.generation++;
    };
  }, [loadListing, revision]);

  useEffect(() => {
    if (!listing?.refreshing || loading || error) return;
    const id = window.setTimeout(() => void loadListing(false, true), POLL_MS);
    return () => clearTimeout(id);
  }, [listing, loading, error, loadListing]);

  useEffect(() => {
    if (!listing?.activeJobs.length) return;
    for (const job of listing.activeJobs) pollingRef.current.add(job.providerType);
  }, [listing]);

  useEffect(() => {
    let cancelled = false;
    const id = window.setInterval(() => {
      const active = Array.from(pollingRef.current).filter(
        (pt) => !pollingRequestsRef.current.has(pt) && Date.now() >= (pollingRetryRef.current[pt]?.nextAt ?? 0)
      );
      active.forEach(async (pt) => {
        pollingRequestsRef.current.add(pt);
        try {
          const st = await getInstallStatus(pt);
          if (cancelled) return;
          clearPollingError(pt);
          const reloadError = st.result?.errors?.join("\n");
          const status = st.state === "installed" && reloadError ? { ...st, state: "failed" as const, error: reloadError } : st;
          setJobs((prev) => ({ ...prev, [pt]: status }));
          recordSample(pt, status);
          if (isTerminalState(status.state)) {
            pollingRef.current.delete(pt);
            if (status.state === "installed") {
              removeJob(pt);
              message.success(t("plugin.market.msg.installed"));
            } else {
              setSpeeds((prev) => {
                const next = { ...prev };
                delete next[pt];
                return next;
              });
              delete lastSampleRef.current[pt];
            }
            await catalogReload();
            if (!cancelled) setRevision((prev) => prev + 1);
          }
        } catch {
          if (cancelled) return;
          const attempts = (pollingRetryRef.current[pt]?.attempts ?? 0) + 1;
          pollingRetryRef.current[pt] = { attempts, nextAt: Date.now() + Math.min(MAX_POLL_DELAY_MS, POLL_MS * 2 ** Math.min(attempts, 5)) };
          setPollingErrors((prev) => ({ ...prev, [pt]: t("plugin.market.error.status_failed") }));
        } finally {
          pollingRequestsRef.current.delete(pt);
        }
      });
    }, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [t, message, catalogReload, clearPollingError, recordSample, removeJob]);

  const handleInstall = async (entry: MarketEntry) => {
    const pt = entry.provider_type;
    try {
      const initial = await installPlugin(pt);
      setJobs((prev) => ({ ...prev, [pt]: initial }));
      startPolling(pt);
    } catch (err) {
      if (err && typeof err === "object" && "status" in err && err.status === 409) {
        startPolling(pt);
        return;
      }
      const msg = err instanceof Error ? err.message : t("plugin.market.error.install_failed");
      message.error(msg);
    }
  };

  const handleDelete = async (entry: MarketEntry) => {
    const pt = entry.provider_type;
    setOp(pt, true);
    try {
      const result = await deletePlugin(pt);
      removeJob(pt);
      await catalogReload();
      setRevision((prev) => prev + 1);
      if (result.errors?.length) throw new Error(result.errors.join("\n"));
      message.success(t("plugin.market.msg.deleted"));
    } catch (err) {
      const msg = err instanceof Error ? err.message : t("plugin.market.error.delete_failed");
      message.error(msg);
    } finally {
      setOp(pt, false);
    }
  };

  const confirmDelete = (entry: MarketEntry) => {
    const pt = entry.provider_type;
    const name = entry.display_name_key ? t(entry.display_name_key, pt) : pt;
    modal.confirm({
      title: <span className="text-error">{t("plugin.market.action.delete_modal_title")}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("plugin.market.action.delete_modal_content", { name }) }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: () => handleDelete(entry),
    });
  };

  const handleUpdate = async (entry: MarketEntry) => {
    const pt = entry.provider_type;
    setOp(pt, true);
    try {
      const result = await updatePlugin(pt);
      await catalogReload();
      setRevision((prev) => prev + 1);
      if (result.errors?.length) throw new Error(result.errors.join("\n"));
      message.success(t("plugin.market.msg.updated"));
    } catch (err) {
      const msg = err instanceof Error ? err.message : t("plugin.market.error.update_failed");
      message.error(msg);
    } finally {
      setOp(pt, false);
    }
  };

  const renderActions = (entry: MarketEntry) => {
    const pt = entry.provider_type;
    const job = jobs[pt] ?? entry.job;
    const busy = !!operating[pt];
    const installing = !!job && !isTerminalState(job.state);
    const failed = job?.state === "failed";

    if (pollingErrors[pt]) {
      return (
        <Button size="small" icon={<IconRefresh size={14} />} onClick={() => startPolling(pt)}>
          {t("plugin.market.action.retry")}
        </Button>
      );
    }

    if (installing) {
      const total = job?.total ?? 0;
      const downloaded = job?.downloaded ?? 0;
      const detailParts: string[] = [t(jobStageLabelKey(job!.state))];
      if (total > 0) {
        const pct = Math.min(100, Math.round((downloaded / total) * 100));
        detailParts.push(`${formatBytes(downloaded)} / ${formatBytes(total)} (${pct}%)`);
      } else if (downloaded > 0) {
        detailParts.push(formatBytes(downloaded));
      }
      const sp = speeds[pt];
      if (sp && sp > 0) {
        detailParts.push(`${formatBytes(sp)}/s`);
      }
      const detail = detailParts.join("  ·  ");

      if (job!.state === "downloading" && total > 0) {
        const pct = Math.min(100, Math.max(0, Math.round((downloaded / total) * 100)));
        return (
          <Tooltip title={detail}>
            <Progress type="circle" percent={pct} size={34} strokeWidth={10} />
          </Tooltip>
        );
      }
      return (
        <Tooltip title={detail}>
          <Tag color="processing" className="flex items-center gap-1">
            <IconRefresh size={12} className="animate-spin" />
            {job!.state === "downloading" ? formatBytes(downloaded) : t(jobStageLabelKey(job!.state))}
          </Tag>
        </Tooltip>
      );
    }

    switch (entry.status) {
      case "not_installed":
        return (
          <Button type="primary" size="small" icon={<IconDownload size={14} />} onClick={() => handleInstall(entry)}>
            {failed ? t("plugin.market.action.retry") : t("plugin.market.action.install")}
          </Button>
        );
      case "installed":
        return (
          <Button danger size="small" icon={<IconTrash size={14} />} loading={busy} onClick={() => confirmDelete(entry)}>
            {t("plugin.market.action.delete")}
          </Button>
        );
      case "update_available":
        return (
          <Button type="primary" size="small" icon={<IconRefresh size={14} />} loading={busy} onClick={() => handleUpdate(entry)}>
            {t("plugin.market.action.update")}
          </Button>
        );
      default:
        return null;
    }
  };

  const renderStatus = (entry: MarketEntry) => {
    const job = jobs[entry.provider_type] ?? entry.job;
    const installing = !!job && !isTerminalState(job.state);
    const failed = job?.state === "failed";

    if (installing) {
      const total = job?.total ?? 0;
      const downloaded = job?.downloaded ?? 0;
      if (job!.state === "downloading" && total > 0) {
        const pct = Math.min(100, Math.round((downloaded / total) * 100));
        return (
          <span className="text-xs text-blue-500">
            {t("plugin.market.job.downloading")} · {pct}%
          </span>
        );
      }
      return (
        <span className="flex items-center gap-1.5 text-xs text-blue-500">
          <IconRefresh size={10} className="animate-spin" />
          {t(jobStageLabelKey(job!.state))}
        </span>
      );
    }
    if (failed) {
      return <span className="text-xs text-red-500">{t("plugin.market.job.failed")}</span>;
    }
    return (
      <span className="flex items-center gap-1.5">
        <span className={`inline-block size-2 rounded-full ${statusColor[entry.status] ?? "bg-gray-400"}`} />
        {t(statusLabel[entry.status], entry.status)}
      </span>
    );
  };

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <Segmented<MarketFilter>
            options={[
              { label: t("plugin.market.filter.all"), value: "all" },
              { label: t("plugin.market.filter.installed"), value: "installed" },
              { label: t("plugin.market.filter.not_installed"), value: "not_installed" },
            ]}
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value);
              setPage(1);
            }}
          />
          <Segmented<MarketCapabilityFilter>
            options={[
              { label: t("plugin.market.filter.all"), value: "all" },
              { label: t("plugin.market.capability.deploy"), value: "deploy" },
              { label: t("plugin.market.capability.dns"), value: "dns" },
              { label: t("plugin.market.capability.notify"), value: "notify" },
            ]}
            value={capabilityFilter}
            onChange={(value) => {
              setCapabilityFilter(value);
              setPage(1);
            }}
          />
          <Input.Search
            allowClear
            className="w-48"
            placeholder={t("plugin.market.action.search_placeholder")}
            onSearch={(value) => {
              setKeyword(value.trim());
              setPage(1);
            }}
          />
        </div>
        <Button icon={<IconRefresh size={16} />} onClick={() => void loadListing(true)} loading={loading || listing?.refreshing}>
          {t("plugin.market.action.refresh")}
        </Button>
      </div>

      {listing?.syncedAt && (
        <p className="mb-3 text-xs text-gray-500">{t("plugin.market.synced_at", { time: new Date(listing.syncedAt).toLocaleString(i18n.language) })}</p>
      )}
      {listing?.refreshFailed && <Alert className="mb-4" type="warning" message={t("plugin.market.error.refresh_failed")} />}
      {error ? (
        <Alert
          type="error"
          message={error}
          action={
            <Button size="small" onClick={() => void loadListing(true)}>
              {t("plugin.market.action.retry")}
            </Button>
          }
        />
      ) : loading ? (
        <Skeleton active />
      ) : !listing?.items.length ? (
        <Empty description={t(keyword || statusFilter !== "all" || capabilityFilter !== "all" ? "plugin.market.empty_filtered" : "plugin.market.empty")} />
      ) : (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(min(100%,280px),1fr))] gap-4">
          {listing.items.map((entry) => (
            <Card key={entry.provider_type} hoverable className="flex min-w-0 flex-col">
              <div className="flex h-full flex-col gap-3">
                <div className="flex items-start gap-3">
                  <div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-gray-100 p-2.5 dark:bg-gray-800">
                    {entry.icon_url ? (
                      <img src={entry.icon_url} alt={entry.provider_type} loading="lazy" className="size-full object-contain" />
                    ) : (
                      <IconPuzzle size={22} className="text-gray-400" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-base font-semibold">
                      {entry.display_name_key ? t(entry.display_name_key, entry.provider_type) : entry.provider_type}
                    </div>
                    <div className="mt-0.5 text-xs text-gray-400">
                      {entry.capability
                        ? [t(`plugin.market.capability.${entry.capability}`, entry.capability), entry.deploy_category].filter(Boolean).join(" · ")
                        : entry.deploy_category}
                    </div>
                  </div>
                  <div className="shrink-0">{renderActions(entry)}</div>
                </div>
                <div className="line-clamp-2 min-h-10 text-sm text-gray-500">{entry.description}</div>
                {(pollingErrors[entry.provider_type] || (jobs[entry.provider_type] ?? entry.job)?.error) && (
                  <Alert
                    type={pollingErrors[entry.provider_type] ? "warning" : "error"}
                    message={pollingErrors[entry.provider_type] || (jobs[entry.provider_type] ?? entry.job)?.error}
                  />
                )}
                <div className="mt-auto flex items-center justify-between border-t border-gray-100 pt-2 text-xs text-gray-400 dark:border-gray-800">
                  {renderStatus(entry)}
                  <span>v{entry.version}</span>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
      {listing && listing.totalItems > 0 && !error && (
        <Pagination
          className="mt-4 justify-end"
          current={listing.page}
          pageSize={perPage}
          total={listing.totalItems}
          pageSizeOptions={[15, 30, 60]}
          showSizeChanger
          disabled={loading}
          onChange={(nextPage, nextSize) => {
            setPage(nextSize === perPage ? nextPage : 1);
            setPerPage(nextSize);
          }}
        />
      )}
    </div>
  );
};

export default PluginList;
