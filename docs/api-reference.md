# API Reference

Generated on 2025-08-21 16:36:28 UTC

## Packages


### approval

**Import Path:** `pkg/approval`




#### Types


##### Approval

Approval represents a single approval decision




##### ApprovalConfig

ApprovalConfig represents approval workflow configuration




##### ApprovalManager

ApprovalManager manages approval workflows and requests



**Methods:**


- `ApproveRequest`: ApproveRequest approves an approval request


- `CreateWorkflow`: CreateWorkflow creates a new approval workflow


- `Disable`: Disable disables approval workflows


- `Enable`: Enable enables approval workflows


- `GetRequest`: GetRequest retrieves an approval request by ID


- `GetWorkflows`: GetWorkflows returns all workflows


- `IsEnabled`: IsEnabled returns whether approval workflows are enabled


- `ListPendingRequests`: ListPendingRequests returns all pending approval requests


- `ListRequests`: ListRequests returns all approval requests


- `RejectRequest`: RejectRequest rejects an approval request


- `RequiresApproval`: RequiresApproval checks if an action requires approval


- `SubmitRequest`: SubmitRequest submits an approval request


- `countStageApprovals`: countStageApprovals counts approvals for a specific stage


- `evaluateCondition`: evaluateCondition evaluates a workflow condition


- `matchesWorkflow`: matchesWorkflow checks if a module and action match a workflow


- `processApproval`: processApproval processes an approval or rejection




##### ApprovalRequest

ApprovalRequest represents an approval request



**Methods:**


- `IsExpired`: IsExpired returns whether the approval request has expired




##### Condition

Condition represents a workflow condition




##### Decision

Decision represents an approval decision




##### Stage

Stage represents a workflow stage




##### Status

Status represents the status of an approval request




##### Workflow

Workflow represents an approval workflow







#### Functions


##### TestApprovalManager_ApproveRequest




##### TestApprovalManager_CreateWorkflow




##### TestApprovalManager_EnableDisable




##### TestApprovalManager_ListRequests




##### TestApprovalManager_MultiStageWorkflow




##### TestApprovalManager_New




##### TestApprovalManager_RejectRequest




##### TestApprovalManager_RequiresApproval




##### TestApprovalManager_SubmitRequest




##### TestApprovalRequest_IsExpired




##### contains

Helper function to check if string contains substring



##### containsHelper




##### generateRequestID

generateRequestID generates a unique request ID









---


### audit

**Import Path:** `pkg/audit`




#### Types


##### AuditConfig

AuditConfig represents audit logging configuration




##### AuditEntry

AuditEntry represents a single audit log entry



**Methods:**


- `Validate`: Validate validates the audit entry




##### AuditLogger

AuditLogger manages audit logging



**Methods:**


- `Close`: Close closes the audit logger


- `Disable`: Disable disables audit logging


- `Enable`: Enable enables audit logging


- `IsEnabled`: IsEnabled returns whether audit logging is enabled


- `LogAuthentication`: LogAuthentication logs an authentication event


- `LogPolicyViolation`: LogPolicyViolation logs a policy violation event


- `LogResourceChange`: LogResourceChange logs a resource change event


- `LogSystemEvent`: LogSystemEvent logs a system event


- `LogUserAction`: LogUserAction logs a user action event


- `SetMaxFileSize`: SetMaxFileSize sets the maximum file size before rotation


- `SetMaxFiles`: SetMaxFiles sets the maximum number of rotated files to keep


- `checkRotation`: checkRotation checks if log rotation is needed and performs it


- `ensureFileOpen`: ensureFileOpen ensures the audit log file is open


- `rotateFiles`: rotateFiles rotates the audit log files


- `writeEntry`: writeEntry writes an audit entry to the log file




##### EventType

EventType represents the type of audit event




##### PolicyViolation

PolicyViolation represents a policy violation in audit logs







#### Functions


##### TestAuditEntry_Validate




##### TestAuditLogger_EnableDisable




##### TestAuditLogger_LogPolicyViolation




##### TestAuditLogger_LogResourceChange




##### TestAuditLogger_LogSystemEvent




##### TestAuditLogger_LogUserAction




##### TestAuditLogger_New




##### TestAuditLogger_Rotation




##### getRemoteAddrFromContext




##### getSessionIDFromContext




##### getUserFromContext

Context helper functions









---


### cli

**Import Path:** `pkg/cli`




#### Types


##### ConnectionConfig





##### InventoryConfig





##### Metadata





##### ModuleConfig





##### ModuleSpec





##### ProjectConfig

Configuration structures




##### ProjectSpec





##### ResourceConfig





##### TargetGroup








#### Functions


##### Execute

Execute adds all child commands to the root command and sets flags appropriately.
This is called by main.main(). It only needs to happen once to the rootCmd.



##### countActionResults




##### displayChangeDiff




##### getChangeSymbol




##### init




##### initConfig

initConfig reads in config file and ENV variables if set.



##### runApply




##### runInit




##### runPlan




##### savePlanToFile




##### writeYAMLFile









#### Variables


- `applyModuleFile`: 

- `applyInventoryFile`: 

- `applyDryRun`: 

- `applyAutoApprove`: 

- `planModuleFile`: 

- `planInventoryFile`: 

- `planOutputFile`: 

- `cfgFile`: 

- `verbose`: 

- `applyCmd`: applyCmd represents the apply command


- `initCmd`: initCmd represents the init command


- `planCmd`: planCmd represents the plan command


- `rootCmd`: rootCmd represents the base command when called without any subcommands




---


### compliance

**Import Path:** `pkg/compliance`




#### Types


##### CISUbuntu2004Module

CISUbuntu2004Module implements CIS Ubuntu 20.04 compliance checks



**Methods:**


- `CheckCompliance`: CheckCompliance checks CIS Ubuntu 20.04 compliance


- `Framework`: Framework returns the compliance framework name


- `Version`: Version returns the compliance framework version


- `checkFileResource`: checkFileResource checks file resources against CIS controls


- `checkResource`: checkResource checks a single resource against CIS controls


- `checkServiceResource`: checkServiceResource checks service resources against CIS controls


- `checkUserResource`: checkUserResource checks user resources against CIS controls




##### ComplianceConfig

ComplianceConfig represents compliance configuration




##### ComplianceManager

ComplianceManager manages compliance modules and checks



**Methods:**


- `CheckAllCompliance`: CheckAllCompliance checks compliance against all loaded modules


- `CheckCompliance`: CheckCompliance checks compliance against the first loaded module


