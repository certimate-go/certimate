import { Navigate, Outlet, useLocation } from "react-router-dom";

const ACME = () => {
  const location = useLocation();

  const subpath = location.pathname.split("/")[2];

  if (!subpath) {
    return <Navigate to="/acme/ca" replace />;
  }

  return (
    <>
      <Outlet />
    </>
  );
};

export default ACME;
