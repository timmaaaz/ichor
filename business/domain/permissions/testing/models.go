package testing

import (
	"time"
)

// RestrictedColumns with field names matching NewRestrictedColumn struct
var RestrictedColumns = []map[string]interface{}{
	{
		"TableName":  "roles",
		"ColumnName": "description",
	},
	{
		"TableName":  "users",
		"ColumnName": "email",
	},
	{
		"TableName":  "user_organizations",
		"ColumnName": "is_unit_manager",
	},
	{
		"TableName":  "user_organizations",
		"ColumnName": "role_id",
	},
	{
		"TableName":  "users",
		"ColumnName": "user_id",
	},
	{
		"TableName":  "users",
		"ColumnName": "username",
	},
}

// Users with field names matching NewUser struct
var Users = []map[string]interface{}{
	{
		"Username": "admin",
		"Email":    "admin@example.com",
	},
	{
		"Username": "manager1",
		"Email":    "manager1@example.com",
	},
	{
		"Username": "manager2",
		"Email":    "manager2@example.com",
	},
	{
		"Username": "employee1",
		"Email":    "employee1@example.com",
	},
	{
		"Username": "employee2",
		"Email":    "employee2@example.com",
	},
	{
		"Username": "employee3",
		"Email":    "employee3@example.com",
	},
	{
		"Username": "employee4",
		"Email":    "employee4@example.com",
	},
	{
		"Username": "readonly",
		"Email":    "readonly@example.com",
	},
}

// Roles with field names matching NewRole struct
var Roles = []map[string]interface{}{
	{
		"Name":        "ADMIN",
		"Description": "System Administrator with full access",
	},
	{
		"Name":        "EMPLOYEE",
		"Description": "Regular employee with standard access",
	},
	{
		"Name":        "FINANCE_ADMIN",
		"Description": "Finance Department Administrator",
	},
	{
		"Name":        "HR_ADMIN",
		"Description": "Human Resources Administrator",
	},
	{
		"Name":        "MANAGER",
		"Description": "Department manager with extended privileges",
	},
	{
		"Name":        "READONLY",
		"Description": "Read-only access to specific resources",
	},
	{
		"Name":        "TEMP_ADMIN",
		"Description": "Temporary Administrator access",
	},
}