- `Disable`: Disable disables compliance checking


- `Enable`: Enable enables compliance checking


- `GetLoadedModules`: GetLoadedModules returns the names of loaded compliance modules


- `IsEnabled`: IsEnabled returns whether compliance checking is enabled


- `LoadModule`: LoadModule loads a compliance module by name




##### ComplianceModule

ComplianceModule defines a compliance framework module




##### ComplianceResult

ComplianceResult represents the result of a compliance check




##### ComplianceViolation

ComplianceViolation represents a compliance violation



**Methods:**


- `String`: String returns a string representation of the violation




##### NIST80053Module

NIST80053Module implements NIST 800-53 compliance checks



**Methods:**


- `CheckCompliance`: CheckCompliance checks NIST 800-53 compliance


- `Framework`: Framework returns the compliance framework name


- `Version`: Version returns the compliance framework version


- `checkFileResource`: checkFileResource checks file resources against NIST controls


- `checkResource`: checkResource checks a single resource against NIST controls


- `checkServiceResource`: checkServiceResource checks service resources against NIST controls




##### STIGRHELModule

STIGRHELModule implements STIG RHEL 8 compliance checks



**Methods:**


- `CheckCompliance`: CheckCompliance checks STIG RHEL 8 compliance


- `Framework`: Framework returns the compliance framework name


- `Version`: Version returns the compliance framework version


- `checkResource`: checkResource checks a single resource against STIG controls




##### Severity

Severity represents the severity level of a compliance violation







#### Functions


##### TestComplianceManager_CheckCISCompliance




##### TestComplianceManager_CheckCISViolations




##### TestComplianceManager_CheckMultipleFrameworks




##### TestComplianceManager_CheckNISTCompliance




##### TestComplianceManager_EnableDisable




##### TestComplianceManager_LoadCISModule




##### TestComplianceManager_LoadNISTModule




##### TestComplianceManager_LoadSTIGModule




##### TestComplianceManager_New




##### TestComplianceViolation_String




##### contains

Helper function to check if string contains substring



##### containsHelper










---


### core

**Import Path:** `pkg/core`




#### Types


##### Action

Action represents the type of action to be performed



**Methods:**


- `String`: String returns the string representation of an action




##### Change

Change represents a planned change to a resource




##### ChangeResult

ChangeResult represents the result of executing a single change




##### ExecuteOptions

ExecuteWithOptions executes a plan with additional options




##### ExecutionResult

ExecutionResult represents the result of executing a plan



**Methods:**


- `AddChangeResult`: AddChangeResult adds a change result and updates the summary


- `Finalize`: Finalize finalizes the execution result by setting end time and duration




##### ExecutionSummary

ExecutionSummary provides a summary of execution results




##### Executor

Executor executes plans by applying changes



**Methods:**


- `ExecutePlan`: ExecutePlan executes all changes in a plan


- `ExecutePlanWithOptions`: ExecutePlanWithOptions executes a plan with the given options


- `executeChange`: executeChange executes a single change




##### Module

Module represents a Chisel configuration module



**Methods:**


- `SaveToFile`: SaveModuleToFile saves a module to a YAML file


- `Validate`: Validate validates the module configuration




##### ModuleMetadata

ModuleMetadata contains metadata about the module




##### ModuleSpec

ModuleSpec contains the module specification




##### Plan

Plan represents a collection of planned changes



**Methods:**


- `AddChange`: AddChange adds a change to the plan


- `HasChanges`: HasChanges returns true if the plan has any changes that need to be applied


- `Summary`: Summary returns a summary of the plan




##### PlanSummary

PlanSummary provides a summary of planned changes




##### Planner

Planner creates execution plans for modules



**Methods:**


- `CreatePlan`: CreatePlan creates an execution plan for the given module


- `determineAction`: determineAction determines what action should be taken based on the resource and diff


- `planResource`: planResource creates a plan for a single resource







#### Functions


##### TestAction_String




##### TestExecutionResult_Summary




##### TestExecutor_ExecuteChange




##### TestExecutor_ExecutePlan




##### TestLoadModuleFromFile




##### TestModule_Validate




##### TestPlan_AddChange




##### TestPlan_Summary




##### TestPlanner_CreatePlan










---


### drift

**Import Path:** `pkg/drift`




#### Types


##### DriftConfig

DriftConfig represents configuration for drift detection



**Methods:**


- `Validate`: Validate validates the drift configuration




##### DriftDetector

DriftDetector performs drift detection on resources



**Methods:**


- `CheckDrift`: CheckDrift performs a one-time drift check on a module


- `GetLastReport`: GetLastReport returns the last drift detection report


- `GetReportChannel`: GetReportChannel returns the channel for receiving drift reports


- `IsRunning`: IsRunning returns whether the drift detector is currently running


- `SetConcurrency`: SetConcurrency updates the maximum concurrency for drift checks


- `SetInterval`: SetInterval updates the drift detection interval


- `SetTimeout`: SetTimeout updates the timeout for individual drift checks


- `Start`: Start starts the drift detection scheduler


- `Stop`: Stop stops the drift detection scheduler


- `checkResourceDrift`: checkResourceDrift checks a single resource for drift


- `schedulerLoop`: schedulerLoop runs the drift detection scheduler




##### DriftNotifier

DriftNotifier handles notifications for drift detection



**Methods:**


- `AddChannel`: AddChannel adds a notification channel


- `Disable`: Disable disables notifications


- `Enable`: Enable enables notifications


- `Notify`: Notify sends notifications if drift is detected above threshold




##### DriftReport

DriftReport represents a complete drift detection report




##### DriftResult

DriftResult represents the result of a drift detection check




##### DriftScheduleConfig

DriftScheduleConfig defines the schedule configuration for drift detection



**Methods:**


- `SetDefaults`: SetDefaults sets default values for the configuration


- `Validate`: Validate validates the drift schedule configuration




##### DriftScheduler

DriftScheduler manages scheduled drift detection for multiple modules



**Methods:**


- `AddModule`: AddModule adds a module to the drift detection schedule


- `Disable`: Disable disables the scheduler


- `Enable`: Enable enables the scheduler


- `GetRecentReports`: GetRecentReports returns the most recent drift reports


- `GetScheduledModules`: GetScheduledModules returns a copy of all scheduled modules


- `IsEnabled`: IsEnabled returns whether the scheduler is enabled


- `IsRunning`: IsRunning returns whether the scheduler is currently running


- `RemoveModule`: RemoveModule removes a module from the drift detection schedule


