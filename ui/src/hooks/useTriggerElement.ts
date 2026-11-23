import { type MouseEventHandler, type ReactNode, cloneElement, isValidElement, useMemo } from "react";

export type UseTriggerElementOptions = {
  onClick?: MouseEventHandler;
};

/**
 * 获取一个触发器元素。
 * 通常为配合 Drawer、Modal 等组件使用。
 * @param {ReactNode} trigger
 * @param {UseTriggerElementOptions} [options]
 * @returns {ReactNode}
 */
const useTriggerElement = (trigger: ReactNode, options?: UseTriggerElementOptions) => {
  const onClick = options?.onClick;
  const triggerDom = useMemo<ReactNode>(() => {
    if (!trigger) {
      return null;
    }

    type ElementProps = Pick<UseTriggerElementOptions, "onClick">;
    if (!isValidElement<ElementProps>(trigger)) {
      return trigger;
    }

    return cloneElement(trigger, {
      onClick: (e) => {
        onClick?.(e);
        trigger.props.onClick?.(e);
      },
    });
  }, [trigger, onClick]);
  return triggerDom;
};

export default useTriggerElement;
