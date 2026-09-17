# API Collection

Base URL: `http://localhost:8080` (or your configured port)

## General Note on Pagination
For all list endpoints (e.g., getting all employees, tasks, activities, screenshots, devices, apps, websites), the API supports pagination via query parameters:
- `page`: The page number (default: 1)
- `limit`: The number of items per page (default: 10)

Example:
```bash
curl -X GET "http://localhost:8080/api/v1/employees?page=2&limit=20" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

The response will have the following structure:
```json
{
  "success": true,
  "data": [ ... items ... ],
  "total_count": 100,
  "page": 2,
  "limit": 20,
  "total_pages": 5
}
```

## Public Endpoints

### 1. Hello API
```bash
curl -X GET http://localhost:8080/api/v1/hello
```

### 2. Health Check
```bash
curl -X GET http://localhost:8080/api/v1/health
```

### 3. Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### 4. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```
*Note: If the logged-in user is an Employee, the response will now include `employee_id` and the `employee` object alongside the `token` and `user`.*

### 5. Employee Login (with device check)
```bash
curl -X POST http://localhost:8080/api/v1/auth/employee-login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "employee@example.com",
    "password": "password123",
    "device_fingerprint": "UNIQUE_FINGERPRINT_123"
  }'
```
*Note: Returns `is_device_registered: true` if the device is found, otherwise `false` so the frontend knows to call the Register API.*

## Desktop App Endpoints (Public/Device-Based)

### 6. Check if Device is Registered
```bash
curl -X GET "http://localhost:8080/api/v1/app/device/check?device_fingerprint=UNIQUE_FINGERPRINT_123"
```
*Note: Returns `success: true` and device details if found, otherwise `404 Not Found`.*

### 7. Background Screenshot Upload (No JWT Required)
```bash
curl -X POST http://localhost:8080/api/v1/app/employee-screenshots \
  -H "X-Device-Fingerprint: UNIQUE_FINGERPRINT_123" \
  -F "screenshot=@/path/to/screenshot.jpg" \
  -F "captured_at=2026-09-12T12:00:00Z" \
  -F "notes=Background screenshot"
```
*Note: This relies on the device fingerprint for authorization and links the screenshot to the device and its associated employee.*

### 8. Save Employee Keyboard and Mouse Activity (No JWT Required)
```bash
curl -X POST http://localhost:8080/api/v1/app/employee-activity \
  -H "X-Device-Fingerprint: UNIQUE_FINGERPRINT_123" \
  -H "Content-Type: application/json" \
  -d '{
    "keyboard_activity": 125,
    "mouse_activity": 84,
    "keystrokes": ["H", "e", "l", "l", "o"],
    "mouse_events": [
      {"type": "move", "x": 640, "y": 360},
      {"type": "click", "button": "left", "x": 640, "y": 360}
    ],
    "captured_at": "2026-09-13T12:00:00Z"
  }'
```
*Note: The device fingerprint identifies the employee. Activity counts, keystrokes, mouse events, employee, device, and capture time are saved in the database.*

---

## Protected Endpoints (User)
*Note: Replace `<YOUR_TOKEN>` with the token received from the Login API.*

### 6. Get Current User (Me)
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 6. Get User by ID
```bash
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 7. Update Current User
```bash
curl -X PUT http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe Updated",
    "email": "john.updated@example.com"
  }'
```

### 8. Delete Current User
```bash
curl -X DELETE http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 9. Update User Status (Ban/Unban)
```bash
curl -X PUT http://localhost:8080/api/v1/users/2/status \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "is_banned": true
  }'
```

---

## Protected Endpoints (Options / Dropdowns)

These endpoints are lightweight and designed specifically for populating UI option boxes (select dropdowns). They only return the `id` and `name` of the entities and are **not** paginated.

### 10. Get Department Options
```bash
curl -X GET http://localhost:8080/api/v1/options/departments \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 11. Get Role Options
```bash
curl -X GET http://localhost:8080/api/v1/options/roles \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 12. Get Team Options
```bash
curl -X GET http://localhost:8080/api/v1/options/teams \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 13. Get Task Options
```bash
curl -X GET http://localhost:8080/api/v1/options/tasks \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 14. Get Employee Options
```bash
curl -X GET http://localhost:8080/api/v1/options/employees \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

**Expected Response (for any options endpoint):**
```json
{
  "success": true,
  "data": [
    {
      "id": "cdef1234-5678-90ab-cdef-1234567890ab",
      "name": "Design Team"
    },
    {
      "id": "12345678-90ab-cdef-1234-567890abcdef",
      "name": "Engineering Team"
    }
  ]
}
```

---

## Protected Endpoints (Employee CRUD)

### 10. Create Employee
```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "00000000-0000-0000-0000-000000000000",
    "employee_code": "EMP001",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "password123",
    "designation": "Software Engineer",
    "department_id": "00000000-0000-0000-0000-000000000000",
    "role_id": "00000000-0000-0000-0000-000000000000",
    "status": "active"
  }'