- `SetCheckInterval`: SetCheckInterval sets the interval for checking scheduled modules (useful for testing)


- `Start`: Start starts the drift detection scheduler


- `Stop`: Stop stops the drift detection scheduler


- `addReport`: addReport adds a report to the recent reports list


- `checkAndRunDriftDetection`: checkAndRunDriftDetection checks if any modules need drift detection and runs them


- `runDriftDetectionForModule`: runDriftDetectionForModule runs drift detection for a specific module


- `schedulerLoop`: schedulerLoop runs the main scheduler loop




##### LogNotificationChannel

LogNotificationChannel logs drift reports to stdout



**Methods:**


- `Send`: Send sends a notification to the log


- `Type`: Type returns the notification channel type




##### MockDriftProvider

MockDriftProvider for testing drift detection



**Methods:**


- `Apply`: 

- `Diff`: 

- `Read`: 

- `Type`: 

- `Validate`: 



##### MockProvider

MockProvider for testing drift detection



**Methods:**


- `Apply`: 

- `Diff`: 

- `Read`: 

- `Type`: 

- `Validate`: 



##### NotificationChannel

NotificationChannel represents a notification channel




##### ScheduledModule

ScheduledModule represents a module scheduled for drift detection







#### Functions


##### TestDriftConfig_Validate




##### TestDriftDetector_CheckDrift




##### TestDriftDetector_Configuration




##### TestDriftDetector_StartStop




##### TestDriftNotifier_Notify




##### TestDriftScheduleConfig_Validate




##### TestDriftScheduler_AddModule




##### TestDriftScheduler_AddModuleErrors




##### TestDriftScheduler_EnableDisable




##### TestDriftScheduler_New




##### TestDriftScheduler_RemoveModule




##### TestDriftScheduler_ScheduleExecution




##### TestDriftScheduler_StartStop




##### TestLogNotificationChannel










---


### events

**Import Path:** `pkg/events`




#### Types


##### Event

Event represents a system event




##### EventBus

EventBus manages event publishing and subscription



**Methods:**


- `Close`: Close shuts down the event bus


- `Publish`: Publish publishes an event to the bus


- `PublishSync`: PublishSync publishes an event synchronously


- `Subscribe`: Subscribe subscribes a handler to specific event types


- `Unsubscribe`: Unsubscribe removes a handler from all event types


- `worker`: worker processes events from the buffer




##### EventEmitter

EventEmitter helps emit events during operations



**Methods:**


- `EmitApplyCompleted`: EmitApplyCompleted emits an apply completed event


- `EmitApplyStarted`: EmitApplyStarted emits an apply started event


- `EmitDriftDetected`: EmitDriftDetected emits a drift detected event


- `EmitPlanCompleted`: EmitPlanCompleted emits a plan completed event


- `EmitPlanStarted`: EmitPlanStarted emits a plan started event


- `EmitResourceCompleted`: EmitResourceCompleted emits a resource completed event


- `EmitResourceFailed`: EmitResourceFailed emits a resource failed event


- `EmitResourceStarted`: EmitResourceStarted emits a resource started event




##### EventHandler

EventHandler handles events




##### EventType

EventType represents the type of event




##### FileEventHandler

FileEventHandler writes events to a file



**Methods:**


- `Handle`: Handle handles an event by writing it to a file


- `Name`: Name returns the handler name


- `Types`: Types returns the event types this handler subscribes to




##### LogEventHandler

LogEventHandler logs events to stdout



**Methods:**


- `Handle`: Handle handles an event by logging it


- `Name`: Name returns the handler name


- `Types`: Types returns the event types this handler subscribes to




##### MetricsEventHandler

MetricsEventHandler collects metrics from events



**Methods:**


- `GetMetrics`: GetMetrics returns current metrics


- `Handle`: Handle handles an event by updating metrics


- `Name`: Name returns the handler name


- `Types`: Types returns the event types this handler subscribes to







#### Functions


##### TestEventBus_BufferFull




##### TestEventBus_Close




##### TestEventBus_ConcurrentPublish




##### TestEventBus_PublishSync




##### TestEventBus_SubscribeAndPublish




##### TestEventBus_Unsubscribe




##### TestEventEmitter




##### TestFileEventHandler




##### TestLogEventHandler




##### TestMetricsEventHandler




##### TestNewEvent




##### generateEventID

generateEventID generates a unique event ID









---


### executor

**Import Path:** `pkg/executor`




#### Types


##### DependencyGraph

DependencyGraph represents a directed acyclic graph for dependency resolution



**Methods:**


- `AddEdge`: AddEdge adds a directed edge from 'from' to 'to'


- `AddNode`: AddNode adds a node to the graph


- `TopologicalSort`: TopologicalSort performs a topological sort and returns batches of resources
that can be executed in parallel




##### EnhancedParallelExecutor

EnhancedParallelExecutor extends ParallelExecutor with rollback capabilities



**Methods:**


- `ExecuteWithRollback`: ExecuteWithRollback executes a plan with automatic rollback on failure




##### ErrorRecovery

ErrorRecovery handles different types of errors and suggests recovery actions



**Methods:**


- `RecoverFromError`: RecoverFromError attempts to recover from an error using registered strategies


- `classifyError`: classifyError classifies an error to determine the appropriate recovery strategy


- `registerDefaultStrategies`: registerDefaultStrategies registers built-in recovery strategies




##### ExecutionBatch

ExecutionBatch represents a batch of resources that can be executed in parallel




##### ExecutionPlan

ExecutionPlan represents a plan for parallel execution




##### ExecutionResult

ExecutionResult represents the result of executing a resource




##### ParallelExecutor

ParallelExecutor executes resources in parallel while respecting dependencies



**Methods:**


- `CreateExecutionPlan`: CreateExecutionPlan creates an execution plan that respects dependencies


- `Execute`: Execute executes the plan in parallel


- `executeBatch`: executeBatch executes a single batch of resources in parallel


- `getResourceType`: getResourceType extracts the resource type from a ResourceID


- `hasDependency`: hasDependency determines if one resource depends on another
This is a simple heuristic - in a real implementation, this would parse
explicit dependencies from the resource configuration




##### RecoveryStrategy

RecoveryStrategy defines how to recover from specific error types




##### RollbackAction

RollbackAction represents an action that can be rolled back




##### RollbackExecutor

RollbackExecutor handles rollback operations



**Methods:**


- `CreateRollbackPlan`: CreateRollbackPlan creates a rollback plan from execution results


- `ExecuteRollback`: ExecuteRollback executes a rollback plan