// OrganizationalUnits with field names matching NewOrganizationalUnit struct
// Properly organized to show the hierarchy relationship with level 0 for root
var OrganizationalUnits = []map[string]interface{}{
	// Root level - must be first since others depend on it
	{
		"Name":                  "Company Headquarters",
		"ParentID":              nil, // Root has no parent
		"Level":                 0,   // Root is level 0
		"Path":                  "Company_Headquarters",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "COMPANY",
		"IsActive":              true,
	},
	// Level 1 - Departments under HQ (notice level adjustment)
	{
		"Name":                  "Finance Department",
		"ParentID":              nil, // Will need to be set to HQ ID
		"Level":                 1,   // Level 1 under the root
		"Path":                  "Company_Headquarters.Finance_Department",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "DEPARTMENT",
		"IsActive":              true,
	},
	{
		"Name":                  "HR Department",
		"ParentID":              nil, // Will need to be set to HQ ID
		"Level":                 1,   // Level 1 under the root
		"Path":                  "Company_Headquarters.HR_Department",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "DEPARTMENT",
		"IsActive":              true,
	},
	{
		"Name":                  "Sales Department",
		"ParentID":              nil, // Will need to be set to HQ ID
		"Level":                 1,   // Level 1 under the root
		"Path":                  "Company_Headquarters.Sales_Department",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "DEPARTMENT",
		"IsActive":              true,
	},
	{
		"Name":                  "IT Department",
		"ParentID":              nil, // Will need to be set to HQ ID
		"Level":                 1,   // Level 1 under the root
		"Path":                  "Company_Headquarters.IT_Department",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "DEPARTMENT",
		"IsActive":              true,
	},
	// Level 2 - Teams/Regions under departments (notice level adjustment)
	{
		"Name":                  "Accounting Team",
		"ParentID":              nil, // Will need to be set to Finance Dept ID
		"Level":                 2,   // Level 2 (under Finance department)
		"Path":                  "Company_Headquarters.Finance_Department.Accounting_Team",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "TEAM",
		"IsActive":              true,
	},
	{
		"Name":                  "Payroll Team",
		"ParentID":              nil, // Will need to be set to Finance Dept ID
		"Level":                 2,   // Level 2 (under Finance department)
		"Path":                  "Company_Headquarters.Finance_Department.Payroll_Team",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "TEAM",
		"IsActive":              true,
	},
	{
		"Name":                  "Recruitment Team",
		"ParentID":              nil, // Will need to be set to HR Dept ID
		"Level":                 2,   // Level 2 (under HR department)
		"Path":                  "Company_Headquarters.HR_Department.Recruitment_Team",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "TEAM",
		"IsActive":              true,
	},
	{
		"Name":                  "Benefits Team",
		"ParentID":              nil, // Will need to be set to HR Dept ID
		"Level":                 2,   // Level 2 (under HR department)
		"Path":                  "Company_Headquarters.HR_Department.Benefits_Team",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "TEAM",
		"IsActive":              true,
	},
	{
		"Name":                  "East Region",
		"ParentID":              nil, // Will need to be set to Sales Dept ID
		"Level":                 2,   // Level 2 (under Sales department)
		"Path":                  "Company_Headquarters.Sales_Department.East_Region",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "REGION",
		"IsActive":              true,
	},
	{
		"Name":                  "West Region",
		"ParentID":              nil, // Will need to be set to Sales Dept ID
		"Level":                 2,   // Level 2 (under Sales department)
		"Path":                  "Company_Headquarters.Sales_Department.West_Region",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "REGION",
		"IsActive":              true,
	},
	{
		"Name":                  "Systems Team",
		"ParentID":              nil, // Will need to be set to IT Dept ID
		"Level":                 2,   // Level 2 (under IT department)
		"Path":                  "Company_Headquarters.IT_Department.Systems_Team",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "TEAM",
		"IsActive":              true,
	},
	{
		"Name":                  "Development Team",
		"ParentID":              nil, // Will need to be set to IT Dept ID
		"Level":                 2,   // Level 2 (under IT department)
		"Path":                  "Company_Headquarters.IT_Department.Development_Team",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "TEAM",
		"IsActive":              true,
	},
	// Level 3 - Branches under regions (notice level adjustment)
	{
		"Name":                  "Northeast Branch",
		"ParentID":              nil, // Will need to be set to East Region ID
		"Level":                 3,   // Level 3 (under East region)
		"Path":                  "Company_Headquarters.Sales_Department.East_Region.Northeast_Branch",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "BRANCH",
		"IsActive":              true,
	},
	{
		"Name":                  "Southeast Branch",
		"ParentID":              nil, // Will need to be set to East Region ID
		"Level":                 3,   // Level 3 (under East region)
		"Path":                  "Company_Headquarters.Sales_Department.East_Region.Southeast_Branch",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "BRANCH",
		"IsActive":              true,
	},
	{
		"Name":                  "Northwest Branch",
		"ParentID":              nil, // Will need to be set to West Region ID
		"Level":                 3,   // Level 3 (under West region)
		"Path":                  "Company_Headquarters.Sales_Department.West_Region.Northwest_Branch",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "BRANCH",
		"IsActive":              true,
	},
	{
		"Name":                  "Southwest Branch",
		"ParentID":              nil, // Will need to be set to West Region ID
		"Level":                 3,   // Level 3 (under West region)
		"Path":                  "Company_Headquarters.Sales_Department.West_Region.Southwest_Branch",
		"CanInheritPermissions": true,
		"CanRollupData":         true,
		"UnitType":              "BRANCH",
		"IsActive":              true,
	},
}

// UserRoles with field names matching NewUserRole struct
var UserRoles = []map[string]interface{}{
	{
		"UserID": nil, // Will need to be set programmatically
		"RoleID": nil, // Will need to be set programmatically
	},
}

// TableAccess with field names matching NewTableAccess struct
var TableAccess = []map[string]interface{}{
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "users",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "roles",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "user_roles",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "table_access",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "restricted_columns",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "organizational_units",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "user_organizations",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "org_unit_field_restrictions",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "cross_unit_permissions",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "permission_overrides",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
	{
		"RoleID":    nil, // Will need to be set programmatically
		"TableName": "temporary_unit_access",
		"CanCreate": true,
		"CanRead":   true,
		"CanUpdate": true,
		"CanDelete": true,
	},
}

// UserOrganizations with field names matching NewUserOrganization struct
var UserOrganizations = []map[string]interface{}{
	{
		"UserID":        nil, // Will need to be set programmatically
		"OrgUnitID":     nil, // Will need to be set programmatically
		"RoleID":        nil, // Will need to be set programmatically
		"IsUnitManager": true,
		"StartDate":     time.Now(),
		"EndDate":       time.Time{}, // Null value
		"CreatedBy":     nil,         // Will need to be set programmatically
	},
}

