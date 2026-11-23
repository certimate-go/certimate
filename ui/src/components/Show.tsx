import { Children as ReactChildren, isValidElement } from "react";

export type ShowProps =
  | {
      children: React.ReactNode;
    }
  | {
      when: boolean;
      children: React.ReactNode;
      fallback?: React.ReactNode;
    };

const Show = ({ children, ...props }: ShowProps) => {
  if ("when" in props) {
    const { when, fallback } = props;
    return when ? children : fallback;
  }

  let fallback: React.ReactNode | undefined;

  const cases = ReactChildren.toArray(children);
  for (let i = 0; i < cases.length; i++) {
    type ElementProps = { when: boolean; children: React.ReactNode };
    const child = cases[i];
    if (isValidElement<ElementProps>(child)) {
      if (child.type === Case && child.props.when) {
        return child.props.children;
      } else if (child.type === Default) {
        if (fallback) {
          console.warn("[certimate] multiple Default components found in Show. Only the first will be used.");
          continue;
        }
        fallback = child.props.children;
      }
    }
  }

  return fallback;
};

const Case = ({ children, when }: { children: React.ReactNode; when: boolean }) => {
  return when ? children : null;
};

const Default = ({ children }: { children: React.ReactNode }) => {
  return children;
};

const _default = Object.assign(Show, {
  Case,
  Default,
});

export default _default;
