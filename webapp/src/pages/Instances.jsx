import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";

function Instances() {
  const [instances, setInstances] = useState([]);
  const [filter, setFilter] = useState("all");

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch("/api/instances");
        if (!res.ok) throw new Error("Failed");
        const json = await res.json();
        setInstances(json.instances || []);
      } catch (e) {
        console.error(e);
      }
    };
    load();
  }, []);

  const filtered = useMemo(() => {
    if (filter === "all") return instances;
    return instances.filter(
      (inst) =>
        (inst.overall_fitment_status?.status || "").toLowerCase() === filter
    );
  }, [instances, filter]);

  return (
    <div>
      <header className="topbar">
        <Link to="/" className="home-link">
          <span className="title">SQL Server Fitment Check</span>
        </Link>
      </header>
      <div className="top-nav">
        <Link to="/summary" className="top-tab">
          Summary
        </Link>
        <Link to="/dbservers" className="top-tab">
          Database Servers
        </Link>
        <Link to="/instances" className="top-tab active">
          Instances
        </Link>
        <Link to="/databases" className="top-tab">
          Databases
        </Link>
      </div>

      <div className="container">
        <div className="filter-section">
          <label htmlFor="filter">Filter:</label>
          <select
            id="filter"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          >
            <option value="all">All</option>
            <option value="passed">Instance Fitment = Passed</option>
            <option value="failed">Instance Fitment = Failed</option>
          </select>
        </div>

        <div className="table-heading">{`Viewing ${
          filter === "all" ? "All" : filter[0].toUpperCase() + filter.slice(1)
        } Instances (${filtered.length})`}</div>

        <table id="instanceTable">
          <thead>
            <tr>
              <th>Entity Name</th>
              <th>Type</th>
              <th>Fitment Status</th>
              <th>Databases</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody id="instanceTableBody">
            {filtered.map((inst) => (
              <tr key={inst.entity_name}>
                <td>{inst.entity_name}</td>
                <td>{inst.type}</td>
                <td
                  className={
                    inst.overall_fitment_status?.status === "Passed"
                      ? "status-passed"
                      : "status-failed"
                  }
                >
                  {inst.overall_fitment_status?.status}
                </td>
                <td>{inst.databases_count}</td>
                <td>
                  <span className="action-link">Configuration</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default Instances;
