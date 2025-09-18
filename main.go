package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	port       = ":8080"
	scriptPath = "./script.ps1"
)

// ComprehensiveResult represents the result from PowerShell script
type ComprehensiveResult struct {
	Success        bool        `json:"Success"`
	Target         string      `json:"Target"`
	VMChecks       []CheckItem `json:"VMChecks"`
	InstanceChecks []CheckItem `json:"InstanceChecks"`
	DatabaseChecks []CheckItem `json:"DatabaseChecks"`
	ErrorMessage   string      `json:"ErrorMessage,omitempty"`
}

type CheckItem struct {
	Check    string `json:"Check"`
	Status   string `json:"Status"`
	Message  string `json:"Message"`
	Severity string `json:"Severity"`
}

type CheckResult struct {
	Check    string `json:"Check"`
	Status   string `json:"Status"`
	Message  string `json:"Message"`
	Severity string `json:"Severity"`
}

type BatchResponse struct {
	Timestamp       string                   `json:"timestamp"`
	Total           int                      `json:"total"`
	Passed          int                      `json:"passed"`
	Failed          int                      `json:"failed"`
	VMResults       map[string][]CheckResult `json:"VMResults"`
	InstanceResults map[string][]CheckResult `json:"InstanceResults"`
	DatabaseResults map[string][]CheckResult `json:"DatabaseResults"`
	Summary         SummaryStats             `json:"Summary"`
}

type SummaryStats struct {
	TotalServers   int `json:"TotalServers"`
	TotalInstances int `json:"TotalInstances"`
	TotalDatabases int `json:"TotalDatabases"`
	TotalChecks    int `json:"TotalChecks"`
	PassedChecks   int `json:"PassedChecks"`
	FailedChecks   int `json:"FailedChecks"`
	ErrorChecks    int `json:"ErrorChecks"`
}

type CheckRequest struct {
	Hostnames string `json:"hostnames"`
}

// ===== Globals =====

var lastCheckResults *BatchResponse
var processedVMs int
var totalVMs int
var progressMu sync.Mutex

// ===== Worker Logic =====

func worker(id int, jobs <-chan string, wg *sync.WaitGroup, response *BatchResponse,
	mu *sync.Mutex, totalChecks *int, passedChecks *int, failedChecks *int, errorChecks *int) {

	defer wg.Done()
	for hostname := range jobs {
		logrus.Infof("Worker %d processing hostname: %s", id, hostname)

		psResult, err := runPowerShellScript(hostname)

		mu.Lock()
		if err != nil {
			logrus.Errorf("PowerShell execution failed for %s: %v", hostname, err)
			response.VMResults[hostname] = []CheckResult{{
				Check:    "PowerShell Execution Policy",
				Status:   "ERROR",
				Message:  fmt.Sprintf("Failed to execute checks: %v", err),
				Severity: "CRITICAL",
			}}
			response.Failed++
			*errorChecks++
			*totalChecks++
		} else {
			processResults(hostname, psResult, response, totalChecks, passedChecks, failedChecks, errorChecks)
		}

		progressMu.Lock()
		processedVMs++
		progressMu.Unlock()

		mu.Unlock()
	}
}

func processResults(hostname string, psResult *ComprehensiveResult, response *BatchResponse,
	totalChecks *int, passedChecks *int, failedChecks *int, errorChecks *int) {

	// VM checks
	if len(psResult.VMChecks) > 0 {
		vmResults := make([]CheckResult, 0, len(psResult.VMChecks))
		for _, check := range psResult.VMChecks {
			vmResults = append(vmResults, CheckResult(check))
			*totalChecks++
			switch check.Status {
			case "SUCCESS":
				*passedChecks++
			case "FAILED":
				*failedChecks++
			case "ERROR":
				*errorChecks++
			}
		}
		response.VMResults[hostname] = vmResults
	}

	// Instance checks
	if len(psResult.InstanceChecks) > 0 {
		instanceResults := make([]CheckResult, 0, len(psResult.InstanceChecks))
		for _, check := range psResult.InstanceChecks {
			instanceResults = append(instanceResults, CheckResult(check))
			*totalChecks++
			switch check.Status {
			case "SUCCESS":
				*passedChecks++
			case "FAILED":
				*failedChecks++
			case "ERROR":
				*errorChecks++
			}
		}
		response.InstanceResults[hostname+"\\MSSQLSERVER"] = instanceResults
	}

	// Database checks
	if len(psResult.DatabaseChecks) > 0 {
		databaseResults := make([]CheckResult, 0, len(psResult.DatabaseChecks))
		for _, check := range psResult.DatabaseChecks {
			databaseResults = append(databaseResults, CheckResult(check))
			*totalChecks++
			switch check.Status {
			case "SUCCESS":
				*passedChecks++
			case "FAILED":
				*failedChecks++
			case "ERROR":
				*errorChecks++
			}
		}
		response.DatabaseResults[hostname+"\\TestDB"] = databaseResults
	}

	// Passed/Failed decision
	hasFailure := false
	for _, check := range psResult.VMChecks {
		if check.Status == "FAILED" || check.Status == "ERROR" {
			hasFailure = true
			break
		}
	}
	for _, check := range psResult.InstanceChecks {
		if check.Status == "FAILED" || check.Status == "ERROR" {
			hasFailure = true
			break
		}
	}
	for _, check := range psResult.DatabaseChecks {
		if check.Status == "FAILED" || check.Status == "ERROR" {
			hasFailure = true
			break
		}
	}

	if hasFailure {
		response.Failed++
	} else {
		response.Passed++
	}
}

