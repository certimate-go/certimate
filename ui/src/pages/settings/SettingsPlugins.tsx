import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconDownload, IconPuzzle, IconTrash, IconRefresh } from "@tabler/icons-react";
import { Alert, App, Button, Card, Empty, Input, Progress, Segmented, Skeleton, Tag, Tooltip } from "antd";

import {
  fetchMarketListing,
  installPlugin,
  deletePlugin,
  updatePlugin,
  getInstallStatus,
  isTerminalState,
  type InstallJobStatus,
  type MarketEntry,
} from "@/api/pluginmarket";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

const POLL_MS = 1000;

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

const marketIconSrc = (entry: MarketEntry): string => {
  if (!entry.icon || !entry.release) {
    return "";
  }
  return `https://cdn.jsdelivr.net/gh/${entry.release.repo}@main/${entry.provider_type}/${entry.icon}`;
};

const statusToSegment = (status: string): "installed" | "not_installed" => {
  if (status === "installed" || status === "update_available" || status === "installed_manual") {
    return "installed";
  }
  return "not_installed";
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

const SettingsPlugins = () => {
  const { t } = useTranslation();
  const { message, modal } = App.useApp();
  const catalogReload = usePluginCatalogStore((s) => s.reload);

  const [entries, setEntries] = useState<MarketEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [operating, setOperating] = useState<Record<string, boolean>>({});
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] = useState<"all" | "installed" | "not_installed">("all");
  const [jobs, setJobs] = useState<Record<string, InstallJobStatus>>({});
  const [speeds, setSpeeds] = useState<Record<string, number>>({});
  const pollingRef = useRef<Set<string>>(new Set());
  const lastSampleRef = useRef<Record<string, { bytes: number; t: number }>>({});

  const loadListing = useCallback(async (): Promise<MarketEntry[]> => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchMarketListing();
      setEntries(data);
      return data;
    } catch {
      setError(t("plugin.market.error.load_failed"));
      return [];
    } finally {
      setLoading(false);
    }
  }, [t]);

  const setOp = (pt: string, v: boolean) => {
    setOperating((prev) => ({ ...prev, [pt]: v }));
  };

  const removeJob = (pt: string) => {
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
  };

  const recordSample = (pt: string, status: InstallJobStatus) => {
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
  };

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const data = await loadListing();
      if (cancelled) return;
      for (const entry of data) {
        if (entry.status !== "not_installed") continue;
        getInstallStatus(entry.provider_type)
          .then((st) => {
            if (cancelled) return;
            if (isTerminalState(st.state)) {
              if (st.state === "failed") {
                setJobs((prev) => ({ ...prev, [entry.provider_type]: st }));
              }
              return;
            }
            setJobs((prev) => ({ ...prev, [entry.provider_type]: st }));
            pollingRef.current.add(entry.provider_type);
          })
          .catch(() => {});
      }
    })();
    return () => {
      cancelled = true;
      pollingRef.current.clear();
    };
  }, [loadListing]);

  useEffect(() => {
    const id = window.setInterval(() => {
      const active = Array.from(pollingRef.current);
      if (active.length === 0) return;
      Promise.all(
        active.map(async (pt) => {
          try {
            const st = await getInstallStatus(pt);
            setJobs((prev) => ({ ...prev, [pt]: st }));
            recordSample(pt, st);
            if (st.state === "installed") {
              pollingRef.current.delete(pt);
              removeJob(pt);
              message.success(t("plugin.market.msg.installed"));
              await catalogReload();
              await loadListing();
            } else if (st.state === "failed") {
              pollingRef.current.delete(pt);
              setSpeeds((prev) => {
                if (!prev[pt]) return prev;
                const next = { ...prev };
                delete next[pt];
                return next;
              });
              delete lastSampleRef.current[pt];
            }
          } catch {
            pollingRef.current.delete(pt);
          }
        })
      );
    }, POLL_MS);
    return () => clearInterval(id);
  }, [t, message, catalogReload, loadListing]);

  const handleInstall = async (entry: MarketEntry) => {
    const pt = entry.provider_type;
    try {
      const initial = await installPlugin(pt);
      setJobs((prev) => ({ ...prev, [pt]: initial }));
      pollingRef.current.add(pt);
    } catch (err) {
      if (err && typeof err === "object" && "status" in err && err.status === 409) {
        pollingRef.current.add(pt);
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
      await deletePlugin(pt);
      message.success(t("plugin.market.msg.deleted"));
      removeJob(pt);
      await catalogReload();
      await loadListing();
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
      await updatePlugin(pt);
      message.success(t("plugin.market.msg.updated"));
      await catalogReload();
      await loadListing();
    } catch (err) {
      const msg = err instanceof Error ? err.message : t("plugin.market.error.update_failed");
      message.error(msg);
    } finally {
      setOp(pt, false);
    }
  };

  const renderActions = (entry: MarketEntry) => {
    const pt = entry.provider_type;
    const job = jobs[pt];
    const busy = !!operating[pt];
    const installing = !!job && !isTerminalState(job.state);
    const failed = job?.state === "failed";

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
    const job = jobs[entry.provider_type];
    const installing = !!job && !isTerminalState(job.state);
    const failed = job?.state === "failed";

    if (installing) {
      const total = job?.total ?? 0;
      const downloaded = job?.downloaded ?? 0;
      if (job!.state === "downloading" && total > 0) {
        const pct = Math.min(100, Math.round((downloaded / total) * 100));
        return <span className="text-xs text-blue-500">{t("plugin.market.job.downloading")} · {pct}%</span>;
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
        <span className={`inline-block h-2 w-2 rounded-full ${statusColor[entry.status] ?? "bg-gray-400"}`} />
        {t(statusLabel[entry.status], entry.status)}
      </span>
    );
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton active />
        <Skeleton active />
        <Skeleton active />
      </div>
    );
  }

  if (error) {
    return (
      <Alert
        type="error"
        message={error}
        action={
          <Button size="small" onClick={loadListing}>
            {t("plugin.market.action.retry")}
          </Button>
        }
      />
    );
  }

  const filtered = entries.filter((entry) => {
    if (statusFilter !== "all" && statusToSegment(entry.status) !== statusFilter) {
      return false;
    }
    const kw = keyword.trim().toLowerCase();
    if (kw) {
      const name = (entry.display_name_key ? t(entry.display_name_key, entry.provider_type) : entry.provider_type).toLowerCase();
      const desc = (entry.description ?? "").toLowerCase();
      if (!name.includes(kw) && !entry.provider_type.toLowerCase().includes(kw) && !desc.includes(kw)) {
        return false;
      }
    }
    return true;
  });

  return (
    <div>
      <h2>{t("plugin.market.heading")}</h2>

      <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Segmented
            options={[
              { label: t("plugin.market.filter.all"), value: "all" },
              { label: t("plugin.market.filter.installed"), value: "installed" },
              { label: t("plugin.market.filter.not_installed"), value: "not_installed" },
            ]}
            value={statusFilter}
            onChange={(value) => setStatusFilter(value as typeof statusFilter)}
          />
          <Input.Search
            allowClear
            className="w-48"
            placeholder={t("plugin.market.action.search_placeholder")}
            onSearch={(value) => setKeyword(value)}
          />
        </div>
        <Button icon={<IconRefresh size={16} />} onClick={() => loadListing()} loading={loading}>
          {t("plugin.market.action.refresh")}
        </Button>
      </div>

      {filtered.length === 0 ? (
        <Empty description={entries.length === 0 ? t("plugin.market.empty") : t("plugin.market.empty_filtered")} />
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((entry) => (
            <Card key={entry.provider_type} hoverable className="flex flex-col">
              <div className="flex h-full flex-col gap-3">
                <div className="flex items-start gap-3">
                  <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gray-100 p-2.5 dark:bg-gray-800">
                    {marketIconSrc(entry) ? (
                      <img src={marketIconSrc(entry)} alt={entry.provider_type} className="h-full w-full object-contain" />
                    ) : (
                      <IconPuzzle size={22} className="text-gray-400" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-base font-semibold">
                      {entry.display_name_key ? t(entry.display_name_key, entry.provider_type) : entry.provider_type}
                    </div>
                    <div className="mt-0.5 text-xs text-gray-400">{entry.deploy_category}</div>
                  </div>
                  <div className="shrink-0">{renderActions(entry)}</div>
                </div>
                <div className="line-clamp-2 min-h-[2.5rem] text-sm text-gray-500">{entry.description}</div>
                <div className="mt-auto flex items-center justify-between border-t border-gray-100 pt-2 text-xs text-gray-400 dark:border-gray-800">
                  {renderStatus(entry)}
                  <span>v{entry.version}</span>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};

export default SettingsPlugins;
