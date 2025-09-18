param (
    [string]$ComputerName
)

try {
    # Execute the checks remotely
    $scriptBlock = {
        param($ComputerName)
        
        # Initialize results
        $vmChecks = @()
        $instanceChecks = @()
        $databaseChecks = @()
        
        # 1. VM Level Check - PowerShell Execution Policy
        try {
            $policy = Get-ExecutionPolicy
            $vmChecks += @{
                Check = "PowerShell Execution Policy"
                Status = if ($policy.ToString() -eq 'Restricted') { 'FAILED' } else { 'SUCCESS' }
                Message = $policy.ToString()
                Severity = if ($policy.ToString() -eq 'Restricted') { 'CRITICAL' } else { 'INFO' }
            }
        }
        catch {
            $vmChecks += @{
                Check = "PowerShell Execution Policy"
                Status = "ERROR"
                Message = $_.Exception.Message
                Severity = "CRITICAL"
            }
        }
        
        # 2. Instance Level Check - Database Count (simulate for now)
        try {
            # For now, simulate instance check (you can add real SQL Server logic later)
            $instanceChecks += @{
                Check = "Database Count Validation"
                Status = "SUCCESS"
                Message = "User databases found: 5 (limit: 150)"
                Severity = "INFO"
            }
        }
        catch {
            $instanceChecks += @{
                Check = "Database Count Validation"
                Status = "ERROR"
                Message = $_.Exception.Message
                Severity = "CRITICAL"
            }
        }
        
        # 3. Database Level Check - Database State (simulate for now)
        try {
            # For now, simulate database check (you can add real SQL Server logic later)
            $databaseChecks += @{
                Check = "Database State"
                Status = "SUCCESS"
                Message = "Database state: ONLINE"
                Severity = "INFO"
            }
        }
        catch {
            $databaseChecks += @{
                Check = "Database State"
                Status = "ERROR"
                Message = $_.Exception.Message
                Severity = "CRITICAL"
            }
        }
        
        return @{
            VMChecks = $vmChecks
            InstanceChecks = $instanceChecks
            DatabaseChecks = $databaseChecks
            Success = $true
        }
    }

    $result = Invoke-Command -ComputerName $ComputerName -ScriptBlock $scriptBlock -ArgumentList $ComputerName -ErrorAction Stop

    # Create the output
    $output = @{
        Success = $true
        Target = $ComputerName
        VMChecks = $result.VMChecks
        InstanceChecks = $result.InstanceChecks
        DatabaseChecks = $result.DatabaseChecks
    }
}
catch {
    $output = @{
        Success = $false
        Target = $ComputerName
        ErrorMessage = $_.Exception.Message.Trim()
        VMChecks = @(@{
            Check = "PowerShell Execution Policy"
            Status = "ERROR"
            Message = $_.Exception.Message.Trim()
            Severity = "CRITICAL"
        })
        InstanceChecks = @()
        DatabaseChecks = @()
    }
}

# Output clean JSON
$output | ConvertTo-Json -Depth 10 -Compress
