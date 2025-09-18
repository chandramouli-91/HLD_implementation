import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

function DBServers() {
  const [vms, setVms] = useState([])
  const [filter, setFilter] = useState('all')

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch('/api/dbservers')
        if (!res.ok) throw new Error('Failed')
        const json = await res.json()
        setVms(json.vms || [])
      } catch (e) { console.error(e) }
    }
    load()
  }, [])

  const filtered = useMemo(() => {
    if (filter === 'all') return vms
    return vms.filter(vm => (vm.overall_fitment_status?.status || '').toLowerCase() === filter)
  }, [vms, filter])

  return (
    <div>
      <header className="topbar">
        <Link to="/" className="home-link"><span className="title">SQL Server Fitment Check</span></Link>
      </header>
      <div className="top-nav">
        <Link to="/summary" className="top-tab">Summary</Link>
        <Link to="/dbservers" className="top-tab active">Database Servers</Link>
        <Link to="/instances" className="top-tab">Instances</Link>
        <Link to="/databases" className="top-tab">Databases</Link>
      </div>

      <div className="container">
        <div className="filter-section">
          <label htmlFor="filter">Filter:</label>
          <select id="filter" value={filter} onChange={e => setFilter(e.target.value)}>
            <option value="all">All</option>
            <option value="passed">VM Fitment = Passed</option>
            <option value="failed">VM Fitment = Failed</option>
          </select>
        </div>

        <div className="table-heading">{`Viewing ${filter === 'all' ? 'All' : filter[0].toUpperCase()+filter.slice(1)} Servers (${filtered.length})`}</div>

        <table id="vmTable">
          <thead>
            <tr>
              <th>Entity Name</th>
              <th>Type</th>
              <th>Fitment Status</th>
              <th>Instances</th>
              <th>Databases</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody id="vmTableBody">
            {filtered.map(vm => (
              <tr key={vm.entity_name}>
                <td>{vm.entity_name}</td>
                <td>{vm.type}</td>
                <td className={(vm.overall_fitment_status?.status === 'Passed') ? 'status-passed' : 'status-failed'}>
                  {vm.overall_fitment_status?.status}
                </td>
                <td>{vm.instances_count}</td>
                <td>{vm.databases_count}</td>
                <td><span className="action-link">Configuration</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default DBServers