- `executeWithRetry`: executeWithRetry executes a rollback action with retry logic




##### RollbackPlan

RollbackPlan represents a plan for rolling back changes







#### Functions


##### TestDependencyGraph_TopologicalSort




##### TestEnhancedParallelExecutor_ExecuteWithRollback




##### TestErrorRecovery_ClassifyError




##### TestErrorRecovery_RecoverFromError




##### TestParallelExecutor_CreateExecutionPlan




##### TestParallelExecutor_Execute




##### TestParallelExecutor_getResourceType




##### TestParallelExecutor_hasDependency




##### TestRollbackExecutor_CreateRollbackPlan




##### TestRollbackExecutor_ExecuteRollback




##### TestRollbackExecutor_executeWithRetry




##### contains

contains checks if any of the substrings are present in the main string









---


### integration

**Import Path:** `pkg/integration`




#### Types


##### MockDynamicInventory

MockDynamicInventory for testing



**Methods:**


- `Discover`: 

- `Type`: 

- `Validate`: 






#### Functions


##### TestFullSystemIntegration

TestFullSystemIntegration tests the complete system working together



##### getResourceType

getResourceType extracts resource type from ResourceID









---


### inventory

**Import Path:** `pkg/inventory`




#### Types


##### AWSInventoryProvider

AWSInventoryProvider discovers targets from AWS EC2 instances



**Methods:**


- `Discover`: Discover discovers EC2 instances based on the selector


- `SetMockMode`: SetMockMode enables mock mode for testing


- `Type`: Type returns the provider type


- `Validate`: Validate validates the provider configuration


- `discoverMock`: discoverMock provides mock data for testing


- `discoverReal`: discoverReal discovers real EC2 instances (would use AWS SDK)


- `instanceMatchesSelector`: instanceMatchesSelector checks if an instance matches the selector




##### AzureConfig

AzureConfig represents Azure inventory configuration




##### AzureInventoryProvider

AzureInventoryProvider discovers targets from Azure VMs



**Methods:**


- `Discover`: Discover discovers Azure VMs based on the selector


- `SetMockMode`: SetMockMode enables mock mode for testing


- `Type`: Type returns the provider type


- `Validate`: Validate validates the provider configuration


- `discoverMock`: discoverMock provides mock data for testing


- `discoverReal`: discoverReal discovers real Azure VMs (would use Azure SDK)


- `vmMatchesSelector`: vmMatchesSelector checks if a VM matches the selector




##### DynamicInventory

DynamicInventory represents a dynamic inventory source




##### Inventory

Inventory represents a Chisel inventory configuration



**Methods:**


- `SaveToFile`: SaveToFile saves an inventory to a YAML file


- `Validate`: Validate validates the inventory configuration




##### InventoryRegistry

InventoryRegistry manages dynamic inventory providers



**Methods:**


- `Discover`: Discover discovers targets using the specified provider and selector


- `ListProviders`: ListProviders returns a list of registered provider types


- `Register`: Register registers a dynamic inventory provider




##### MockDynamicInventory

MockDynamicInventory for testing



**Methods:**


- `Discover`: 

- `Type`: 

- `Validate`: 



##### TargetGroup

TargetGroup represents a group of target hosts



**Methods:**


- `GetHosts`: GetHosts returns the list of hosts for this target group


- `Validate`: Validate validates a target group




##### mockAzureVM

mockAzureVM represents a mock Azure VM for testing




##### mockEC2Instance

mockEC2Instance represents a mock EC2 instance for testing







#### Functions


##### MatchesSelector

MatchesSelector checks if a target matches the given selector



##### ParseSelector

ParseSelector parses a selector string into key-value pairs
Format: "key1=value1,key2=value2"



##### TestAWSInventoryProvider_Discover




##### TestAWSInventoryProvider_DiscoverWithInvalidSelector




##### TestAWSInventoryProvider_DiscoverWithoutMockMode




##### TestAWSInventoryProvider_Type




##### TestAWSInventoryProvider_Validate




##### TestAzureInventoryProvider_Discover




##### TestAzureInventoryProvider_DiscoverWithInvalidSelector




##### TestAzureInventoryProvider_DiscoverWithoutMockMode




##### TestAzureInventoryProvider_Type




##### TestAzureInventoryProvider_Validate




##### TestInventoryRegistry_Discover




##### TestInventoryRegistry_Register




##### TestInventory_Validate




##### TestLoadInventoryFromFile




##### TestMatchesSelector




##### TestParseAWSSelector




##### TestParseAzureSelector




##### TestParseSelector




##### TestTargetGroup_GetHosts




##### parseAWSSelector

parseAWSSelector parses AWS tag selector format



##### parseAzureSelector

parseAzureSelector parses Azure tag selector format









---


### monitoring

**Import Path:** `pkg/monitoring`




#### Types


##### AggregatedMetrics

AggregatedMetrics represents aggregated system metrics




##### ExecutionMetrics

ExecutionMetrics represents metrics for a module execution



**Methods:**


- `Validate`: Validate validates the execution metrics




##### MetricsCollector

MetricsCollector collects and aggregates system metrics



**Methods:**


- `Disable`: Disable disables metrics collection


- `Enable`: Enable enables metrics collection


- `ExportToFile`: ExportToFile exports metrics to a JSON file


- `GetMetrics`: GetMetrics returns aggregated metrics


- `GetMetricsByTimeRange`: GetMetricsByTimeRange returns metrics for a specific time range


- `GetPrometheusMetrics`: GetPrometheusMetrics returns metrics in Prometheus format


- `IsEnabled`: IsEnabled returns whether metrics collection is enabled


- `RecordExecution`: RecordExecution records execution metrics


- `RecordPolicyViolation`: RecordPolicyViolation records policy violation metrics


- `RecordResource`: RecordResource records resource metrics




##### MonitoringConfig

MonitoringConfig represents monitoring configuration




##### PolicyViolationMetrics

PolicyViolationMetrics represents metrics for policy violations




##### ResourceMetrics

ResourceMetrics represents metrics for a resource operation







#### Functions


##### TestExecutionMetrics_Validate




##### TestMetricsCollector_EnableDisable




##### TestMetricsCollector_ExportToFile




##### TestMetricsCollector_GetMetricsByTimeRange




##### TestMetricsCollector_GetPrometheusMetrics




##### TestMetricsCollector_New




##### TestMetricsCollector_RecordExecution




##### TestMetricsCollector_RecordFailedExecution




##### TestMetricsCollector_RecordPolicyViolation




