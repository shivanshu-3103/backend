# Desktop App Screenshot Flow

This document explains the flow for the desktop application to check device registration and upload screenshots in different scenarios (logged in vs. logged out).

## Flow Overview

1. **On App Startup**: The desktop app should check if the device is registered in the backend.
2. **If Logged In**: The app uses the standard authenticated API with the user's JWT token to upload screenshots.
3. **If Logged Out (Background)**: The app uses the unauthenticated (device-based) API with the `X-Device-Fingerprint` header to upload screenshots.

---

## 1. Check if Device is Registered

Before sending background screenshots, the app should verify that the device is still recognized by the server. 

**API Endpoint:** `GET /api/v1/app/device/check`

**Curl Command:**
```bash
curl -X GET "http://localhost:8080/api/v1/app/device/check?device_fingerprint=UNIQUE_FINGERPRINT_123"
```

**Expected Response (Success - 200 OK):**
```json
{
  "success": true,
  "message": "Device is registered",
  "device_id": "uuid-of-device",
  "employee_id": "uuid-of-employee"
}
```
*If this returns a `404 Not Found` or `401 Unauthorized`, the app should stop background tracking and prompt the user to log in and register the device again.*

---

## 2. Upload Screenshot WHEN LOGGED IN (Foreground)

When the user is actively logged into the app, you have their JWT token. You should use the standard authenticated endpoint.

**API Endpoint:** `POST /api/v1/employee-screenshots`

**Curl Command:**
```bash
curl -X POST http://localhost:8080/api/v1/employee-screenshots \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN_HERE>" \
  -F "screenshot=@/path/to/local/screenshot.jpg" \
  -F "captured_at=2026-09-12T12:00:00Z" \
  -F "notes=Active session screenshot"
```

**Why use this?**
- It explicitly uses the current session of the user.
- Best practice for standard data uploads when the user is authenticated.

---

## 3. Upload Screenshot WHEN LOGGED OUT (Background)

When the user closes the app or logs out, but you still need to send tracked screenshots in the background, you no longer have a valid JWT token. Instead, you use the device fingerprint to authorize the upload. The backend will automatically link the screenshot to the employee who owns this device.

**API Endpoint:** `POST /api/v1/app/employee-screenshots`

**Curl Command:**
```bash
curl -X POST http://localhost:8080/api/v1/app/employee-screenshots \
  -H "X-Device-Fingerprint: UNIQUE_FINGERPRINT_123" \
  -F "screenshot=@/path/to/local/screenshot.jpg" \
  -F "captured_at=2026-09-12T12:00:00Z" \
  -F "notes=Background screenshot (logged out)"
```

**Why use this?**
- No JWT token is required (no `Authorization: Bearer` header).
- It relies entirely on the custom `X-Device-Fingerprint` header.
- As long as the device is registered in the database, the backend accepts the screenshot.
