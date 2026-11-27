import { useTranslation } from "react-i18next";
import { IconBook, IconBrandGithub, IconBrandTelegram, IconCoin, IconMessageChatbot } from "@tabler/icons-react";
import { Badge, Button, Card, List, Tooltip, Typography } from "antd";

import { APP_DOCUMENT_URL, APP_DOWNLOAD_URL, APP_REPO_URL, APP_VERSION } from "@/domain/app";
import { useVersionChecker } from "@/hooks";

const SettingsAbout = () => {
  const { t } = useTranslation();

  const { hasUpdate } = useVersionChecker();

  const handleDownloadClick = () => {
    window.open(APP_DOWNLOAD_URL, "_blank");
  };

  const handleDocumentClick = () => {
    window.open(APP_DOCUMENT_URL, "_blank");
  };

  const handleGithubClick = () => {
    window.open(APP_REPO_URL, "_blank");
  };

  const handleTelegramClick = () => {
    window.open("https://t.me/+ZXphsppxUg41YmVl", "_blank");
  };

  const handleDonateClick = () => {
    window.open("https://profile.ikit.fun/sponsors/", "_blank");
  };

  const handleFeedbackClick = () => {
    window.open(APP_REPO_URL + "/issues", "_blank");
  };

  return (
    <>
      <div>
        <Card>
          <h2 className="text-center text-2xl font-semibold">Certimate</h2>
          <div className="mt-6 mb-4">
            <div className="flex items-center justify-center gap-2">
              <p className="text-sm font-normal whitespace-normal text-foreground/70">Version: {APP_VERSION}</p>
              <Badge className="cursor-pointer" count={hasUpdate ? t("settings.about.version.new") : void 0} onClick={handleDownloadClick} />
            </div>
          </div>
          <div className="mb-2 flex flex-wrap items-center justify-center gap-2">
            <Tooltip title={t("settings.about.socials.document")}>
              <Button type="text" icon={<IconBook size="1.5em" onClick={handleDocumentClick} />} />
            </Tooltip>
            <Tooltip title={t("settings.about.socials.github")}>
              <Button type="text" icon={<IconBrandGithub size="1.5em" onClick={handleGithubClick} />} />
            </Tooltip>
            <Tooltip title={t("settings.about.socials.telegram")}>
              <Button type="text" icon={<IconBrandTelegram size="1.5em" onClick={handleTelegramClick} />} />
            </Tooltip>
            <Tooltip title={t("settings.about.socials.donate")}>
              <Button type="text" icon={<IconCoin size="1.5em" onClick={handleDonateClick} />} />
            </Tooltip>
          </div>
        </Card>
      </div>

      <div>
        <Card style={{ background: "transparent" }} size="small">
          <div className="flex items-center gap-4 px-4">
            <span className="relative -top-2.5">
              <IconMessageChatbot size="1.5em" />
            </span>
            <div className="mt-4 mb-0 flex flex-1 items-center overflow-hidden">
              <div className="flex-1">
                <h2 className="truncate text-base font-semibold">{t("settings.about.feedback.title")}</h2>
                <p className="text-sm font-normal whitespace-normal text-foreground/70">{t("settings.about.feedback.subtitle")}</p>
              </div>
              <Button className="relative -top-2.5" onClick={handleFeedbackClick}>
                {t("settings.about.feedback.button")}
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </>
  );
};

export default SettingsAbout;