##### TestMetricsCollector_RecordResourceMetrics




##### contains

Helper functions



##### containsHelper




##### fileExists




##### removeFile










---


### notifications

**Import Path:** `pkg/notifications`




#### Types


##### ConsoleChannel

ConsoleChannel writes notifications to stdout/stderr



**Methods:**


- `Name`: Name returns the channel name


- `Send`: Send writes a notification to the console


- `Type`: Type returns the channel type




##### EmailChannel

EmailChannel sends notifications via email



**Methods:**


- `Name`: Name returns the channel name


- `Send`: Send sends a notification via email


- `Type`: Type returns the channel type




##### FileChannel

FileChannel writes notifications to a file



**Methods:**


- `Name`: Name returns the channel name


- `Send`: Send writes a notification to the file


- `Type`: Type returns the channel type




##### Notification

Notification represents a notification message



**Methods:**


- `AddData`: AddData adds data to the notification


- `AddTag`: AddTag adds a tag to the notification




##### NotificationChannel

NotificationChannel represents a notification delivery channel




##### NotificationEventHandler

NotificationEventHandler handles events and generates notifications



**Methods:**


- `Handle`: Handle handles an event by generating appropriate notifications


- `Name`: Name returns the handler name


- `Types`: Types returns the event types this handler subscribes to


- `eventToNotification`: eventToNotification converts an event to a notification




##### NotificationLevel

NotificationLevel represents the severity level




##### NotificationManager

NotificationManager manages notification channels and routing



**Methods:**


- `AddChannel`: AddChannel adds a notification channel


- `AddRule`: AddRule adds a notification rule


- `Disable`: Disable disables the notification manager


- `Enable`: Enable enables the notification manager


- `GetChannels`: GetChannels returns a list of registered channel names


- `GetRules`: GetRules returns a copy of all notification rules


- `IsEnabled`: IsEnabled returns whether the notification manager is enabled


- `RemoveChannel`: RemoveChannel removes a notification channel


- `Send`: Send sends a notification through appropriate channels


- `SendToChannels`: SendToChannels sends a notification to specific channels


- `findMatchingRules`: findMatchingRules finds rules that match the notification


- `ruleMatches`: ruleMatches checks if a rule matches a notification




##### NotificationRule

NotificationRule defines when and how to send notifications




##### NotificationTemplate

NotificationTemplate defines how to format notifications




##### RateLimitConfig

RateLimitConfig defines rate limiting for notifications




##### RateLimiter

RateLimiter implements token bucket rate limiting



**Methods:**


- `Allow`: Allow checks if an operation is allowed under the rate limit




##### SlackChannel

SlackChannel sends notifications to Slack via webhook



**Methods:**


- `Name`: Name returns the channel name


- `Send`: Send sends a notification to Slack


- `Type`: Type returns the channel type




##### TestNotificationChannel

TestNotificationChannel for testing



**Methods:**


- `GetSentNotifications`: 

- `Name`: 

- `Send`: 

- `SetShouldFail`: 

- `Type`: 



##### WebhookChannel

WebhookChannel sends notifications to a webhook URL



**Methods:**


- `Name`: Name returns the channel name


- `Send`: Send sends a notification to the webhook


- `Type`: Type returns the channel type







#### Functions


##### TestConsoleChannel_Send




##### TestEmailChannel_Properties




##### TestFileChannel_AppendMode




##### TestFileChannel_DirectoryCreation




##### TestFileChannel_Send




##### TestNewNotification




##### TestNotificationEventHandler




##### TestNotificationManager_AddChannel




##### TestNotificationManager_AddChannelErrors




##### TestNotificationManager_AddRule




##### TestNotificationManager_AddRuleErrors




##### TestNotificationManager_ChannelFailure




##### TestNotificationManager_DisabledRule




##### TestNotificationManager_EnableDisable




##### TestNotificationManager_RateLimit




##### TestNotificationManager_RemoveChannel




##### TestNotificationManager_Send




##### TestNotificationManager_SendNoMatchingRules




##### TestNotificationManager_SendToChannels




##### TestNotificationManager_SendWhenDisabled




##### TestNotificationManager_SendWithConditions




##### TestNotification_AddData




##### TestNotification_AddTag




##### TestRateLimiter




##### TestSlackChannel_Send




##### TestWebhookChannel_Send




##### TestWebhookChannel_SendError




##### TestWebhookChannel_Timeout




##### generateNotificationID

generateNotificationID generates a unique notification ID



##### min

min returns the minimum of two integers









---


### policy

**Import Path:** `pkg/policy`




#### Types


##### PolicyConfig

PolicyConfig represents policy engine configuration




##### PolicyEngine

PolicyEngine manages and evaluates policies



**Methods:**


- `Disable`: Disable disables the policy engine


- `Enable`: Enable enables the policy engine


- `EvaluateModule`: EvaluateModule evaluates all resources in a module against loaded policies


- `EvaluateResource`: EvaluateResource evaluates a single resource against all loaded policies


- `GetLoadedPolicies`: GetLoadedPolicies returns the names of all loaded policies


- `IsEnabled`: IsEnabled returns whether the policy engine is enabled


- `LoadFromConfig`: LoadFromConfig loads policies from configuration


- `LoadPolicy`: LoadPolicy loads a policy from content


- `LoadPolicyFromFile`: LoadPolicyFromFile loads a policy from a file


- `RemovePolicy`: RemovePolicy removes a policy


- `evaluateModuleAgainstPolicy`: evaluateModuleAgainstPolicy evaluates a module against a specific policy


- `evaluateResourceAgainstPolicy`: evaluateResourceAgainstPolicy evaluates a resource against a specific policy




##### PolicyInput

PolicyInput represents input data for policy evaluation



**Methods:**


- `ToJSON`: ToJSON converts policy input to JSON for OPA evaluation




##### PolicyResult

PolicyResult represents the result of a policy evaluation




##### PolicyViolation

PolicyViolation represents a policy violation



**Methods:**


- `String`: String returns a string representation of the policy violation







#### Functions


##### TestPolicyEngine_EnableDisable




##### TestPolicyEngine_EvaluateModule




##### TestPolicyEngine_EvaluateResource




##### TestPolicyEngine_EvaluateWhenDisabled




##### TestPolicyEngine_LoadPolicy




##### TestPolicyEngine_LoadPolicyFromFile




##### TestPolicyEngine_New




##### TestPolicyEngine_RemovePolicy




##### TestPolicyViolation_String




##### contains

Helper function to check if string contains substring



##### containsHelper




