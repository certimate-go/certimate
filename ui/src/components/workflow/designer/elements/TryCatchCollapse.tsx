import {
  type CustomLabelProps,
  FlowNodeRenderData,
  FlowNodeTransformData,
  FlowRendererRegistry,
  FlowTextKey,
  useBaseColor,
} from "@flowgram.ai/fixed-layout-editor";

export interface TryCatchCollapseProps extends CustomLabelProps {}

const TryCatchCollapse = ({ node, ...props }: TryCatchCollapseProps) => {
  const { baseColor, baseActivatedColor } = useBaseColor();

  const nodeRendererService = node.getService(FlowRendererRegistry);
  const nodeRenderData = node.getData(FlowNodeRenderData)!;
  const nodeTransformData = node.getData(FlowNodeTransformData)!;

  const handleMouseEnter = () => {
    nodeRenderData.activated = true;
  };

  const handleMouseLeave = () => {
    nodeRenderData.activated = false;
  };

  if (!nodeTransformData || !nodeTransformData.parent) {
    return <></>;
  }

  const width = nodeTransformData.inputPoint.x - nodeTransformData.parent.inputPoint.x;
  const height = 40;
  return (
    <div
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{
        width,
        height,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 6,
      }}
    >
      <div
        data-label-id={props.labelId}
        style={{
          fontSize: 12,
          color: nodeRenderData.activated || nodeRenderData.lineActivated ? baseActivatedColor : baseColor,
          textAlign: "center",
          lineHeight: "20px",
          whiteSpace: "nowrap",
          backgroundColor: "var(--g-editor-background)",
        }}
      >
        {nodeRendererService.getText(FlowTextKey.CATCH_TEXT)}
      </div>
    </div>
  );
};

export default TryCatchCollapse;
