package v1

import (
	"encoding/json"
	"net/http"

	"go-api/middleware"
	"go-api/repository"
	"go-api/services"

	"gorm.io/gorm"
)

func RegisterRoutes(mux *http.ServeMux, db *gorm.DB) {

	employeeRepo := repository.NewEmployeeRepository(db)

	authHandler := &AuthHandler{
		DB:           db,
		EmployeeRepo: employeeRepo,
	}

	userRepo := repository.NewUserRepository(db)
	userHandler := &UserHandler{
		Repo: userRepo,
	}

	emailService := services.NewEmailService()
	employeeHandler := &EmployeeHandler{
		Repo:         employeeRepo,
		EmailService: emailService,
	}

	employeeDeviceRepo := repository.NewEmployeeDeviceRepository(db)
	employeeDeviceHandler := &EmployeeDeviceHandler{
		DeviceRepo:   employeeDeviceRepo,
		EmployeeRepo: employeeRepo,
		DB:           db,
	}

	employeeScreenshotRepo := repository.NewEmployeeScreenshotRepository(db)
	employeeScreenshotHandler := &EmployeeScreenshotHandler{
		Repo:       employeeScreenshotRepo,
		DeviceRepo: employeeDeviceRepo,
		DB:         db,
	}

	employeeActivityRepo := repository.NewEmployeeActivityRepository(db)
	employeeActivityHandler := &EmployeeActivityHandler{
		Repo:       employeeActivityRepo,
		DeviceRepo: employeeDeviceRepo,
		DB:         db,
	}

	applicationRepo := repository.NewApplicationRepository(db)
	applicationHandler := &ApplicationHandler{
		Repo:       applicationRepo,
		DeviceRepo: employeeDeviceRepo,
		DB:         db,
	}

	websiteRepo := repository.NewWebsiteRepository(db)
	websiteHandler := &WebsiteHandler{
		Repo:       websiteRepo,
		DeviceRepo: employeeDeviceRepo,
		DB:         db,
	}

	teamRepo := repository.NewTeamRepository(db)
	teamHandler := &TeamHandler{
		Repo: teamRepo,
		DB:   db,
	}

	taskRepo := repository.NewTaskRepository(db)
	taskAssignmentRepo := repository.NewTaskAssignmentRepository(db)
	taskHandler := &TaskHandler{
		Repo:           taskRepo,
		AssignmentRepo: taskAssignmentRepo,
		DB:             db,
	}

	departmentRepo := repository.NewDepartmentRepository(db)
	departmentHandler := &DepartmentHandler{
		Repo: departmentRepo,
	}

	roleRepo := repository.NewRoleRepository(db)
	roleHandler := &RoleHandler{
		Repo: roleRepo,
	}

	optionsHandler := &OptionsHandler{
		DB: db,
	}

	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceHandler := &AttendanceHandler{
		Repo: attendanceRepo,
		DB:   db,
	}

	// --------------------------------
	// Public
	// --------------------------------

	mux.HandleFunc("GET /api/v1/hello", helloHandler)

	mux.HandleFunc(
		"GET /api/v1/health",
		healthHandler(db),
	)

	mux.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/employee-login",
		authHandler.EmployeeLogin,
	)

	mux.HandleFunc(
		"POST /api/v1/test-email",
		employeeHandler.TestEmail,
	)

	// App (Device-based, non-JWT)
	mux.HandleFunc(
		"GET /api/v1/app/device/check",
		employeeDeviceHandler.CheckDevice,
	)

	mux.HandleFunc(
		"POST /api/v1/app/employee-screenshots",
		employeeScreenshotHandler.AppCreate,
	)

	mux.HandleFunc(
		"POST /api/v1/app/employee-activity",
		employeeActivityHandler.AppCreate,
	)

	mux.HandleFunc(
		"POST /api/v1/app/applications",
		applicationHandler.AppCreate,
	)

	mux.HandleFunc(
		"POST /api/v1/app/websites",
		websiteHandler.AppCreate,
	)

	// --------------------------------
	// Protected
	// --------------------------------

	mux.Handle(
		"GET /api/v1/users/me",
		middleware.JWT(http.HandlerFunc(userHandler.Me)),
	)

	mux.Handle(
		"GET /api/v1/users/{id}",
		middleware.JWT(http.HandlerFunc(userHandler.Get)),
	)

	mux.Handle(
		"PUT /api/v1/users/me",
		middleware.JWT(http.HandlerFunc(userHandler.Update)),
	)

	mux.Handle(
		"DELETE /api/v1/users/me",
		middleware.JWT(http.HandlerFunc(userHandler.Delete)),
	)
	mux.Handle(
		"PUT /api/v1/users/{id}/status",
		middleware.JWT(http.HandlerFunc(userHandler.UpdateStatus)),
	)

	// Options Routes (For Dropdowns)
	mux.Handle(
		"GET /api/v1/options/departments",
		middleware.JWT(http.HandlerFunc(optionsHandler.Departments)),
	)
	mux.Handle(
		"GET /api/v1/options/roles",
		middleware.JWT(http.HandlerFunc(optionsHandler.Roles)),
	)
	mux.Handle(
		"GET /api/v1/options/teams",
		middleware.JWT(http.HandlerFunc(optionsHandler.Teams)),
	)
	mux.Handle(
		"GET /api/v1/options/tasks",
		middleware.JWT(http.HandlerFunc(optionsHandler.Tasks)),
	)
	mux.Handle(
		"GET /api/v1/options/employees",
		middleware.JWT(http.HandlerFunc(optionsHandler.Employees)),
	)

	// Employee Routes
	mux.Handle(
		"POST /api/v1/employees",
		middleware.JWT(http.HandlerFunc(employeeHandler.Create)),
	)
	mux.Handle(
		"GET /api/v1/employees",
		middleware.JWT(http.HandlerFunc(employeeHandler.List)),
	)
	mux.Handle(
		"GET /api/v1/employees/{id}",
		middleware.JWT(http.HandlerFunc(employeeHandler.Get)),
	)
	mux.Handle(
		"PUT /api/v1/employees/{id}",
		middleware.JWT(http.HandlerFunc(employeeHandler.Update)),
	)
	mux.Handle(
		"DELETE /api/v1/employees/{id}",
		middleware.JWT(http.HandlerFunc(employeeHandler.Delete)),
	)

	// Employee Device Routes
	mux.Handle(
		"POST /api/v1/employee-devices/register",
		middleware.JWT(http.HandlerFunc(employeeDeviceHandler.Register)),
	)
	mux.Handle(
		"GET /api/v1/employee-devices",
		middleware.JWT(http.HandlerFunc(employeeDeviceHandler.ListMyDevices)),
	)
	mux.Handle(
		"DELETE /api/v1/employee-devices/{id}",
		middleware.JWT(http.HandlerFunc(employeeDeviceHandler.DeleteDevice)),
	)

	// Admin Employee Device Routes
	mux.Handle(
		"GET /api/v1/admin/employee-devices",
		middleware.JWT(http.HandlerFunc(employeeDeviceHandler.ListAllDevices)),
	)
	mux.Handle(
		"GET /api/v1/admin/employees/{id}/devices",
		middleware.JWT(http.HandlerFunc(employeeDeviceHandler.ListDevicesByEmployee)),
	)

	// Employee screenshot routes
	mux.Handle(
		"POST /api/v1/employee-screenshots",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.Create)),
	)
	mux.Handle(
		"GET /api/v1/employee-screenshots",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.ListMine)),
	)
	mux.Handle(
		"GET /api/v1/employee-screenshots/{id}",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.Get)),
	)
	mux.Handle(
		"GET /api/v1/employee-screenshots/{id}/download",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.Download)),
	)

	// Employee activity routes
	mux.Handle(
		"GET /api/v1/employee-activities",
		middleware.JWT(http.HandlerFunc(employeeActivityHandler.ListMine)),
	)

	// Attendance routes (Employee)
	mux.Handle(
		"POST /api/v1/attendance/clock-in",
		middleware.JWT(http.HandlerFunc(attendanceHandler.ClockIn)),
	)
	mux.Handle(
		"POST /api/v1/attendance/clock-out",
		middleware.JWT(http.HandlerFunc(attendanceHandler.ClockOut)),
	)
	mux.Handle(
		"POST /api/v1/attendance/break-start",
		middleware.JWT(http.HandlerFunc(attendanceHandler.BreakStart)),
	)
	mux.Handle(
		"POST /api/v1/attendance/break-end",
		middleware.JWT(http.HandlerFunc(attendanceHandler.BreakEnd)),
	)
	mux.Handle(
		"GET /api/v1/attendance/timesheet",
		middleware.JWT(http.HandlerFunc(attendanceHandler.GetTimesheet)),
	)

	// Admin screenshot routes
	mux.Handle(
		"GET /api/v1/admin/employee-screenshots",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.AdminList)),
	)
	mux.Handle(
		"GET /api/v1/admin/employees/{employee_id}/screenshots",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.AdminListByEmployee)),
	)
	mux.Handle(
		"GET /api/v1/admin/employee-screenshots/{id}/download",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.AdminDownload)),
	)
	mux.Handle(
		"GET /api/v1/admin/employee-screenshots/{id}",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.AdminGet)),
	)
	mux.Handle(
		"PUT /api/v1/admin/employee-screenshots/{id}",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.Update)),
	)
	mux.Handle(
		"DELETE /api/v1/admin/employee-screenshots/{id}",
		middleware.JWT(http.HandlerFunc(employeeScreenshotHandler.Delete)),
	)

	// Admin activity routes
	mux.Handle(
		"GET /api/v1/admin/employee-activities",
		middleware.JWT(http.HandlerFunc(employeeActivityHandler.AdminList)),
	)
	mux.Handle(
		"GET /api/v1/admin/employees/{employee_id}/activities",
		middleware.JWT(http.HandlerFunc(employeeActivityHandler.AdminListByEmployee)),
	)

	// Admin attendance routes
	mux.Handle(
		"GET /api/v1/admin/attendance",
		middleware.JWT(http.HandlerFunc(attendanceHandler.AdminGetAllTimesheets)),
	)
	mux.Handle(
		"GET /api/v1/admin/employees/{id}/attendance",
		middleware.JWT(http.HandlerFunc(attendanceHandler.AdminGetEmployeeTimesheet)),
	)

	// Admin application routes
	mux.Handle(
		"GET /api/v1/admin/applications",
		middleware.JWT(http.HandlerFunc(applicationHandler.AdminList)),
	)
	mux.Handle(
		"GET /api/v1/admin/employees/{employee_id}/applications",
		middleware.JWT(http.HandlerFunc(applicationHandler.AdminListByEmployee)),
	)
	mux.Handle(
		"PUT /api/v1/admin/applications/{id}",
		middleware.JWT(http.HandlerFunc(applicationHandler.AdminUpdate)),
	)
	mux.Handle(
		"DELETE /api/v1/admin/applications/{id}",
		middleware.JWT(http.HandlerFunc(applicationHandler.AdminDelete)),
	)

	// Admin website routes
	mux.Handle(
		"GET /api/v1/admin/websites",
		middleware.JWT(http.HandlerFunc(websiteHandler.AdminList)),
	)
	mux.Handle(
		"GET /api/v1/admin/employees/{employee_id}/websites",
		middleware.JWT(http.HandlerFunc(websiteHandler.AdminListByEmployee)),
	)
	mux.Handle(
		"PUT /api/v1/admin/websites/{id}",
		middleware.JWT(http.HandlerFunc(websiteHandler.AdminUpdate)),
	)
	mux.Handle(
		"DELETE /api/v1/admin/websites/{id}",
		middleware.JWT(http.HandlerFunc(websiteHandler.AdminDelete)),
	)

	// Admin Team Routes
	mux.Handle(
		"POST /api/v1/admin/teams",
		middleware.JWT(http.HandlerFunc(teamHandler.AdminCreateTeam)),
	)
	mux.Handle(
		"GET /api/v1/admin/teams",
		middleware.JWT(http.HandlerFunc(teamHandler.AdminListTeams)),
	)
	mux.Handle(
		"PUT /api/v1/admin/teams/{id}",
		middleware.JWT(http.HandlerFunc(teamHandler.AdminUpdateTeam)),
	)
	mux.Handle(
		"DELETE /api/v1/admin/teams/{id}",
		middleware.JWT(http.HandlerFunc(teamHandler.AdminDeleteTeam)),
	)

	// Admin Task Routes
	mux.Handle(
		"POST /api/v1/admin/tasks",
		middleware.JWT(http.HandlerFunc(taskHandler.AdminCreateTask)),
	)
	mux.Handle(
		"GET /api/v1/admin/tasks",
		middleware.JWT(http.HandlerFunc(taskHandler.AdminListTasks)),
	)
	mux.Handle(
		"PUT /api/v1/admin/tasks/{id}",
		middleware.JWT(http.HandlerFunc(taskHandler.AdminUpdateTask)),
	)
	mux.Handle(
		"DELETE /api/v1/admin/tasks/{id}",
		middleware.JWT(http.HandlerFunc(taskHandler.AdminDeleteTask)),
	)

	// Admin Task Assignment Routes
	mux.Handle(
		"POST /api/v1/admin/tasks/{id}/assignments",
		middleware.JWT(http.HandlerFunc(taskHandler.AdminAssignTask)),
	)
	mux.Handle(
		"GET /api/v1/admin/tasks/{id}/assignments",
		middleware.JWT(http.HandlerFunc(taskHandler.AdminListTaskAssignments)),
	)
	mux.Handle(
		"DELETE /api/v1/admin/tasks/assignments/{assignment_id}",
		middleware.JWT(http.HandlerFunc(taskHandler.AdminDeleteTaskAssignment)),
	)

	// Department Routes
	mux.Handle(
		"POST /api/v1/departments",
		middleware.JWT(http.HandlerFunc(departmentHandler.Create)),
	)

	mux.Handle(
		"GET /api/v1/departments",
		middleware.JWT(http.HandlerFunc(departmentHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/departments/{id}",
		middleware.JWT(http.HandlerFunc(departmentHandler.Get)),
	)

	mux.Handle(
		"PUT /api/v1/departments/{id}",
		middleware.JWT(http.HandlerFunc(departmentHandler.Update)),
	)

	mux.Handle(
		"DELETE /api/v1/departments/{id}",
		middleware.JWT(http.HandlerFunc(departmentHandler.Delete)),
	)

	// Roles by Department
	mux.Handle(
		"GET /api/v1/departments/{id}/roles",
		middleware.JWT(http.HandlerFunc(roleHandler.ListByDepartment)),
	)
	// Role Routes
	mux.Handle(
		"POST /api/v1/roles",
		middleware.JWT(http.HandlerFunc(roleHandler.Create)),
	)

	mux.Handle(
		"GET /api/v1/roles",
		middleware.JWT(http.HandlerFunc(roleHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/roles/{id}",
		middleware.JWT(http.HandlerFunc(roleHandler.Get)),
	)

	mux.Handle(
		"PUT /api/v1/roles/{id}",
		middleware.JWT(http.HandlerFunc(roleHandler.Update)),
	)

	mux.Handle(
		"DELETE /api/v1/roles/{id}",
		middleware.JWT(http.HandlerFunc(roleHandler.Delete)),
	)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"version": "v1",
		"message": "Hello API v1 is running",
	})
}

func healthHandler(db *gorm.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {

			w.WriteHeader(http.StatusServiceUnavailable)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":  false,
				"server":   "running",
				"database": "disconnected",
				"version":  "v1",
			})

			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"server":   "running",
			"database": "connected",
			"version":  "v1",
		})
	}
}

func jsonError(w http.ResponseWriter, message string, status int) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
