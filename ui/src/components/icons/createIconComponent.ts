import { createElement, forwardRef } from "react";
import { type Icon, type IconNode, type IconProps } from "@tabler/icons-react";

const defaultAttrs = {
  outline: {
    xmlns: "http://www.w3.org/2000/svg",
    width: 24,
    height: 24,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 2,
    strokeLinecap: "round",
    strokeLinejoin: "round",
  },
  filled: {
    xmlns: "http://www.w3.org/2000/svg",
    width: 24,
    height: 24,
    viewBox: "0 0 24 24",
    fill: "currentColor",
    stroke: "none",
  },
};

const createIconComponent = (type: "outline" | "filled", iconName: string, iconAttrs: Record<string, any>, iconNode: IconNode) => {
  const Component = forwardRef<Icon, IconProps>(({ color = "currentColor", size = 24, stroke = 2, title, className, children, ...rest }: IconProps, ref) =>
    createElement(
      "svg",
      {
        ref,
        ...defaultAttrs[type],
        ...iconAttrs,
        width: size,
        height: size,
        className: ["icon", className],
        ...(type === "filled"
          ? {
              fill: color,
            }
          : {
              strokeWidth: stroke,
              stroke: color,
            }),
        ...rest,
      },
      [
        title && createElement("title", { key: "svg-title" }, title),
        ...iconNode.map(([tag, attrs]) => createElement(tag, attrs)),
        ...(Array.isArray(children) ? children : [children]),
      ]
    )
  );

  Component.displayName = iconName;

  return Component;
};

export default createIconComponent;
