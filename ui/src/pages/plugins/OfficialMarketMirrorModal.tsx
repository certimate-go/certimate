import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Alert, Form, Input, Modal, Radio, Select, Skeleton } from "antd";

import { type MarketMirror, type MarketMirrorSettings, fetchOfficialMirror, saveOfficialMirror } from "@/api/pluginmarket";
import { unwrapErrMsg } from "@/utils/error";

interface OfficialMarketMirrorModalProps {
  onClose: () => void;
  onSaved: () => void;
}

const OfficialMarketMirrorModal = ({ onClose, onSaved }: OfficialMarketMirrorModalProps) => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<MarketMirrorSettings>();
  const [mode, setMode] = useState<MarketMirror["mode"]>("default");
  const [preset, setPreset] = useState("");
  const [baseUrl, setBaseUrl] = useState("");
  const [error, setError] = useState<string>();
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let pending = true;
    fetchOfficialMirror()
      .then((result) => {
        if (!pending) return;
        setSettings(result);
        setMode(result.mirror.mode);
        if (result.mirror.mode === "custom") setBaseUrl(result.mirror.baseUrl);
        if (result.mirror.mode === "preset") setPreset(result.mirror.preset);
      })
      .catch((reason: unknown) => {
        if (pending) setError(reason instanceof Error ? unwrapErrMsg(reason) : t("plugin.market.mirror.load_failed"));
      });
    return () => {
      pending = false;
    };
  }, [t]);

  const submit = async () => {
    setSaving(true);
    setError(undefined);
    const mirror: MarketMirror = mode === "custom" ? { mode, baseUrl: baseUrl.trim() } : mode === "preset" ? { mode, preset } : { mode: "default" };
    try {
      await saveOfficialMirror(mirror);
      onSaved();
    } catch (reason) {
      setError(reason instanceof Error ? unwrapErrMsg(reason) : t("plugin.market.mirror.save_failed"));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal
      open
      title={t("plugin.market.mirror.title")}
      onCancel={onClose}
      onOk={() => void submit()}
      confirmLoading={saving}
      closable={!saving}
      mask={{ closable: !saving }}
      keyboard={!saving}
      cancelButtonProps={{ disabled: saving }}
      okButtonProps={{ disabled: !settings || (mode === "preset" && !preset) || (mode === "custom" && !baseUrl.trim()) }}
      okText={t("plugin.market.mirror.save")}
    >
      <p>{t("plugin.market.mirror.help")}</p>
      {error && <Alert className="mb-3" type="error" title={error} />}
      {!settings && !error && <Skeleton active />}
      {settings && (
        <Form layout="vertical" disabled={saving}>
          {settings.error && <Alert className="mb-3" type="warning" title={settings.error} />}
          <Form.Item label={t("plugin.market.mirror.mode")}>
            <Radio.Group
              value={mode}
              onChange={({ target }) => {
                if (target.value === "default" || target.value === "preset" || target.value === "custom") setMode(target.value);
              }}
              options={[
                { label: t("plugin.market.mirror.default"), value: "default" },
                { label: t("plugin.market.mirror.preset"), value: "preset", disabled: !settings.presets.length },
                { label: t("plugin.market.mirror.custom"), value: "custom" },
              ]}
            />
          </Form.Item>
          {mode === "default" && <p className="text-xs text-gray-500">{settings.defaultBase}</p>}
          {mode === "preset" && (
            <Form.Item label={t("plugin.market.mirror.preset")} htmlFor="official-mirror-preset">
              <Select
                id="official-mirror-preset"
                value={preset || undefined}
                onChange={setPreset}
                options={settings.presets.map(({ id, name, baseUrl }) => ({ value: id, label: `${name} · ${baseUrl}` }))}
              />
            </Form.Item>
          )}
          {mode === "custom" && (
            <Form.Item label={t("plugin.market.mirror.base")} htmlFor="official-mirror-url">
              <Input
                id="official-mirror-url"
                value={baseUrl}
                onChange={(event) => setBaseUrl(event.target.value)}
                maxLength={1024}
                placeholder="https://cdn.example.com"
              />
            </Form.Item>
          )}
        </Form>
      )}
    </Modal>
  );
};

export default OfficialMarketMirrorModal;
