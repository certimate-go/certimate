import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { AutoComplete, Input } from "antd";

import { type MarketEntry, type MarketQuery, type PluginTarget, fetchMarketListing } from "@/api/pluginmarket";
import OfficialBadge from "./OfficialBadge";

interface PluginMarketSearchProps {
  draft: string;
  filters: Pick<MarketQuery, "status" | "capability" | "source" | "locale">;
  onDraftChange: (value: string) => void;
  onSearch: (keyword: string) => void;
  onSelect: (target: PluginTarget) => void;
}

const PluginMarketSearch = ({ draft, filters: { status, capability, source, locale }, onDraftChange, onSearch, onSelect }: PluginMarketSearchProps) => {
  const { t, i18n } = useTranslation();
  const [suggestionKeyword, setSuggestionKeyword] = useState("");
  const [suggestions, setSuggestions] = useState<{ query: string; entries: MarketEntry[] }>({ query: "", entries: [] });
  const [open, setOpen] = useState(false);
  const suggestionQuery = JSON.stringify([suggestionKeyword, status, capability, source, locale]);
  const entries = suggestions.query === suggestionQuery ? suggestions.entries : [];
  const generation = useRef(0);
  const focused = useRef(false);

  const clearSuggestions = () => {
    generation.current++;
    setSuggestions({ query: "", entries: [] });
    setOpen(false);
  };

  useEffect(() => {
    let cancelled = false;
    const request = ++generation.current;
    const timer = window.setTimeout(async () => {
      if (!suggestionKeyword.trim()) return;
      try {
        const result = await fetchMarketListing({
          page: 1,
          perPage: 10,
          keyword: suggestionKeyword.trim(),
          status,
          capability,
          source,
          locale,
          refresh: false,
        });
        if (cancelled || request !== generation.current) return;
        setSuggestions({ query: suggestionQuery, entries: result.items });
        setOpen(focused.current);
      } catch {
        if (!cancelled && request === generation.current) setSuggestions({ query: "", entries: [] });
      }
    }, 250);
    return () => {
      clearTimeout(timer);
      cancelled = true;
    };
  }, [suggestionKeyword, status, capability, source, locale, suggestionQuery]);

  const options = entries.map((entry) => {
    const name = entry.display_name_key && i18n.exists(entry.display_name_key) ? t(entry.display_name_key, entry.provider_type) : entry.provider_type;
    return {
      value: `${entry.source.id}/${entry.provider_type}`,
      entry,
      label: (
        <div className="min-w-0">
          <div className="flex items-center gap-1">
            <span className="min-w-0 truncate">{name}</span>
            {entry.official && <OfficialBadge />}
          </div>
          <div className="truncate text-xs text-gray-500">
            {entry.provider_type} · {entry.source.repository || t("plugin.market.sources.local")}
          </div>
        </div>
      ),
    };
  });

  return (
    <AutoComplete
      className="max-w-80 min-w-32 flex-1"
      value={draft}
      options={options}
      open={open && !!suggestionKeyword.trim() && entries.length > 0}
      defaultActiveFirstOption={false}
      filterOption={false}
      onOpenChange={setOpen}
      onFocus={() => {
        focused.current = true;
      }}
      onBlur={() => {
        focused.current = false;
        setOpen(false);
      }}
      onSearch={(value) => {
        clearSuggestions();
        onDraftChange(value);
        setSuggestionKeyword(value);
      }}
      onSelect={(_, option) => {
        clearSuggestions();
        setSuggestionKeyword("");
        onDraftChange(option.entry.provider_type);
        onSelect({ providerType: option.entry.provider_type, source: option.entry.source.id });
      }}
    >
      <Input.Search
        allowClear
        className="text-sm placeholder:text-sm"
        aria-label={t("plugin.market.action.search_placeholder")}
        placeholder={t("plugin.market.action.search_placeholder")}
        size="large"
        onSearch={(value, event) => {
          if (event && "key" in event && event.key === "Enter" && event.currentTarget instanceof HTMLInputElement) {
            const active = event.currentTarget.getAttribute("aria-activedescendant");
            if (active && document.getElementById(active)) return;
          }
          clearSuggestions();
          setSuggestionKeyword("");
          onSearch(value.trim());
        }}
      />
    </AutoComplete>
  );
};

export default PluginMarketSearch;
