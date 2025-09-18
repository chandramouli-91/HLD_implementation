import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import "./index.css";
import Home from "./pages/Home.jsx";
import Summary from "./pages/Summary.jsx";
import DBServers from "./pages/DBServers.jsx";
import Instances from "./pages/Instances.jsx";
import Databases from "./pages/Databases.jsx";
import "@nutanix-ui/prism-reactjs/dist/index.css";

const router = createBrowserRouter([
  { path: "/", element: <Home /> },
  { path: "/summary", element: <Summary /> },
  { path: "/dbservers", element: <DBServers /> },
  { path: "/instances", element: <Instances /> },
  { path: "/databases", element: <Databases /> },
]);

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>
);
