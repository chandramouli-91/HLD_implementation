import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

function Databases() {
  const [databases, setDatabases] = useState([])
  const [filter, setFilter] = useState('all')

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch('/api/databases')
        if (!res.ok) throw new Error('Failed')
        const json = await res.json()
        setDatabases(json.databases || [])
      } catch (e) { console.error(e) }
    }
    load()
  }, [])

  const filtered = useMemo(() => {
    if (filter === 'all') return databases
    return databases.filter(db => (db.overall_fitment_status?.status || '').toLowerCase() === filter)
  }, [databases, filter])

  return (
    <div>
      <header className="topbar">
        <Link to="/" className="home-link"><span className="title">SQL Server Fitment Check</span></Link>
      </header>
      <div className="top-nav">
        <Link to="/summary" className="top-tab">Summary</Link>
        <Link to="/dbservers" className="top-tab">Database Servers</Link>
        <Link to="/instances" className="top-tab">Instances</Link>
        <Link to="/databases" className="top-tab active">Databases</Link>
      </div>

      <div className="container">
        <div className="filter-section">
          <label htmlFor="filter">Filter:</label>
          <select id="filter" value={filter} onChange={e => setFilter(e.target.value)}>
            <option value="all">All</option>
            <option value="passed">Database Fitment = Passed</option>
            <option value="failed">Database Fitment = Failed</option>
          </select>
        </div>

        <div className="table-heading">{`Viewing ${filter === 'all' ? 'All' : filter[0].toUpperCase()+filter.slice(1)} Databases (${filtered.length})`}</div>

        <table id="table">
          <thead>
            <tr>
              <th>Entity Name</th>
              <th>Database VM</th>
              <th>Fitment Status</th>
              <th>Instance</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody id="tableBody">
            {filtered.map(db => (
              <tr key={db.entity_name}>
                <td>{db.entity_name}</td>
                <td>{db.type}</td>
                <td className={(db.overall_fitment_status?.status === 'Passed') ? 'status-passed' : 'status-failed'}>
                  {db.overall_fitment_status?.status}
                </td>
                <td>{db.parent_instance}</td>
                <td><span className="action-link">Configuration</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default Databases