```

### 11. List All Employees
```bash
curl -X GET http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 12. Get Employee by ID
```bash
curl -X GET http://localhost:8080/api/v1/employees/<EMPLOYEE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 13. Update Employee
```bash
curl -X PUT http://localhost:8080/api/v1/employees/<EMPLOYEE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe Updated",
    "email": "jane.updated@example.com",
    "employee_code": "EMP001",
    "designation": "Senior Software Engineer",
    "organization_id": "00000000-0000-0000-0000-000000000000",
    "department_id": "00000000-0000-0000-0000-000000000000",
    "role_id": "00000000-0000-0000-0000-000000000000",
    "status": "active"
  }'
```

### 14. Delete Employee
```bash
curl -X DELETE http://localhost:8080/api/v1/employees/<EMPLOYEE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Protected Endpoints (Department CRUD)

### 15. Create Department
```bash
curl -X POST http://localhost:8080/api/v1/departments \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering",
    "status": "active"
  }'
```

### 16. List All Departments
```bash
curl -X GET http://localhost:8080/api/v1/departments \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 17. Get Department by ID
```bash
curl -X GET http://localhost:8080/api/v1/departments/<DEPARTMENT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 18. Update Department
```bash
curl -X PUT http://localhost:8080/api/v1/departments/<DEPARTMENT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering & Technology",
    "status": "inactive"
  }'
```

### 19. Delete Department
```bash
curl -X DELETE http://localhost:8080/api/v1/departments/<DEPARTMENT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Protected Endpoints (Role CRUD)

### 20. Create Role
```bash
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Manager",
    "status": "active",
    "department_id": "<DEPARTMENT_UUID>"
  }'
```

### 21. List All Active Roles
```bash
curl -X GET http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 22. List Active Roles by Department
```bash
curl -X GET http://localhost:8080/api/v1/departments/<DEPARTMENT_UUID>/roles \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 22. Get Role by ID
```bash
curl -X GET http://localhost:8080/api/v1/roles/<ROLE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 23. Update Role
```bash
curl -X PUT http://localhost:8080/api/v1/roles/<ROLE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Manager",
    "status": "inactive"
  }'
```

### 24. Delete Role
```bash
curl -X DELETE http://localhost:8080/api/v1/roles/<ROLE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Protected Endpoints (Employee Devices)

### 25. Register Employee Device
```bash
curl -X POST http://localhost:8080/api/v1/employee-devices/register \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "device_fingerprint": "UNIQUE_FINGERPRINT_123",
    "machine_guid_hash": "hash123",
    "bios_serial_hash": "hash123",
    "motherboard_serial_hash": "hash123",
    "device_name": "Desktop-PC",
    "manufacturer": "Dell",
    "model": "XPS 15",
    "os_name": "Windows 11",
    "os_version": "10.0.22621",
    "architecture": "x64",
    "cpu_name": "Intel Core i7",
    "ram_total_mb": 16384,
    "disk_total_gb": 512,
    "app_version": "1.0.0"
  }'
```

### 26. List My Devices
```bash
curl -X GET http://localhost:8080/api/v1/employee-devices \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 27. Delete/Revoke My Device
```bash
curl -X DELETE http://localhost:8080/api/v1/employee-devices/<DEVICE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Protected Endpoints (Admin Device Management)

### 28. List All Devices Across Company
```bash
curl -X GET http://localhost:8080/api/v1/admin/employee-devices \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 29. List All Devices for a Specific Employee
```bash
curl -X GET http://localhost:8080/api/v1/admin/employees/<EMPLOYEE_UUID>/devices \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Protected Endpoints (Employee Screenshots)

### 30. Upload/Save Employee Screenshot
```bash
curl -X POST http://localhost:8080/api/v1/employee-screenshots \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -F "screenshot=@/path/to/screenshot.jpg" \
  -F "captured_at=2026-09-12T12:00:00Z" \
  -F "notes=Optional note"
```

### 31. List My Screenshots
```bash
curl -X GET http://localhost:8080/api/v1/employee-screenshots \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 32. Get My Screenshot Info
```bash
curl -X GET http://localhost:8080/api/v1/employee-screenshots/<SCREENSHOT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 33. Download My Screenshot
```bash
curl -X GET http://localhost:8080/api/v1/employee-screenshots/<SCREENSHOT_UUID>/download \
  -H "Authorization: Bearer <YOUR_TOKEN>" --output screenshot.jpg
