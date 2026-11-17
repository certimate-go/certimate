import { useTranslation } from "react-i18next";
import { IconRosetteDiscountCheckFilled } from "@tabler/icons-react";
import { Tooltip } from "antd";

const OfficialBadge = () => {
  const { t } = useTranslation();
  return (
    <Tooltip title={t("plugin.market.sources.official_tooltip")} trigger={["hover", "focus"]}>
      <span role="img" aria-label={t("plugin.market.sources.official")} tabIndex={0} className="inline-flex shrink-0 align-middle text-blue-500">
        <IconRosetteDiscountCheckFilled size={15} aria-hidden />
      </span>
    </Tooltip>
  );
};

export default OfficialBadge;
