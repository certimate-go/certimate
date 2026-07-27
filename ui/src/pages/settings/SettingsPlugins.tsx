import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconDownload, IconPuzzle, IconTrash, IconRefresh } from "@tabler/icons-react";
import { Alert, Button, Card, Empty, Skeleton, Tag, App } from "antd";

import { fetchMarketListing, installPlugin, deletePlugin, updatePlugin, type MarketEntry } from "@/api/pluginmarket";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

const statusLabel: Record<string, string> = {
  not_installed: "plugin.market.status.not_installed",
  installed: "plugin.market.status.installed",
  update_available: "plugin.market.status.update_available",
  installed_manual: "plugin.market.status.installed_manual",
  unsupported_platform: "plugin.market.status.unsupported_platform",
};

const SettingsPlugins = () => {
  const { t } = useTranslation();
  const { message } = App.useApp();
  const catalogReload = usePluginCatalogStore((s) => s.reload);

  const [entries, setEntries] = useState<MarketEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [operating, setOperating] = useState<Record<string, boolean>>({});

  const loadListing = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchMarketListing();
      setEntries(data);
    } catch {
      setError(t("plugin.market.error.load_failed"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadListing();
  }, [loadListing]);

  const setOp = (pt: string, v: boolean) => {
    setOperating((prev) => ({ ...prev, [pt]: v }));
  };

  const handleInstall = async (entry: MarketEntry) => {
    const pt = entry.provider_type;
    setOp(pt, true);
    try {
      await installPlugin(pt);
      message.success(t("plugin.market.msg.installed"));
      await catalogReload();
      await loadListing();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("plugin.market.error.install_failed");
      message.error(msg);
    } finally {
      setOp(pt, false);
    }
  };

  const handleDelete = async (entry: MarketEntry) => {
    const pt = entry.provider_type;
    setOp(pt, true);
    try {
      await deletePlugin(pt);
      message.success(t("plugin.market.msg.deleted"));
      await catalogReload();
      await loadListing();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("plugin.market.error.delete_failed");
      message.error(msg);
    } finally {
      setOp(pt, false);
    }
  };

  const handleUpdate = async (entry: MarketEntry) => {
    const pt = entry.provider_type;
    setOp(pt, true);
    try {
      await updatePlugin(pt);
      message.success(t("plugin.market.msg.updated"));
      await catalogReload();
      await loadListing();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("plugin.market.error.update_failed");
      message.error(msg);
    } finally {
      setOp(pt, false);
    }
  };

  const renderActions = (entry: MarketEntry) => {
    const pt = entry.provider_type;
    const busy = operating[pt];

    switch (entry.status) {
      case "not_installed":
        return (
          <Button type="primary" icon={<IconDownload size={16} />} loading={busy} onClick={() => handleInstall(entry)}>
            {t("plugin.market.action.install")}
          </Button>
        );
      case "installed":
        return (
          <Button danger icon={<IconTrash size={16} />} loading={busy} onClick={() => handleDelete(entry)}>
            {t("plugin.market.action.delete")}
          </Button>
        );
      case "update_available":
        return (
          <div className="flex items-center gap-2">
            <Tag color="orange">{t("plugin.market.label.update_available")}</Tag>
            <Button type="primary" icon={<IconRefresh size={16} />} loading={busy} onClick={() => handleUpdate(entry)}>
              {t("plugin.market.action.update")} {entry.version}
            </Button>
          </div>
        );
      case "installed_manual":
        return (
          <Tag>{t("plugin.market.label.manual")}</Tag>
        );
      case "unsupported_platform":
        return (
          <Tag color="default">{t("plugin.market.label.unsupported_platform")}</Tag>
        );
      default:
        return null;
    }
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

  if (entries.length === 0) {
    return <Empty description={t("plugin.market.empty")} />;
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <Button icon={<IconRefresh size={16} />} onClick={loadListing} loading={loading}>
          {t("plugin.market.action.refresh")}
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        {entries.map((entry) => (
          <Card
            key={entry.provider_type}
            hoverable
            className="flex flex-col"
            cover={
              entry.icon ? (
                <div className="flex h-32 items-center justify-center bg-gray-50 dark:bg-gray-800">
                  <img src={entry.icon} alt={entry.provider_type} className="max-h-24 max-w-full object-contain" />
                </div>
              ) : (
                <div className="flex h-32 items-center justify-center bg-gray-50 dark:bg-gray-800">
                  <IconPuzzle size={48} className="text-gray-400" />
                </div>
              )
            }
          >
            <div className="flex flex-col gap-2">
              <Card.Meta
                title={entry.display_name_key ? t(entry.display_name_key, entry.provider_type) : entry.provider_type}
                description={entry.description || entry.provider_type}
              />
              <div className="flex items-center gap-2 text-sm text-gray-500">
                <span>v{entry.version}</span>
                <Tag>{t(statusLabel[entry.status], entry.status)}</Tag>
              </div>
              <div className="mt-2">{renderActions(entry)}</div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
};

export default SettingsPlugins;