// ===== API Handlers =====

// func handleCheck(c *gin.Context) {
// 	var req CheckRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
// 		return
// 	}
// 	if req.Hostnames == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Hostnames cannot be empty"})
// 		return
// 	}

// 	// Parse hostnames
// 	hostnames := strings.Split(req.Hostnames, ",")
// 	for i := range hostnames {
// 		hostnames[i] = strings.TrimSpace(hostnames[i])
// 	}

// 	logrus.Infof("Processing checks for %d hostnames: %v", len(hostnames), hostnames)

// 	// Reset counters
// 	progressMu.Lock()
// 	processedVMs = 0
// 	totalVMs = len(hostnames)
// 	progressMu.Unlock()

// 	// Init response
// 	response := BatchResponse{
// 		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
// 		Total:           len(hostnames),
// 		Passed:          0,
// 		Failed:          0,
// 		VMResults:       make(map[string][]CheckResult),
// 		InstanceResults: make(map[string][]CheckResult),
// 		DatabaseResults: make(map[string][]CheckResult),
// 	}

// 	var totalChecks, passedChecks, failedChecks, errorChecks int
// 	var mu sync.Mutex

// 	jobs := make(chan string, len(hostnames))
// 	var wg sync.WaitGroup

// 	// Start workers
// 	numWorkers := 10
// 	wg.Add(numWorkers)
// 	for w := 1; w <= numWorkers; w++ {
// 		go worker(w, jobs, &wg, &response, &mu, &totalChecks, &passedChecks, &failedChecks, &errorChecks)
// 	}

// 	// Push jobs
// 	for _, h := range hostnames {
// 		jobs <- h
// 	}
// 	close(jobs)

// 	wg.Wait()

// 	// Summary
// 	response.Summary = SummaryStats{
// 		TotalServers:   len(hostnames),
// 		TotalInstances: len(response.InstanceResults),
// 		TotalDatabases: len(response.DatabaseResults),
// 		TotalChecks:    totalChecks,
// 		PassedChecks:   passedChecks,
// 		FailedChecks:   failedChecks,
// 		ErrorChecks:    errorChecks,
// 	}

// 	lastCheckResults = &response

// 	c.JSON(http.StatusOK, response)
// }