##### extractPolicyName

extractPolicyName extracts policy name from file path









---


### providers

**Import Path:** `pkg/providers`




#### Types


##### FileProvider

FileProvider manages file resources



**Methods:**


- `Apply`: Apply applies the changes to bring the resource to desired state


- `Diff`: Diff compares desired vs current state and returns the differences


- `Read`: Read reads the current state of the file resource


- `Type`: Type returns the resource type this provider handles


- `Validate`: Validate validates the file resource configuration


- `createFile`: createFile creates a new file


- `deleteFile`: deleteFile removes a file


- `resolveContent`: resolveContent resolves the content for a file, handling templates if specified


- `setFileAttributes`: setFileAttributes sets file mode, owner, and group


- `updateFile`: updateFile updates an existing file


- `writeFileContent`: writeFileContent writes content to a file




##### KubernetesConfig

KubernetesConfig represents Kubernetes provider configuration




##### KubernetesProvider

KubernetesProvider manages Kubernetes resources



**Methods:**


- `Apply`: Apply applies changes to a Kubernetes resource


- `Diff`: Diff compares desired and current state


- `Read`: Read reads the current state of a Kubernetes resource


- `SetMockMode`: SetMockMode enables mock mode for testing


- `Type`: Type returns the provider type


- `Validate`: Validate validates a Kubernetes resource


- `applyMock`: applyMock applies changes in mock mode


- `applyReal`: applyReal applies changes to actual Kubernetes cluster


- `compareProperties`: compareProperties compares two property maps and returns changes


- `readMock`: readMock provides mock data for testing


- `readReal`: readReal reads from actual Kubernetes cluster


- `validateConfigMap`: validateConfigMap validates a ConfigMap resource


- `validateDeployment`: validateDeployment validates a Deployment resource


- `validateSecret`: validateSecret validates a Secret resource


- `validateService`: validateService validates a Service resource




##### MockSSHConnection

MockSSHConnection is a mock implementation of SSH connection for testing



**Methods:**


- `Close`: 

- `Connect`: 

- `Execute`: 



##### PkgProvider

PkgProvider manages package resources



**Methods:**


- `Apply`: Apply applies the changes to bring the package to desired state


- `Diff`: Diff compares desired vs current state and returns the differences


- `Read`: Read reads the current state of the package


- `Type`: Type returns the resource type this provider handles


- `Validate`: Validate validates the package resource configuration


- `installPackage`: installPackage installs a package


- `isPackageInstalled`: isPackageInstalled checks if a package is installed and returns its version


- `removePackage`: removePackage removes a package


- `updatePackage`: updatePackage updates a package to the latest version




##### ServiceProvider

ServiceProvider manages service resources



**Methods:**


- `Apply`: Apply applies the changes to bring the service to desired state


- `Diff`: Diff compares desired vs current state and returns the differences


- `Read`: Read reads the current state of the service


- `Type`: Type returns the resource type this provider handles


- `Validate`: Validate validates the service resource configuration


- `disableService`: disableService disables a service from starting at boot


- `enableService`: enableService enables a service to start at boot


- `isServiceActive`: isServiceActive checks if a service is currently active/running


- `isServiceEnabled`: isServiceEnabled checks if a service is enabled to start at boot


- `startService`: startService starts a service


- `stopService`: stopService stops a service


- `updateService`: updateService updates the service state and enabled status




##### ShellProvider

ShellProvider manages shell command execution resources



**Methods:**


- `Apply`: Apply applies the changes by executing the shell command


- `Diff`: Diff compares desired vs current state and returns the differences


- `Read`: Read reads the current state to determine if the command should run


- `Type`: Type returns the resource type this provider handles


- `Validate`: Validate validates the shell resource configuration


- `buildCommand`: buildCommand builds the full command with user, cwd, and other context


- `commandSucceeds`: commandSucceeds checks if a command executes successfully


- `executeCommand`: executeCommand executes the shell command with proper context


- `fileExists`: fileExists checks if a file or directory exists




##### UserProvider

UserProvider manages user resources



**Methods:**


- `Apply`: Apply applies the changes to bring the user to desired state


- `Diff`: Diff compares desired vs current state and returns the differences


- `Read`: Read reads the current state of the user


- `Type`: Type returns the resource type this provider handles


- `Validate`: Validate validates the user resource configuration


- `createUser`: createUser creates a new user


- `deleteUser`: deleteUser removes a user


- `getUserInfo`: getUserInfo gets detailed information about a user


- `setUserGroups`: setUserGroups sets the groups for a user


- `updateUser`: updateUser updates an existing user


- `userExists`: userExists checks if a user exists







#### Functions


##### TestFileProvider_Apply_WithTemplate




##### TestFileProvider_Diff_CreateFile




##### TestFileProvider_Diff_DeleteFile




##### TestFileProvider_Diff_NoChanges




##### TestFileProvider_Diff_UpdateContent




##### TestFileProvider_Read_FileExists




##### TestFileProvider_Read_FileNotExists




##### TestFileProvider_Type




##### TestFileProvider_Validate




##### TestKubernetesProvider_ApplyDeployment




##### TestKubernetesProvider_DiffDeployment




##### TestKubernetesProvider_ReadDeployment




##### TestKubernetesProvider_SupportedKinds




##### TestKubernetesProvider_Type




##### TestKubernetesProvider_UnsupportedKind




##### TestKubernetesProvider_ValidateConfigMap




##### TestKubernetesProvider_ValidateDeployment




##### TestKubernetesProvider_ValidateSecret




##### TestKubernetesProvider_ValidateService




##### TestKubernetesProvider_WithoutMockMode




##### TestPkgProvider_Apply




##### TestPkgProvider_Diff




##### TestPkgProvider_Read




##### TestPkgProvider_Type




##### TestPkgProvider_Validate




##### TestServiceProvider_Apply




##### TestServiceProvider_Diff




##### TestServiceProvider_Read




##### TestServiceProvider_Type




##### TestServiceProvider_Validate




##### TestShellEscape




##### TestShellProvider_Apply




##### TestShellProvider_Diff




##### TestShellProvider_Read




##### TestShellProvider_Type




##### TestShellProvider_Validate




##### TestUserProvider_Apply




##### TestUserProvider_Diff




##### TestUserProvider_Read




##### TestUserProvider_Type




##### TestUserProvider_Validate




##### shellEscape

shellEscape escapes a string for safe use in shell commands








#### Variables


- `_`: Ensure MockSSHConnection implements ssh.Executor




---


### rbac

**Import Path:** `pkg/rbac`




