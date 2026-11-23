import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { EditorState, FlowLayoutDefault, useClientContext, usePlaygroundTools, useRefresh } from "@flowgram.ai/fixed-layout-editor";
import { IconHandStop, IconLayoutCards, IconMatrix, IconMaximize, IconMinus, IconPlus } from "@tabler/icons-react";
import { Button, type ButtonProps, Dropdown, Tooltip } from "antd";

import Show from "@/components/Show";
import { mergeCls } from "@/utils/css";

import Minimap from "./Minimap";

export interface ToolbarProps {
  className?: string;
  style?: React.CSSProperties;
  size?: ButtonProps["size"];
  showLayout?: boolean;
  showMinimap?: boolean;
  showMouseState?: boolean;
  showZoom?: boolean;
  showZoomFit?: boolean;
  showZoomLevel?: boolean;
}

const Toolbar = ({
  className,
  style,
  size,
  showLayout = true,
  showMinimap = true,
  showMouseState = true,
  showZoom = true,
  showZoomFit = true,
  showZoomLevel = true,
}: ToolbarProps) => {
  const { t } = useTranslation();

  const ctx = useClientContext();
  const { playground } = ctx;

  const tools = usePlaygroundTools({ minZoom: 0.1, maxZoom: 3 });

  const refresh = useRefresh();

  useEffect(() => {
    const d = playground.config.onReadonlyOrDisabledChange(() => refresh());

    return () => d.dispose();
  }, [playground]);

  const buttonIconSize = size === "large" ? "1.5em" : size === "small" ? "1em" : "1.25em";

  const [isMinimapVisible, setIsMinimapVisible] = useState(() => window.screen.availWidth >= 1024);

  const [isMouseFriendly, setIsMouseFriendly] = useState(() => playground.editorState.is(EditorState.STATE_MOUSE_FRIENDLY_SELECT.id));

  const handleToggleLayout = () => {
    if (tools.isVertical) {
      tools.changeLayout(FlowLayoutDefault.HORIZONTAL_FIXED_LAYOUT);
    } else {
      tools.changeLayout(FlowLayoutDefault.VERTICAL_FIXED_LAYOUT);
    }
  };

  const handleToggleMinimap = () => {
    setIsMinimapVisible((prev) => !prev);
  };

  const handleToggleMouseFriendly = () => {
    if (isMouseFriendly) {
      playground.editorState.changeState(EditorState.STATE_SELECT.id);
      setIsMouseFriendly(false);
    } else {
      playground.editorState.changeState(EditorState.STATE_MOUSE_FRIENDLY_SELECT.id);
      setIsMouseFriendly(true);
    }
  };

  return (
    <div className={className} style={style}>
      <div className="relative flex items-center gap-2">
        <Show when={showMouseState}>
          <Tooltip title={isMouseFriendly ? t("workflow.detail.design.toolbar.drag_mode") : t("workflow.detail.design.toolbar.pointer_mode")}>
            <Button
              ghost={isMouseFriendly}
              icon={<IconHandStop size={buttonIconSize} />}
              size={size}
              type={isMouseFriendly ? "primary" : "default"}
              onClick={handleToggleMouseFriendly}
            />
          </Tooltip>
        </Show>

        <Show when={showZoom}>
          <Tooltip title={t("workflow.detail.design.toolbar.zoomout")}>
            <Button icon={<IconMinus size={buttonIconSize} />} size={size} onClick={() => tools.zoomout()} />
          </Tooltip>
          <Show when={showZoomLevel}>
            <Dropdown
              menu={{
                items: [
                  ...[200, 100, 75, 50, 25].map((zoom) => ({
                    key: `${zoom}%`,
                    label: `${zoom}%`,
                    onClick: () => tools.updateZoom(zoom / 100),
                  })),
                  {
                    type: "divider",
                  },
                  {
                    key: "auto",
                    label: t("workflow.detail.design.toolbar.auto_fit"),
                    onClick: () => tools.fitView(),
                  },
                ],
              }}
              trigger={["click"]}
            >
              <Button className="w-16 text-center" size={size}>
                {Math.round(tools.zoom * 100)}%
              </Button>
            </Dropdown>
          </Show>
          <Tooltip title={t("workflow.detail.design.toolbar.zoomin")}>
            <Button icon={<IconPlus size={buttonIconSize} />} size={size} onClick={() => tools.zoomin()} />
          </Tooltip>
          <Show when={showZoomFit}>
            <Tooltip title={t("workflow.detail.design.toolbar.auto_fit")}>
              <Button icon={<IconMaximize size={buttonIconSize} />} size={size} onClick={() => tools.fitView()} />
            </Tooltip>
          </Show>
        </Show>

        <Show when={showLayout}>
          <Tooltip title={tools.isVertical ? t("workflow.detail.design.toolbar.vertical_layout") : t("workflow.detail.design.toolbar.horizontal_layout")}>
            <Button
              icon={<IconLayoutCards className={mergeCls({ ["rotate-90"]: tools.isVertical })} size={buttonIconSize} />}
              size={size}
              onClick={handleToggleLayout}
            />
          </Tooltip>
        </Show>

        <Show when={showMinimap}>
          <Tooltip title={t("workflow.detail.design.toolbar.minimap")}>
            <Button
              icon={<IconMatrix size={buttonIconSize} />}
              ghost={isMinimapVisible}
              type={isMinimapVisible ? "primary" : "default"}
              size={size}
              onClick={handleToggleMinimap}
            />
          </Tooltip>
          {isMinimapVisible && <Minimap className="absolute right-0 bottom-[42px]" />}
        </Show>
      </div>
    </div>
  );
};

export default Toolbar;