func handleCheck(c *gin.Context) {
	var req CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if req.Hostnames == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hostnames cannot be empty"})
		return
	}

	// Parse hostnames
	hostnames := strings.Split(req.Hostnames, ",")
	for i := range hostnames {
		hostnames[i] = strings.TrimSpace(hostnames[i])
	}

	logrus.Infof("Processing checks for %d hostnames: %v", len(hostnames), hostnames)

	// Reset counters
	progressMu.Lock()
	processedVMs = 0
	totalVMs = len(hostnames)
	progressMu.Unlock()

	// Init response struct
	response := &BatchResponse{
		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
		Total:           len(hostnames),
		Passed:          0,
		Failed:          0,
		VMResults:       make(map[string][]CheckResult),
		InstanceResults: make(map[string][]CheckResult),
		DatabaseResults: make(map[string][]CheckResult),
	}
	lastCheckResults = response

	go func() {
		var totalChecks, passedChecks, failedChecks, errorChecks int
		var mu sync.Mutex
		jobs := make(chan string, len(hostnames))
		var wg sync.WaitGroup

		// Start workers
		numWorkers := 10
		wg.Add(numWorkers)
		for w := 1; w <= numWorkers; w++ {
			go worker(w, jobs, &wg, response, &mu, &totalChecks, &passedChecks, &failedChecks, &errorChecks)
		}

		// Send jobs
		for _, h := range hostnames {
			jobs <- h
		}
		close(jobs)

		// Wait for workers in background
		wg.Wait()

		// Finalize summary once done
		mu.Lock()
		response.Summary = SummaryStats{
			TotalServers:   len(hostnames),
			TotalInstances: len(response.InstanceResults),
			TotalDatabases: len(response.DatabaseResults),
			TotalChecks:    totalChecks,
			PassedChecks:   passedChecks,
			FailedChecks:   failedChecks,
			ErrorChecks:    errorChecks,
		}
		mu.Unlock()

		logrus.Infof("All workers finished. Summary ready for %d hosts.", len(hostnames))
	}()

	c.JSON(http.StatusOK, gin.H{"status": "started", "total": len(hostnames)})
}

func getProgress(c *gin.Context) {
	progressMu.Lock()
	defer progressMu.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"processed": processedVMs,
		"total":     totalVMs,
	})
}

// ===== PowerShell runner =====

func runPowerShellScript(hostname string) (*ComprehensiveResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell.exe",
		"-ExecutionPolicy", "Bypass",
		"-File", scriptPath,
		"-ComputerName", hostname)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("PowerShell execution failed: %v, output: %s", err, string(output))
	}

	logrus.Infof("PowerShell raw output for %s: %s", hostname, string(output))

	var result ComprehensiveResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PowerShell output: %v, raw output: %s", err, string(output))
	}
	return &result, nil
}

// handleSummaryAPI returns summary data for the new UI
func handleSummaryAPI(c *gin.Context) {
	if lastCheckResults == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No check results available"})
		return
	}

	// Transform data to match new UI expectations
	summary := gin.H{
		"summary": gin.H{
			"entity_wise": gin.H{
				"total_vms":        lastCheckResults.Summary.TotalServers,
				"failed_vms":       lastCheckResults.Failed,
				"total_instances":  lastCheckResults.Summary.TotalInstances,
				"failed_instances": calculateFailedInstances(),
				"total_databases":  lastCheckResults.Summary.TotalDatabases,
				"failed_databases": calculateFailedDatabases(),
			},
			"check_wise": []gin.H{
				{
					"type": "Database Server",
					"categories": []gin.H{
						{
							"category_name": "VM Checks",
							"passed":        countChecksByStatus("VM", "SUCCESS"),
							"failed":        countChecksByStatus("VM", "FAILED") + countChecksByStatus("VM", "ERROR"),
							"check": []gin.H{
								{
									"check_name": "PowerShell Execution Policy",
									"passed":     countChecksByStatus("VM", "SUCCESS"),
									"failed":     countChecksByStatus("VM", "FAILED") + countChecksByStatus("VM", "ERROR"),
								},
							},
						},
					},
				},
				{
					"type": "Instance",
					"categories": []gin.H{
						{
							"category_name": "Instance Checks",
							"passed":        countChecksByStatus("Instance", "SUCCESS"),
							"failed":        countChecksByStatus("Instance", "FAILED") + countChecksByStatus("Instance", "ERROR"),
							"check": []gin.H{
								{
									"check_name": "Database Count Validation",
									"passed":     countChecksByStatus("Instance", "SUCCESS"),
									"failed":     countChecksByStatus("Instance", "FAILED") + countChecksByStatus("Instance", "ERROR"),
								},
							},
						},
					},
				},
				{
					"type": "Database",
					"categories": []gin.H{
						{
							"category_name": "Database Checks",
							"passed":        countChecksByStatus("Database", "SUCCESS"),
							"failed":        countChecksByStatus("Database", "FAILED") + countChecksByStatus("Database", "ERROR"),
							"check": []gin.H{
								{
									"check_name": "Database State",
									"passed":     countChecksByStatus("Database", "SUCCESS"),
									"failed":     countChecksByStatus("Database", "FAILED") + countChecksByStatus("Database", "ERROR"),
								},
							},
						},
					},
				},
			},
		},
	}

	c.JSON(http.StatusOK, summary)
}