#### Types


##### Permission

Permission represents a specific permission



**Methods:**


- `String`: String returns the string representation of the permission




##### RBACConfig

RBACConfig represents RBAC configuration




##### RBACManager

RBACManager manages roles, users, and permissions



**Methods:**


- `AssignRole`: AssignRole assigns a role to a user


- `CheckPermission`: CheckPermission checks if a user has a specific permission for a resource


- `CreateRole`: CreateRole creates a new role


- `CreateUser`: CreateUser creates a new user


- `DeleteRole`: DeleteRole deletes a role


- `DeleteUser`: DeleteUser deletes a user


- `Disable`: Disable disables RBAC (allows all operations)


- `Enable`: Enable enables RBAC


- `GetRole`: GetRole retrieves a role by name


- `GetUser`: GetUser retrieves a user by username


- `GetUserPermissions`: GetUserPermissions returns all permissions for a user


- `IsEnabled`: IsEnabled returns whether RBAC is enabled


- `ListActiveUsers`: ListActiveUsers returns all active users


- `ListRoles`: ListRoles returns all roles


- `ListUsers`: ListUsers returns all users


- `RevokeRole`: RevokeRole revokes a role from a user


- `UpdateLastLogin`: UpdateLastLogin updates the last login time for a user


- `UpdateRole`: UpdateRole updates an existing role


- `UpdateUser`: UpdateUser updates an existing user


- `createDefaultRoles`: createDefaultRoles creates default system roles




##### Role

Role represents a role with a set of permissions



**Methods:**


- `HasPermission`: HasPermission checks if the role has a specific permission




##### User

User represents a user in the RBAC system







#### Functions


##### TestPermission_String




##### TestRBACManager_AssignRole




##### TestRBACManager_CheckPermission




##### TestRBACManager_CreateRole




##### TestRBACManager_CreateUser




##### TestRBACManager_DeleteRole




##### TestRBACManager_DeleteUser




##### TestRBACManager_EnableDisable




##### TestRBACManager_ListRoles




##### TestRBACManager_ListUsers




##### TestRBACManager_New




##### TestRBACManager_RevokeRole










---


### secrets

**Import Path:** `pkg/secrets`




#### Types


##### AWSSecretsProvider

AWSSecretsProvider implements AWS Secrets Manager integration



**Methods:**


- `DeleteSecret`: DeleteSecret deletes a secret from AWS Secrets Manager


- `GetSecret`: GetSecret retrieves a secret from AWS Secrets Manager


- `ListSecrets`: ListSecrets lists secrets from AWS Secrets Manager


- `SetSecret`: SetSecret sets a secret in AWS Secrets Manager


- `Type`: Type returns the provider type




##### MockSecretsProvider

MockSecretsProvider for testing



**Methods:**


- `DeleteSecret`: 

- `GetSecret`: 

- `ListSecrets`: 

- `SetSecret`: 

- `Type`: 



##### Secret

Secret represents a secret value




##### SecretsConfig

SecretsConfig represents secrets management configuration




##### SecretsManager

SecretsManager manages multiple secret providers



**Methods:**


- `DeleteSecret`: DeleteSecret deletes a secret


- `Disable`: Disable disables the secrets manager


- `Enable`: Enable enables the secrets manager


- `GetProviders`: GetProviders returns the list of registered provider types


- `GetSecret`: GetSecret retrieves a secret by path


- `IsEnabled`: IsEnabled returns whether the secrets manager is enabled


- `ListSecrets`: ListSecrets lists secrets with optional prefix


- `RegisterProvider`: RegisterProvider registers a secrets provider


- `ResolveSecrets`: ResolveSecrets resolves secret references in a data structure


- `SetSecret`: SetSecret sets a secret value


- `resolveMap`: resolveMap resolves secrets in a map


- `resolveSlice`: resolveSlice resolves secrets in a slice


- `resolveString`: resolveString resolves secret references in a string


- `resolveValue`: resolveValue recursively resolves secrets in any value




##### SecretsProvider

SecretsProvider defines the interface for secret providers




##### VaultProvider

VaultProvider implements HashiCorp Vault integration



**Methods:**


- `DeleteSecret`: DeleteSecret deletes a secret from Vault


- `GetSecret`: GetSecret retrieves a secret from Vault


- `ListSecrets`: ListSecrets lists secrets from Vault


- `SetSecret`: SetSecret sets a secret in Vault


- `Type`: Type returns the provider type







#### Functions


##### TestParseSecretPath




##### TestSecretsManager_DeleteSecret




##### TestSecretsManager_EnableDisable




##### TestSecretsManager_GetSecret




##### TestSecretsManager_ListSecrets




##### TestSecretsManager_New




##### TestSecretsManager_RegisterProvider




##### TestSecretsManager_ResolveSecrets




##### TestSecretsManager_SetSecret




##### parseSecretPath

parseSecretPath parses a secret path in the format "provider://path"









---


### ssh

**Import Path:** `pkg/ssh`




#### Types


##### Connection

Connection represents an SSH connection



**Methods:**


- `Close`: Close closes the SSH connection


- `Connect`: Connect establishes the SSH connection


- `Execute`: Execute runs a command on the remote host


- `buildClientConfig`: buildClientConfig creates the SSH client configuration


- `loadPrivateKey`: loadPrivateKey loads a private key from file




##### ConnectionConfig

ConnectionConfig holds SSH connection configuration



**Methods:**


- `SetDefaults`: SetDefaults sets default values for connection config


- `Validate`: Validate validates the connection configuration




##### ExecuteResult

ExecuteResult holds the result of command execution



**Methods:**


- `Success`: Success returns true if the command executed successfully




##### Executor

Executor defines the interface for executing commands remotely




##### LocalExecutor

LocalExecutor executes commands locally for testing



**Methods:**


- `Close`: Close is a no-op for local execution


- `Connect`: Connect is a no-op for local execution


- `Execute`: Execute runs a command locally using shell




##### MockExecutor

MockExecutor is a mock implementation of the Executor interface for testing



**Methods:**


- `Close`: Close closes the connection (mock implementation)


- `Connect`: Connect establishes a connection (mock implementation)


- `Execute`: Execute executes a command (mock implementation)




##### RealSSHConnection

RealSSHConnection implements the Executor interface using real SSH connections



**Methods:**


- `Close`: Close closes the SSH connection


- `Connect`: Connect establishes the SSH connection


- `Execute`: Execute executes a command over SSH


- `connectWithTimeout`: connectWithTimeout establishes SSH connection with timeout


