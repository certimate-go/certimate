import { useEffect, useState } from "react";
import { IconPuzzle } from "@tabler/icons-react";

import { fetchOfficialMarketIcon } from "@/api/pluginmarket";

interface OfficialMarketIconProps {
  path: string;
  fallbackColor: string;
}

const OfficialMarketIcon = ({ path, fallbackColor }: OfficialMarketIconProps) => {
  const [icon, setIcon] = useState<{ path: string; url: string }>();

  useEffect(() => {
    const abort = new AbortController();
    let url: string | undefined;
    fetchOfficialMarketIcon(path, abort.signal)
      .then((body) => {
        if (abort.signal.aborted) return;
        url = URL.createObjectURL(body);
        setIcon({ path, url });
      })
      .catch(() => {});
    return () => {
      abort.abort();
      if (url) URL.revokeObjectURL(url);
    };
  }, [path]);

  return icon?.path === path ? <img src={icon.url} alt="" className="size-full object-contain" /> : <IconPuzzle size={20} color={fallbackColor} />;
};

export default OfficialMarketIcon;