// handleDBServersAPI returns database servers data for the new UI
func handleDBServersAPI(c *gin.Context) {
	if lastCheckResults == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No check results available"})
		return
	}

	var vms []gin.H
	for hostname, vmChecks := range lastCheckResults.VMResults {
		hasFailure := false
		for _, check := range vmChecks {
			if check.Status == "FAILED" || check.Status == "ERROR" {
				hasFailure = true
				break
			}
		}

		// compute per-VM instance/database counts
		instancesCount := 0
		databasesCount := 0
		for instanceName := range lastCheckResults.InstanceResults {
			if strings.HasPrefix(instanceName, hostname+"\\") {
				instancesCount++
			}
		}
		for dbName := range lastCheckResults.DatabaseResults {
			if strings.HasPrefix(dbName, hostname+"\\") {
				databasesCount++
			}
		}

		vm := gin.H{
			"entity_name": hostname,
			"type":        "Database VM",
			"overall_fitment_status": gin.H{
				"status": func() string {
					if hasFailure {
						return "Failed"
					}
					return "Passed"
				}(),
			},
			"instances_count": instancesCount,
			"databases_count": databasesCount,
			"instances":       []gin.H{},
		}

		// Add instances for this VM
		for instanceName, instanceChecks := range lastCheckResults.InstanceResults {
			if strings.HasPrefix(instanceName, hostname+"\\") {
				instHasFailure := false
				for _, check := range instanceChecks {
					if check.Status == "FAILED" || check.Status == "ERROR" {
						instHasFailure = true
						break
					}
				}

				// strip VM prefix from instance name
				shortInstance := instanceName
				if idx := strings.Index(instanceName, "\\"); idx != -1 {
					shortInstance = instanceName[idx+1:]
				}

				// count DBs belonging to this VM (demo approximation)
				instDBCount := 0
				for dbName := range lastCheckResults.DatabaseResults {
					if strings.HasPrefix(dbName, hostname+"\\") {
						instDBCount++
					}
				}

				instance := gin.H{
					"entity_name": shortInstance,
					"type":        "SQL Server Instance",
					"overall_fitment_status": gin.H{
						"status": func() string {
							if instHasFailure {
								return "Failed"
							}
							return "Passed"
						}(),
					},
					"databases_count": instDBCount,
					"databases":       []gin.H{},
				}

				// Add databases for this instance
				for dbName, dbChecks := range lastCheckResults.DatabaseResults {
					if strings.HasPrefix(dbName, hostname+"\\") {
						dbHasFailure := false
						for _, check := range dbChecks {
							if check.Status == "FAILED" || check.Status == "ERROR" {
								dbHasFailure = true
								break
							}
						}

						// strip VM prefix from DB name
						shortDB := dbName
						if idx := strings.Index(dbName, "\\"); idx != -1 {
							shortDB = dbName[idx+1:]
						}

						database := gin.H{
							"entity_name": shortDB,
							"type":        "Database",
							"overall_fitment_status": gin.H{
								"status": func() string {
									if dbHasFailure {
										return "Failed"
									}
									return "Passed"
								}(),
							},
						}

						instance["databases"] = append(instance["databases"].([]gin.H), database)
					}
				}

				vm["instances"] = append(vm["instances"].([]gin.H), instance)
			}
		}

		vms = append(vms, vm)
	}

	c.JSON(http.StatusOK, gin.H{"vms": vms})
}