// OrgUnitColumnAccess with field names matching NewOrgUnitFieldRestriction struct
var OrgUnitColumnAccess = []map[string]interface{}{
	{
		"OrgUnitID":             nil, // Will need to be set programmatically
		"TableName":             "users",
		"ColumnName":            "email",
		"CanInheritPermissions": false,
		"CanRollupData":         false,
		"CanRead":               true,
		"CanUpdate":             false,
	},
	{
		"OrgUnitID":             nil, // Will need to be set programmatically
		"TableName":             "users",
		"ColumnName":            "user_id",
		"CanInheritPermissions": false,
		"CanRollupData":         false,
		"CanRead":               true,
		"CanUpdate":             false,
	},
	{
		"OrgUnitID":             nil, // Will need to be set programmatically
		"TableName":             "user_roles",
		"ColumnName":            "role_id",
		"CanInheritPermissions": false,
		"CanRollupData":         false,
		"CanRead":               true,
		"CanUpdate":             false,
	},
	{
		"OrgUnitID":             nil, // Will need to be set programmatically
		"TableName":             "user_organizations",
		"ColumnName":            "start_date",
		"CanInheritPermissions": false,
		"CanRollupData":         false,
		"CanRead":               true,
		"CanUpdate":             false,
	},
}

// CrossUnitPermissions with field names matching NewCrossUnitPermission struct
var CrossUnitPermissions = []map[string]interface{}{
	{
		"SourceUnitID":   nil, // Will need to be set programmatically
		"TargetUnitID":   nil, // Will need to be set programmatically
		"PermissionType": "READ",
		"GrantedBy":      nil, // Will need to be set programmatically
		"ValidFrom":      time.Now(),
		"ValidUntil":     time.Now().AddDate(1, 0, 0), // 1 year
		"Reason":         "Finance needs to read HR data for payroll processing",
	},
	{
		"SourceUnitID":   nil, // Will need to be set programmatically
		"TargetUnitID":   nil, // Will need to be set programmatically
		"PermissionType": "READ",
		"GrantedBy":      nil, // Will need to be set programmatically
		"ValidFrom":      time.Now(),
		"ValidUntil":     time.Now().AddDate(1, 0, 0), // 1 year
		"Reason":         "HR needs to read Finance data for budget planning",
	},
	{
		"SourceUnitID":   nil, // Will need to be set programmatically
		"TargetUnitID":   nil, // Will need to be set programmatically
		"PermissionType": "WRITE",
		"GrantedBy":      nil, // Will need to be set programmatically
		"ValidFrom":      time.Now(),
		"ValidUntil":     time.Now().AddDate(1, 0, 0), // 1 year
		"Reason":         "Payroll needs to update Benefits data",
	},
	{
		"SourceUnitID":   nil, // Will need to be set programmatically
		"TargetUnitID":   nil, // Will need to be set programmatically
		"PermissionType": "READ",
		"GrantedBy":      nil, // Will need to be set programmatically
		"ValidFrom":      time.Now(),
		"ValidUntil":     time.Now().AddDate(1, 0, 0), // 1 year
		"Reason":         "Cross-regional data sharing",
	},
}

// PermissionOverrides with field names matching NewPermissionOverride struct
var PermissionOverrides = []map[string]interface{}{
	{
		"UserID":     nil, // Will need to be set programmatically
		"TableName":  "users",
		"ColumnName": nil, // Optional
		"OrgUnitID":  nil, // Will need to be set programmatically
		"CanCreate":  false,
		"CanRead":    true,
		"CanUpdate":  false,
		"CanDelete":  false,
		"Reason":     "Temporary access for user data review project",
		"GrantedBy":  nil, // Will need to be set programmatically
		"ValidFrom":  time.Now(),
		"ValidUntil": time.Now().AddDate(0, 1, 0), // 1 month
	},
	{
		"UserID":     nil, // Will need to be set programmatically
		"TableName":  "users",
		"ColumnName": nil, // Optional
		"OrgUnitID":  nil, // Will need to be set programmatically
		"CanCreate":  false,
		"CanRead":    true,
		"CanUpdate":  false,
		"CanDelete":  false,
		"Reason":     "Covering for employee on leave",
		"GrantedBy":  nil, // Will need to be set programmatically
		"ValidFrom":  time.Now(),
		"ValidUntil": time.Now().AddDate(0, 0, 14), // 14 days
	},
}

// TemporaryUnitAccess with field names matching NewTemporaryUnitAccess struct
var TemporaryUnitAccess = []map[string]interface{}{
	{
		"UserID":         nil, // Will need to be set programmatically
		"OrgUnitID":      nil, // Will need to be set programmatically
		"PermissionType": "READ",
		"Reason":         "Cross-departmental project collaboration",
		"GrantedBy":      nil, // Will need to be set programmatically
		"ValidFrom":      time.Now(),
		"ValidUntil":     time.Now().AddDate(0, 3, 0), // 3 months
	},
	{
		"UserID":         nil, // Will need to be set programmatically
		"OrgUnitID":      nil, // Will need to be set programmatically
		"PermissionType": "ADMIN",
		"Reason":         "Temporary team lead coverage",
		"GrantedBy":      nil, // Will need to be set programmatically
		"ValidFrom":      time.Now(),
		"ValidUntil":     time.Now().AddDate(0, 1, 15), // 45 days
	},
}