- `createAuthMethods`: createAuthMethods creates authentication methods based on configuration


- `createHostKeyCallback`: createHostKeyCallback creates the host key verification callback


- `createKeyAuth`: createKeyAuth creates authentication from private key string


- `createKeyFileAuth`: createKeyFileAuth creates authentication from private key file


- `createSSHConfig`: createSSHConfig creates the SSH client configuration


- `trySSHAgent`: trySSHAgent attempts to use SSH agent for authentication




##### SSHConnectionPool

SSHConnectionPool manages a pool of SSH connections for efficiency



**Methods:**


- `CloseAll`: CloseAll closes all connections in the pool


- `GetConnection`: GetConnection gets or creates a connection for the given configuration







#### Functions


##### TestConnectionConfig_SetDefaults




##### TestConnectionConfig_SetDefaultsDoesNotOverride




##### TestConnectionConfig_Validate




##### TestConnection_ConnectWithInvalidConfig




##### TestConnection_ExecuteWithoutConnection




##### TestExecuteResult_Success




##### TestNewConnection










---


### templating

**Import Path:** `pkg/templating`




#### Types


##### TemplateEngine

TemplateEngine provides template rendering capabilities



**Methods:**


- `AddFunction`: AddFunction adds a custom function to the template engine


- `Render`: Render renders a template string with the given variables


- `RenderFile`: RenderFile renders a template file with the given variables


- `RenderFileToFile`: RenderFileToFile renders a template file and writes the result to another file


- `RenderToFile`: RenderToFile renders a template and writes the result to a file


- `addBuiltinFunctions`: addBuiltinFunctions adds commonly used template functions







#### Functions


##### TestTemplateEngine_AddFunction




##### TestTemplateEngine_Render




##### TestTemplateEngine_RenderFile




##### TestTemplateEngine_WithBuiltinFunctions










---


### types

**Import Path:** `pkg/types`




#### Types


##### ConnectionConfig

ConnectionConfig represents connection configuration for a target




##### DiffAction

DiffAction represents the type of change needed




##### Inventory

Inventory represents the complete inventory configuration




##### MockProvider

MockProvider is a test implementation of the Provider interface



**Methods:**


- `Apply`: 

- `Diff`: 

- `Read`: 

- `Type`: 

- `Validate`: 



##### Provider

Provider defines the interface that all resource providers must implement




##### ProviderRegistry

ProviderRegistry manages available providers



**Methods:**


- `Get`: Get retrieves a provider for the given resource type


- `Register`: Register registers a provider for a resource type


- `Types`: Types returns all registered provider types




##### Resource

Resource represents a unit of infrastructure state



**Methods:**


- `ResourceID`: ResourceID returns a unique identifier for the resource


- `Validate`: Validate checks if the resource configuration is valid




##### ResourceDiff

ResourceDiff represents the difference between current and desired state




##### ResourceState

ResourceState represents the desired state of a resource




##### Target

Target represents a target host for configuration management



**Methods:**


- `AddLabel`: AddLabel adds or updates a label on the target


- `Address`: Address returns the target address in host:port format


- `GetKeyFile`: GetKeyFile returns the effective SSH key file for the target


- `GetPassword`: GetPassword returns the effective password for the target


- `GetPort`: GetPort returns the effective port for the target


- `GetUser`: GetUser returns the effective username for the target


- `HasLabel`: HasLabel checks if the target has a specific label with the given value


- `RemoveLabel`: RemoveLabel removes a label from the target


- `Validate`: Validate validates the target configuration




##### TargetGroup

TargetGroup represents a group of targets







#### Functions


##### TestProviderRegistry_Get




##### TestProviderRegistry_Register




##### TestProviderRegistry_RegisterDuplicate




##### TestProviderRegistry_Types




##### TestResource_ResourceID




##### TestResource_Validate




##### isValidHostname

isValidHostname validates a hostname according to RFC standards









---


### webui

**Import Path:** `pkg/webui`




#### Types


##### ExecutionRecord

ExecutionRecord represents an execution record for the UI




##### WebUIConfig

WebUIConfig represents web UI configuration




##### WebUIServer

WebUIServer provides a web interface for Chisel



**Methods:**


- `AddExecution`: AddExecution adds an execution record to the UI


- `AddModule`: AddModule adds a module to the UI


- `Start`: Start starts the web UI server


- `Stop`: Stop stops the web UI server


- `handleExecutions`: handleExecutions handles execution list requests


- `handleHealth`: handleHealth handles health check requests


- `handleIndex`: handleIndex handles the main dashboard page


- `handleModuleDetail`: handleModuleDetail handles individual module detail requests


- `handleModules`: handleModules handles module list requests


- `handleNotFound`: handleNotFound handles 404 requests


- `handleStatic`: handleStatic handles static file requests


- `handleStatistics`: handleStatistics handles statistics requests


- `withCORS`: withCORS adds CORS headers to API responses


- `writeJSON`: writeJSON writes a JSON response







#### Functions


##### TestWebUIServer_CORSHeaders




##### TestWebUIServer_ExecutionsEndpoint




##### TestWebUIServer_HealthEndpoint




##### TestWebUIServer_JSONResponse




##### TestWebUIServer_ModuleDetailEndpoint




##### TestWebUIServer_ModulesEndpoint




##### TestWebUIServer_New




##### TestWebUIServer_NotFoundEndpoint




##### TestWebUIServer_StaticFiles




##### TestWebUIServer_StatisticsEndpoint










---


### winrm

**Import Path:** `pkg/winrm`




#### Types


##### ConnectionConfig

ConnectionConfig holds WinRM connection configuration



**Methods:**


- `SetDefaults`: SetDefaults sets default values for the connection config


- `Validate`: Validate validates the connection configuration




##### ExecuteResult

ExecuteResult represents the result of a WinRM command execution



**Methods:**


- `Success`: Success returns true if the command executed successfully




##### Executor

Executor interface compatibility




##### WinRMConnection

WinRMConnection represents a WinRM connection



**Methods:**


- `Close`: Close closes the WinRM connection


- `Connect`: Connect establishes the WinRM connection


- `Execute`: Execute executes a command over WinRM


- `attemptConnection`: attemptConnection attempts to establish a WinRM connection







#### Functions


##### TestNewWinRMConnection




##### TestWinRMConnectionConfig_SetDefaults




##### TestWinRMConnectionConfig_Validate




##### TestWinRMConnection_Connect




##### TestWinRMConnection_Execute




##### TestWinRMExecuteResult_Success










---