```

---

## Protected Endpoints (Admin Employee Screenshots)

### 34. List All Screenshots Across Company (with optional filters)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/employee-screenshots?employee_id=<EMPLOYEE_UUID>&department_id=<DEPARTMENT_UUID>" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```
*Note: Omit both query parameters to return screenshots for all employees. Use `employee_id` to filter one employee or `department_id` to filter one department.*

### 35. List All Screenshots for One Employee
```bash
curl -X GET http://localhost:8080/api/v1/admin/employees/<EMPLOYEE_UUID>/screenshots \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 36. Get Any Screenshot Info
```bash
curl -X GET http://localhost:8080/api/v1/admin/employee-screenshots/<SCREENSHOT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 37. Download Any Screenshot
```bash
curl -X GET http://localhost:8080/api/v1/admin/employee-screenshots/<SCREENSHOT_UUID>/download \
  -H "Authorization: Bearer <YOUR_TOKEN>" --output screenshot.jpg
```

### 38. Update Screenshot Details
```bash
curl -X PUT http://localhost:8080/api/v1/admin/employee-screenshots/<SCREENSHOT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -F "notes=Updated notes" \
  -F "captured_at=2026-09-12T12:05:00Z"
```

### 39. Delete Screenshot
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/employee-screenshots/<SCREENSHOT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

*Admin screenshot endpoints accept admin users with `user_type` 0 or 1. Employee users have `user_type` 2 and receive `403 Forbidden`.*

---

## Protected Endpoints (Admin Employee Activities)

### 40. List All Activities Across Company (with optional filters)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/employee-activities?employee_id=<EMPLOYEE_UUID>&department_id=<DEPARTMENT_UUID>" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```
*Note: Omit both query parameters to return activities for all employees. Use `employee_id` to filter one employee or `department_id` to filter one department.*

**Expected Response (Success):**
```json
{
  "success": true,
  "data": [
    {
      "id": "e8d69f0b-5f16-43b6-9f4a-9b7e7c4f1c1f",
      "employee_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
      "employee": {
        "id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
        "name": "Jane Doe",
        "email": "jane@handdy.com",
        "department_id": "d1c2b3a4-1234-5678-90ab-cdef12345678"
      },
      "employee_device_id": "f5e4d3c2-b1a0-9876-5432-10fedcba0987",
      "keyboard_activity": 145,
      "mouse_activity": 230,
      "keystrokes": [
        { "key": "Enter", "timestamp": "2026-09-14T10:15:20Z" },
        { "key": "a", "timestamp": "2026-09-14T10:15:21Z" }
      ],
      "mouse_events": [
        { "x": 450, "y": 300, "type": "click", "timestamp": "2026-09-14T10:15:25Z" }
      ],
      "captured_at": "2026-09-14T10:15:30Z",
      "created_at": "2026-09-14T10:15:31Z"
    }
  ],
  "total_count": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1
}
```
      "id": "e8d69f0b-5f16-43b6-9f4a-9b7e7c4f1c1f",
      "employee_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
      "employee": {
        "id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
        "name": "Jane Doe",
        "email": "jane@handdy.com",
        "department_id": "d1c2b3a4-1234-5678-90ab-cdef12345678"
      },
      "employee_device_id": "f5e4d3c2-b1a0-9876-5432-10fedcba0987",
      "keyboard_activity": 145,
      "mouse_activity": 230,
      "keystrokes": [
        { "key": "Enter", "timestamp": "2026-09-14T10:15:20Z" },
        { "key": "a", "timestamp": "2026-09-14T10:15:21Z" }
      ],
      "mouse_events": [
        { "x": 450, "y": 300, "type": "click", "timestamp": "2026-09-14T10:15:25Z" }
      ],
      "captured_at": "2026-09-14T10:15:30Z",
      "created_at": "2026-09-14T10:15:31Z"
    }
  ]
}
```

### 41. List All Activities for One Employee
```bash
curl -X GET http://localhost:8080/api/v1/admin/employees/<EMPLOYEE_UUID>/activities \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

**Expected Response (Success):**
```json
{
  "success": true,
  "data": [
    // ... Same activity objects as above ...
  ],
  "total_count": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1
}
```

*Admin activity endpoints accept admin users with `user_type` 0 or 1. Employee users receive `403 Forbidden`.*

---

## App Endpoints (Applications & Websites)
*Note: These endpoints require the `X-Device-Fingerprint` header.*

