import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconCheck, IconChevronDown, IconDots, IconDownload, IconFilter, IconPuzzle, IconRefresh, IconTrash, IconX } from "@tabler/icons-react";
import { Alert, App, Button, Card, Dropdown, Empty, Input, type MenuProps, Pagination, Progress, Select, Skeleton, Tag, Tooltip, theme } from "antd";
import { ClientResponseError } from "pocketbase";

import {
  type InstallJobStatus,
  type MarketCapabilityFilter,
  type MarketEntry,
  type MarketFilter,
  type MarketPage,
  type PluginTarget,
  accessDependencyErrorSchema,
  deletePlugin,
  fetchMarketListing,
  getInstallStatus,
  installPlugin,
  isTerminalState,
  marketConflictSchema,
  updatePlugin,
} from "@/api/pluginmarket";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import MarketSourcesModal from "./MarketSourcesModal";
import OfficialBadge from "./OfficialBadge";

const POLL_MS = 1000;
const MAX_POLL_DELAY_MS = 30000;

const statusLabel: Record<string, string> = {
  installed: "plugin.market.status.installed",
  update_available: "plugin.market.status.update_available",
  installed_manual: "plugin.market.status.installed_manual",
  unsupported_platform: "plugin.market.status.unsupported_platform",
  ownership_conflict: "plugin.market.status.ownership_conflict",
};

