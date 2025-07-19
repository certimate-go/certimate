import { Outlet } from "react-router-dom";
import { Layout } from "antd";

const ErrorLayout = ({ children }: { children?: React.ReactNode }) => {
  return (
    <Layout className="h-screen">
      <div className="relative">{children || <Outlet />}</div>
    </Layout>
  );
};

export default ErrorLayout;