### 42. Save Employee Application Usage
```bash
curl -X POST http://localhost:8080/api/v1/app/applications \
  -H "X-Device-Fingerprint: <FINGERPRINT>" \
  -H "Content-Type: application/json" \
  -d '{
    "application_name": "Visual Studio Code",
    "executable_name": "Code.exe",
    "window_title": "index.js - project - Visual Studio Code",
    "started_at": "2026-09-14T10:00:00Z",
    "ended_at": "2026-09-14T10:05:00Z",
    "duration_seconds": 300
  }'
```

### 43. Save Employee Website Usage
```bash
curl -X POST http://localhost:8080/api/v1/app/websites \
  -H "X-Device-Fingerprint: <FINGERPRINT>" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "github.com",
    "url": "https://github.com/my/project",
    "page_title": "Project Overview",
    "started_at": "2026-09-14T10:05:00Z",
    "ended_at": "2026-09-14T10:10:00Z",
    "duration_seconds": 300
  }'
```

---

## Protected Endpoints (Admin Applications & Websites)

### 44. List All Applications (with optional filters)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/applications?employee_id=<EMPLOYEE_UUID>&department_id=<DEPARTMENT_UUID>" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 45. List All Applications for One Employee
```bash
curl -X GET http://localhost:8080/api/v1/admin/employees/<EMPLOYEE_UUID>/applications \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 46. Update Application
```bash
curl -X PUT http://localhost:8080/api/v1/admin/applications/<APP_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{ "window_title": "Updated Title" }'
```

### 47. Delete Application
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/applications/<APP_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 48. List All Websites (with optional filters)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/websites?employee_id=<EMPLOYEE_UUID>&department_id=<DEPARTMENT_UUID>" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 49. List All Websites for One Employee
```bash
curl -X GET http://localhost:8080/api/v1/admin/employees/<EMPLOYEE_UUID>/websites \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 50. Update Website
```bash
curl -X PUT http://localhost:8080/api/v1/admin/websites/<WEBSITE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{ "page_title": "Updated Title" }'
```

### 51. Delete Website
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/websites/<WEBSITE_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Teams

### 52. Create Team
```bash
curl -X POST http://localhost:8080/api/v1/admin/teams \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{ "name": "Engineering Team" }'
```

### 53. List All Teams
```bash
curl -X GET http://localhost:8080/api/v1/admin/teams \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 54. Update Team
```bash
curl -X PUT http://localhost:8080/api/v1/admin/teams/<TEAM_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{ "name": "Design Team" }'
```

### 55. Delete Team (Soft Delete)
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/teams/<TEAM_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Tasks

### 56. Create Task
```bash
curl -X POST http://localhost:8080/api/v1/admin/tasks \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Frontend Development",
    "type": "work",
    "idle_logout_enabled": true,
    "idle_logout_minutes": 15,
    "status": "active"
  }'
```

### 57. List All Tasks
```bash
curl -X GET http://localhost:8080/api/v1/admin/tasks \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 58. Update Task
```bash
curl -X PUT http://localhost:8080/api/v1/admin/tasks/<TASK_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Development",
    "status": "inactive"
  }'
```

### 59. Delete Task (Soft Delete)
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/tasks/<TASK_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## Task Assignments

### 60. Assign Task
```bash
curl -X POST http://localhost:8080/api/v1/admin/tasks/<TASK_UUID>/assignments \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "assign_type": "employee",
    "assign_id": "<EMPLOYEE_UUID>"
  }'
```
*Note: `assign_type` can be `"company"`, `"team"`, or `"employee"`. For `"company"`, `assign_id` can be omitted or null.*

### 61. List Assignments for a Task
```bash
curl -X GET http://localhost:8080/api/v1/admin/tasks/<TASK_UUID>/assignments \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 62. Remove Task Assignment
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/tasks/assignments/<ASSIGNMENT_UUID> \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---
## Attendance & Timesheet Endpoints (Employee)

### 63. Clock In
```bash
curl -X POST http://localhost:8080/api/v1/attendance/clock-in \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"work_type": "remote"}'
```

### 64. Clock Out
```bash
curl -X POST http://localhost:8080/api/v1/attendance/clock-out \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 65. Start Break
```bash
curl -X POST http://localhost:8080/api/v1/attendance/break-start \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Lunch", "other_reason": ""}'
```

### 66. End Break
```bash
curl -X POST http://localhost:8080/api/v1/attendance/break-end \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 67. Get My Timesheet
```bash
curl -X GET "http://localhost:8080/api/v1/attendance/timesheet?page=1&limit=10" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---
## Attendance & Timesheet Endpoints (Admin)

### 68. List All Attendances (Admin)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/attendance?page=1&limit=10" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 69. List Attendance For Specific Employee (Admin)
```bash
curl -X GET "http://localhost:8080/api/v1/admin/employees/<EMPLOYEE_UUID>/attendance?page=1&limit=10" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```