const statusColor: Record<Exclude<MarketEntry["status"], "not_installed">, string> = {
  installed: "success",
  update_available: "processing",
  installed_manual: "default",
  unsupported_platform: "default",
  ownership_conflict: "warning",
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

const pluginTarget = (entry: MarketEntry): PluginTarget => ({ providerType: entry.provider_type, source: entry.source.id });
const cardKey = ({ source, providerType }: PluginTarget) => `${source}/${providerType}`;

interface MarketCardActionsProps {
  entry: MarketEntry;
  job?: InstallJobStatus;
  pollingError?: string;
  speed?: number;
  busy: boolean;
  onRetry: () => void;
  onInstall: () => void;
  onUpdate: () => void;
  onDelete: () => void;
}

const MarketCardActions = ({ entry, job, pollingError, speed, busy, onRetry, onInstall, onUpdate, onDelete }: MarketCardActionsProps) => {
  const { t } = useTranslation();
  const installing = !!job && !isTerminalState(job.state);
  const failed = job?.state === "failed";

  if (entry.canInstall && (pollingError || failed)) {
    return (
      <Button size="small" icon={<IconRefresh size={14} />} loading={busy} onClick={onRetry}>
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
    const sp = speed;
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

  if (entry.canInstall) {
    return (
      <Button color="primary" variant="filled" size="small" icon={<IconDownload size={14} />} loading={busy} onClick={onInstall}>
        {t("plugin.market.action.install")}
      </Button>
    );
  }
  const items: MenuProps["items"] = [];
  if (entry.canUpdate) items.push({ key: "update", label: t("plugin.market.action.update"), icon: <IconRefresh size={14} />, onClick: onUpdate });
  if (entry.canDelete) items.push({ key: "delete", label: t("plugin.market.action.delete"), icon: <IconTrash size={14} />, danger: true, onClick: onDelete });
  if (!items.length) return null;
  const name = entry.display_name_key ? t(entry.display_name_key, entry.provider_type) : entry.provider_type;
  return (
    <Dropdown autoFocus placement="bottomRight" trigger={["click"]} menu={{ items }} disabled={busy}>
      <Button
        type="text"
        size="small"
        icon={<IconDots size={16} />}
        loading={busy}
        aria-haspopup="menu"
        aria-label={t("plugin.market.action.more", { name })}
      />
    </Dropdown>
  );
};

const PluginMarket = () => {
  const { t, i18n } = useTranslation();
  const { message, modal } = App.useApp();
  const { token } = theme.useToken();
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
  const [sourceFilter, setSourceFilter] = useState("");
  const [sourcesOpen, setSourcesOpen] = useState(false);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [jobs, setJobs] = useState<Record<string, InstallJobStatus>>({});
  const [speeds, setSpeeds] = useState<Record<string, number>>({});
  const [pollingErrors, setPollingErrors] = useState<Record<string, string>>({});
  const pollingRef = useRef(new Map<string, PluginTarget & { jobId?: string }>());
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
        const data = await fetchMarketListing({
          page,
          perPage,
          keyword,
          status: statusFilter,
          capability: capabilityFilter,
          locale: i18n.language,
          refresh,
          source: sourceFilter,
        });
        if (request === listingRequestRef.current.generation) {
          if (sourceFilter && !data.sources.some(({ id }) => id === sourceFilter)) {
            setSourceFilter("");
            setPage(1);
          }
          setListing(data);
        }
      } catch {
        if (request === listingRequestRef.current.generation) setError(t("plugin.market.error.load_failed"));
      } finally {
        if (request === listingRequestRef.current.generation) setLoading(false);
      }
    },
    [t, page, perPage, keyword, statusFilter, capabilityFilter, sourceFilter, i18n.language]
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
    (target: PluginTarget, jobId?: string) => {
      const key = cardKey(target);
      clearPollingError(key);
      pollingRef.current.set(key, { ...target, jobId });
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
    for (const job of listing.activeJobs) {
      const target = { providerType: job.providerType, source: job.source.id };
      const key = cardKey(target);
      if (pollingRef.current.get(key)?.jobId !== job.id) {
        pollingRef.current.set(key, { ...target, jobId: job.id });
      }
    }
  }, [listing]);

  useEffect(() => {
    let cancelled = false;
    const id = window.setInterval(() => {
      const active = Array.from(pollingRef.current).filter(
        ([pt]) => !pollingRequestsRef.current.has(pt) && Date.now() >= (pollingRetryRef.current[pt]?.nextAt ?? 0)
      );
      active.forEach(async ([pt, target]) => {
        pollingRequestsRef.current.add(pt);
        try {
          const st = await getInstallStatus(target.providerType);
          if (cancelled || pollingRef.current.get(pt) !== target) return;
          if (st.source.id !== target.source || (target.jobId && st.id !== target.jobId)) {
            pollingRef.current.delete(pt);
            removeJob(pt);
            setRevision((prev) => prev + 1);
            return;
          }
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
          if (cancelled || pollingRef.current.get(pt) !== target) return;
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
    const target = pluginTarget(entry);
    const pt = cardKey(target);
    setOp(pt, true);
    try {
      const initial = await installPlugin(target);
      setJobs((prev) => ({ ...prev, [pt]: initial }));
      startPolling(target, initial.id);
    } catch (err) {
      const conflict = marketConflictSchema.safeParse(err instanceof ClientResponseError ? err.response.data : undefined);
      if (conflict.success && conflict.data.reason === "install_in_progress") {
        startPolling(target);
        return;
      }
      const msg = err instanceof Error ? err.message : t("plugin.market.error.install_failed");
      message.error(msg);
    } finally {
      setOp(pt, false);
    }
  };

  const handleDelete = async (entry: MarketEntry) => {
    const target = pluginTarget(entry);
    const pt = cardKey(target);
    setOp(pt, true);
    try {
      const result = await deletePlugin(target);
      removeJob(pt);
      await catalogReload();
      setRevision((prev) => prev + 1);
      if (result.errors?.length) throw new Error(result.errors.join("\n"));
      message.success(t("plugin.market.msg.deleted"));
    } catch (err) {
      const dependency = accessDependencyErrorSchema.safeParse(err instanceof ClientResponseError ? err.response.data : undefined);
      if (dependency.success) {
        modal.warning({
          title: t("plugin.market.error.delete_dependency_title"),
          content: t("plugin.market.error.delete_dependency", {
            name: dependency.data.providerType,
            accessType: dependency.data.accessType,
            dependents: dependency.data.dependents.join(", "),
          }),
          okText: t("common.button.confirm"),
        });
        return;
      }
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
      content: t("plugin.market.action.delete_modal_content", { name }),
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
    const target = pluginTarget(entry);
    const pt = cardKey(target);
    setOp(pt, true);
    try {
      const result = await updatePlugin(target);
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

  const retryInstall = (entry: MarketEntry, job?: InstallJobStatus) => {
    const target = pluginTarget(entry);
    if (pollingErrors[cardKey(target)]) startPolling(target, job?.id);
    else void handleInstall(entry);
  };

  const renderStatus = (entry: MarketEntry) => {
    const key = cardKey(pluginTarget(entry));
    const job = jobs[key] ?? entry.job;
    const pollingError = pollingErrors[key];
    const errorDetails = [job?.error, pollingError].filter(Boolean).join("\n");
    const installing = !!job && !isTerminalState(job.state);

    if (job?.state === "failed" || errorDetails) {
      const failureLabel = t(pollingError ? "plugin.market.job.status_unavailable" : "plugin.market.job.failed");
      if (!errorDetails) return <span className="text-xs text-error">{failureLabel}</span>;
      return (
        <Tooltip title={<span className="wrap-anywhere whitespace-pre-wrap">{errorDetails}</span>} trigger={["hover", "focus", "click"]}>
          <button
            type="button"
            className="cursor-help appearance-none border-0 bg-transparent p-0 text-left text-xs text-error underline decoration-dotted underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            {failureLabel}
          </button>
        </Tooltip>
      );
    }
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
    if (entry.status === "not_installed") return null;

    return (
      <Tag color={statusColor[entry.status]} variant="filled" className="m-0 rounded-full whitespace-normal">
        {t(statusLabel[entry.status], entry.status)}
      </Tag>
    );
  };

  const statusOptions: { value: MarketFilter; label: string }[] = [
    { value: "all", label: t("plugin.market.filter.all") },
    { value: "installed", label: t("plugin.market.filter.installed") },
    { value: "not_installed", label: t("plugin.market.filter.not_installed") },
  ];
  const capabilityOptions: { value: MarketCapabilityFilter; label: string }[] = [
    { value: "all", label: t("plugin.market.filter.all") },
    { value: "deploy", label: t("plugin.market.capability.deploy") },
    { value: "dns", label: t("plugin.market.capability.dns") },
    { value: "notify", label: t("plugin.market.capability.notify") },
  ];
  const changeStatus = (value: MarketFilter) => {
    setStatusFilter(value);
    setPage(1);
  };
  const changeCapability = (value: MarketCapabilityFilter) => {
    setCapabilityFilter(value);
    setPage(1);
  };
  const activeFilters = [
    { value: statusFilter, options: statusOptions, clear: () => changeStatus("all"), label: t("plugin.market.filter.status") },
    { value: capabilityFilter, options: capabilityOptions, clear: () => changeCapability("all"), label: t("plugin.market.filter.capability") },
  ].filter(({ value }) => value !== "all");

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <Dropdown
          autoFocus
          placement="bottomLeft"
          trigger={["click"]}
          open={filtersOpen}
          onOpenChange={setFiltersOpen}
          menu={{
            selectable: true,
            multiple: true,
            selectedKeys: [`status-${statusFilter}`, `capability-${capabilityFilter}`],
            items: [
              {
                type: "group",
                label: t("plugin.market.filter.status"),
                children: statusOptions.map(({ value, label }) => ({
                  key: `status-${value}`,
                  label,
                  icon: <IconCheck size={14} className={value === statusFilter ? undefined : "invisible"} />,
                  onClick: () => changeStatus(value),
                })),
              },
              {
                type: "group",
                label: t("plugin.market.filter.capability"),
                children: capabilityOptions.map(({ value, label }) => ({
                  key: `capability-${value}`,
                  label,
                  icon: <IconCheck size={14} className={value === capabilityFilter ? undefined : "invisible"} />,
                  onClick: () => changeCapability(value),
                })),
              },
            ],
          }}
        >
          <Button className="text-sm" icon={<IconFilter size="1.25em" />} size="large" aria-haspopup="menu" aria-expanded={filtersOpen}>
            {t("plugin.market.filter.label")}
            <IconChevronDown size={14} />
          </Button>
        </Dropdown>
        <Select
          size="large"
          className="min-w-40"
          aria-label={t("plugin.market.sources.filter")}
          value={sourceFilter}
          options={[
            { value: "", label: t("plugin.market.sources.all") },
            ...(listing?.sources ?? []).map((source) => ({ value: source.id, label: source.repository })),
          ]}
          onChange={(value) => {
            setSourceFilter(value);
            setPage(1);
          }}
        />
        <Input.Search
          allowClear
          className="max-w-80 min-w-32 flex-1 text-sm placeholder:text-sm"
          aria-label={t("plugin.market.action.search_placeholder")}
          placeholder={t("plugin.market.action.search_placeholder")}
          size="large"
          onSearch={(value) => {
            setKeyword(value.trim());
            setPage(1);
          }}
        />
        <Button
          className="text-sm"
          icon={<IconRefresh size="1.25em" />}
          size="large"
          onClick={() => void loadListing(true)}
          loading={loading || listing?.refreshing}
        >
          {t("plugin.market.action.refresh")}
        </Button>
        <Button className="text-sm" size="large" onClick={() => setSourcesOpen(true)}>
          {t("plugin.market.sources.manage")}
        </Button>
      </div>
      {sourcesOpen && (
        <MarketSourcesModal
          sources={listing?.sources ?? []}
          onClose={() => setSourcesOpen(false)}
          onChange={(removed) => {
            listingRequestRef.current.generation++;
            if (removed) {
              for (const [key, target] of pollingRef.current) {
                if (target.source === removed) {
                  pollingRef.current.delete(key);
                  removeJob(key);
                }
              }
              setListing(
                (current) =>
                  current && {
                    ...current,
                    sources: current.sources.filter((source) => source.id !== removed),
                    items: current.items.filter((entry) => entry.source.id !== removed),
                    activeJobs: current.activeJobs.filter((job) => job.source.id !== removed),
                  }
              );
              setSourceFilter((current) => (current === removed ? "" : current));
              setPage(1);
            }
            setRevision((previous) => previous + 1);
          }}
        />
      )}
      {activeFilters.length > 0 && (
        <div className="mb-3 flex flex-wrap gap-2">
          {activeFilters.map(({ value, options, clear, label }) => (
            <Tag key={label} closeIcon={<IconX size={12} aria-label={t("plugin.market.filter.clear", { filter: label })} />} onClose={clear} className="m-0">
              {options.find((option) => option.value === value)?.label}
            </Tag>
          ))}
        </div>
      )}

      {listing?.syncedAt && (
        <p className="mb-3 text-xs text-gray-500">{t("plugin.market.synced_at", { time: new Date(listing.syncedAt).toLocaleString(i18n.language) })}</p>
      )}
      {listing?.sources
        .filter((source) => source.error && (!sourceFilter || source.id === sourceFilter))
        .map((source) => (
          <Alert
            key={source.id}
            className="mb-3"
            type="warning"
            title={`${source.repository} · ${t(source.stale ? "plugin.market.sources.stale" : "plugin.market.sources.unavailable")}`}
            description={source.error}
          />
        ))}
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
        <Empty
          description={t(
            keyword || sourceFilter || statusFilter !== "all" || capabilityFilter !== "all" ? "plugin.market.empty_filtered" : "plugin.market.empty"
          )}
        />
      ) : (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(min(100%,280px),1fr))] gap-4">
          {listing.items.map((entry) => {
            const name = entry.display_name_key && i18n.exists(entry.display_name_key) ? t(entry.display_name_key, entry.provider_type) : entry.provider_type;
            const key = cardKey(pluginTarget(entry));
            const job = jobs[key] ?? entry.job;
            return (
              <Card
                key={cardKey(pluginTarget(entry))}
                size="small"
                className="flex min-w-0 flex-col"
                classNames={{ body: "flex grow flex-col gap-2" }}
                styles={{ body: { padding: 16 } }}
              >
                <div className="flex items-start gap-3 sm:min-h-15">
                  <div className="flex size-11 shrink-0 items-center justify-center rounded-lg p-2" style={{ background: token.colorFillQuaternary }}>
                    {entry.icon_url ? (
                      <img src={entry.icon_url} alt="" loading="lazy" className="size-full object-contain" />
                    ) : (
                      <IconPuzzle size={20} color={token.colorTextTertiary} />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-start gap-1">
                      <Tooltip title={name} trigger={["hover", "focus"]}>
                        <h2 className="m-0 line-clamp-2 text-sm/5 font-semibold wrap-anywhere" tabIndex={0}>
                          {name}
                        </h2>
                      </Tooltip>
                      {entry.official && (
                        <div className="flex h-5 shrink-0 items-center">
                          <OfficialBadge />
                        </div>
                      )}
                    </div>
                    <div className="text-xs wrap-anywhere" style={{ color: token.colorTextSecondary }}>
                      {entry.source.repository || t("plugin.market.sources.local")}
                      {entry.canDelete && !entry.sourceAvailable && ` · ${t("plugin.market.sources.removed")}`}
                      {entry.canDelete && entry.sourceAvailable && !entry.listed && ` · ${t("plugin.market.sources.unlisted")}`}
                      {entry.status === "ownership_conflict" && (
                        <div>
                          {t("plugin.market.sources.installed_from", { repository: entry.installedOrigin?.repository || t("plugin.market.sources.local") })}
                        </div>
                      )}
                    </div>
                    <div className="mt-0.5 text-xs/4 wrap-anywhere" style={{ color: token.colorTextSecondary }}>
                      {entry.capability
                        ? [t(`plugin.market.capability.${entry.capability}`, entry.capability), entry.deploy_category].filter(Boolean).join(" · ")
                        : entry.deploy_category}
                    </div>
                  </div>
                </div>
                {entry.description && (
                  <div className="line-clamp-2 text-xs" style={{ color: token.colorTextSecondary }}>
                    {entry.description}
                  </div>
                )}
                <div className="mt-auto flex min-h-8 flex-wrap items-center justify-between gap-x-2 gap-y-1">
                  <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs" style={{ color: token.colorTextSecondary }}>
                    {entry.version && <span className="wrap-anywhere tabular-nums">v{entry.version}</span>}
                    {renderStatus(entry)}
                  </div>
                  <div className="ml-auto flex items-center">
                    <MarketCardActions
                      entry={entry}
                      job={job}
                      pollingError={pollingErrors[key]}
                      speed={speeds[key]}
                      busy={!!operating[key]}
                      onRetry={() => retryInstall(entry, job)}
                      onInstall={() => void handleInstall(entry)}
                      onUpdate={() => void handleUpdate(entry)}
                      onDelete={() => confirmDelete(entry)}
                    />
                  </div>
                </div>
              </Card>
            );
          })}
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
