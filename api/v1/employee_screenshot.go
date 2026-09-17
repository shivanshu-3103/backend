package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go-api/middleware"
	"go-api/models"
	"go-api/repository"
	"go-api/utils"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"golang.org/x/image/draw"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const maxScreenshotSize = 10 << 20 // 10 MB

type EmployeeScreenshotHandler struct {
	Repo       repository.EmployeeScreenshotRepository
	DeviceRepo repository.EmployeeDeviceRepository
	DB         *gorm.DB
}

func (h *EmployeeScreenshotHandler) Create(w http.ResponseWriter, r *http.Request) {
	employee, err := h.employeeFromRequest(r)
	if err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	screenshot, err := parseScreenshotForm(r, employee.ID, employee.DepartmentID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Repo.Create(screenshot); err != nil {
		utils.LogError("Failed to save authenticated screenshot: %v", err)
		jsonError(w, "Failed to save screenshot", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Foreground screenshot saved: EmployeeID=%s, FilePath=%s", screenshot.EmployeeID, screenshot.FilePath)
	writeScreenshotJSON(w, http.StatusCreated, map[string]interface{}{"success": true, "screenshot": screenshot})
}

func (h *EmployeeScreenshotHandler) AppCreate(w http.ResponseWriter, r *http.Request) {
	deviceFingerprint := r.Header.Get("X-Device-Fingerprint")
	if deviceFingerprint == "" {
		jsonError(w, "X-Device-Fingerprint header is required", http.StatusUnauthorized)
		return
	}

	device, err := h.DeviceRepo.FindByFingerprint(deviceFingerprint)
	if err != nil {
		jsonError(w, "Device not registered", http.StatusUnauthorized)
		return
	}

	screenshot, err := parseScreenshotForm(r, device.EmployeeID, device.Employee.DepartmentID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	screenshot.EmployeeDeviceID = &device.ID

	if err := h.Repo.Create(screenshot); err != nil {
		utils.LogError("Failed to save device screenshot: %v", err)
		jsonError(w, "Failed to save screenshot", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Background screenshot saved: DeviceID=%s, EmployeeID=%s, FilePath=%s", device.ID, device.EmployeeID, screenshot.FilePath)
	writeScreenshotJSON(w, http.StatusCreated, map[string]interface{}{"success": true, "screenshot": screenshot})
}

func (h *EmployeeScreenshotHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	employee, err := h.employeeFromRequest(r)
	if err != nil {
		jsonError(w, "Employee record not found", http.StatusNotFound)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	screenshots, totalCount, err := h.Repo.FindByEmployeeID(employee.ID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch screenshots", http.StatusInternalServerError)
		return
	}
	presignScreenshots(screenshots)
	writeScreenshotJSON(w, http.StatusOK, utils.PaginatedResponse{
		Success:    true,
		Data:       screenshots,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *EmployeeScreenshotHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, departmentID, err := screenshotFilters(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	screenshots, totalCount, err := h.Repo.FindAll(employeeID, departmentID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch screenshots", http.StatusInternalServerError)
		return
	}
	presignScreenshots(screenshots)
	writeScreenshotJSON(w, http.StatusOK, utils.PaginatedResponse{
		Success:    true,
		Data:       screenshots,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *EmployeeScreenshotHandler) AdminListByEmployee(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, err := uuid.Parse(r.PathValue("employee_id"))
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	screenshots, totalCount, err := h.Repo.FindByEmployeeID(employeeID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch employee screenshots", http.StatusInternalServerError)
		return
	}
	presignScreenshots(screenshots)
	writeScreenshotJSON(w, http.StatusOK, utils.PaginatedResponse{
		Success:    true,
		Data:       screenshots,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *EmployeeScreenshotHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid screenshot ID", http.StatusBadRequest)
		return
	}
	screenshot, err := h.Repo.FindByID(id)
	if err != nil {
		jsonError(w, "Screenshot not found", http.StatusNotFound)
		return
	}
	if !h.canAccessScreenshot(w, r, screenshot) {
		return
	}
	presignScreenshot(screenshot)
	writeScreenshotJSON(w, http.StatusOK, map[string]interface{}{"success": true, "screenshot": screenshot})
}

func (h *EmployeeScreenshotHandler) AdminGet(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	h.Get(w, r)
}

func (h *EmployeeScreenshotHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid screenshot ID", http.StatusBadRequest)
		return
	}
	screenshot, err := h.Repo.FindByID(id)
	if err != nil {
		jsonError(w, "Screenshot not found", http.StatusNotFound)
		return
	}
	if !h.canAccessScreenshot(w, r, screenshot) {
		return
	}
	if strings.HasPrefix(screenshot.FilePath, "http://") || strings.HasPrefix(screenshot.FilePath, "https://") {
		presignScreenshot(screenshot)
		http.Redirect(w, r, screenshot.FilePath, http.StatusTemporaryRedirect)
		return
	}
	w.Header().Set("Content-Type", screenshot.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, screenshot.Filename))
	http.ServeFile(w, r, screenshot.FilePath)
}

func (h *EmployeeScreenshotHandler) AdminDownload(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	h.Download(w, r)
}

func (h *EmployeeScreenshotHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid screenshot ID", http.StatusBadRequest)
		return
	}
	screenshot, err := h.Repo.FindByID(id)
	if err != nil {
		jsonError(w, "Screenshot not found", http.StatusNotFound)
		return
	}
	if err := r.ParseMultipartForm(maxScreenshotSize); err != nil && err != http.ErrNotMultipart {
		jsonError(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}
	if notes := r.FormValue("notes"); notes != "" {
		screenshot.Notes = notes
	}
	if capturedAt := r.FormValue("captured_at"); capturedAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339, capturedAt)
		if parseErr != nil {
			jsonError(w, "captured_at must be RFC3339", http.StatusBadRequest)
			return
		}
		screenshot.CapturedAt = parsed
	}
	if file, header, fileErr := r.FormFile("screenshot"); fileErr == nil {
		defer file.Close()
		data, readErr := io.ReadAll(io.LimitReader(file, maxScreenshotSize+1))
		if readErr != nil || int64(len(data)) > maxScreenshotSize {
			jsonError(w, "Screenshot must be 10 MB or smaller", http.StatusBadRequest)
			return
		}

		screenshot.ContentType = header.Header.Get("Content-Type")
		if screenshot.ContentType == "" {
			screenshot.ContentType = http.DetectContentType(data)
		}

		fileURL, uploadErr := uploadScreenshot(data, header.Filename, screenshot.ContentType)
		if uploadErr != nil {
			utils.LogError("Screenshot upload failed: %v", uploadErr)
			jsonError(w, "Failed to upload screenshot", http.StatusInternalServerError)
			return
		}

		screenshot.FileSize, screenshot.Filename, screenshot.FilePath = int64(len(data)), header.Filename, fileURL
	}
	if err := h.Repo.Update(screenshot); err != nil {
		jsonError(w, "Failed to update screenshot", http.StatusInternalServerError)
		return
	}
	writeScreenshotJSON(w, http.StatusOK, map[string]interface{}{"success": true, "screenshot": screenshot})
}

func (h *EmployeeScreenshotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid screenshot ID", http.StatusBadRequest)
		return
	}

	screenshot, err := h.Repo.FindByID(id)
	if err == nil && screenshot.FilePath != "" && !strings.HasPrefix(screenshot.FilePath, "http") {
		os.Remove(screenshot.FilePath)
	}

	if err := h.Repo.Delete(id); err != nil {
		jsonError(w, "Failed to delete screenshot", http.StatusInternalServerError)
		return
	}
	writeScreenshotJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Screenshot deleted successfully"})
}

func (h *EmployeeScreenshotHandler) employeeFromRequest(r *http.Request) (*models.Employee, error) {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil {
		return nil, err
	}
	var employee models.Employee
	err := h.DB.Where("email = ?", user.Email).First(&employee).Error
	return &employee, err
}

func (h *EmployeeScreenshotHandler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil || user.UserType == 2 {
		jsonError(w, "Admin access required", http.StatusForbidden)
		return false
	}
	return true
}

func (h *EmployeeScreenshotHandler) canAccessScreenshot(w http.ResponseWriter, r *http.Request, screenshot *models.EmployeeScreenshot) bool {
	var user models.User
	if h.DB.First(&user, middleware.GetUserID(r)).Error == nil && user.UserType != 2 {
		return true
	}
	employee, err := h.employeeFromRequest(r)
	if err != nil || employee.ID != screenshot.EmployeeID {
		jsonError(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func parseScreenshotForm(r *http.Request, employeeID, departmentID uuid.UUID) (*models.EmployeeScreenshot, error) {
	if err := r.ParseMultipartForm(maxScreenshotSize); err != nil {
		return nil, fmt.Errorf("request must be multipart/form-data")
	}
	file, header, err := r.FormFile("screenshot")
	if err != nil {
		return nil, fmt.Errorf("screenshot file is required")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxScreenshotSize+1))
	if err != nil || int64(len(data)) > maxScreenshotSize {
		return nil, fmt.Errorf("screenshot must be 10 MB or smaller")
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("screenshot must be an image")
	}
	capturedAt := time.Now().UTC()
	if value := r.FormValue("captured_at"); value != "" {
		capturedAt, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, fmt.Errorf("captured_at must be RFC3339")
		}
	}

	// Compress the image before uploading to hit 10-20KB sizes with better quality
	compressedData, compressErr := compressScreenshot(data)
	if compressErr == nil && len(compressedData) > 0 {
		data = compressedData
		contentType = "image/jpeg"
		
		// Safely change extension to .jpg
		lastDotIndex := strings.LastIndex(header.Filename, ".")
		if lastDotIndex > 0 {
			header.Filename = header.Filename[:lastDotIndex] + ".jpg"
		} else {
			header.Filename += ".jpg"
		}
	}

	fileURL, err := uploadScreenshot(data, header.Filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload screenshot")
	}

	return &models.EmployeeScreenshot{EmployeeID: employeeID, DepartmentID: departmentID, Filename: header.Filename, ContentType: contentType, FileSize: int64(len(data)), FilePath: fileURL, CapturedAt: capturedAt, Notes: r.FormValue("notes")}, nil
}

func compressScreenshot(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Downscale to a reasonable size to hit the ~10-20KB target with decent quality
	targetWidth := 1280
	if width > targetWidth {
		ratio := float64(targetWidth) / float64(width)
		targetHeight := int(float64(height) * ratio)

		dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
		draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		img = dst
	}

	var buf bytes.Buffer
	// Encode as JPEG with higher quality (65) to achieve 10-20KB
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 65})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func uploadScreenshot(data []byte, filename string, contentType string) (string, error) {
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL != "" {
		cld, err := cloudinary.NewFromURL(cloudinaryURL)
		if err != nil {
			return "", fmt.Errorf("invalid CLOUDINARY_URL: %w", err)
		}
		result, err := cld.Upload.Upload(context.Background(), bytes.NewReader(data), uploader.UploadParams{
			Folder:   "handdy/screenshots",
			PublicID: uuid.NewString(),
		})
		if err != nil {
			return "", err
		}
		if result.SecureURL == "" {
			return "", fmt.Errorf("Cloudinary returned an empty secure URL for %s", filename)
		}
		return result.SecureURL, nil
	}

	awsRegion := os.Getenv("AWS_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("S3_BUCKET_NAME")

	if awsRegion == "" || awsAccessKey == "" || awsSecretKey == "" || bucketName == "" {
		return "", fmt.Errorf("neither CLOUDINARY_URL nor full AWS S3 credentials are provided")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")),
	)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	upl := manager.NewUploader(client)

	key := fmt.Sprintf("handdy/screenshots/%s-%s", uuid.NewString(), filename)

	result, err := upl.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return result.Location, nil
}

func screenshotFilters(r *http.Request) (*uuid.UUID, *uuid.UUID, error) {
	parse := func(name string) (*uuid.UUID, error) {
		value := r.URL.Query().Get(name)
		if value == "" {
			return nil, nil
		}
		id, err := uuid.Parse(value)
		return &id, err
	}
	employeeID, err := parse("employee_id")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid employee_id")
	}
	departmentID, err := parse("department_id")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid department_id")
	}
	return employeeID, departmentID, nil
}

func writeScreenshotJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func getPresignClient() *s3.PresignClient {
	awsRegion := os.Getenv("AWS_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if awsRegion == "" || awsAccessKey == "" || awsSecretKey == "" {
		return nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")),
	)
	if err != nil {
		return nil
	}
	return s3.NewPresignClient(s3.NewFromConfig(cfg))
}

func presignScreenshots(screenshots []models.EmployeeScreenshot) {
	bucketName := os.Getenv("S3_BUCKET_NAME")
	awsRegion := os.Getenv("AWS_REGION")
	if bucketName == "" || awsRegion == "" {
		return
	}

	presignClient := getPresignClient()
	if presignClient == nil {
		return
	}

	prefix := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", bucketName, awsRegion)

	for i := range screenshots {
		if strings.HasPrefix(screenshots[i].FilePath, prefix) {
			key := strings.TrimPrefix(screenshots[i].FilePath, prefix)
			req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			}, s3.WithPresignExpires(15*time.Minute))
			if err == nil {
				screenshots[i].FilePath = req.URL
			}
		}
	}
}

func presignScreenshot(screenshot *models.EmployeeScreenshot) {
	if screenshot == nil {
		return
	}
	bucketName := os.Getenv("S3_BUCKET_NAME")
	awsRegion := os.Getenv("AWS_REGION")
	if bucketName == "" || awsRegion == "" {
		return
	}

	presignClient := getPresignClient()
	if presignClient == nil {
		return
	}

	prefix := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", bucketName, awsRegion)

	if strings.HasPrefix(screenshot.FilePath, prefix) {
		key := strings.TrimPrefix(screenshot.FilePath, prefix)
		req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(15*time.Minute))
		if err == nil {
			screenshot.FilePath = req.URL
		}
	}
}
