import { useState } from "react";
import { useTranslation } from "react-i18next";
import { IconRefresh, IconTrash } from "@tabler/icons-react";
import { Alert, Button, Form, Input, Modal, Popconfirm } from "antd";
import { ClientResponseError } from "pocketbase";

import { type MarketSourceStatus, addMarketSource, marketConflictSchema, refreshMarketSource, removeMarketSource } from "@/api/pluginmarket";
import { unwrapErrMsg } from "@/utils/error";
import OfficialBadge from "./OfficialBadge";

interface MarketSourcesModalProps {
  sources: MarketSourceStatus[];
  onClose: () => void;
  onChange: (removed?: string) => void;
}

const MarketSourcesModal = ({ sources, onClose, onChange }: MarketSourcesModalProps) => {
  const { t, i18n } = useTranslation();
  const [repository, setRepository] = useState("");
  const [pending, setPending] = useState<string>();
  const [error, setError] = useState<string>();

  const run = async (key: string, action: () => Promise<void>) => {
    setPending(key);
    setError(undefined);
    try {
      await action();
    } catch (error) {
      const conflict = marketConflictSchema.safeParse(error instanceof ClientResponseError ? error.response.data : undefined);
      if (conflict.success && conflict.data.reason === "source_busy") setError(t("plugin.market.sources.busy"));
      else if (conflict.success && conflict.data.reason === "duplicate_source") setError(t("plugin.market.sources.duplicate"));
      else setError(error instanceof Error ? unwrapErrMsg(error) : t("plugin.market.error.load_failed"));
    } finally {
      setPending(undefined);
    }
  };

  return (
    <Modal open title={t("plugin.market.sources.manage")} onCancel={onClose} footer={null}>
      <Form
        layout="vertical"
        onFinish={() =>
          void run("add", async () => {
            await addMarketSource(repository.trim());
            setRepository("");
            onChange();
          })
        }
      >
        <Form.Item label={t("plugin.market.sources.repository")} htmlFor="market-repository" extra={t("plugin.market.sources.convention")}>
          <div className="flex flex-wrap gap-2">
            <Input
              id="market-repository"
              className="min-w-36 flex-1"
              value={repository}
              onChange={(event) => setRepository(event.target.value)}
              placeholder="https://github.com/owner/repository"
            />
            <Button type="primary" htmlType="submit" loading={pending === "add"} disabled={!repository.trim() || !!pending}>
              {t("plugin.market.sources.add")}
            </Button>
          </div>
        </Form.Item>
      </Form>
      {error && <Alert className="mb-3" type="error" title={error} />}
      <ul className="m-0 list-none space-y-3 p-0">
        {sources.map((source) => (
          <li key={source.id} className="rounded-lg border border-solid border-gray-200 p-3 dark:border-gray-700">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <span className="font-medium wrap-anywhere">
                {source.repository} {source.official && <OfficialBadge />}
              </span>
              <div className="flex gap-1">
                <Button
                  size="small"
                  aria-label={t("plugin.market.sources.refresh_one", { repository: source.repository })}
                  icon={<IconRefresh size={14} />}
                  loading={pending === `refresh:${source.id}` || source.refreshing}
                  disabled={!!pending}
                  onClick={() =>
                    void run(`refresh:${source.id}`, async () => {
                      await refreshMarketSource(source.id);
                      onChange();
                    })
                  }
                />
                {!source.official && (
                  <Popconfirm
                    title={t("plugin.market.sources.remove_confirm")}
                    onConfirm={() =>
                      run(`remove:${source.id}`, async () => {
                        await removeMarketSource(source.id);
                        onChange(source.id);
                      })
                    }
                  >
                    <Button
                      size="small"
                      danger
                      icon={<IconTrash size={14} />}
                      aria-label={t("plugin.market.sources.remove_one", { repository: source.repository })}
                      disabled={!!pending}
                    />
                  </Popconfirm>
                )}
              </div>
            </div>
            {source.syncedAt && (
              <div className="text-xs text-gray-500">{t("plugin.market.synced_at", { time: new Date(source.syncedAt).toLocaleString(i18n.language) })}</div>
            )}
            {source.error && (
              <div role="status" className="mt-1 text-xs wrap-anywhere text-error">
                {source.stale ? t("plugin.market.sources.stale") : t("plugin.market.sources.unavailable")} · {source.error}
              </div>
            )}
          </li>
        ))}
      </ul>
    </Modal>
  );
};

export default MarketSourcesModal;
