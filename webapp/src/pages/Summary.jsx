import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

function Summary() {
  const [data, setData] = useState(null)

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch('/api/summary')
        if (!res.ok) throw new Error('Failed')
        const json = await res.json()
        setData(json)
      } catch (e) {
        console.error(e)
      }
    }
    load()
  }, [])

  const [activeTab, setActiveTab] = useState('Database Server')

  const checkWise = data?.summary?.check_wise || []
  const entity = data?.summary?.entity_wise

  return (
    <div>
      <header className="topbar">
        <Link to="/" className="home-link"><span className="title">SQL Server Fitment Check</span></Link>
      </header>
      <div className="top-nav">
        <Link to="/summary" className="top-tab active">Summary</Link>
        <Link to="/dbservers" className="top-tab">Database Servers</Link>
        <Link to="/instances" className="top-tab">Instances</Link>
        <Link to="/databases" className="top-tab">Databases</Link>
      </div>

      <div className="container">
        <div className="card">
          <h2>Entities Wise Summary</h2>
          <div className="summary-block">
            <div>
              <div className="summary-count" id="server-count">{entity?.total_vms ?? '-'}</div>
              <div className="summary-text">Total DBServers Checked</div>
            </div>
            <div className="summary-details">
              <h3>Check Status by Database Servers</h3>
              <div className="progress-bar">
                <div className="progress-fill failed" style={{ width: `${entity ? (entity.failed_vms / (entity.total_vms || 1)) * 100 : 0}%` }}></div>
                <div className="progress-fill passed" style={{ width: `${entity ? 100 - ((entity.failed_vms / (entity.total_vms || 1)) * 100) : 0}%` }}></div>
              </div>
              <p><span className="failed">{entity?.failed_vms ?? '-'}</span> / <span>{entity?.total_vms ?? '-'}</span> Failed</p>
            </div>
          </div>

          <div className="summary-block">
            <div>
              <div className="summary-count" id="instance-count">{entity?.total_instances ?? '-'}</div>
              <div className="summary-text">Total Instances Checked</div>
            </div>
            <div className="summary-details">
              <h3>Check Status by Instances</h3>
              <div className="progress-bar">
                <div className="progress-fill failed" style={{ width: `${entity ? (entity.failed_instances / (entity.total_instances || 1)) * 100 : 0}%` }}></div>
                <div className="progress-fill passed" style={{ width: `${entity ? 100 - ((entity.failed_instances / (entity.total_instances || 1)) * 100) : 0}%` }}></div>
              </div>
              <p><span className="failed">{entity?.failed_instances ?? '-'}</span> / <span>{entity?.total_instances ?? '-'}</span> Failed</p>
            </div>
          </div>

          <div className="summary-block">
            <div>
              <div className="summary-count" id="database-count">{entity?.total_databases ?? '-'}</div>
              <div className="summary-text">Total Databases Checked</div>
            </div>
            <div className="summary-details">
              <h3>Check Status by Databases</h3>
              <div className="progress-bar">
                <div className="progress-fill failed" style={{ width: `${entity ? (entity.failed_databases / (entity.total_databases || 1)) * 100 : 0}%` }}></div>
                <div className="progress-fill passed" style={{ width: `${entity ? 100 - ((entity.failed_databases / (entity.total_databases || 1)) * 100) : 0}%` }}></div>
              </div>
              <p><span className="failed">{entity?.failed_databases ?? '-'}</span> / <span>{entity?.total_databases ?? '-'}</span> Failed</p>
            </div>
          </div>

          <div className="legend">
            <span className="failed">✘ Failed</span>
            <span className="passed">✔ Passed</span>
          </div>
        </div>

        <div className="card">
          <h2>Check Wise Summary</h2>
          <div className="tabs">
            {['Database Server','Instance','Database'].map(t => (
              <button key={t} className={`tab ${activeTab===t ? 'active' : ''}`} onClick={() => setActiveTab(t)}>{t} Checks</button>
            ))}
          </div>
          <table className="summary-table">
            <thead>
              <tr><th>Checks</th><th>Passed</th><th>Failed</th></tr>
            </thead>
            <tbody id="checks-table-body">
              {(checkWise.find(x => x.type === activeTab)?.categories || []).flatMap(cat => [
                <tr key={cat.category_name} className="category-row">
                  <td style={{ cursor: 'pointer' }}>{cat.category_name}</td>
                  <td className="passed">✔ {cat.passed}</td>
                  <td className="failed">✘ {cat.failed}</td>
                  <td className="toggle-cell"><span className="toggle">▼</span></td>
                </tr>,
                ...cat.check.map(ch => (
                  <tr key={`${cat.category_name}-${ch.check_name}`} className="check-row">
                    <td style={{ paddingLeft: 20 }}>{ch.check_name}</td>
                    <td className="passed">✔ {ch.passed}</td>
                    <td className="failed">✘ {ch.failed}</td>
                    <td></td>
                  </tr>
                ))
              ])}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export default Summary