// handleInstancesAPI returns instances data for the new UI
func handleInstancesAPI(c *gin.Context) {
	if lastCheckResults == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No check results available"})
		return
	}

	var instances []gin.H
	for instanceName, instanceChecks := range lastCheckResults.InstanceResults {
		hasFailure := false
		for _, check := range instanceChecks {
			if check.Status == "FAILED" || check.Status == "ERROR" {
				hasFailure = true
				break
			}
		}

		// Split VM and instance names
		vmName := instanceName
		shortInstance := instanceName
		if idx := strings.Index(instanceName, "\\"); idx != -1 {
			vmName = instanceName[:idx]
			shortInstance = instanceName[idx+1:]
		}

		// Build databases list for this instance (approximate by VM prefix)
		var dbs []gin.H
		dbCount := 0
		for dbName, dbChecks := range lastCheckResults.DatabaseResults {
			if strings.HasPrefix(dbName, vmName+"\\") {
				dbCount++

				dbHasFailure := false
				for _, check := range dbChecks {
					if check.Status == "FAILED" || check.Status == "ERROR" {
						dbHasFailure = true
						break
					}
				}

				shortDB := dbName
				if idx := strings.Index(dbName, "\\"); idx != -1 {
					shortDB = dbName[idx+1:]
				}

				dbs = append(dbs, gin.H{
					"entity_name": shortDB,
					"type":        "Database",
					"overall_fitment_status": gin.H{
						"status": func() string {
							if dbHasFailure {
								return "Failed"
							}
							return "Passed"
						}(),
					},
				})
			}
		}

		instance := gin.H{
			"entity_name": shortInstance,
			"type":        "SQL Server Instance",
			"overall_fitment_status": gin.H{
				"status": func() string {
					if hasFailure {
						return "Failed"
					}
					return "Passed"
				}(),
			},
			"databases_count": dbCount,
			"databases":       dbs,
		}

		instances = append(instances, instance)
	}

	c.JSON(http.StatusOK, gin.H{"instances": instances})
}

// handleDatabasesAPI returns databases data for the new UI
func handleDatabasesAPI(c *gin.Context) {
	if lastCheckResults == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No check results available"})
		return
	}

	var databases []gin.H
	for dbName, dbChecks := range lastCheckResults.DatabaseResults {
		hasFailure := false
		for _, check := range dbChecks {
			if check.Status == "FAILED" || check.Status == "ERROR" {
				hasFailure = true
				break
			}
		}

		// Strip VM prefix to show only database name and derive parent instance name
		shortDB := dbName
		if idx := strings.Index(dbName, "\\"); idx != -1 {
			shortDB = dbName[idx+1:]
		}

		database := gin.H{
			"entity_name": shortDB,
			"type":        "Database",
			"overall_fitment_status": gin.H{
				"status": func() string {
					if hasFailure {
						return "Failed"
					}
					return "Passed"
				}(),
			},
			"parent_instance": "MSSQLSERVER",
		}

		databases = append(databases, database)
	}

	c.JSON(http.StatusOK, gin.H{"databases": databases})
}

// Helper functions
func calculateFailedInstances() int {
	if lastCheckResults == nil {
		return 0
	}

	failed := 0
	for _, instanceChecks := range lastCheckResults.InstanceResults {
		for _, check := range instanceChecks {
			if check.Status == "FAILED" || check.Status == "ERROR" {
				failed++
				break
			}
		}
	}
	return failed
}

func calculateFailedDatabases() int {
	if lastCheckResults == nil {
		return 0
	}

	failed := 0
	for _, dbChecks := range lastCheckResults.DatabaseResults {
		for _, check := range dbChecks {
			if check.Status == "FAILED" || check.Status == "ERROR" {
				failed++
				break
			}
		}
	}
	return failed
}

func countChecksByStatus(category, status string) int {
	if lastCheckResults == nil {
		return 0
	}

	count := 0
	var results map[string][]CheckResult

	switch category {
	case "VM":
		results = lastCheckResults.VMResults
	case "Instance":
		results = lastCheckResults.InstanceResults
	case "Database":
		results = lastCheckResults.DatabaseResults
	default:
		return 0
	}

	for _, checks := range results {
		for _, check := range checks {
			if check.Status == status {
				count++
			}
		}
	}

	return count
}

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("Starting NDB PreCheck Service...")

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

    // Serve React build
    router.Static("/assets", "./webapp/dist/assets")
	router.StaticFile("/vite.svg", "./webapp/dist/vite.svg")
	router.StaticFile("/clipboard.png", "./webapp/dist/clipboard.png")

    // SPA fallback: let React handle client routes at root
    router.NoRoute(func(c *gin.Context) {
        path := c.Request.URL.Path
        if strings.HasPrefix(path, "/api/") {
            c.Status(http.StatusNotFound)
            return
        }
        c.File("./webapp/dist/index.html")
    })

	// APIs
	router.POST("/api/check", handleCheck)
	router.GET("/api/progress", getProgress)

	router.GET("/api/summary", handleSummaryAPI)
	router.GET("/api/dbservers", handleDBServersAPI)
	router.GET("/api/instances", handleInstancesAPI)
	router.GET("/api/databases", handleDatabasesAPI)

	logrus.Infof("Server starting on port %s", port)
	router.Run(port)
}
